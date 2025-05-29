package model

import "time"

// Content represents a post/content shared by creators
type Content struct {
	ID          string    `json:"id"`
	CreatorID   string    `json:"creator_id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	ContentType string    `json:"content_type"` // "image", "video", "text", "live"
	MediaURL    string    `json:"media_url,omitempty"`
	Price       float64   `json:"price"`     // 0 for free content
	IsLocked    bool      `json:"is_locked"` // Premium content requiring payment
	ViewCount   int       `json:"view_count"`
	LikeCount   int       `json:"like_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Tags        []string  `json:"tags,omitempty"`
}

// Content types for simulation
var ContentTypes = []string{
	"image",
	"video",
	"text",
	"live",
	"gallery",
}
