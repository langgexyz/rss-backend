package repository_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"rss-backend/internal/repository"
	"rss-backend/internal/testutil"
)

func TestFeedRepository_Create(t *testing.T) {
	db := testutil.SetupMySQL(t)
	repo := repository.NewFeedRepository(db)

	feed, err := repo.Create("https://example.com/rss.xml")
	require.NoError(t, err)
	assert.NotZero(t, feed.ID)
	assert.Equal(t, "pending", feed.FetchStatus)
}

func TestFeedRepository_Create_DuplicateURL(t *testing.T) {
	db := testutil.SetupMySQL(t)
	repo := repository.NewFeedRepository(db)

	_, err := repo.Create("https://example.com/rss.xml")
	require.NoError(t, err)

	_, err = repo.Create("https://example.com/rss.xml")
	assert.Error(t, err, "重复 URL 应返回错误")
}

func TestFeedRepository_List(t *testing.T) {
	db := testutil.SetupMySQL(t)
	repo := repository.NewFeedRepository(db)

	repo.Create("https://a.com/rss")
	repo.Create("https://b.com/rss")

	feeds, err := repo.List()
	require.NoError(t, err)
	assert.Len(t, feeds, 2)
}

func TestFeedRepository_UpdateStatus(t *testing.T) {
	db := testutil.SetupMySQL(t)
	repo := repository.NewFeedRepository(db)

	feed, _ := repo.Create("https://example.com/rss.xml")
	err := repo.UpdateStatus(feed.ID, "success", "", nil, nil, "My Blog")
	require.NoError(t, err)

	updated, _ := repo.GetByID(feed.ID)
	assert.Equal(t, "success", updated.FetchStatus)
	assert.Equal(t, "My Blog", updated.Title)
}
