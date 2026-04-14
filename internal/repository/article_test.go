package repository_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"rss-backend/internal/fetcher"
	"rss-backend/internal/repository"
)

func TestArticleRepository_Upsert_Idempotent(t *testing.T) {
	db := setupTestDB(t)
	feedRepo := repository.NewFeedRepository(db)
	artRepo := repository.NewArticleRepository(db)

	feed, _ := feedRepo.Create("https://example.com/rss1")
	arts := []fetcher.NormalizedArticle{
		{GUIDHash: "aaaa1111aaaa1111aaaa1111aaaa1111aaaa1111aaaa1111aaaa1111aaaa1111", Title: "Hello", Link: "https://x.com/1", Content: "body", PublishedAt: time.Now()},
	}

	err := artRepo.Upsert(feed.ID, arts)
	require.NoError(t, err)

	err = artRepo.Upsert(feed.ID, arts)
	require.NoError(t, err)

	result, err := artRepo.List(repository.ArticleFilter{Page: 1, PageSize: 10})
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.Total, "幂等入库后应只有一条记录")
}

func TestArticleRepository_Upsert_DoesNotOverwriteUserState(t *testing.T) {
	db := setupTestDB(t)
	feedRepo := repository.NewFeedRepository(db)
	artRepo := repository.NewArticleRepository(db)

	feed, _ := feedRepo.Create("https://example.com/rss2")
	arts := []fetcher.NormalizedArticle{
		{GUIDHash: "bbbb2222bbbb2222bbbb2222bbbb2222bbbb2222bbbb2222bbbb2222bbbb2222", Title: "Title", Link: "https://x.com/1", Content: "body", PublishedAt: time.Now()},
	}
	artRepo.Upsert(feed.ID, arts)

	result, _ := artRepo.List(repository.ArticleFilter{Page: 1, PageSize: 10})
	artID := result.Items[0].ID
	artRepo.Update(artID, map[string]interface{}{"is_read": true})

	artRepo.Upsert(feed.ID, arts)

	updated, _ := artRepo.GetByID(artID)
	assert.True(t, updated.IsRead, "is_read 不应被 upsert 覆盖")
}

func TestArticleRepository_Upsert_ProtectsFullContent(t *testing.T) {
	db := setupTestDB(t)
	feedRepo := repository.NewFeedRepository(db)
	artRepo := repository.NewArticleRepository(db)

	feed, _ := feedRepo.Create("https://example.com/rss3")
	arts := []fetcher.NormalizedArticle{
		{GUIDHash: "cccc3333cccc3333cccc3333cccc3333cccc3333cccc3333cccc3333cccc3333", Title: "Title", Content: "full content", PublishedAt: time.Now()},
	}
	artRepo.Upsert(feed.ID, arts)

	result, _ := artRepo.List(repository.ArticleFilter{Page: 1, PageSize: 10})
	artID := result.Items[0].ID
	artRepo.Update(artID, map[string]interface{}{"is_full_content": true})

	arts[0].Content = "short summary"
	artRepo.Upsert(feed.ID, arts)

	updated, _ := artRepo.GetByID(artID)
	assert.Equal(t, "full content", updated.Content, "is_full_content=1 时 content 不应被覆盖")
}

func TestArticleRepository_List_FilterByFeed(t *testing.T) {
	db := setupTestDB(t)
	feedRepo := repository.NewFeedRepository(db)
	artRepo := repository.NewArticleRepository(db)

	feed1, _ := feedRepo.Create("https://a.com/rss4")
	feed2, _ := feedRepo.Create("https://b.com/rss4")

	artRepo.Upsert(feed1.ID, []fetcher.NormalizedArticle{
		{GUIDHash: "dddd4444dddd4444dddd4444dddd4444dddd4444dddd4444dddd4444dddd4444", Title: "A1", PublishedAt: time.Now()},
	})
	artRepo.Upsert(feed2.ID, []fetcher.NormalizedArticle{
		{GUIDHash: "eeee5555eeee5555eeee5555eeee5555eeee5555eeee5555eeee5555eeee5555", Title: "B1", PublishedAt: time.Now()},
	})

	result, err := artRepo.List(repository.ArticleFilter{FeedID: &feed1.ID, Page: 1, PageSize: 10})
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.Total)
	assert.Equal(t, "A1", result.Items[0].Title)
}
