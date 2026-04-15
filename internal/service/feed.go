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
	ftSvc    *FulltextService
}

// ft 为可选参数，传入后会在抓取文章后自动后台获取全文
func NewFeedService(fr *repository.FeedRepository, ar *repository.ArticleRepository, f *fetcher.Fetcher, ft ...*FulltextService) *FeedService {
	svc := &FeedService{feedRepo: fr, artRepo: ar, fetcher: f}
	if len(ft) > 0 {
		svc.ftSvc = ft[0]
	}
	return svc
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

	// 后台自动抓取全文（限速：每篇间隔 500ms，避免对目标站点造成压力）
	if s.ftSvc != nil {
		go s.autoFetchFulltext(feedID)
	}
}

var ErrTooSoon = fmt.Errorf("too soon to refresh")

func (s *FeedService) RefreshFeed(id uint, minIntervalSecs int) error {
	feed, err := s.feedRepo.GetByID(id)
	if err != nil {
		return err
	}
	// 首次 last_fetched_at 为 nil 时不限制
	if feed.LastFetchedAt != nil {
		elapsed := time.Since(*feed.LastFetchedAt).Seconds()
		if elapsed < float64(minIntervalSecs) {
			return ErrTooSoon
		}
	}
	go s.triggerFetch(id, feed.URL)
	return nil
}

func (s *FeedService) DeleteFeed(id uint) error {
	return s.feedRepo.Delete(id)
}

// autoFetchFulltext 对该订阅源下尚未获取全文的文章逐篇抓取，每篇间隔 500ms
func (s *FeedService) autoFetchFulltext(feedID uint) {
	result, err := s.artRepo.List(repository.ArticleFilter{FeedID: &feedID, Page: 1, PageSize: 200})
	if err != nil {
		return
	}
	for _, art := range result.Items {
		if art.IsFullContent || len([]rune(art.Content)) >= 500 || art.Link == "" {
			continue
		}
		s.ftSvc.FetchFulltext(art.ID)
		time.Sleep(500 * time.Millisecond)
	}
}
