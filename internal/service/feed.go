package service

import (
	"context"
	"fmt"
	"time"

	"rss-backend/internal/fetcher"
	"rss-backend/internal/model"
	"rss-backend/internal/repository"
)

type FeedService struct {
	feedRepo *repository.FeedRepository
	artRepo  *repository.ArticleRepository
	fetcher  *fetcher.Fetcher
}

func NewFeedService(fr *repository.FeedRepository, ar *repository.ArticleRepository, f *fetcher.Fetcher) *FeedService {
	return &FeedService{feedRepo: fr, artRepo: ar, fetcher: f}
}

func (s *FeedService) CreateFeed(url string) (*model.Feed, error) {
	feed, err := s.feedRepo.Create(url)
	if err != nil {
		return nil, err
	}
	go s.triggerFetch(feed.ID, url)
	return feed, nil
}

func (s *FeedService) ListFeeds() ([]model.Feed, error) {
	return s.feedRepo.List()
}

func (s *FeedService) GetFeed(id uint) (*model.Feed, error) {
	return s.feedRepo.GetByID(id)
}

func (s *FeedService) triggerFetch(feedID uint, url string) {
	fetchSemaphore <- struct{}{}
	defer func() { <-fetchSemaphore }()

	s.feedRepo.UpdateStatus(feedID, "fetching", "", nil, nil, "")

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	nf, err := s.fetcher.Fetch(ctx, url)
	if err != nil {
		s.feedRepo.UpdateStatus(feedID, "failed", fmt.Sprintf("%.512s", err.Error()), nil, nil, "")
		return
	}

	if err := s.artRepo.Upsert(feedID, nf.Articles); err != nil {
		s.feedRepo.UpdateStatus(feedID, "failed", "upsert articles failed", nil, nil, "")
		return
	}

	now := time.Now().UTC()
	s.feedRepo.UpdateStatus(feedID, "success", "", &now, &nf.UpdatedAt, nf.Title)
}
