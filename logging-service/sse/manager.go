package sse

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/portfolio-sim/backend/services/logging-service/redis"
)

type Client struct {
	ID       string
	Channels []string
	Send     chan []byte
}

type Manager struct {
	redis   *redis.Client
	clients map[string]*Client
	mu      sync.RWMutex
}

func NewManager(r *redis.Client) *Manager {
	m := &Manager{
		redis:   r,
		clients: make(map[string]*Client),
	}
	go m.subscribe()
	return m
}

func (m *Manager) Register(c *Client) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.clients[c.ID] = c
}

func (m *Manager) Unregister(c *Client) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.clients, c.ID)
}

func (m *Manager) subscribe() {
	ctx := context.Background()
	pubsub := m.redis.Subscribe(ctx, "logs:*")
	defer pubsub.Close()
	for msg := range pubsub.Channel() {
		m.mu.RLock()
		for _, c := range m.clients {
			for _, ch := range c.Channels {
				if strings.HasPrefix(msg.Channel, ch) || ch == "*" {
					select {
					case c.Send <- []byte(fmt.Sprintf("event: %s\ndata: %s\n\n", msg.Channel, msg.Payload)):
					default:
					}
				}
			}
		}
		m.mu.RUnlock()
	}
}

func ParseChannels(param string) []string {
	if param == "" {
		return nil
	}
	return strings.Split(param, ",")
}

func timeNow() time.Time {
	return time.Now().UTC()
}