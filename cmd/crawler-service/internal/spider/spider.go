package spider

import (
	"context"
	"log"
	"net/url"
	"sync"

	"github.com/wrrn/crawler/cmd/crawler-service/internal/site"
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

	wg *sync.WaitGroup
}

// Crawl starts a spider crawling across a site.
func (s *Spider) Crawl(u *url.URL) {

	s.wg.Add(1)
	defer s.wg.Done()
	// seenPaths is a the cache of paths that we have already seen. It is used to
	// limit the amount of time we spend looking at duplicate paths.
	seenPaths := make(map[string]bool)

	// foundURLs is the channel on which the workers send the paths as it finds
	// them. Setting the buffer length to an arbitrary number. There is probably a
	// better way of determining this number.
	foundURLs := make(chan *url.URL, 100)

	tree := site.Tree{Root: u.Hostname()}
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
		case url := <-foundURLs: // Wait for workers to send back paths they have found
			if seenPaths[url.Path] {
				continue
			}

			// Add the path to our site tree
			tree.Add(url.Path)
			seenPaths[url.Path] = true

			// Spin off a worker to start crawling the path. It would be better to use
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

func (s *Spider) SiteTree() site.Tree {

	return s.tree
}
