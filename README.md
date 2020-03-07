# Crawler
Web Crawler as a gRPC service

The application consists of a command line client and a service which does the actual web crawling. For each URL, the
Web Crawler, creates a "site tree", which is a tree of links with the root of
the tree being the root URL. The crawler only follow links on the 
domain of the provided URL and does not follow external links.
The command line client provides the following operations:
```shell
$ crawl -start www.example.com # signals the service to start crawling www.example.com
$ crawl -stop www.example.com # signals the service to stop crawling www.example.com
$ crawl -list # shows the current "site tree" for all crawled URLs.
```

## Building the client and server
```shell
make all
```
The service and client can be built using go build, but I have provided a
Makefile to simplify building both the client and the service. Running `make
all` will build both the service and the client, and dump two binaries
(crawler-service, and crawler receptively) into the
project's root.

## Starting the service
To start the service run

```shell
./crawler-service
```

This will run the crawler on localhost:5555.
