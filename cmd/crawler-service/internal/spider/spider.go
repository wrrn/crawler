package spider

import "github.com/wrrn/crawler/cmd/crawler-service/internal/site"

// Spider crawls through a web page and builds a site tree.
type Spider struct {
	URL string
}

// Crawl starts a spider crawling across a site.
func (s *Spider) Crawl() {

}

// Stops a spider from crawling across a site.
func (s *Spider) Stop() error {
	return nil
}

func (s *Spider) SiteTree() site.Tree {
	return site.Tree{}
}
