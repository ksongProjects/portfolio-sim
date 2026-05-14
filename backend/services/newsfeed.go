package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/portfolio-sim/backend/logging"
)

type NewsFeedService struct {
	newsFeedURL string
	client      *http.Client
	logger      *logging.Client
}

func NewNewsFeedService(newsFeedURL string, logClient *logging.Client) *NewsFeedService {
	if newsFeedURL == "" {
		newsFeedURL = "http://news-feed-service:8080"
	}
	return &NewsFeedService{
		newsFeedURL: newsFeedURL,
		client:      &http.Client{Timeout: 10 * time.Second},
		logger:      logClient,
	}
}

func (s *NewsFeedService) doRequest(ctx context.Context, method, url string, body io.Reader, contentType string, expectedStatus int) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != expectedStatus {
		return nil, fmt.Errorf("%s %s failed: %d - %s", method, url, resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func (s *NewsFeedService) GetChannels(ctx context.Context) ([]byte, error) {
	url := fmt.Sprintf("%s/api/channels", s.newsFeedURL)
	return s.doRequest(ctx, http.MethodGet, url, nil, "", http.StatusOK)
}

func (s *NewsFeedService) GetLatestVideos(ctx context.Context, channelID string) ([]byte, error) {
	url := fmt.Sprintf("%s/api/videos/latest?channel_id=%s", s.newsFeedURL, channelID)
	return s.doRequest(ctx, http.MethodGet, url, nil, "", http.StatusOK)
}

func (s *NewsFeedService) GetStoredVideos(ctx context.Context) ([]byte, error) {
	url := fmt.Sprintf("%s/api/videos", s.newsFeedURL)
	return s.doRequest(ctx, http.MethodGet, url, nil, "", http.StatusOK)
}

func (s *NewsFeedService) AnalyzeVideo(ctx context.Context, videoID, title string) error {
	url := fmt.Sprintf("%s/api/videos/analyze", s.newsFeedURL)
	body, _ := json.Marshal(map[string]string{"video_id": videoID, "title": title})
	_, err := s.doRequest(ctx, http.MethodPost, url, bytes.NewBuffer(body), "application/json", http.StatusNoContent)
	return err
}
