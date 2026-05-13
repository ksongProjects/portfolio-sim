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

type SentimentResult struct {
	Sentiment      string   `json:"sentiment"`
	RelatedTickers []string `json:"related_tickers"`
}

type Manager struct {
	redis        *redis.Client
	pgx          *pgxpool.Pool
	scrapeTimer  *time.Timer
	mu           sync.Mutex
	wg           sync.WaitGroup
	logClient    interface{ Info(ctx context.Context, msg string) error; Error(ctx context.Context, msg string) error }
	geminiClient interface {
		AnalyzeArticle(ctx context.Context, title, summary string) (string, []string, error)
	}
}

func NewManager(redisClient *redis.Client, pgx *pgxpool.Pool, logClient interface{ Info(ctx context.Context, msg string) error; Error(ctx context.Context, msg string) error }, geminiClient interface{ AnalyzeArticle(ctx context.Context, title, summary string) (string, []string, error) }) *Manager {
	return &Manager{
		redis:        redisClient,
		pgx:          pgx,
		logClient:    logClient,
		geminiClient: geminiClient,
	}
}

type NewsArticle struct {
	ID             uuid.UUID `json:"id"`
	Tickers        []string  `json:"tickers"`
	Source         string    `json:"source"`
	SourceType     string    `json:"source_type"`
	Title          string    `json:"title"`
	URL            string    `json:"url"`
	Summary        string    `json:"summary,omitempty"`
	Content        string    `json:"content,omitempty"`
	Sentiment      string    `json:"sentiment,omitempty"`
	SentimentValue string    `json:"sentiment_value,omitempty"`
	PublishedAt    time.Time `json:"published_at,omitempty"`
	FetchedAt      time.Time `json:"fetched_at"`
	Channel        string    `json:"channel,omitempty"`
}

type NewsArticleFormatter struct{}

func NewNewsArticleFormatter() *NewsArticleFormatter {
	return &NewsArticleFormatter{}
}

func (f *NewsArticleFormatter) FromRSS(article NewsArticle, sentiment, sentimentValue string, tickers []string) NewsArticle {
	return NewsArticle{
		ID:             article.ID,
		Tickers:        tickers,
		Source:         article.Source,
		SourceType:     "rss",
		SourceURL:      article.URL,
		Title:          article.Title,
		URL:            article.URL,
		Summary:        article.Summary,
		Sentiment:      sentiment,
		SentimentValue: sentimentValue,
		PublishedAt:    article.PublishedAt,
		FetchedAt:      time.Now().UTC(),
	}
}

func (f *NewsArticleFormatter) FromYouTube(videoID, title, channelName, transcript, sentiment, sentimentValue string, tickers []string, publishedAt time.Time) NewsArticle {
	return NewsArticle{
		ID:             uuid.New(),
		Tickers:        tickers,
		Source:         channelName,
		SourceType:     "youtube",
		SourceURL:      fmt.Sprintf("https://youtube.com/watch?v=%s", videoID),
		Title:          title,
		URL:            fmt.Sprintf("https://youtube.com/watch?v=%s", videoID),
		Summary:        truncateString(transcript, 500),
		Content:        transcript,
		Sentiment:      sentiment,
		SentimentValue: sentimentValue,
		PublishedAt:    publishedAt,
		FetchedAt:      time.Now().UTC(),
		Channel:        channelName,
	}
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
			Tickers:     []string{},
			Source:      feed.Name,
			SourceType:  "rss",
			SourceURL:   item.Link,
			Title:       item.Title,
			URL:         item.Link,
			Summary:     truncateString(item.Description, 500),
			PublishedAt: timePtr(item.Published),
			FetchedAt:   time.Now().UTC(),
		}

		if m.geminiClient != nil {
			sentiment, sentimentValue, tickers, err := m.geminiClient.AnalyzeArticle(ctx, item.Title, truncateString(item.Description, 1000))
			if err == nil {
				article.Sentiment = sentiment
				article.SentimentValue = sentimentValue
				article.Tickers = tickers
			} else if m.logClient != nil {
				m.logClient.Error(ctx, fmt.Sprintf("Gemini analysis failed for %s: %v", article.Title, err))
			}
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
	tickersJSON, _ := json.Marshal(article.Tickers)
	_, err := m.pgx.Exec(ctx, `
		INSERT INTO news_articles (id, tickers, source, source_type, title, url, summary, content, sentiment, sentiment_value, published_at, fetched_at, channel)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (url) DO UPDATE SET
			title = EXCLUDED.title,
			summary = EXCLUDED.summary,
			content = EXCLUDED.content,
			tickers = EXCLUDED.tickers,
			sentiment = EXCLUDED.sentiment,
			sentiment_value = EXCLUDED.sentiment_value,
			fetched_at = EXCLUDED.fetched_at
	`, article.ID, tickersJSON, article.Source, article.SourceType, article.Title, article.URL, article.Summary, article.Content, article.Sentiment, article.SentimentValue, article.PublishedAt, article.FetchedAt, article.Channel)
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