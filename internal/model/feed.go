package model

import "time"

type Feed struct {
	ID              uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	URL             string     `gorm:"type:varchar(500);not null;uniqueIndex" json:"url"`
	Title           string     `gorm:"type:varchar(512);not null;default:''" json:"title"`
	Description     string     `gorm:"type:text" json:"description"`
	SiteURL         string     `gorm:"type:varchar(2048)" json:"site_url"`
	FetchStatus     string     `gorm:"type:enum('pending','fetching','success','failed');not null;default:'pending'" json:"fetch_status"`
	FetchError      string     `gorm:"type:varchar(1024)" json:"fetch_error"`
	LastFetchedAt   *time.Time `json:"last_fetched_at"`
	SourceUpdatedAt *time.Time `json:"source_updated_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}
