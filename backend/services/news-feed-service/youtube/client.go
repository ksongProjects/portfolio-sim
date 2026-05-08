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
	captions, err := c.svc.Captions.List([]string{"snippet"}, videoID).Do()
	if err != nil {
		return "", fmt.Errorf("list captions: %w", err)
	}

	if len(captions.Items) == 0 {
		return "", fmt.Errorf("no captions for video %s", videoID)
	}

	caption := captions.Items[0]
	track, err := c.svc.Captions.Download(caption.Id).Do()
}
