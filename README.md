# dn-shorty

Simple URL shortener by Golang.

## Installation

1\. Create Database

2\. Set `Env` before run for PowerShell

    $Env:host = "http://localhost:8080"
    
    $Env:db = "mongodb://localhost:27017"
    
    $Env:dbName = "shorty" 

3\. Run App

    go run .

Build App

    go build

## Generating short URLs

pass in a `url` query parameter.

    http://localhost:8080/?url=https://www.google.com

return a shortened URL

    http://localhost:8080/ge

## Generating short URLs

* Go 1.15+
* MongoDB