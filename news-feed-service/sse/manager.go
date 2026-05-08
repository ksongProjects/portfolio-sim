package sse

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
)

type Manager struct {
	redis *redis.Client
	subs  map[string]chan []byte
}

func NewManager(redisClient *redis.Client) *Manager {
	return &Manager{
		redis: redisClient,
		subs:  make(map[string]chan []byte),
	}
}

func (m *Manager) Subscribe(channel string) (<-chan []byte, func()) {
	ch := make(chan []byte, 100)
	m.subs[channel] = ch

	unsubscribe := func() {
		delete(m.subs, channel)
		close(ch)
	}

	return ch, unsubscribe
}

func (m *Manager) Publish(ctx context.Context, channel string, data interface{}) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return m.redis.Publish(ctx, channel, payload).Err()
}
