package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Url struct {
	Id         int       `bson:"_id"`
	Url        string    `bson:"url"`
	CreatedAt  time.Time `bson:"createdAt"`
	AccessedAt time.Time `bson:"accessedAt,omitempty"`
	Hits       int       `bson:"hits"`
}

type Counter struct {
	Id    string `bson:"_id"`
	Value int    `bson:"sequence_value"`
}

var (
	urls        *mongo.Collection
	cc          *mongo.Collection
	ctx, cancel = context.WithCancel(context.Background())

	chars  string = os.Getenv("chars")
	base   int    = len(chars)
	host   string = os.Getenv("host")
	port   string = os.Getenv("port")
	db     string = os.Getenv("db")
	dbName string = os.Getenv("dbName")
)

func init() {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(db))
	urls = client.Database(dbName).Collection("urls")
	cc = client.Database(dbName).Collection("counters")

	defer func() {
		if err != nil {
			cancel()
		}
	}()

	databases, err := client.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	if !contains(databases, dbName) {
		fmt.Printf("database %s is not exist.", dbName)
		os.Exit(1)
	}

	collections, err := client.Database(dbName).ListCollectionNames(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}

	if !contains(collections, "urls") {
		client.Database(dbName).CreateCollection(ctx, "urls")
	}

	if !contains(collections, "counters") {
		client.Database(dbName).CreateCollection(ctx, "counters")

		result, err := cc.InsertOne(ctx, bson.M{"_id": "urls", "sequence_value": 100})
		if err != nil {
			cancel()
			fmt.Println(err)
		}
		fmt.Println(result)
	}

}

func encode(n int) string {

	m := n % base
	r := []rune(chars)

	if n-m == 0 {
		return string(r[n-1:][0])
	}

	var a string = ""

	for m > 0 || n > 0 {
		a = (string(r[m-1:][0])) + a
		n = (n - m) / base
		m = n % base
	}

	return a
}

func decode(s string) int {

	l := len(s)
	r := []rune(s)
	var n int = 0

	for i := 0; i < l; i++ {
		n += (strings.Index(chars, string(r[i:][0])) + 1) * int(math.Pow(float64(base), float64(l-i-1)))
	}

	return n
}

func sanitizeUrl(s string) string {
	s = url.QueryEscape(s)
	return s
}

func getNextSequenceValue() int {
	result := &Counter{}

	ctx, cancel = context.WithCancel(context.Background())

	err := cc.FindOneAndUpdate(ctx, bson.M{"_id": "urls"}, bson.M{"$inc": bson.M{"sequence_value": 1}}).Decode(&result)
	if err != nil {
		cancel()
		fmt.Println(err)
	}

	return result.Value
}

func store(s string) int {

	var data Url

	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	data.Id = getNextSequenceValue()
	data.Url = s
	data.CreatedAt = time.Now()
	data.Hits = 0

	result, err := urls.InsertOne(ctx, data)
	if err != nil {
		cancel()
		fmt.Println(err)
	}

	return int(result.InsertedID.(int32))
}

func find(s string) (*Url, error) {
	result := &Url{}

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := urls.FindOne(ctx, bson.M{"url": s}).Decode(&result)
	if err != nil {
		return nil, err
	}

	if err == mongo.ErrNoDocuments {
		return nil, err
	} else if err != nil {
		return nil, err
	}

	return result, nil
}

func findUrl(s int) (*Url, error) {
	result := &Url{}

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	a := time.Now()

	err := urls.FindOneAndUpdate(ctx, bson.M{"_id": s}, bson.M{"$set": bson.M{"accessedAt": a}, "$inc": bson.M{"hits": 1}}).Decode(&result)
	if err != nil {
		return nil, err
	}

	if err == mongo.ErrNoDocuments {
		return nil, err
	} else if err != nil {
		return nil, err
	}

	return result, nil
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func TrimSuffix(s, suffix string) string {
	if strings.HasSuffix(s, suffix) {
		s = s[:len(s)-len(suffix)]
	}
	return s
}

func handler(w http.ResponseWriter, r *http.Request) {

	query := r.URL.Query()
	url := query.Get("url")
	urlPath := r.URL.Path[1:]

	if len(urlPath) != 0 {
		u, err := findUrl(decode(urlPath))
		if err != nil {
			cancel()
			fmt.Println(err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		redirect(w, r, u.Url)
		return
	}

	if len(url) != 0 {

		u, err := find(url)
		if err == mongo.ErrNoDocuments {
			cancel()
			fmt.Fprintf(w, "%s/%s", host, encode(store(url)))
			return
		} else if err != nil {
			cancel()
			w.WriteHeader(http.StatusNotFound)
			return
		}

		fmt.Fprintf(w, "%s/%s", host, encode(u.Id))
		return
	}

	w.WriteHeader(http.StatusNotFound)
}

func redirect(w http.ResponseWriter, r *http.Request, url string) {
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func main() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/favicon.ico", faviconHandler)
	http.ListenAndServe(":"+port, nil)
}
