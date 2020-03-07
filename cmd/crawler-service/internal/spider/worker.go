package spider

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// crawl crawls the html at the given url and writes the local URLs to the
// channel as it finds them.
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

		// Only send the path if it is a relative path (the host is empty), or the
		// host matches our host.
		foundURLs <- href
	}

	return nil
}
