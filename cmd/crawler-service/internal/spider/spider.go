package spider

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/wrrn/crawler/cmd/crawler-service/internal/site"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// New creates a new spider to crawl a site and build a site tree. Call it's
// Crawl() method to start the crawling.
func New() *Spider {
	return &Spider{
		stop: make(chan struct{}),
		wg:   &sync.WaitGroup{},
	}
}

// Spider crawls through a web page and builds a site tree.
type Spider struct {
	// stop is closed to notify all the workers to stop crawling.
	stop chan struct{}

	// tree is the Tree that is being built.
	tree site.Tree

	// wg is used in the Stop method so that it blocks until the Crawl method finishes.
	wg *sync.WaitGroup
}

// Crawl starts a spider crawling across a site.
func (s *Spider) Crawl(u *url.URL) {

	s.wg.Add(1)
	defer s.wg.Done()
	// seenPaths is a the cache of paths that we have already seen. It is used to
	// limit the amount of time we spend looking at duplicate paths.
	seenPaths := make(map[string]bool)

	// foundURLs is the channel on which the workers send the URLs as it finds
	// them. Setting the buffer length to an arbitrary number to prevent to many
	// goroutines blocking at once. There is probably a better way of determining
	// this number.
	foundURLs := make(chan *url.URL, 100)

	tree := site.Tree{Value: u.Hostname()}
	var stopped bool
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	// Populate our foundURLs channel with our initial url to get things started.
	foundURLs <- u
	for {
		if stopped {
			break
		}

		select {
		case url := <-foundURLs: // Wait for workers to send back URLs they have found
			if seenPaths[url.Path] {
				continue
			}

			// Add the path to our site tree
			tree.Add(url.Path)
			seenPaths[url.Path] = true

			// Spin off a worker to start crawling the URL. It would be better to use
			// a worker pool for this, so that an indeterminate amount of workers
			// aren't spun off. I could add an elastic worker pool.
			wg.Add(1)
			go func() {
				if err := crawl(ctx, url, foundURLs); err != nil {
					log.Println("Got an error while crawling %s: %v", url, err)
				}
				wg.Done()
			}()

		case <-s.stop: // Listen for the stop signal.
			stopped = true
		}
	}

	// Cancel any log running requests so that we terminate sooner.
	cancel()

	// Wait for all the goroutines to finish up and close the foundPaths so that
	// the for loop below will exit.
	go func() {
		wg.Wait()
		close(foundURLs)
	}()

	// Drain the foundPaths channel.
	for url := range foundURLs {
		if seenPaths[url.Path] {
			continue
		}

		// Add the path to our site tree
		tree.Add(url.Path)
	}

	s.tree = tree
}

// Stops a spider from crawling across a site.
func (s *Spider) Stop() {
	close(s.stop)
	// Wait for all calls to Crawl to finish
	s.wg.Wait()
}

// SiteTree returns the spider's site tree. If stop has not been called then it will return nothing.
func (s *Spider) SiteTree() site.Tree {
	return s.tree
}

// crawl crawls the html at the given url and writes the local URLs to the
// foundURLs channel as it finds them.
func crawl(ctx context.Context, u *url.URL, foundURLs chan<- *url.URL) error {
	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), strings.NewReader(u.Query().Encode()))
	if err != nil {
		return errors.Wrapf(err, "failed to build request for %s", u)
	}

	// Set the accept header so we have chance of the server not sending back some huge binary.
	req.Header.Set("accept", "text/html")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrapf(err, "failed to retrieve the body from %s", u)
	}

	// There isn't anything for us to do with a resource that isn't a HTML page.
	if !strings.Contains(resp.Header.Get("content-type"), "text/html") {
		return nil
	}

	// parse the html
	t := html.NewTokenizer(resp.Body)
	for tokenType := t.Next(); t.Err() == nil; tokenType = t.Next() {
		if tokenType != html.StartTagToken {
			continue
		}

		token := t.Token()
		if token.DataAtom != atom.A {
			continue
		}

		var hrefVal string
		for _, attr := range token.Attr {
			if attr.Key == atom.Href.String() {
				hrefVal = attr.Val
				break
			}
		}

		// Parses the URL in the context of given url. The provided hrefVal may be
		// relative or absolute.
		href, err := u.Parse(hrefVal)
		if err != nil {
			return errors.Wrapf(err, "failed to parse href %s", hrefVal)
		}

		if href == nil {
			continue
		}

		// We are only interested in links to the current host.
		if href.Hostname() != u.Hostname() {
			continue
		}

		// Put the URL on the queue of work for the spider to do.
		foundURLs <- href
	}

	return nil
}
