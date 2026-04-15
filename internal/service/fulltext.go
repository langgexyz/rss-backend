package service

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/html"
	"rss-backend/internal/repository"
)

type FulltextService struct {
	artRepo *repository.ArticleRepository
	client  *http.Client
}

func NewFulltextService(ar *repository.ArticleRepository) *FulltextService {
	return &FulltextService{
		artRepo: ar,
		client:  &http.Client{Timeout: 15 * time.Second},
	}
}

// FetchFulltext judges if article is summary-only, fetches full content if needed, and updates the content field
func (s *FulltextService) FetchFulltext(id uint) (*repository.ArticleWithFeedTitle, error) {
	art, err := s.artRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Already has full content, return directly
	if art.IsFullContent || len([]rune(art.Content)) >= 500 {
		return art, nil
	}

	if art.Link == "" {
		return art, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, art.Link, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 RSS-Reader/1.0")
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch fulltext: %w", err)
	}
	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parse html: %w", err)
	}

	content := extractMainContent(doc)
	if content == "" {
		return art, nil
	}

	s.artRepo.Update(id, map[string]interface{}{
		"content":          content,
		"is_full_content": true,
	})
	art.Content = content
	art.IsFullContent = true
	return art, nil
}

// noiseTags 是正文里不需要的 UI 元素，渲染前统一删除
var noiseTag = map[string]bool{
	"script": true, "style": true, "nav": true, "header": true,
	"footer": true, "aside": true, "button": true, "form": true,
	"input": true, "select": true, "textarea": true, "iframe": true,
	"noscript": true, "figure": true, "figcaption": true,
}

func extractMainContent(doc *html.Node) string {
	if n := findNode(doc, "article"); n != nil {
		stripNoise(n)
		return nodeText(n)
	}
	if n := findNode(doc, "main"); n != nil {
		stripNoise(n)
		return nodeText(n)
	}
	return longestDiv(doc)
}

// stripNoise 从节点树中原地删除所有噪声元素
func stripNoise(n *html.Node) {
	var toRemove []*html.Node
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode && noiseTag[node.Data] {
			toRemove = append(toRemove, node)
			return // 子节点一并删除，不再递归
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	for _, node := range toRemove {
		node.Parent.RemoveChild(node)
	}
}

func findNode(n *html.Node, tag string) *html.Node {
	if n.Type == html.ElementNode && n.Data == tag {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if found := findNode(c, tag); found != nil {
			return found
		}
	}
	return nil
}

func nodeText(n *html.Node) string {
	var buf strings.Builder
	if err := html.Render(&buf, n); err != nil {
		return ""
	}
	return strings.TrimSpace(buf.String())
}

func longestDiv(doc *html.Node) string {
	var longest string
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "div" {
			t := nodeText(n)
			if len(t) > len(longest) {
				longest = t
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return longest
}
