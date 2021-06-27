package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"net/http"
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
	urls *mongo.Collection
	cc   *mongo.Collection
	ctx  = context.Background()

	chars  string = os.Getenv("chars")
	base   int    = len(chars)
	host   string = os.Getenv("host")
	port   string = os.Getenv("port")
	db     string = os.Getenv("db")
	dbName string = os.Getenv("dbName")
)

func errorHandler(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func redirect(w http.ResponseWriter, r *http.Request, url string) {
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func init() {
	log.Println("Database initialization ...")
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(db))
	errorHandler(err)

	urls = client.Database(dbName).Collection("urls")
	cc = client.Database(dbName).Collection("counters")

	databases, err := client.ListDatabaseNames(ctx, bson.M{"name": dbName})
	errorHandler(err)
	if !contains(databases, dbName) {
		collections, err := client.Database(dbName).ListCollectionNames(ctx, bson.M{})
		errorHandler(err)
		if !contains(collections, "urls") {
			client.Database(dbName).CreateCollection(ctx, "urls")
			log.Println("Collection 'urls' Create")
		}

		if !contains(collections, "counters") {
			client.Database(dbName).CreateCollection(ctx, "counters")
			_, err := cc.InsertOne(ctx, bson.M{"_id": "urls", "sequence_value": 100})
			errorHandler(err)
			log.Println("Collection 'counters' Create")
		}
	}
	log.Println("Database initialization completed.")
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

func getNextSequenceValue() int {
	result := &Counter{}

	errorHandler(cc.FindOneAndUpdate(ctx, bson.M{"_id": "urls"}, bson.M{"$inc": bson.M{"sequence_value": 1}}).Decode(&result))
	return result.Value
}

func store(s string) int {
	var data Url

	data.Id = getNextSequenceValue()
	data.Url = s
	data.CreatedAt = time.Now()
	data.Hits = 0

	result, err := urls.InsertOne(ctx, data)
	errorHandler(err)

	return int(result.InsertedID.(int32))
}

func find(s string) (*Url, error) {
	result := &Url{}

	err := urls.FindOne(ctx, bson.M{"url": s}).Decode(&result)
	return result, err
}

func findUrl(s int) (*Url, error) {
	result := &Url{}
	a := time.Now()

	err := urls.FindOneAndUpdate(ctx, bson.M{"_id": s}, bson.M{"$set": bson.M{"accessedAt": a}, "$inc": bson.M{"hits": 1}}).Decode(&result)
	return result, err
}

func handler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	url := query.Get("url")
	urlPath := r.URL.Path[1:]

	if len(urlPath) != 0 {
		u, err := findUrl(decode(urlPath))

		if err == mongo.ErrNoDocuments {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		errorHandler(err)

		redirect(w, r, u.Url)
		return
	}

	if len(url) != 0 {
		u, err := find(url)
		if err == mongo.ErrNoDocuments {
			fmt.Fprintf(w, "%s/%s", host, encode(store(url)))
			return
		} else if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		fmt.Fprintf(w, "%s/%s", host, encode(u.Id))
		return
	}
	w.WriteHeader(http.StatusNotFound)
}

func main() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/favicon.ico", faviconHandler)

	log.Printf("Start HTTP server on port %s.\n", port)
	err := http.ListenAndServe(":"+port, nil)
	errorHandler(err)
}
