package model

import "time"

// Creator represents a content creator on the platform
type Creator struct {
	ID              string    `json:"id"`
	Username        string    `json:"username"`
	DisplayName     string    `json:"display_name"`
	Email           string    `json:"email"`
	IsVerified      bool      `json:"is_verified"`
	SubscriberCount int       `json:"subscriber_count"`
	MonthlyPrice    float64   `json:"monthly_price"`
	CreatedAt       time.Time `json:"created_at"`
	IsOnline        bool      `json:"is_online"`
	Category        string    `json:"category"`
	ProfilePic      string    `json:"profile_pic,omitempty"`
}

// Creator categories for simulation
var CreatorCategories = []string{
	"fitness",
	"lifestyle",
	"art",
	"music",
	"gaming",
	"cooking",
	"fashion",
	"photography",
	"education",
	"entertainment",
}
