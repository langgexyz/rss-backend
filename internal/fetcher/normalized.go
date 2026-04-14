package fetcher

import "time"

type NormalizedFeed struct {
	Title       string
	Description string
	SiteURL     string
	UpdatedAt   time.Time
	Articles    []NormalizedArticle
}

// GUIDHash 降级链：SHA256(原始guid/id) → SHA256(link) → SHA256(title+pubDate字符串)
type NormalizedArticle struct {
	GUIDHash    string
	Title       string
	Link        string
	Content     string
	Author      string
	PublishedAt time.Time
}
