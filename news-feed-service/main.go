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
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := database.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer db.Close()

	redisClient, err := redis.NewClient(cfg.RedisAddr)
	if err != nil {
		log.Fatalf("connect redis: %v", err)
	}
	defer redisClient.Close()

	logClient := logging.NewClient(db.Pool)
	logWriter := logging.NewLogWriter(logClient)

	geminiClient := gemini.NewClient(cfg.GeminiAPIKey)
	youtubeClient, _ := youtube.NewClient(cfg.YouTubeAPIKey)

	feedManager := feed.NewManager(redisClient.Redis())
	sseManager := sse.NewManager(redisClient.Redis())

	go feedManager.StartScheduler(context.Background(), time.Duration(cfg.ScrapeIntervalMin)*time.Minute)

	go processScrapeNewsJobs(context.Background(), redisClient.Redis(), feedManager, logWriter)
	go processTranscribeJobs(context.Background(), redisClient.Redis(), youtubeClient, geminiClient, sseManager, logWriter)

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
	})

	port := cfg.Port
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{Addr: ":" + port, Handler: nil}
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

func processScrapeNewsJobs(ctx context.Context, rdb *redis.Client, mgr *feed.Manager, logWriter *logging.LogWriter) {
	for {
		result, err := rdb.BRPop(ctx, 0, "queue:scrape-news").Result()
		if err != nil {
			continue
		}
		fmt.Fprintf(logWriter, "processing scrape job: %s\n", result[1])
		_ = mgr.ScrapeFeeds(ctx)
	}
}

type transcribeJob struct {
	VideoID string `json:"video_id"`
}

func processTranscribeJobs(ctx context.Context, rdb *redis.Client, ytClient *youtube.Client, gemClient *gemini.Client, sseMgr *sse.Manager, logWriter *logging.LogWriter) {
	for {
		result, err := rdb.BRPop(ctx, 0, "queue:transcribe").Result()
		if err != nil {
			continue
		}

		var job transcribeJob
		if err := json.Unmarshal([]byte(result[1]), &job); err != nil {
			fmt.Fprintf(logWriter, "parse job error: %v\n", err)
			continue
		}

		fmt.Fprintf(logWriter, "processing transcribe job: %s\n", job.VideoID)

		if ytClient == nil || gemClient == nil {
			continue
		}

		transcript, err := ytClient.GetTranscript(ctx, job.VideoID)
		if err != nil {
			fmt.Fprintf(logWriter, "transcript error: %v\n", err)
			continue
		}

		summary, err := gemClient.Summarize(ctx, transcript)
		if err != nil {
			fmt.Fprintf(logWriter, "summarize error: %v\n", err)
			continue
		}

		_ = sseMgr.Publish(ctx, "news:videos", map[string]string{
			"youtube_id": job.VideoID,
			"summary":    summary,
		})
	}
}
