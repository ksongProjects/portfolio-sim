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

func (c *Client) SummarizeVideo(ctx context.Context, title, transcript string) (string, error) {
	prompt := fmt.Sprintf(`Summarize this financial YouTube video for a portfolio dashboard.

Title: %s
Transcript or description:
%s

Return a concise plain-text summary in 3-5 bullets. Include mentioned companies, tickers, market direction, and key risks when present.`, title, transcript)
	return c.Summarize(ctx, prompt)
}

func (c *Client) AnalyzeArticle(ctx context.Context, title, summary string) (string, string, []string, error) {
	if c.apiKey == "" {
		return "neutral", "", nil, fmt.Errorf("api key required")
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(c.apiKey))
	if err != nil {
		return "neutral", "", nil, fmt.Errorf("create client: %w", err)
	}
	defer client.Close()

	prompt := fmt.Sprintf(`Analyze this financial news article.

Title: %s
Summary: %s

Respond with JSON only in this exact format:
{"sentiment": "bullish|bearish|neutral", "sentiment_value": "0.0 to 1.0", "related_tickers": ["TICKER1", "TICKER2"]}

Rules:
- sentiment: "bullish" for positive outlook, "bearish" for negative, "neutral" for mixed/neutral
- sentiment_value: a number from 0.0 to 1.0 where 0.0 is very bearish, 0.5 is neutral, 1.0 is very bullish
- related_tickers: stock ticker symbols mentioned or strongly implied (max 3, use empty array if none found)
- Only include tickers that are explicit in the article`, title, summary)

	model := client.GenerativeModel("gemini-2.5-flash")
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "neutral", "", nil, fmt.Errorf("generate content: %w", err)
	}

	if len(resp.Candidates) == 0 {
		return "neutral", "", nil, fmt.Errorf("no candidates")
	}

	raw := fmt.Sprintf("%v", resp.Candidates[0].Content)
	raw = strings.TrimSpace(raw)
	raw = strings.Trim(raw, "```")
	raw = strings.TrimPrefix(raw, "json")
	raw = strings.TrimSpace(raw)

	sentiment := "neutral"
	sentimentValue := "0.5"
	tickers := []string{}

	raw = strings.ToLower(raw)
	if strings.Contains(raw, "bearish") {
		sentiment = "bearish"
	} else if strings.Contains(raw, "bullish") {
		sentiment = "bullish"
	}

	if strings.Contains(raw, "0.3") {
		sentimentValue = "0.3"
	} else if strings.Contains(raw, "0.7") {
		sentimentValue = "0.7"
	} else if strings.Contains(raw, "0.1") {
		sentimentValue = "0.1"
	} else if strings.Contains(raw, "0.9") {
		sentimentValue = "0.9"
	} else if strings.Contains(raw, "0.0") {
		sentimentValue = "0.0"
	} else if strings.Contains(raw, "1.0") {
		sentimentValue = "1.0"
	}

	if idx := strings.Index(raw, "related_tickers"); idx != -1 {
		start := strings.Index(raw[idx:], "[")
		end := strings.Index(raw[idx:], "]")
		if start != -1 && end != -1 && start < end {
			tickersStr := raw[idx+start : idx+end+1]
			tickersStr = strings.ReplaceAll(tickersStr, "\"", "")
			tickersStr = strings.ReplaceAll(tickersStr, " ", "")
			tickersStr = strings.TrimPrefix(tickersStr, "[")
			tickersStr = strings.TrimSuffix(tickersStr, "]")
			if tickersStr != "" {
				parts := strings.Split(tickersStr, ",")
				for _, p := range parts {
					p = strings.TrimSpace(p)
					if p != "" {
						tickers = append(tickers, strings.ToUpper(p))
					}
				}
			}
		}
	}

	return sentiment, sentimentValue, tickers, nil
}
