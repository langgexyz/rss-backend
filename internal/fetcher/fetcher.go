package fetcher

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"time"

	"github.com/mmcdole/gofeed"
)

type Fetcher struct {
	client  *http.Client
	parser  *gofeed.Parser
	maxArts int
}

func New(timeoutSecs, maxArticles int) *Fetcher {
	return &Fetcher{
		client:  &http.Client{Timeout: time.Duration(timeoutSecs) * time.Second},
		parser:  gofeed.NewParser(),
		maxArts: maxArticles,
	}
}

func (f *Fetcher) Fetch(ctx context.Context, url string) (*NormalizedFeed, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	feed, err := f.parser.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parse feed: %w", err)
	}
	return f.normalize(feed), nil
}

func (f *Fetcher) normalize(feed *gofeed.Feed) *NormalizedFeed {
	nf := &NormalizedFeed{
		Title:       feed.Title,
		Description: feed.Description,
		SiteURL:     feed.Link,
	}
	if feed.UpdatedParsed != nil {
		nf.UpdatedAt = feed.UpdatedParsed.UTC()
	} else if feed.PublishedParsed != nil {
		nf.UpdatedAt = feed.PublishedParsed.UTC()
	}

	items := feed.Items
	if len(items) > f.maxArts {
		items = items[:f.maxArts]
	}

	for _, item := range items {
		na := NormalizedArticle{
			GUIDHash: computeGUIDHash(item),
			Title:    item.Title,
			Link:     item.Link,
			Author:   authorName(item),
			Content:  itemContent(item),
		}
		switch {
		case item.PublishedParsed != nil:
			na.PublishedAt = item.PublishedParsed.UTC()
		case item.UpdatedParsed != nil:
			na.PublishedAt = item.UpdatedParsed.UTC()
		default:
			na.PublishedAt = time.Now().UTC()
		}
		nf.Articles = append(nf.Articles, na)
	}
	return nf
}

func computeGUIDHash(item *gofeed.Item) string {
	if item.GUID != "" {
		return sha256Hex(item.GUID)
	}
	if item.Link != "" {
		return sha256Hex(item.Link)
	}
	return sha256Hex(item.Title + item.Published)
}

func sha256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", h)
}

func authorName(item *gofeed.Item) string {
	if item.Author != nil {
		return item.Author.Name
	}
	return ""
}

func itemContent(item *gofeed.Item) string {
	if item.Content != "" {
		return item.Content
	}
	return item.Description
}
