package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/wrrn/crawler/cmd/crawler-service/internal/service"
	"github.com/wrrn/crawler/pkg/crawler"
	"google.golang.org/grpc"
)

func main() {
	var (
		listenAddr = flag.String("listen-address", ":5555", "the address that the service should listen on")
	)

	flag.Parse()
	listener, err := net.Listen("tcp", *listenAddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed listen on %s", *listenAddr)
		os.Exit(1)
	}

	server := grpc.NewServer()
	crawler.RegisterCrawlerServer(server, service.New())

	server.Serve(listener)
}
