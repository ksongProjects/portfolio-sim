package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/portfolio-sim/backend/logging"
)

type NewsFeedService struct {
	newsFeedURL string
	client      *http.Client
	logger      *logging.Client
}

type VideoSummaryRequest struct {
	VideoID string `json:"video_id"`
	Title   string `json:"title,omitempty"`
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
	requestURL := fmt.Sprintf("%s/api/channels", s.newsFeedURL)
	return s.doRequest(ctx, http.MethodGet, requestURL, nil, "", http.StatusOK)
}

func (s *NewsFeedService) SearchChannels(ctx context.Context, query string) ([]byte, error) {
	requestURL := fmt.Sprintf("%s/api/channels/search?q=%s", s.newsFeedURL, url.QueryEscape(query))
	return s.doRequest(ctx, http.MethodGet, requestURL, nil, "", http.StatusOK)
}

func (s *NewsFeedService) GetLatestVideos(ctx context.Context, channelID string, limit int) ([]byte, error) {
	if limit <= 0 {
		limit = 10
	}
	requestURL := fmt.Sprintf("%s/api/videos/latest?channel_id=%s&limit=%d", s.newsFeedURL, url.QueryEscape(channelID), limit)
	return s.doRequest(ctx, http.MethodGet, requestURL, nil, "", http.StatusOK)
}

func (s *NewsFeedService) GetStoredVideos(ctx context.Context) ([]byte, error) {
	requestURL := fmt.Sprintf("%s/api/videos", s.newsFeedURL)
	return s.doRequest(ctx, http.MethodGet, requestURL, nil, "", http.StatusOK)
}

func (s *NewsFeedService) AnalyzeVideo(ctx context.Context, videoID, title string) error {
	requestURL := fmt.Sprintf("%s/api/videos/analyze", s.newsFeedURL)
	body, _ := json.Marshal(map[string]string{"video_id": videoID, "title": title})
	_, err := s.doRequest(ctx, http.MethodPost, requestURL, bytes.NewBuffer(body), "application/json", http.StatusNoContent)
	return err
}

func (s *NewsFeedService) SummarizeVideos(ctx context.Context, videos []VideoSummaryRequest) ([]byte, error) {
	requestURL := fmt.Sprintf("%s/api/videos/summarize", s.newsFeedURL)
	body, _ := json.Marshal(map[string][]VideoSummaryRequest{"videos": videos})
	return s.doRequest(ctx, http.MethodPost, requestURL, bytes.NewBuffer(body), "application/json", http.StatusOK)
}

func (s *NewsFeedService) AddChannel(ctx context.Context, channelID, name string) error {
	requestURL := fmt.Sprintf("%s/api/channels", s.newsFeedURL)
	body, _ := json.Marshal(map[string]string{"channel_id": channelID, "name": name})
	_, err := s.doRequest(ctx, http.MethodPost, requestURL, bytes.NewBuffer(body), "application/json", http.StatusNoContent)
	return err
}

func (s *NewsFeedService) DeleteChannel(ctx context.Context, id string) error {
	requestURL := fmt.Sprintf("%s/api/channels?id=%s", s.newsFeedURL, url.QueryEscape(id))
	_, err := s.doRequest(ctx, http.MethodDelete, requestURL, nil, "", http.StatusNoContent)
	return err
}
