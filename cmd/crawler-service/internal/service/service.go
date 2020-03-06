package service

import (
	"context"
	"sync"

	"github.com/wrrn/crawler/cmd/crawler-service/internal/site"
	"github.com/wrrn/crawler/cmd/crawler-service/internal/spider"
	pb "github.com/wrrn/crawler/pkg/crawler"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service struct {
	activeSpiders map[string]*spider.Spider
	spidersLock   *sync.Mutex

	trees     map[string]site.Tree
	treesLock *sync.RWMutex
}

// Start signals the service to start crawling the given URL.
func (s *Service) Start(_ context.Context, req *pb.StartRequest) (*pb.StartResponse, error) {
	if len(req.GetUrl()) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Cannot pass in an empty URL")
	}

	// Adding more url validation here would be good.

	spider := spider.New()
	go spider.Crawl(req.GetUrl())

	s.addSpider(req.GetUrl(), spider)

	return &pb.StartResponse{}, nil
}

// Stop signals the service to stop crawling the given URL.
func (s *Service) Stop(_ context.Context, req *pb.StopRequest) (*pb.StopResponse, error) {
	spider, found := s.removeSpider(req.GetUrl())
	if !found {
		return nil, status.Errorf(codes.InvalidArgument, "Start crawling %s before calling Stop", req.GetUrl())
	}

	spider.Stop()

	s.addTree(req.GetUrl(), spider.SiteTree())

	return &pb.StopResponse{}, nil
}

// Show the current site tree for all the given URLs.
func (s *Service) List(context.Context, *pb.ListRequest) (*pb.ListResponse, error) {
	return &pb.ListResponse{SiteTrees: s.getProtoTrees()}, nil
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
func (s *Service) removeSpider(url string) (*spider.Spider, bool) {
	s.spidersLock.Lock()
	spider, ok := s.activeSpiders[url]
	delete(s.activeSpiders, url)
	s.spidersLock.Unlock()

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
	// TODO(wh): Implement
	s.treesLock.RUnlock()
	return []*pb.SiteTree{}
}
