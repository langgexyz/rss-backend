package model

import "time"

type Article struct {
	ID            uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	FeedID        uint       `gorm:"not null;index;uniqueIndex:uq_feed_guid_hash" json:"feed_id"`
	GUIDHash      string     `gorm:"type:char(64);not null;uniqueIndex:uq_feed_guid_hash" json:"guid_hash"`
	Title         string     `gorm:"type:varchar(1024);not null;default:''" json:"title"`
	Link          string     `gorm:"type:varchar(2048)" json:"link"`
	Content       string     `gorm:"type:mediumtext" json:"content"`
	Author        string     `gorm:"type:varchar(512)" json:"author"`
	PublishedAt   *time.Time `json:"published_at"`
	IsRead        bool       `gorm:"not null;default:false" json:"is_read"`
	IsStarred     bool       `gorm:"not null;default:false" json:"is_starred"`
	IsFullContent bool       `gorm:"not null;default:false" json:"is_full_content"`
	CreatedAt     time.Time  `json:"created_at"`
}
