package spider

import (
	"log"
	"sync"

	"github.com/wrrn/crawler/cmd/crawler-service/internal/site"
)

func New() *Spider {
	return &Spider{
		stop: make(chan struct{}),
	}
}

// Spider crawls through a web page and builds a site tree.
type Spider struct {
	// stop is closed to notify all the workers to stop crawling.
	stop chan struct{}

	// tree is the Tree that is being built.
	tree site.Tree
}

// Crawl starts a spider crawling across a site.
func (s *Spider) Crawl(url string) {
	// seenPaths is a the cache of paths that we have already seen. It is used to
	// limit the amount of time we spend looking at duplicate paths.
	seenPaths := make(map[string]bool)

	// foundPaths is the channel on which the workers send the paths as it finds
	// them. Setting this to an arbitrary number. There is probably a better way
	// of determining this number.
	foundPaths := make(chan string, 100)

	tree := site.Tree{}
	var stopped bool
	var wg sync.WaitGroup

	for {
		if stopped {
			break
		}

		select {
		case path := <-foundPaths: // Wait for workers to send back paths they have found
			if seenPaths[path] {
				continue
			}

			// Add the path to our site tree
			tree.Add(path)

			// Spin off a worker to start crawling the path. It would be better to use
			// a worker pool for this, so that an indeterminate amount of workers
			// aren't spun off. I could add an elastic worker pool.
			wg.Add(1)
			go func() {
				if err := (worker{baseURL: url, foundPaths: foundPaths}).crawl(path); err != nil {
					log.Println("Got an error while crawling %s: %v", path, err)
				}
				wg.Done()
			}()

		case <-s.stop: // Listen for the stop signal.
			stopped = true
		}
	}

	// Wait for all the goroutines to finish up and close the foundPaths so that
	// the for loop below will exit.
	go func() {
		wg.Wait()
		close(foundPaths)
	}()

	// Drain the foundPaths channel.
	for path := range foundPaths {
		if seenPaths[path] {
			continue
		}

		// Add the path to our site tree
		tree.Add(path)
	}

	s.tree = tree
}

// Stops a spider from crawling across a site.
func (s *Spider) Stop() {
	close(s.stop)

}

func (s *Spider) SiteTree() site.Tree {
	return s.tree
}
