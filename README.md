# dn-shorty

Simple URL shortener by Golang.

## Installation

1\. Create Database

2\. Install required dependency package

    go get go.mongodb.org/mongo-driver

3\. Set `Env` before run for PowerShell

    $Env:chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

    $Env:host = "http://localhost:8080"
    
    $Env:db = "mongodb://localhost:27017"
    
    $Env:dbName = "shorty" 

4\. Run App

    go run .

or Install

    go install

After installing, the `shorty` command can be run directly.


## Generating short URLs

pass in a `url` query parameter.

    http://localhost:8080/?url=https://www.google.com

return a shortened URL

    http://localhost:8080/ge

## Requirement

* Go 1.15+
* MongoDB

## Docker Support

1\. Pull form repository

    docker pull ghcr.io/dnxza/dn-shorty:latest

or Pull form `Docker Hub` repository

    docker pull dnratthee/dn-shorty:latest

2\. copy `Env`

    cp .env.init .env

3\. edit `Env`

    nano .env

4\. run container

    docker run -d -p 80:80 --env-file ./.env ghcr.io/dnxza/dn-shorty:latest

or 

(`if you pull form Docker hub repository`)

    docker run -d -p 80:80 --env-file ./.env dnratthee/dn-shorty:latest