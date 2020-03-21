package service

import (
	"context"
	"net/url"
	"strings"
	"sync"

	"github.com/wrrn/crawler/cmd/crawler-service/internal/site"
	"github.com/wrrn/crawler/cmd/crawler-service/internal/spider"
	pb "github.com/wrrn/crawler/pkg/crawler"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// New returns a new Service that implements crawler.CrawlerService.
func New() *Service {
	return &Service{
		activeSpiders: map[string]*spider.Spider{},
		spidersLock:   sync.RWMutex{},
		trees:         map[string]site.Tree{},
		treesLock:     sync.RWMutex{},
	}
}

// Service accepts incoming gRPC requests to start and stop crawling urls, and
// to list the site trees for all of the parsed URLs.
type Service struct {
	activeSpiders map[string]*spider.Spider
	spidersLock   sync.RWMutex

	// Use a map here so that we can just overwrite the existing site trees, when
	// we receive a new request. Also gives us faster lookups for start and stop.
	trees     map[string]site.Tree
	treesLock sync.RWMutex
}

// Start signals the service to start crawling the given URL.
func (s *Service) Start(_ context.Context, req *pb.StartRequest) (*pb.StartResponse, error) {
	url, err := parseURL(req.GetUrl())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%s was not a valid URL", req.GetUrl())
	}

	if _, found := s.getSpider(req.GetUrl()); found {
		return nil, status.Errorf(codes.FailedPrecondition, "Already crawling %s", req.GetUrl())
	}

	spider := spider.New()
	go spider.Crawl(url)

	s.addSpider(req.GetUrl(), spider)

	return &pb.StartResponse{}, nil
}

// Stop signals the service to stop crawling the given URL.
func (s *Service) Stop(_ context.Context, req *pb.StopRequest) (*pb.StopResponse, error) {
	spider, found := s.getSpider(req.GetUrl())
	if !found {
		return nil, status.Errorf(codes.InvalidArgument, "Start crawling %s before calling Stop", req.GetUrl())
	}

	spider.Stop()
	s.addTree(req.GetUrl(), spider.SiteTree())
	s.removeSpider(req.GetUrl())

	return &pb.StopResponse{}, nil
}

// Show the current site tree for all the given URLs.
func (s *Service) List(context.Context, *pb.ListRequest) (*pb.ListResponse, error) {
	return &pb.ListResponse{SiteTrees: s.getProtoTrees()}, nil
}

// parseURL will convert the string into an url.URL. It will return errors if
// the given string is empty, the string is not a parsable url, or the host
// cannot be determined. If the string doesn't have a scheme (https, http, ...)
// then http will be added as the scheme.
func parseURL(raw string) (*url.URL, error) {
	if len(raw) == 0 {
		return nil, errEmptyURL
	}

	url, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}

	if len(url.Host) == 0 && len(url.Path) == 0 {
		return nil, errUnparsableURL
	}

	if len(url.Host) == 0 {
		parts := strings.SplitN(url.Path, "/", 2)
		if len(parts) == 1 {
			url.Host = parts[0]
			url.Path = ""
		} else {
			url.Host = parts[0]
			url.Path = parts[1]
		}
	}

	if len(url.Scheme) == 0 {
		url.Scheme = "http"
	}

	return url, nil
}

// addSpider will associate the give spider with give URL so that it can stopped
// later.
func (s *Service) addSpider(url string, spider *spider.Spider) {
	s.spidersLock.Lock()
	s.activeSpiders[url] = spider
	s.spidersLock.Unlock()
}

// removeSpider removes the spider from active spider pool and returns the
// spider. False will be returned if there isn't a spider crawling the given
// url.
func (s *Service) removeSpider(url string) {
	s.spidersLock.Lock()
	delete(s.activeSpiders, url)
	s.spidersLock.Unlock()
}

func (s *Service) getSpider(url string) (*spider.Spider, bool) {
	s.spidersLock.RLock()
	spider, ok := s.activeSpiders[url]
	s.spidersLock.RUnlock()
	return spider, ok
}

// addTree will add a tree for the give URL to the cache of site trees.
func (s *Service) addTree(url string, tree site.Tree) {
	s.treesLock.Lock()
	s.trees[url] = tree
	s.treesLock.Unlock()
}

func (s *Service) getProtoTrees() []*pb.SiteTree {
	s.treesLock.RLock()
	defer s.treesLock.RUnlock()

	trees := make([]*pb.SiteTree, 0, len(s.trees))
	for site, tree := range s.trees {
		trees = append(trees, &pb.SiteTree{
			Url:  site,
			Tree: treeToProto(tree),
		})
	}

	return trees
}

func treeToProto(t site.Tree) *pb.Tree {
	return &pb.Tree{
		Name:     t.Value,
		Children: treesToProto(t.Children),
	}
}

func treesToProto(trees []*site.Tree) []*pb.Tree {
	if len(trees) == 0 {
		return []*pb.Tree{}
	}

	protoTrees := make([]*pb.Tree, 0, len(trees))
	for _, t := range trees {
		protoTrees = append(protoTrees, treeToProto(*t))
	}

	return protoTrees
}
