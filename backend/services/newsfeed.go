package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/portfolio-sim/backend/logging"
)

type NewsFeedService struct {
	newsFeedURL string
	client      *http.Client
	logger      *logging.Client
}

func NewNewsFeedService(newsFeedURL string, logClient *logging.Client) *NewsFeedService {
	if newsFeedURL == "" {
		newsFeedURL = "http://localhost:8082"
	}
	return &NewsFeedService{
		newsFeedURL: newsFeedURL,
		client:      &http.Client{Timeout: 10 * 1e9},
		logger:      logClient,
	}
}

func (s *NewsFeedService) GetChannels(ctx context.Context) ([]byte, error) {
	url := fmt.Sprintf("%s/api/channels", s.newsFeedURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func (s *NewsFeedService) GetLatestVideos(ctx context.Context, channelID string) ([]byte, error) {
	url := fmt.Sprintf("%s/api/videos/latest?channel_id=%s", s.newsFeedURL, channelID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func (s *NewsFeedService) GetStoredVideos(ctx context.Context) ([]byte, error) {
	url := fmt.Sprintf("%s/api/videos", s.newsFeedURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func (s *NewsFeedService) AnalyzeVideo(ctx context.Context, videoID, title string) error {
	url := fmt.Sprintf("%s/api/videos/analyze", s.newsFeedURL)
	body, _ := json.Marshal(map[string]string{"video_id": videoID, "title": title})
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("analyze failed: %d", resp.StatusCode)
	}
	return nil
}