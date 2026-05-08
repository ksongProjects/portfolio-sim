package feed

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/mmcdole/gofeed"
	"github.com/redis/go-redis/v9"
)

type Manager struct {
	redis       *redis.Client
	pgx         interface{}
	gemini      interface{}
	scrapeTimer *time.Timer
	mu          sync.Mutex
}

func NewManager(redisClient *redis.Client) *Manager {
	return &Manager{
		redis: redisClient,
	}
}

type NewsArticle struct {
	ID          uuid.UUID `json:"id"`
	TickerIDs   []string  `json:"ticker_ids"`
	Source      string    `json:"source"`
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	Summary     string    `json:"summary,omitempty"`
	Sentiment   string    `json:"sentiment,omitempty"`
	PublishedAt time.Time `json:"published_at,omitempty"`
	FetchedAt   time.Time `json:"fetched_at"`
}

func (m *Manager) ScrapeFeeds(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	parser := gofeed.NewParser()
	feeds, err := m.getActiveFeeds(ctx)
	if err != nil {
		return fmt.Errorf("get feeds: %w", err)
	}

	for _, feed := range feeds {
		go m.scrapeFeed(ctx, parser, feed)
	}

	return nil
}

type rssFeed struct {
	Name string
	URL  string
}

func (m *Manager) getActiveFeeds(ctx context.Context) ([]rssFeed, error) {
	_ = m.redis
	return []rssFeed{}, nil
}

func (m *Manager) scrapeFeed(ctx context.Context, parser *gofeed.Parser, feed rssFeed) error {
	parsed, err := parser.ParseURLWithContext(feed.URL, ctx)
	if err != nil {
		return fmt.Errorf("parse feed %s: %w", feed.URL, err)
	}

	for _, item := range parsed.Items {
		article := NewsArticle{
			ID:          uuid.New(),
			Source:      feed.Name,
			Title:       item.Title,
			URL:         item.Link,
			PublishedAt: timePtr(item.Published),
			FetchedAt:   time.Now().UTC(),
		}
		if err := m.storeArticle(ctx, article); err != nil {
			continue
		}
		m.publishArticle(ctx, article)
	}

	return nil
}

func timePtr(t string) time.Time {
	if t == "" {
		return time.Time{}
	}
	parsed, _ := time.Parse(time.RFC1123, t)
	if parsed.IsZero() {
		parsed, _ = time.Parse(time.RFC822, t)
	}
	return parsed
}

func (m *Manager) storeArticle(ctx context.Context, article NewsArticle) error {
	data, _ := json.Marshal(article)
	_ = m.redis.Publish(ctx, "news:articles", data).Err()
	return nil
}

func (m *Manager) publishArticle(ctx context.Context, article NewsArticle) {
	data, _ := json.Marshal(article)
	m.redis.Publish(ctx, "news:articles", data)
}

func (m *Manager) StartScheduler(ctx context.Context, interval time.Duration) {
	m.scrapeTimer = time.NewTimer(interval)
	go func() {
		for {
			select {
			case <-m.scrapeTimer.C:
				_ = m.ScrapeFeeds(ctx)
				m.scrapeTimer.Reset(interval)
			case <-ctx.Done():
				return
			}
		}
	}()
}
