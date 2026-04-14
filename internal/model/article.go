package model

import "time"

type Article struct {
	ID            uint       `gorm:"primaryKey;autoIncrement"`
	FeedID        uint       `gorm:"not null;index;uniqueIndex:uq_feed_guid_hash"`
	GUIDHash      string     `gorm:"type:char(64);not null;uniqueIndex:uq_feed_guid_hash"`
	Title         string     `gorm:"type:varchar(1024);not null;default:''"`
	Link          string     `gorm:"type:varchar(2048)"`
	Content       string     `gorm:"type:mediumtext"`
	Author        string     `gorm:"type:varchar(512)"`
	PublishedAt   *time.Time
	IsRead        bool       `gorm:"not null;default:false"`
	IsStarred     bool       `gorm:"not null;default:false"`
	IsFullContent bool       `gorm:"not null;default:false"`
	CreatedAt     time.Time
}
