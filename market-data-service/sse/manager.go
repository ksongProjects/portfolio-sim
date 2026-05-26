package sse

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/portfolio-sim/market-data-service/redis"
)

type Manager struct {
	redis *redis.Client
}

func NewManager(r *redis.Client) *Manager {
	return &Manager{redis: r}
}

func (m *Manager) PublishTick(ticker string, data interface{}) error {
	ctx := context.Background()
	stream := fmt.Sprintf("stream:market:ticks:%s", ticker)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if err := m.redis.XAdd(ctx, stream, "data", string(jsonData)); err != nil {
		return err
	}

	channel := fmt.Sprintf("market:ticks:%s", ticker)
	return m.redis.Publish(ctx, channel, jsonData)
}

func (m *Manager) PublishOptionChain(ticker string, data interface{}) error {
	ctx := context.Background()
	channel := fmt.Sprintf("market:chains:%s", ticker)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return m.redis.Publish(ctx, channel, jsonData)
}

func (m *Manager) PublishIntradayBar(ticker string, interval string, data interface{}) error {
	ctx := context.Background()
	channel := fmt.Sprintf("market:bars:%s:%s", ticker, interval)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return m.redis.Publish(ctx, channel, jsonData)
}

type BackfillRequest struct {
	Ticker       string    `json:"ticker"`
	Source       string    `json:"source"`
	DataType     string    `json:"data_type"`
	Interval     string    `json:"interval,omitempty"`
	StartDate    time.Time `json:"start_date"`
	EndDate      time.Time `json:"end_date"`
	RequestedAt  time.Time `json:"requested_at"`
}

func (m *Manager) EnqueueBackfill(req *BackfillRequest) error {
	ctx := context.Background()
	jsonData, err := json.Marshal(req)
	if err != nil {
		return err
	}
	return m.redis.LPush(ctx, "queue:backfill", jsonData)
}
