package gemini

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type Client struct {
	apiKey string
}

func NewClient(apiKey string) *Client {
	return &Client{apiKey: apiKey}
}

func (c *Client) Summarize(ctx context.Context, text string) (string, error) {
	if c.apiKey == "" {
		return "", fmt.Errorf("api key required")
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(c.apiKey))
	if err != nil {
		return "", fmt.Errorf("create client: %w", err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-2.5-flash")
	resp, err := model.GenerateContent(ctx, genai.Text(text))
	if err != nil {
		return "", fmt.Errorf("generate content: %w", err)
	}

	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("no candidates")
	}

	return fmt.Sprintf("%v", resp.Candidates[0].Content), nil
}

type SentimentResult struct {
	Sentiment    string   `json:"sentiment"`
	RelatedTickers []string `json:"related_tickers"`
}

func (c *Client) AnalyzeArticle(ctx context.Context, title, summary string) (*SentimentResult, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("api key required")
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(c.apiKey))
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}
	defer client.Close()

	prompt := fmt.Sprintf(`Analyze this financial news article.

Title: %s
Summary: %s

Respond with JSON only in this exact format:
{"sentiment": "bullish|bearish|neutral", "related_tickers": ["TICKER1", "TICKER2"]}

Rules:
- sentiment: "bullish" for positive outlook, "bearish" for negative, "neutral" for mixed/neutral
- related_tickers: stock ticker symbols mentioned or strongly implied (max 3, use "N/A" if none found)
- Only include tickers that are explicit in the article`, title, summary)

	model := client.GenerativeModel("gemini-2.5-flash")
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("generate content: %w", err)
	}

	if len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("no candidates")
	}

	raw := fmt.Sprintf("%v", resp.Candidates[0].Content)
	raw = strings.TrimSpace(raw)
	raw = strings.Trim(raw, "```")
	raw = strings.TrimPrefix(raw, "json")
	raw = strings.TrimSpace(raw)

	result := &SentimentResult{
		Sentiment:    "neutral",
		RelatedTickers: []string{},
	}

	if strings.Contains(strings.ToLower(raw), "bearish") {
		result.Sentiment = "bearish"
	} else if strings.Contains(strings.ToLower(raw), "bullish") {
		result.Sentiment = "bullish"
	}

	return result, nil
}
