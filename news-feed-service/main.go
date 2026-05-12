package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/portfolio-sim/news-feed-service/config"
	"github.com/portfolio-sim/news-feed-service/database"
	"github.com/portfolio-sim/news-feed-service/feed"
	"github.com/portfolio-sim/news-feed-service/gemini"
	"github.com/portfolio-sim/news-feed-service/logging"
	"github.com/portfolio-sim/news-feed-service/redis"
	"github.com/portfolio-sim/news-feed-service/sse"
	"github.com/portfolio-sim/news-feed-service/youtube"
)

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

	geminiClient := gemini.NewClient(cfg.GeminiAPIKey)
	youtubeClient, _ := youtube.NewClient(cfg.YouTubeAPIKey)

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

	mux.HandleFunc("/api/channels", func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Pool.Query(r.Context(), "SELECT id::text, channel_id, name, youtube_handle FROM youtube_channels WHERE is_active = true")
		if err != nil {
			http.Error(w, "failed to query channels", http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		var channels []map[string]string
		for rows.Next() {
			var id, channelID, name, handle string
			if err := rows.Scan(&id, &channelID, &name, &handle); err != nil {
				continue
			}
			channels = append(channels, map[string]string{"id": id, "channel_id": channelID, "name": name, "handle": handle})
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(channels)
	})

	mux.HandleFunc("/api/videos/latest", func(w http.ResponseWriter, r *http.Request) {
		if ytClient == nil {
			http.Error(w, "youtube client not configured", http.StatusInternalServerError)
			return
		}
		channelID := r.URL.Query().Get("channel_id")
		if channelID == "" {
			http.Error(w, "channel_id required", http.StatusBadRequest)
			return
		}
		videos, err := ytClient.GetLatestVideos(r.Context(), channelID, 20)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to fetch videos: %v", err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(videos)
	})

	mux.HandleFunc("/api/videos/analyze", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			VideoID string `json:"video_id"`
			Title   string `json:"title"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		details, err := ytClient.GetVideoDetails(r.Context(), req.VideoID)
		if err != nil {
			http.Error(w, "failed to get video details", http.StatusInternalServerError)
			return
		}

		transcript, _ := ytClient.GetVideoCaption(r.Context(), req.VideoID)
		if transcript == "" {
			transcript = details.Description
		}

		result, err := geminiClient.AnalyzeArticle(r.Context(), req.Title, transcript)
		if err != nil {
			http.Error(w, "failed to analyze video", http.StatusInternalServerError)
			return
		}

		tickersJSON, _ := json.Marshal(result.RelatedTickers)
		_, err = db.Pool.Exec(r.Context(), `
			INSERT INTO news_videos (youtube_id, title, channel, transcript_text, summary_text, sentiment, tickers, published_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (youtube_id) DO UPDATE SET
				transcript_text = EXCLUDED.transcript_text,
				summary_text = $5,
				sentiment = EXCLUDED.sentiment,
				tickers = EXCLUDED.tickers
		`, req.VideoID, req.Title, details.ChannelName, transcript, "", result.Sentiment, tickersJSON, details.PublishedAt)
		if err != nil {
			log.Printf("failed to store video: %v", err)
			http.Error(w, "failed to store video", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})

	mux.HandleFunc("/api/videos", func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Pool.Query(r.Context(), `
			SELECT id::text, youtube_id, title, channel, summary_text, sentiment, tickers, published_at
			FROM news_videos
			ORDER BY published_at DESC
			LIMIT 50
		`)
		if err != nil {
			http.Error(w, "failed to query videos", http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		var videos []map[string]interface{}
		for rows.Next() {
			var id, ytID, title, channel, summary, sentiment string
			var tickers []byte
			var publishedAt *time.Time
			if err := rows.Scan(&id, &ytID, &title, &channel, &summary, &sentiment, &tickers, &publishedAt); err != nil {
				continue
			}
			videos = append(videos, map[string]interface{}{
				"id": id, "youtube_id": ytID, "title": title, "channel": channel,
				"summary": summary, "sentiment": sentiment, "tickers": string(tickers), "published_at": publishedAt,
			})
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(videos)
	})

	wrappedMux := logging.LoggingMiddleware(mux, logClient)

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		transcript, err := ytClient.GetTranscript(ctx, job.VideoID)
		if err != nil {
			fmt.Println("transcript error:", err)
			continue
		}

		summary, err := gemClient.Summarize(ctx, transcript)
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
