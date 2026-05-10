package feed

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mmcdole/gofeed"
	"github.com/redis/go-redis/v9"
)

type Manager struct {
	redis       *redis.Client
	pgx         *pgxpool.Pool
	scrapeTimer *time.Timer
	mu          sync.Mutex
	wg          sync.WaitGroup
	logClient   interface{ Info(ctx context.Context, msg string) error; Error(ctx context.Context, msg string) error }
}

func NewManager(redisClient *redis.Client, pgx *pgxpool.Pool, logClient interface{ Info(ctx context.Context, msg string) error; Error(ctx context.Context, msg string) error }) *Manager {
	return &Manager{
		redis:     redisClient,
		pgx:       pgx,
		logClient: logClient,
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
		m.wg.Add(1)
		go func(f rssFeed) {
			defer m.wg.Done()
			_ = m.scrapeFeed(ctx, parser, f)
		}(feed)
	}
	m.wg.Wait()

	return nil
}

type rssFeed struct {
	Name string
	URL  string
}

func (m *Manager) getActiveFeeds(ctx context.Context) ([]rssFeed, error) {
	if m.logClient != nil {
		m.logClient.Info(ctx, "Querying active feeds...")
	}
	rows, err := m.pgx.Query(ctx, "SELECT name, url FROM rss_feeds WHERE is_active = true")
	if err != nil {
		if m.logClient != nil {
			m.logClient.Error(ctx, fmt.Sprintf("Query feeds error: %v", err))
		}
		return nil, fmt.Errorf("query feeds: %w", err)
	}
	defer rows.Close()

	var feeds []rssFeed
	for rows.Next() {
		var f rssFeed
		if err := rows.Scan(&f.Name, &f.URL); err != nil {
			continue
		}
		feeds = append(feeds, f)
	}
	if m.logClient != nil {
		m.logClient.Info(ctx, fmt.Sprintf("Found %d active feeds", len(feeds)))
	}
	return feeds, nil
}

func (m *Manager) scrapeFeed(ctx context.Context, parser *gofeed.Parser, feed rssFeed) error {
	if m.logClient != nil {
		m.logClient.Info(ctx, fmt.Sprintf("Scraping feed: %s", feed.URL))
	}
	parsed, err := parser.ParseURLWithContext(feed.URL, ctx)
	if err != nil {
		return fmt.Errorf("parse feed %s: %w", feed.URL, err)
	}

	if m.logClient != nil {
		m.logClient.Info(ctx, fmt.Sprintf("Feed %s parsed, items count: %d", feed.Name, len(parsed.Items)))
	}
	for i, item := range parsed.Items {
		if item.Link == "" {
			if m.logClient != nil {
				m.logClient.Info(ctx, fmt.Sprintf("Item %d: no link, skipping", i))
			}
			continue
		}
		if m.logClient != nil {
			m.logClient.Info(ctx, fmt.Sprintf("Item %d: title=%s, link=%s", i, item.Title, item.Link))
		}
		article := NewsArticle{
			ID:          uuid.New(),
			TickerIDs:   []string{},
			Source:      feed.Name,
			Title:       item.Title,
			URL:         item.Link,
			Summary:     truncateString(item.Description, 500),
			PublishedAt: timePtr(item.Published),
			FetchedAt:   time.Now().UTC(),
		}
		if err := m.storeArticle(ctx, article); err != nil {
			if m.logClient != nil {
				m.logClient.Error(ctx, fmt.Sprintf("Failed to store article: %v", err))
			}
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

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func (m *Manager) storeArticle(ctx context.Context, article NewsArticle) error {
	if m.logClient != nil {
		m.logClient.Info(ctx, fmt.Sprintf("Storing article: %s from %s", article.Title, article.Source))
	}
	_, err := m.pgx.Exec(ctx, `
		INSERT INTO news_articles (id, ticker_ids, source, title, url, summary, sentiment, published_at, fetched_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (url) DO UPDATE SET
			title = EXCLUDED.title,
			summary = EXCLUDED.summary,
			fetched_at = EXCLUDED.fetched_at
	`, article.ID, article.TickerIDs, article.Source, article.Title, article.URL, article.Summary, article.Sentiment, article.PublishedAt, article.FetchedAt)
	if err != nil {
		if m.logClient != nil {
			m.logClient.Error(ctx, fmt.Sprintf("Store article error: %v", err))
		}
		return err
	}
	if m.logClient != nil {
		m.logClient.Info(ctx, fmt.Sprintf("Article stored successfully: %s", article.Title))
	}
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