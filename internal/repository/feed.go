package repository

import (
	"time"

	"gorm.io/gorm"
	"rss-backend/internal/model"
)

type FeedRepository struct {
	db *gorm.DB
}

func NewFeedRepository(db *gorm.DB) *FeedRepository {
	return &FeedRepository{db: db}
}

func (r *FeedRepository) Create(url string) (*model.Feed, error) {
	feed := &model.Feed{URL: url}
	return feed, r.db.Create(feed).Error
}

func (r *FeedRepository) List() ([]model.Feed, error) {
	var feeds []model.Feed
	return feeds, r.db.Order("created_at DESC").Find(&feeds).Error
}

func (r *FeedRepository) GetByID(id uint) (*model.Feed, error) {
	var feed model.Feed
	return &feed, r.db.First(&feed, id).Error
}

func (r *FeedRepository) UpdateStatus(id uint, status, errMsg string, lastFetched, sourceUpdated *time.Time, title string) error {
	updates := map[string]interface{}{
		"fetch_status": status,
		"fetch_error":  errMsg,
	}
	if lastFetched != nil {
		updates["last_fetched_at"] = lastFetched
	}
	if sourceUpdated != nil && !sourceUpdated.IsZero() {
		updates["source_updated_at"] = sourceUpdated
	}
	if title != "" {
		updates["title"] = title
	}
	return r.db.Model(&model.Feed{}).Where("id = ?", id).Updates(updates).Error
}

func (r *FeedRepository) Delete(id uint) error {
	return r.db.Delete(&model.Feed{}, id).Error
}
