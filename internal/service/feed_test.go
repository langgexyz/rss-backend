package service_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"rss-backend/internal/fetcher"
	"rss-backend/internal/repository"
	"rss-backend/internal/service"
	"rss-backend/internal/testutil"
)

func newTestFeedService(t *testing.T) (*service.FeedService, *repository.FeedRepository, *repository.ArticleRepository) {
	t.Helper()
	db := testutil.SetupMySQL(t)
	service.InitSemaphore(1)
	feedRepo := repository.NewFeedRepository(db)
	artRepo := repository.NewArticleRepository(db)
	f := fetcher.New(5, 10)
	svc := service.NewFeedService(feedRepo, artRepo, f)
	return svc, feedRepo, artRepo
}

func TestFeedService_CreateFeed_ReturnsImmediately(t *testing.T) {
	svc, feedRepo, _ := newTestFeedService(t)

	feed, err := svc.CreateFeed("https://example.com/rss.xml")
	require.NoError(t, err)
	assert.NotZero(t, feed.ID)
	assert.Equal(t, "pending", feed.FetchStatus)

	// Verify record also exists in database
	got, err := feedRepo.GetByID(feed.ID)
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/rss.xml", got.URL)
}

func TestFeedService_CreateFeed_DuplicateURL(t *testing.T) {
	svc, _, _ := newTestFeedService(t)

	_, err := svc.CreateFeed("https://example.com/rss.xml")
	require.NoError(t, err)

	_, err = svc.CreateFeed("https://example.com/rss.xml")
	assert.Error(t, err, "重复 URL 应返回错误")
}

func TestFeedService_RefreshFeed_TooSoon(t *testing.T) {
	svc, feedRepo, _ := newTestFeedService(t)

	feed, err := svc.CreateFeed("https://example.com/rss.xml")
	require.NoError(t, err)

	// Set last_fetched_at to 5 seconds ago
	now := time.Now().UTC().Add(-5 * time.Second)
	err = feedRepo.UpdateStatus(feed.ID, "success", "", &now, nil, "")
	require.NoError(t, err)

	// minInterval = 300 seconds, refreshing after 5s should return ErrTooSoon
	err = svc.RefreshFeed(feed.ID, 300)
	assert.ErrorIs(t, err, service.ErrTooSoon)
}

func TestFeedService_RefreshFeed_FirstTime_NoLimit(t *testing.T) {
	svc, _, _ := newTestFeedService(t)

	feed, err := svc.CreateFeed("https://example.com/rss.xml")
	require.NoError(t, err)

	// last_fetched_at is nil (first time), should not trigger ErrTooSoon.
	// triggerFetch runs in background goroutine and may fail (no real RSS server),
	// but RefreshFeed itself should return nil (not ErrTooSoon).
	err = svc.RefreshFeed(feed.ID, 300)
	assert.NotErrorIs(t, err, service.ErrTooSoon, "首次刷新不应受间隔限制")
}
