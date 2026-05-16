package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/portfolio-sim/news-feed-service/config"
	"github.com/portfolio-sim/news-feed-service/database"
	"github.com/portfolio-sim/news-feed-service/feed"
	"github.com/portfolio-sim/news-feed-service/gemini"
	"github.com/portfolio-sim/news-feed-service/logging"
	"github.com/portfolio-sim/news-feed-service/redis"
	"github.com/portfolio-sim/news-feed-service/sse"
	"github.com/portfolio-sim/news-feed-service/youtube"
)

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

type videoSummaryRequest struct {
	VideoID string `json:"video_id"`
	Title   string `json:"title,omitempty"`
}

type videoSummaryResult struct {
	VideoID          string `json:"video_id"`
	Title            string `json:"title,omitempty"`
	Summary          string `json:"summary,omitempty"`
	Status           string `json:"status"`
	Error            string `json:"error,omitempty"`
	TranscriptSource string `json:"transcript_source,omitempty"`
}

func fetchProviderKey(providerID string) string {
	mainAPIURL := os.Getenv("MAIN_API_URL")
	if mainAPIURL == "" {
		mainAPIURL = "http://main-api:8080"
	}
	internalToken := os.Getenv("INTERNAL_API_TOKEN")
	if internalToken == "" {
		log.Printf("fetch provider key %s: INTERNAL_API_TOKEN not configured", providerID)
		return ""
	}
	url := fmt.Sprintf("%s/internal/providers/%s", mainAPIURL, providerID)

	for attempt := 1; attempt <= 12; attempt++ {
		key, retry := fetchProviderKeyOnce(providerID, url, internalToken)
		if key != "" {
			return key
		}
		if !retry {
			return ""
		}
		time.Sleep(2 * time.Second)
	}
	return ""
}

func fetchProviderKeyOnce(providerID, url, internalToken string) (string, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", false
	}
	req.Header.Set("X-Internal-API-Token", internalToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("fetch provider key %s: %v", providerID, err)
		return "", true
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("fetch provider key %s: status %d", providerID, resp.StatusCode)
		return "", resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500
	}

	var result struct {
		APIKey string `json:"api_key"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("fetch provider key %s: decode error: %v", providerID, err)
		return "", false
	}
	return result.APIKey, false
}

func summarizeVideo(ctx context.Context, db *database.Postgres, ytClient *youtube.Client, gemClient *gemini.Client, req videoSummaryRequest) videoSummaryResult {
	videoID := strings.TrimSpace(req.VideoID)
	result := videoSummaryResult{VideoID: videoID, Status: "error"}
	if videoID == "" {
		result.Error = "video_id required"
		return result
	}
	if ytClient == nil {
		result.Error = "youtube client not configured"
		return result
	}
	if gemClient == nil {
		result.Error = "gemini client not configured"
		return result
	}

	details, err := ytClient.GetVideoDetails(ctx, videoID)
	if err != nil {
		result.Error = fmt.Sprintf("failed to get video details: %v", err)
		return result
	}
	if details == nil {
		result.Error = "video not found"
		return result
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		title = details.Title
	}
	result.Title = title

	transcript, transcriptErr := ytClient.GetTranscript(ctx, videoID)
	transcript = strings.TrimSpace(transcript)
	transcriptSource := "captions"
	if transcript == "" {
		transcript = strings.TrimSpace(details.Description)
		transcriptSource = "description"
	}
	if transcript == "" {
		if transcriptErr != nil {
			result.Error = fmt.Sprintf("no transcript or description available: %v", transcriptErr)
		} else {
			result.Error = "no transcript or description available"
		}
		return result
	}

	summary, err := gemClient.SummarizeVideo(ctx, title, transcript)
	if err != nil {
		result.Error = fmt.Sprintf("failed to summarize video: %v", err)
		return result
	}

	sentiment := "neutral"
	sentimentValue := "0.5"
	tickers := []string{}
	if analyzedSentiment, analyzedValue, analyzedTickers, err := gemClient.AnalyzeArticle(ctx, title, transcript); err == nil {
		sentiment = analyzedSentiment
		if analyzedValue != "" {
			sentimentValue = analyzedValue
		}
		tickers = analyzedTickers
	} else {
		log.Printf("video sentiment analysis failed: %v", err)
	}

	tickersJSON, _ := json.Marshal(tickers)
	articleID := uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO news_articles (id, tickers, source, source_type, title, url, summary, content, sentiment, sentiment_value, published_at, fetched_at, channel)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), $12)
		ON CONFLICT (url) DO UPDATE SET
			source = EXCLUDED.source,
			source_type = EXCLUDED.source_type,
			title = EXCLUDED.title,
			summary = EXCLUDED.summary,
			content = EXCLUDED.content,
			sentiment = EXCLUDED.sentiment,
			sentiment_value = EXCLUDED.sentiment_value,
			tickers = EXCLUDED.tickers,
			published_at = EXCLUDED.published_at,
			fetched_at = EXCLUDED.fetched_at,
			channel = EXCLUDED.channel
	`, articleID, tickersJSON, details.ChannelName, "youtube",
		title, fmt.Sprintf("https://youtube.com/watch?v=%s", videoID),
		truncateString(summary, 2000), transcript, sentiment, sentimentValue, details.PublishedAt, details.ChannelName)
	if err != nil {
		result.Error = fmt.Sprintf("failed to store video summary: %v", err)
		return result
	}

	result.Summary = summary
	result.Status = "ok"
	result.TranscriptSource = transcriptSource
	return result
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	redisClient, err := redis.NewClient(cfg.RedisAddr)
	if err != nil {
		log.Fatalf("connect redis: %v", err)
	}
	defer redisClient.Close()

	db, err := database.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect postgres: %v", err)
	}
	defer db.Close()

	logURL := os.Getenv("LOGGING_SERVICE_URL")
	if logURL == "" {
		logURL = "http://main-api:8080/api/logs"
	}
	logClient := logging.NewClient("news-feed-service", logURL)

	geminiAPIKey := cfg.GeminiAPIKey
	if geminiAPIKey == "" {
		geminiAPIKey = fetchProviderKey("gemini")
	}
	geminiClient := gemini.NewClient(geminiAPIKey)

	youtubeAPIKey := cfg.YouTubeAPIKey
	if youtubeAPIKey == "" {
		youtubeAPIKey = fetchProviderKey("youtube")
	}
	youtubeClient, err := youtube.NewClient(youtubeAPIKey)
	if err != nil {
		log.Printf("youtube client failed to init: %v", err)
	}

	feedManager := feed.NewManager(redisClient.Redis(), db.Pool, logClient, geminiClient)
	sseManager := sse.NewManager(redisClient.Redis())

	go feedManager.StartScheduler(context.Background(), time.Duration(cfg.ScrapeIntervalMin)*time.Minute)

	go processScrapeNewsJobs(context.Background(), redisClient, feedManager)
	go processTranscribeJobs(context.Background(), redisClient, youtubeClient, geminiClient, sseManager, logClient)

	port := cfg.Port
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
	})
	mux.HandleFunc("/health-simple", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
	})
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
	})
	mux.HandleFunc("/api/scrape", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		ctx := r.Context()
		log.Println("Starting feed scrape...")
		if err := feedManager.ScrapeFeeds(ctx); err != nil {
			log.Printf("Scrape error: %v", err)
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
	})

	mux.HandleFunc("/api/channels/search", func(w http.ResponseWriter, r *http.Request) {
		if youtubeClient == nil {
			http.Error(w, "youtube client not configured", http.StatusInternalServerError)
			return
		}
		query := strings.TrimSpace(r.URL.Query().Get("q"))
		if query == "" {
			http.Error(w, "q parameter required", http.StatusBadRequest)
			return
		}
		channels, err := youtubeClient.SearchChannels(r.Context(), query)
		if err != nil {
			http.Error(w, fmt.Sprintf("search failed: %v", err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(channels)
	})

	mux.HandleFunc("/api/channels", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			rows, err := db.Pool.Query(r.Context(), "SELECT id::text, channel_id, name, COALESCE(youtube_handle, '') FROM youtube_channels WHERE is_active = true")
			if err != nil {
				http.Error(w, "failed to query channels", http.StatusInternalServerError)
				return
			}
			defer rows.Close()
			channels := []map[string]string{}
			for rows.Next() {
				var id, channelID, name, handle string
				if err := rows.Scan(&id, &channelID, &name, &handle); err != nil {
					continue
				}
				channels = append(channels, map[string]string{"id": id, "channel_id": channelID, "name": name, "handle": handle})
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(channels)
			return
		}
		if r.Method == "POST" {
			var req struct {
				ChannelID string `json:"channel_id"`
				Name      string `json:"name"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "invalid request", http.StatusBadRequest)
				return
			}
			req.ChannelID = strings.TrimSpace(req.ChannelID)
			req.Name = strings.TrimSpace(req.Name)
			if req.ChannelID == "" || req.Name == "" {
				http.Error(w, "channel_id and name required", http.StatusBadRequest)
				return
			}
			_, err := db.Pool.Exec(r.Context(), `
				INSERT INTO youtube_channels (channel_id, name) VALUES ($1, $2)
				ON CONFLICT (channel_id) DO UPDATE SET name = $2
			`, req.ChannelID, req.Name)
			if err != nil {
				http.Error(w, "failed to add channel", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})

	mux.HandleFunc("/api/videos/latest", func(w http.ResponseWriter, r *http.Request) {
		if youtubeClient == nil {
			http.Error(w, "youtube client not configured", http.StatusInternalServerError)
			return
		}
		channelID := strings.TrimSpace(r.URL.Query().Get("channel_id"))
		if channelID == "" {
			http.Error(w, "channel_id required", http.StatusBadRequest)
			return
		}
		limit := int64(10)
		if rawLimit := r.URL.Query().Get("limit"); rawLimit != "" {
			parsed, err := strconv.ParseInt(rawLimit, 10, 64)
			if err != nil || parsed <= 0 {
				http.Error(w, "limit must be a positive integer", http.StatusBadRequest)
				return
			}
			if parsed > 50 {
				parsed = 50
			}
			limit = parsed
		}
		videos, err := youtubeClient.GetLatestVideos(r.Context(), channelID, limit)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to fetch videos: %v", err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(videos)
	})

	mux.HandleFunc("/api/videos/summarize", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Videos []videoSummaryRequest `json:"videos"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		if len(req.Videos) == 0 {
			http.Error(w, "videos required", http.StatusBadRequest)
			return
		}

		results := make([]videoSummaryResult, 0, len(req.Videos))
		for _, video := range req.Videos {
			results = append(results, summarizeVideo(r.Context(), db, youtubeClient, geminiClient, video))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string][]videoSummaryResult{"results": results})
	})

	mux.HandleFunc("/api/videos/analyze", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req videoSummaryRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		result := summarizeVideo(r.Context(), db, youtubeClient, geminiClient, req)
		if result.Status != "ok" {
			http.Error(w, result.Error, http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})

	wrappedMux := logging.LoggingMiddleware(mux, logClient)

	srv := &http.Server{
		Addr: ":" + port,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" || r.URL.Path == "/api/health" || r.URL.Path == "/api/scrape" {
				mux.ServeHTTP(w, r)
				return
			}
			wrappedMux.ServeHTTP(w, r)
		}),
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	log.Printf("news-feed-service listening on :%s", port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}

func processScrapeNewsJobs(ctx context.Context, rdb *redis.Client, mgr *feed.Manager) {
	for {
		result, err := rdb.BRPop(ctx, 0, "queue:scrape-news")
		if err != nil {
			continue
		}
		fmt.Println("processing scrape job:", result[1])
		_ = mgr.ScrapeFeeds(ctx)
	}
}

type transcribeJob struct {
	VideoID string `json:"video_id"`
}

func processTranscribeJobs(ctx context.Context, rdb *redis.Client, ytClient *youtube.Client, gemClient *gemini.Client, sseMgr *sse.Manager, logClient *logging.Client) {
	for {
		result, err := rdb.BRPop(ctx, 0, "queue:transcribe")
		if err != nil {
			continue
		}

		var job transcribeJob
		if err := json.Unmarshal([]byte(result[1]), &job); err != nil {
			fmt.Println("parse job error:", err)
			continue
		}

		fmt.Println("processing transcribe job:", job.VideoID)

		if ytClient == nil || gemClient == nil {
			continue
		}

		title := job.VideoID
		transcript, err := ytClient.GetTranscript(ctx, job.VideoID)
		if err != nil || strings.TrimSpace(transcript) == "" {
			details, detailsErr := ytClient.GetVideoDetails(ctx, job.VideoID)
			if detailsErr != nil || details == nil || strings.TrimSpace(details.Description) == "" {
				fmt.Println("transcript error:", err)
				continue
			}
			title = details.Title
			transcript = details.Description
		}

		summary, err := gemClient.SummarizeVideo(ctx, title, transcript)
		if err != nil {
			fmt.Println("summarize error:", err)
			continue
		}

		_ = sseMgr.Publish(ctx, "news:videos", map[string]string{
			"youtube_id": job.VideoID,
			"summary":    summary,
		})
	}
}
