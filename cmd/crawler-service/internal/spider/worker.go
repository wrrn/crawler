package spider

// worker is the entity that does the parsing the of site and sends any paths it
// finds on the found paths channel.
type worker struct {
	baseURL    string
	foundPaths chan<- string
}

// crawl crawls the html at the given path and writes the paths it finds on the channel.
func (w worker) crawl(path string) error {
	// convert path to URL
	// http.GetURL
	// parse the html
	// as I find an href that is related to the base url write it to the channel.
	return nil
}
