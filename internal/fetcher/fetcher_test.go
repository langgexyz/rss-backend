package fetcher_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"rss-backend/internal/fetcher"
)

const sampleRSS2 = `<?xml version="1.0"?>
<rss version="2.0">
  <channel>
    <title>Test Feed</title>
    <link>https://example.com</link>
    <lastBuildDate>Mon, 14 Apr 2026 10:00:00 +0000</lastBuildDate>
    <item>
      <title>Article One</title>
      <link>https://example.com/1</link>
      <guid>guid-001</guid>
      <description>Summary of article one</description>
      <author>Alice</author>
      <pubDate>Mon, 14 Apr 2026 09:00:00 +0000</pubDate>
    </item>
    <item>
      <title>Article Two</title>
      <link>https://example.com/2</link>
      <description>Summary of article two</description>
    </item>
  </channel>
</rss>`

func TestFetch_RSS2(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		w.Write([]byte(sampleRSS2))
	}))
	defer srv.Close()

	f := fetcher.New(10, 100)
	feed, err := f.Fetch(context.Background(), srv.URL)

	require.NoError(t, err)
	assert.Equal(t, "Test Feed", feed.Title)
	assert.Equal(t, "https://example.com", feed.SiteURL)
	assert.False(t, feed.UpdatedAt.IsZero())
	assert.Len(t, feed.Articles, 2)

	a := feed.Articles[0]
	assert.Equal(t, "Article One", a.Title)
	assert.Equal(t, "https://example.com/1", a.Link)
	assert.Equal(t, "Alice", a.Author)
	assert.NotEmpty(t, a.GUIDHash)
	assert.Len(t, a.GUIDHash, 64)
}

func TestFetch_GUIDHash_Fallback(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(sampleRSS2))
	}))
	defer srv.Close()

	f := fetcher.New(10, 100)
	feed, err := f.Fetch(context.Background(), srv.URL)
	require.NoError(t, err)

	a1 := feed.Articles[0]
	a2 := feed.Articles[1]
	assert.NotEqual(t, a1.GUIDHash, a2.GUIDHash)
}

func TestFetch_MaxArticles(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(sampleRSS2))
	}))
	defer srv.Close()

	f := fetcher.New(10, 1)
	feed, err := f.Fetch(context.Background(), srv.URL)
	require.NoError(t, err)
	assert.Len(t, feed.Articles, 1)
}

func TestFetch_InvalidURL(t *testing.T) {
	f := fetcher.New(1, 100)
	_, err := f.Fetch(context.Background(), "http://127.0.0.1:1")
	assert.Error(t, err)
}
