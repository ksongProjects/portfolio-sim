package youtube

import (
	"context"
	"fmt"
	"time"

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

type Video struct {
	ID           string
	Title        string
	Description  string
	ChannelID    string
	ChannelName  string
	PublishedAt  time.Time
	ThumbURL     string
}

func (c *Client) GetLatestVideos(ctx context.Context, channelID string, maxResults int64) ([]Video, error) {
	if maxResults <= 0 {
		maxResults = 10
	}

	call := c.svc.Search.List([]string{"snippet"}).ChannelId(channelID).Order("date").Type("video").MaxResults(maxResults)
	resp, err := call.Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("search videos: %w", err)
	}

	videos := make([]Video, 0, len(resp.Items))
	for _, item := range resp.Items {
		if item.Id.VideoId == "" {
			continue
		}
		publishedAt, _ := time.Parse(time.RFC3339, item.Snippet.PublishedAt)
		videos = append(videos, Video{
			ID:          item.Id.VideoId,
			Title:       item.Snippet.Title,
			Description: item.Snippet.Description,
			ChannelID:   item.Snippet.ChannelId,
			ChannelName: item.Snippet.ChannelTitle,
			PublishedAt: publishedAt,
			ThumbURL:    item.Snippet.Thumbnails.Default.Url,
		})
	}
	return videos, nil
}

func (c *Client) GetVideoCaption(ctx context.Context, videoID string) (string, error) {
	call := c.svc.Captions.List([]string{"snippet"}, videoID)
	resp, err := call.Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("list captions: %w", err)
	}

	if len(resp.Items) == 0 {
		return "", nil
	}

	for _, cap := range resp.Items {
		if cap.Snippet.TrackKind == "standard" || cap.Snippet.TrackKind == "ASR" {
			return "", fmt.Errorf("caption download requires OAuth - not implemented")
		}
	}
	return "", nil
}

func (c *Client) GetVideoDetails(ctx context.Context, videoID string) (*Video, error) {
	call := c.svc.Videos.List([]string{"snippet"}).Id(videoID)
	resp, err := call.Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get video: %w", err)
	}
	if len(resp.Items) == 0 {
		return nil, nil
	}
	item := resp.Items[0]
	publishedAt, _ := time.Parse(time.RFC3339, item.Snippet.PublishedAt)
	return &Video{
		ID:          item.Id,
		Title:       item.Snippet.Title,
		Description: item.Snippet.Description,
		ChannelID:   item.Snippet.ChannelId,
		ChannelName: item.Snippet.ChannelTitle,
		PublishedAt: publishedAt,
		ThumbURL:    item.Snippet.Thumbnails.Default.Url,
}, nil
}

func (c *Client) GetTranscript(ctx context.Context, videoID string) (string, error) {
	return "", fmt.Errorf("transcript download requires OAuth - not implemented via API key")
}

type Channel struct {
	ID   string
	Name string
	Handle string
}

func (c *Client) SearchChannels(ctx context.Context, query string) ([]Channel, error) {
	call := c.svc.Search.List([]string{"snippet"}).Q(query).Type("channel").MaxResults(10)
	resp, err := call.Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("search channels: %w", err)
	}
	channels := make([]Channel, 0, len(resp.Items))
	for _, item := range resp.Items {
		if item.Id.ChannelId == "" {
			continue
		}
		channels = append(channels, Channel{
			ID:     item.Id.ChannelId,
			Name:   item.Snippet.Title,
			Handle: item.Snippet.ChannelTitle,
		})
	}
	return channels, nil
}