package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/wrrn/crawler/pkg/crawler"
	"google.golang.org/grpc"
)

var (
	errTooManyCommands = fmt.Errorf("Too many commands used. Use one of the following command flags: %v", commandNames)
	errNoCommand       = fmt.Errorf("No command used. Use one of the following command flags: %v", commandNames)
	commands           = map[string]bool{"start": true, "stop": true, "list": true}
	commandNames       = []string{"-start", "-stop", "-list"}
)

func main() {
	var (
		insecure = flag.Bool("insecure", false, "use an insecure client connection")
		// TODO(wh): Remove insecure
		_          = insecure
		serverAddr = flag.String("service-addr", "localhost:5555", "the address of the crawler-service")
		startURL   = flag.String("start", "", "the url to start crawling")
		stopURL    = flag.String("stop", "", "the url to stop crawling")
		list       = flag.Bool("list", false, "show the current site tree for all crawled URLs")
	)

	flag.Parse()

	if err := validateFlags(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		flag.Usage()
		os.Exit(1)
		// We don't need to return here because os.Exit will handle that for us.
	}

	// TODO(wh): Use a tls cert
	//	Create the client
	conn, err := grpc.Dial(*serverAddr, grpc.WithInsecure())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to dial %s: %v\n", *serverAddr, err)
		os.Exit(2)
		// We don't need to return here because os.Exit will handle that for us.
	}

	client := crawler.NewCrawlerClient(conn)
	_ = client
	ctx, cancel := context.WithCancel(context.Background())

	// Listen for incoming signals so that we can cancel slow requests
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, os.Kill)
		// We don't care if we get an interrupt or a kill we will just do the same thing.
		<-c

		cancel()
	}()
	// Send the information over the wire
	// TODO(wh): Is there better way to handle this than with a switch statement.
	switch {
	case len(*startURL) > 0:
		// We don't care about the output of Start because it returns an empty response.
		_, err := client.Start(ctx, &crawler.StartRequest{Url: *startURL})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to send the start request to %s: %v", *serverAddr, err)
		}
	case len(*stopURL) > 0:
		// We don't care about the output of Stop because it returns an empty response.
		_, err := client.Stop(ctx, &crawler.StopRequest{Url: *stopURL})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to send the stop request to %s: %v", *serverAddr, err)
		}

	case *list:
		listResponse, err := client.List(ctx, &crawler.ListRequest{})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to send the list request to %s: %v", *serverAddr, err)
		}

		printSiteTree(listResponse)
	}

}

func validateFlags() error {
	var commandsSeen int8
	flag.Visit(func(f *flag.Flag) {
		if commands[f.Name] {
			commandsSeen++
		}
	})

	if commandsSeen > 1 {
		return errTooManyCommands
	}

	if commandsSeen == 0 {
		return errNoCommand
	}

	return nil
}

func printSiteTree(listResponse *crawler.ListResponse) {
	fmt.Println(listResponse)
}
