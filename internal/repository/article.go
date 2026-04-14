package repository

import (
	"gorm.io/gorm"
	"rss-backend/internal/fetcher"
	"rss-backend/internal/model"
)

type ArticleRepository struct {
	db *gorm.DB
}

func NewArticleRepository(db *gorm.DB) *ArticleRepository {
	return &ArticleRepository{db: db}
}

func (r *ArticleRepository) Upsert(feedID uint, articles []fetcher.NormalizedArticle) error {
	for _, a := range articles {
		err := r.db.Exec(`
			INSERT INTO articles (feed_id, guid_hash, title, link, content, author, published_at)
			VALUES (?, ?, ?, ?, ?, ?, ?)
			ON DUPLICATE KEY UPDATE
				title   = VALUES(title),
				link    = VALUES(link),
				author  = VALUES(author),
				content = IF(is_full_content = 1, content, VALUES(content))
		`, feedID, a.GUIDHash, a.Title, a.Link, a.Content, a.Author, a.PublishedAt).Error
		if err != nil {
			return err
		}
	}
	return nil
}

type ArticleFilter struct {
	FeedID   *uint
	Starred  *bool
	Unread   *bool
	Page     int
	PageSize int
}

type ArticleWithFeedTitle struct {
	model.Article
	FeedTitle string `json:"feed_title"`
}

type ArticleListResult struct {
	Total int64
	Items []ArticleWithFeedTitle
}

func (r *ArticleRepository) List(f ArticleFilter) (*ArticleListResult, error) {
	query := r.db.Model(&model.Article{}).
		Select("articles.*, feeds.title AS feed_title").
		Joins("LEFT JOIN feeds ON feeds.id = articles.feed_id").
		Order("articles.published_at DESC")

	if f.FeedID != nil {
		query = query.Where("articles.feed_id = ?", *f.FeedID)
	}
	if f.Starred != nil {
		query = query.Where("articles.is_starred = ?", *f.Starred)
	}
	if f.Unread != nil && *f.Unread {
		query = query.Where("articles.is_read = 0")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var items []ArticleWithFeedTitle
	offset := (f.Page - 1) * f.PageSize
	err := query.Limit(f.PageSize).Offset(offset).Scan(&items).Error
	return &ArticleListResult{Total: total, Items: items}, err
}

func (r *ArticleRepository) GetByID(id uint) (*ArticleWithFeedTitle, error) {
	var art ArticleWithFeedTitle
	err := r.db.Model(&model.Article{}).
		Select("articles.*, feeds.title AS feed_title").
		Joins("LEFT JOIN feeds ON feeds.id = articles.feed_id").
		Where("articles.id = ?", id).
		Scan(&art).Error
	if art.ID == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &art, err
}

func (r *ArticleRepository) Update(id uint, updates map[string]interface{}) (*model.Article, error) {
	if err := r.db.Model(&model.Article{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return nil, err
	}
	var art model.Article
	return &art, r.db.First(&art, id).Error
}
