package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"onlyfans-event-publisher/internal/model"

	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/kmsg"
)

// PlatformPublisher handles publishing content and creator events to Redpanda
type PlatformPublisher struct {
	client       *kgo.Client
	contentTopic string
	creatorTopic string
}

// NewPlatformPublisher creates a new platform publisher
func NewPlatformPublisher(ctx context.Context, brokers, contentTopic, creatorTopic string) (*PlatformPublisher, error) {
	// Create Redpanda client options
	opts := []kgo.Opt{
		kgo.SeedBrokers(strings.Split(brokers, ",")...),
		kgo.AllowAutoTopicCreation(),
		kgo.ProducerBatchMaxBytes(1024 * 1024), // 1MB
		kgo.ProducerLinger(5 * time.Millisecond),
		kgo.RecordRetries(3),
		kgo.RetryTimeout(10 * time.Second),
	}

	// Create client
	client, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Redpanda client: %w", err)
	}

	// Test connection
	if err := checkConnection(ctx, client); err != nil {
		client.Close()
		return nil, err
	}

	return &PlatformPublisher{
		client:       client,
		contentTopic: contentTopic,
		creatorTopic: creatorTopic,
	}, nil
}

// checkConnection verifies the connection to Redpanda
func checkConnection(ctx context.Context, client *kgo.Client) error {
	// Attempt to list topics to check connection
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Request a list of topics to verify connection
	req := kmsg.MetadataRequest{
		Topics: []kmsg.MetadataRequestTopic{},
	}

	resp, err := client.Request(ctx, &req)
	if err != nil {
		return fmt.Errorf("failed to connect to Redpanda: %w", err)
	}

	metaResp := resp.(*kmsg.MetadataResponse)
	if len(metaResp.Brokers) == 0 {
		return fmt.Errorf("no brokers found in Redpanda cluster")
	}

	return nil
}

// PublishContent publishes a content post to the content topic
func (p *PlatformPublisher) PublishContent(ctx context.Context, content model.Content) error {
	// Marshal content to JSON
	data, err := json.Marshal(content)
	if err != nil {
		return fmt.Errorf("failed to marshal content: %w", err)
	}

	// Create record with content ID as key
	record := &kgo.Record{
		Topic: p.contentTopic,
		Key:   []byte(content.ID),
		Value: data,
	}

	// Produce record
	result := p.client.ProduceSync(ctx, record)
	if err := result.FirstErr(); err != nil {
		return fmt.Errorf("failed to produce content record: %w", err)
	}

	return nil
}

// PublishCreator publishes a creator update to the creator topic
func (p *PlatformPublisher) PublishCreator(ctx context.Context, creator model.Creator) error {
	// Marshal creator to JSON
	data, err := json.Marshal(creator)
	if err != nil {
		return fmt.Errorf("failed to marshal creator: %w", err)
	}

	// Create record with creator ID as key
	record := &kgo.Record{
		Topic: p.creatorTopic,
		Key:   []byte(creator.ID),
		Value: data,
	}

	// Produce record
	result := p.client.ProduceSync(ctx, record)
	if err := result.FirstErr(); err != nil {
		return fmt.Errorf("failed to produce creator record: %w", err)
	}

	return nil
}

// PublishContentBatch publishes multiple content posts to the content topic
func (p *PlatformPublisher) PublishContentBatch(ctx context.Context, contents []model.Content) error {
	if len(contents) == 0 {
		return nil
	}

	// Create records
	records := make([]*kgo.Record, len(contents))
	for i, content := range contents {
		// Marshal content to JSON
		data, err := json.Marshal(content)
		if err != nil {
			return fmt.Errorf("failed to marshal content: %w", err)
		}

		// Create record
		records[i] = &kgo.Record{
			Topic: p.contentTopic,
			Key:   []byte(content.ID),
			Value: data,
		}
	}

	// Produce records
	results := p.client.ProduceSync(ctx, records...)
	for _, result := range results {
		if err := result.Err; err != nil {
			return fmt.Errorf("failed to produce content batch: %w", err)
		}
	}

	return nil
}

// PublishCreatorBatch publishes multiple creator updates to the creator topic
func (p *PlatformPublisher) PublishCreatorBatch(ctx context.Context, creators []model.Creator) error {
	if len(creators) == 0 {
		return nil
	}

	// Create records
	records := make([]*kgo.Record, len(creators))
	for i, creator := range creators {
		// Marshal creator to JSON
		data, err := json.Marshal(creator)
		if err != nil {
			return fmt.Errorf("failed to marshal creator: %w", err)
		}

		// Create record
		records[i] = &kgo.Record{
			Topic: p.creatorTopic,
			Key:   []byte(creator.ID),
			Value: data,
		}
	}

	// Produce records
	results := p.client.ProduceSync(ctx, records...)
	for _, result := range results {
		if err := result.Err; err != nil {
			return fmt.Errorf("failed to produce creator batch: %w", err)
		}
	}

	return nil
}

// PublishMixed publishes both content and creator updates in a single batch
func (p *PlatformPublisher) PublishMixed(ctx context.Context, contents []model.Content, creators []model.Creator) error {
	var records []*kgo.Record

	// Add content records
	for _, content := range contents {
		data, err := json.Marshal(content)
		if err != nil {
			return fmt.Errorf("failed to marshal content: %w", err)
		}

		records = append(records, &kgo.Record{
			Topic: p.contentTopic,
			Key:   []byte(content.ID),
			Value: data,
		})
	}

	// Add creator records
	for _, creator := range creators {
		data, err := json.Marshal(creator)
		if err != nil {
			return fmt.Errorf("failed to marshal creator: %w", err)
		}

		records = append(records, &kgo.Record{
			Topic: p.creatorTopic,
			Key:   []byte(creator.ID),
			Value: data,
		})
	}

	if len(records) == 0 {
		return nil
	}

	// Produce all records
	results := p.client.ProduceSync(ctx, records...)
	for _, result := range results {
		if err := result.Err; err != nil {
			return fmt.Errorf("failed to produce mixed batch: %w", err)
		}
	}

	return nil
}

// GetTopics returns the configured topics
func (p *PlatformPublisher) GetTopics() (contentTopic, creatorTopic string) {
	return p.contentTopic, p.creatorTopic
}

// Close closes the Redpanda client
func (p *PlatformPublisher) Close() {
	if p.client != nil {
		p.client.Close()
	}
}