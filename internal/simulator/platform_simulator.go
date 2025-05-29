package simulator

import (
	"fmt"
	"math/rand"
	"onlyfans-event-publisher/internal/model"
	"time"
)

// PlatformSimulator simulates creator and content activity
type PlatformSimulator struct {
	creators             []model.Creator
	contentCounts        []int     // Number of content posted by each creator
	activityLevels       []float64 // Activity level for each creator (0-1)
	lastPostTimes        []time.Time
	subscriberTrends     []float64 // Subscriber growth trend
	engagementRates      []float64 // Base engagement rate per creator
	abnormalActivityProb float64
	rng                  *rand.Rand
}

// NewPlatformSimulator creates a new platform simulator
func NewPlatformSimulator(numCreators int, abnormalProb float64) *PlatformSimulator {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Create creators
	creators := make([]model.Creator, numCreators)
	contentCounts := make([]int, numCreators)
	activityLevels := make([]float64, numCreators)
	lastPostTimes := make([]time.Time, numCreators)
	subscriberTrends := make([]float64, numCreators)
	engagementRates := make([]float64, numCreators)

	baseTime := time.Now().Add(-time.Hour * 24 * 30) // Start 30 days ago

	// Initialize creators with realistic data
	for i := 0; i < numCreators; i++ {
		subscriberCount := r.Intn(10000) + 100       // 100-10,100 subscribers
		monthlyPrice := float64(r.Intn(45)+5) + 0.99 // $5.99-$49.99

		creators[i] = model.Creator{
			ID:              fmt.Sprintf("creator-%d", i),
			Username:        fmt.Sprintf("user_%d", i),
			DisplayName:     fmt.Sprintf("Creator %d", i),
			Email:           fmt.Sprintf("creator%d@platform.com", i),
			IsVerified:      r.Float64() < 0.3, // 30% verified
			SubscriberCount: subscriberCount,
			MonthlyPrice:    monthlyPrice,
			CreatedAt:       baseTime.Add(time.Duration(r.Intn(720)) * time.Hour), // Random creation time
			IsOnline:        r.Float64() < 0.4,                                    // 40% online initially
			Category:        model.CreatorCategories[r.Intn(len(model.CreatorCategories))],
			ProfilePic:      fmt.Sprintf("https://cdn.platform.com/profiles/creator-%d.jpg", i),
		}

		// Initialize activity patterns
		activityLevels[i] = r.Float64()*0.8 + 0.2 // 0.2-1.0 activity level
		lastPostTimes[i] = time.Now().Add(-time.Duration(r.Intn(48)) * time.Hour)
		subscriberTrends[i] = (r.Float64() - 0.5) * 0.02 // -1% to +1% daily trend
		engagementRates[i] = r.Float64()*0.15 + 0.05     // 5%-20% engagement rate
		contentCounts[i] = r.Intn(50) + 10               // Start with 10-60 posts
	}

	return &PlatformSimulator{
		creators:             creators,
		contentCounts:        contentCounts,
		activityLevels:       activityLevels,
		lastPostTimes:        lastPostTimes,
		subscriberTrends:     subscriberTrends,
		engagementRates:      engagementRates,
		abnormalActivityProb: abnormalProb,
		rng:                  r,
	}
}

// GetCreators returns the list of simulated creators
func (s *PlatformSimulator) GetCreators() []model.Creator {
	return s.creators
}

// GenerateCreatorUpdates generates creator status updates
func (s *PlatformSimulator) GenerateCreatorUpdates() []model.Creator {
	var updates []model.Creator

	for i := range s.creators {
		// 20% chance of creator update per cycle
		if s.rng.Float64() < 0.2 {
			updates = append(updates, s.generateCreatorUpdate(i))
		}
	}

	return updates
}

// GenerateContent generates new content posts
func (s *PlatformSimulator) GenerateContent() []model.Content {
	var content []model.Content

	for i := range s.creators {
		// Check if creator should post based on activity level and time since last post
		timeSincePost := time.Since(s.lastPostTimes[i])
		shouldPost := s.shouldCreatorPost(i, timeSincePost)

		if shouldPost {
			newContent := s.generateCreatorContent(i)
			content = append(content, newContent)
			s.lastPostTimes[i] = time.Now()
			s.contentCounts[i]++
		}
	}

	return content
}

// generateCreatorUpdate generates an updated creator profile
func (s *PlatformSimulator) generateCreatorUpdate(creatorIndex int) model.Creator {
	creator := s.creators[creatorIndex]

	// Update subscriber count based on trend
	subscriberChange := float64(creator.SubscriberCount) * s.subscriberTrends[creatorIndex]
	creator.SubscriberCount = int(float64(creator.SubscriberCount) + subscriberChange)
	if creator.SubscriberCount < 0 {
		creator.SubscriberCount = 0
	}

	// Update online status (60% chance of change)
	if s.rng.Float64() < 0.6 {
		creator.IsOnline = !creator.IsOnline
	}

	// Occasionally adjust monthly price (5% chance)
	if s.rng.Float64() < 0.05 {
		priceChange := (s.rng.Float64() - 0.5) * 10 // ±$5 change
		creator.MonthlyPrice = clamp(creator.MonthlyPrice+priceChange, 4.99, 99.99)
	}

	// Update trends occasionally
	if s.rng.Float64() < 0.1 {
		s.subscriberTrends[creatorIndex] += (s.rng.Float64() - 0.5) * 0.01
		s.subscriberTrends[creatorIndex] = clamp(s.subscriberTrends[creatorIndex], -0.05, 0.05)
	}

	// Update the stored creator
	s.creators[creatorIndex] = creator
	return creator
}

// shouldCreatorPost determines if a creator should post content
func (s *PlatformSimulator) shouldCreatorPost(creatorIndex int, timeSincePost time.Duration) bool {
	activityLevel := s.activityLevels[creatorIndex]

	// Base posting frequency: highly active creators post every 2-6 hours
	// Less active creators post every 12-48 hours
	baseInterval := time.Duration(2+((1-activityLevel)*46)) * time.Hour

	// Add randomness
	if timeSincePost < baseInterval/2 {
		return false // Too soon
	}

	// Probability increases with time
	probability := float64(timeSincePost) / float64(baseInterval)

	// Abnormal activity (posting spree)
	if s.rng.Float64() < s.abnormalActivityProb {
		probability *= 5 // Much higher chance during abnormal activity
	}

	return s.rng.Float64() < probability
}

// generateCreatorContent generates new content from a creator
func (s *PlatformSimulator) generateCreatorContent(creatorIndex int) model.Content {
	creator := s.creators[creatorIndex]
	contentID := fmt.Sprintf("content-%s-%d", creator.ID, s.contentCounts[creatorIndex])

	contentType := model.ContentTypes[s.rng.Intn(len(model.ContentTypes))]

	// Generate realistic engagement based on creator's subscriber count and engagement rate
	baseViews := int(float64(creator.SubscriberCount) * s.engagementRates[creatorIndex])
	viewCount := baseViews + s.rng.Intn(baseViews/2)                   // ±25% variation
	likeCount := int(float64(viewCount) * (0.1 + s.rng.Float64()*0.2)) // 10-30% like rate

	// Determine if content should be locked/premium
	isLocked := s.rng.Float64() < 0.4 // 40% premium content
	var price float64
	if isLocked {
		price = float64(s.rng.Intn(25)+5) + 0.99 // $5.99-$29.99 for premium
	}

	// Generate media URL based on content type
	var mediaURL string
	if contentType != "text" {
		mediaURL = fmt.Sprintf("https://cdn.platform.com/%s/%s.%s",
			contentType, contentID, getFileExtension(contentType))
	}

	// Generate tags
	tags := generateTags(creator.Category, contentType, s.rng)

	return model.Content{
		ID:          contentID,
		CreatorID:   creator.ID,
		Title:       generateContentTitle(contentType, creator.Category, s.rng),
		Description: generateContentDescription(contentType, s.rng),
		ContentType: contentType,
		MediaURL:    mediaURL,
		Price:       price,
		IsLocked:    isLocked,
		ViewCount:   viewCount,
		LikeCount:   likeCount,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Tags:        tags,
	}
}

// Helper functions
func getFileExtension(contentType string) string {
	switch contentType {
	case "image", "gallery":
		return "jpg"
	case "video", "live":
		return "mp4"
	default:
		return "txt"
	}
}

func generateContentTitle(contentType, category string, rng *rand.Rand) string {
	titles := map[string][]string{
		"fitness":   {"Workout Session", "Training Day", "Gym Time", "Fitness Tips"},
		"lifestyle": {"Daily Vibe", "Life Update", "Casual Moment", "Behind Scenes"},
		"art":       {"New Creation", "Art Process", "Creative Session", "Artistic Vision"},
		"music":     {"New Track", "Studio Session", "Live Performance", "Music Moment"},
		"gaming":    {"Gaming Session", "New Game", "Epic Win", "Game Review"},
	}

	categoryTitles, exists := titles[category]
	if !exists {
		categoryTitles = []string{"New Post", "Update", "Fresh Content", "Latest"}
	}

	return categoryTitles[rng.Intn(len(categoryTitles))]
}

func generateContentDescription(contentType string, rng *rand.Rand) string {
	descriptions := []string{
		"Check out my latest content!",
		"Hope you enjoy this one ❤️",
		"What do you think about this?",
		"Been working on this for a while...",
		"Exclusive content just for you!",
		"",
	}
	return descriptions[rng.Intn(len(descriptions))]
}

func generateTags(category, contentType string, rng *rand.Rand) []string {
	baseTags := []string{category, contentType}

	additionalTags := []string{"new", "exclusive", "hot", "trending", "premium", "special"}

	// Add 0-3 additional tags
	numAdditional := rng.Intn(4)
	for i := 0; i < numAdditional; i++ {
		tag := additionalTags[rng.Intn(len(additionalTags))]
		// Avoid duplicates
		exists := false
		for _, existing := range baseTags {
			if existing == tag {
				exists = true
				break
			}
		}
		if !exists {
			baseTags = append(baseTags, tag)
		}
	}

	return baseTags
}

// clamp ensures a value is within bounds
func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
