# dn-shorty

Simple URL shortener by Golang.

## Installation

1\. Create Database

2\. Install required dependency package

    go get go.mongodb.org/mongo-driver

3\. Set `Env` before run for PowerShell

    $Env:host = "http://localhost:8080"
    
    $Env:db = "mongodb://localhost:27017"
    
    $Env:dbName = "shorty" 

4\. Run App

    go run .

Build App

    go build

## Generating short URLs

pass in a `url` query parameter.

    http://localhost:8080/?url=https://www.google.com

return a shortened URL

    http://localhost:8080/ge

## Requirement

* Go 1.15+
* MongoDB
