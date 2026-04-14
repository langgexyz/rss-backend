package model

import "time"

type Feed struct {
	ID              uint       `gorm:"primaryKey;autoIncrement"`
	URL             string     `gorm:"type:varchar(500);not null;uniqueIndex"`
	Title           string     `gorm:"type:varchar(512);not null;default:''"`
	Description     string     `gorm:"type:text"`
	SiteURL         string     `gorm:"type:varchar(2048)"`
	FetchStatus     string     `gorm:"type:enum('pending','fetching','success','failed');not null;default:'pending'"`
	FetchError      string     `gorm:"type:varchar(1024)"`
	LastFetchedAt   *time.Time
	SourceUpdatedAt *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
