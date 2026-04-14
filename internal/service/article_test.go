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

func seedArticle(t *testing.T, artRepo *repository.ArticleRepository, feedRepo *repository.FeedRepository, feedURL string) uint {
	t.Helper()
	feed, err := feedRepo.Create(feedURL)
	require.NoError(t, err)
	err = artRepo.Upsert(feed.ID, []fetcher.NormalizedArticle{
		{
			GUIDHash:    "ffff1234ffff1234ffff1234ffff1234ffff1234ffff1234ffff1234ffff1234",
			Title:       "Test Article",
			Link:        "https://example.com/article",
			Content:     "article content",
			PublishedAt: time.Now(),
		},
	})
	require.NoError(t, err)
	result, err := artRepo.List(repository.ArticleFilter{Page: 1, PageSize: 1})
	require.NoError(t, err)
	require.NotEmpty(t, result.Items)
	return result.Items[0].ID
}

func TestArticleService_GetArticle_AutoMarksRead(t *testing.T) {
	db := testutil.SetupMySQL(t)
	feedRepo := repository.NewFeedRepository(db)
	artRepo := repository.NewArticleRepository(db)
	svc := service.NewArticleService(artRepo)

	artID := seedArticle(t, artRepo, feedRepo, "https://example.com/rss-a")

	// First fetch → auto mark as read
	art, err := svc.GetArticle(artID)
	require.NoError(t, err)
	assert.True(t, art.IsRead, "GetArticle 应自动标记已读")

	// Should also be persisted in database
	fromDB, err := artRepo.GetByID(artID)
	require.NoError(t, err)
	assert.True(t, fromDB.IsRead, "is_read 应持久化到数据库")
}

func TestArticleService_UpdateArticle_FieldWhitelist(t *testing.T) {
	db := testutil.SetupMySQL(t)
	feedRepo := repository.NewFeedRepository(db)
	artRepo := repository.NewArticleRepository(db)
	svc := service.NewArticleService(artRepo)

	artID := seedArticle(t, artRepo, feedRepo, "https://example.com/rss-b")

	// Try updating a disallowed field (content) and an allowed field (is_starred)
	updated, err := svc.UpdateArticle(artID, map[string]interface{}{
		"is_starred": true,
		"content":    "injected content", // not in whitelist, should be ignored
	})
	require.NoError(t, err)
	assert.True(t, updated.IsStarred, "is_starred 应被更新")
	assert.Equal(t, "article content", updated.Content, "content 不应被更新")
}

func TestArticleService_UpdateArticle_EmptyUpdates(t *testing.T) {
	db := testutil.SetupMySQL(t)
	feedRepo := repository.NewFeedRepository(db)
	artRepo := repository.NewArticleRepository(db)
	svc := service.NewArticleService(artRepo)

	artID := seedArticle(t, artRepo, feedRepo, "https://example.com/rss-c")

	// Passing an empty map (no valid fields) should return original article without error
	art, err := svc.UpdateArticle(artID, map[string]interface{}{})
	require.NoError(t, err)
	assert.NotNil(t, art)
}
