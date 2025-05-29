package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"onlyfans-event-publisher/internal/config"
	"onlyfans-event-publisher/internal/publisher"
	"onlyfans-event-publisher/internal/simulator"
)

func main() {
	// Setup logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting OnlyFans Event Publisher...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Configuration loaded:")
	log.Printf("  Redpanda Brokers: %s", cfg.RedpandaBrokers)
	log.Printf("  Content Topic: %s", cfg.ContentTopic)
	log.Printf("  Creator Topic: %s", cfg.CreatorTopic)
	log.Printf("  Number of Creators: %d", cfg.NumCreators)
	log.Printf("  Interval: %dms", cfg.IntervalMs)
	log.Printf("  Abnormal Probability: %.2f", cfg.AbnormalProbability)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create platform simulator
	log.Printf("Initializing platform simulator with %d creators...", cfg.NumCreators)
	sim := simulator.NewPlatformSimulator(cfg.NumCreators, cfg.AbnormalProbability)

	// Log initial creators
	creators := sim.GetCreators()
	log.Printf("Created %d creators:", len(creators))
	for i, creator := range creators {
		if i < 3 { // Log first 3 creators as examples
			log.Printf("  - %s (%s): %d subscribers, $%.2f/month, %s",
				creator.Username, creator.DisplayName, creator.SubscriberCount,
				creator.MonthlyPrice, creator.Category)
		}
	}
	if len(creators) > 3 {
		log.Printf("  ... and %d more creators", len(creators)-3)
	}

	// Create platform publisher
	log.Println("Connecting to Redpanda cluster...")
	pub, err := publisher.NewPlatformPublisher(ctx, cfg.RedpandaBrokers, cfg.ContentTopic, cfg.CreatorTopic)
	if err != nil {
		log.Fatalf("Failed to create publisher: %v", err)
	}
	defer pub.Close()

	contentTopic, creatorTopic := pub.GetTopics()
	log.Printf("Connected to Redpanda - Content Topic: %s, Creator Topic: %s", contentTopic, creatorTopic)

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Statistics tracking
	stats := &Statistics{
		StartTime: time.Now(),
	}

	// Main simulation loop
	ticker := time.NewTicker(time.Duration(cfg.IntervalMs) * time.Millisecond)
	defer ticker.Stop()

	log.Println("Starting simulation loop...")
	log.Println("Press Ctrl+C to stop gracefully")

	for {
		select {
		case <-ctx.Done():
			log.Println("Context cancelled, shutting down...")
			return

		case <-sigChan:
			log.Println("Received shutdown signal")
			printFinalStats(stats)
			cancel()
			return

		case <-ticker.C:
			if err := runSimulationCycle(ctx, sim, pub, stats); err != nil {
				log.Printf("Error in simulation cycle: %v", err)
				// Continue running even if there's an error
			}

			// Print stats every minute
			if stats.Cycles%60 == 0 && stats.Cycles > 0 {
				printPeriodicStats(stats)
			}
		}
	}
}

// Statistics holds runtime statistics
type Statistics struct {
	StartTime        time.Time
	Cycles           int64
	ContentPublished int64
	CreatorUpdates   int64
	PublishErrors    int64
	LastContentCount int
	LastCreatorCount int
}

// runSimulationCycle runs one cycle of the simulation
func runSimulationCycle(ctx context.Context, sim *simulator.PlatformSimulator, pub *publisher.PlatformPublisher, stats *Statistics) error {
	// Generate content and creator updates
	newContent := sim.GenerateContent()
	creatorUpdates := sim.GenerateCreatorUpdates()

	stats.Cycles++
	stats.LastContentCount = len(newContent)
	stats.LastCreatorCount = len(creatorUpdates)

	// Publish to Redpanda if we have data
	if len(newContent) > 0 || len(creatorUpdates) > 0 {
		// Use mixed publishing for efficiency
		if err := pub.PublishMixed(ctx, newContent, creatorUpdates); err != nil {
			stats.PublishErrors++
			return fmt.Errorf("failed to publish events: %w", err)
		}

		stats.ContentPublished += int64(len(newContent))
		stats.CreatorUpdates += int64(len(creatorUpdates))

		// Log activity
		if len(newContent) > 0 && len(creatorUpdates) > 0 {
			log.Printf("Published %d content posts and %d creator updates", len(newContent), len(creatorUpdates))
		} else if len(newContent) > 0 {
			log.Printf("Published %d content posts", len(newContent))
		} else if len(creatorUpdates) > 0 {
			log.Printf("Published %d creator updates", len(creatorUpdates))
		}

		// Log some sample content for debugging
		if len(newContent) > 0 {
			sample := newContent[0]
			log.Printf("  Sample content: '%s' by %s (%s) - %d views, %d likes",
				sample.Title, sample.CreatorID, sample.ContentType, sample.ViewCount, sample.LikeCount)
		}
	}

	return nil
}

// printPeriodicStats prints statistics every minute
func printPeriodicStats(stats *Statistics) {
	uptime := time.Since(stats.StartTime)
	avgContentPerMin := float64(stats.ContentPublished) / uptime.Minutes()
	avgCreatorPerMin := float64(stats.CreatorUpdates) / uptime.Minutes()

	log.Printf("=== Statistics (Uptime: %v) ===", uptime.Round(time.Second))
	log.Printf("Cycles: %d", stats.Cycles)
	log.Printf("Content Published: %d (%.1f/min)", stats.ContentPublished, avgContentPerMin)
	log.Printf("Creator Updates: %d (%.1f/min)", stats.CreatorUpdates, avgCreatorPerMin)
	log.Printf("Publish Errors: %d", stats.PublishErrors)
	log.Printf("Last Cycle: %d content, %d creators", stats.LastContentCount, stats.LastCreatorCount)
	log.Printf("===============================")
}

// printFinalStats prints final statistics on shutdown
func printFinalStats(stats *Statistics) {
	uptime := time.Since(stats.StartTime)

	log.Println("=== Final Statistics ===")
	log.Printf("Total Uptime: %v", uptime.Round(time.Second))
	log.Printf("Total Cycles: %d", stats.Cycles)
	log.Printf("Content Published: %d", stats.ContentPublished)
	log.Printf("Creator Updates: %d", stats.CreatorUpdates)
	log.Printf("Total Events: %d", stats.ContentPublished+stats.CreatorUpdates)
	log.Printf("Publish Errors: %d", stats.PublishErrors)

	if uptime.Minutes() > 0 {
		log.Printf("Average Events/min: %.1f", float64(stats.ContentPublished+stats.CreatorUpdates)/uptime.Minutes())
	}

	if stats.Cycles > 0 {
		log.Printf("Success Rate: %.2f%%", float64(stats.Cycles-stats.PublishErrors)/float64(stats.Cycles)*100)
	}

	log.Println("========================")
	log.Println("Shutdown complete")
}
