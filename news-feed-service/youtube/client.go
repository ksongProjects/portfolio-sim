package youtube

import (
	"context"
	"fmt"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type Client struct {
	apiKey string
	svc    *youtube.Service
}

func NewClient(apiKey string) (*Client, error) {
	svc, err := youtube.NewService(context.Background(), option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("create youtube service: %w", err)
	}
	return &Client{apiKey: apiKey, svc: svc}, nil
}

func (c *Client) GetTranscript(ctx context.Context, videoID string) (string, error) {
	return "", fmt.Errorf("transcript download not implemented - requires OAuth")
}
