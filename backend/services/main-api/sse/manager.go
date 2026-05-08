package sse

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/portfolio-sim/backend/internal/redis"
)

type Event struct {
	Channel string
	Data    []byte
}

type Subscriber struct {
	ID       string
	Events   chan Event
	ctx      context.Context
	cancel   context.CancelFunc
}

func (s *Subscriber) IDs() string { return s.ID }

type Manager struct {
	subs       map[string]*Subscriber
	mu         sync.RWMutex
	redis      *redis.Client
	ps         *redis.PubSub
	handlers   map[string]func(Event)
	register   chan *Subscriber
	unregister chan *Subscriber
	subscribe  chan struct {
		subscriber *Subscriber
		channels   []string
	}
}

func NewManager(r *redis.Client) *Manager {
	m := &Manager{
		subs:       make(map[string]*Subscriber),
		redis:      r,
		handlers:   make(map[string]func(Event)),
		register:   make(chan *Subscriber, 100),
		unregister: make(chan *Subscriber, 100),
		subscribe:  make(chan struct{ subscriber *Subscriber; channels []string }, 100),
	}
	go m.run()
	return m
}

func (m *Manager) run() {
	for {
		select {
		case sub := <-m.register:
			m.mu.Lock()
			sub.ctx, sub.cancel = context.WithCancel(context.Background())
			m.subs[sub.ID] = sub
			m.mu.Unlock()
		case sub := <-m.unregister:
			m.mu.Lock()
			if s, ok := m.subs[sub.ID]; ok && s == sub {
				sub.cancel()
				delete(m.subs, sub.ID)
			}
			m.mu.Unlock()
		case req := <-m.subscribe:
			go m.handleSubscribe(req.subscriber, req.channels)
		}
	}
}

func (m *Manager) handleSubscribe(sub *Subscriber, channels []string) {
	if m.ps == nil {
		m.ps = m.redis.Subscribe(context.Background(), channels...)
	}
	ch := m.ps.Channel()
	for {
		select {
		case <-sub.ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			data, _ := json.Marshal(map[string]interface{}{
				"channel": msg.Channel,
				"data":    msg.Payload,
			})
			select {
			case sub.Events <- Event{Channel: msg.Channel, Data: data}:
			case <-time.After(time.Second):
			case <-sub.ctx.Done():
				return
			}
		}
	}
}

func (m *Manager) Register(id string) *Subscriber {
	sub := &Subscriber{
		ID:     id,
		Events: make(chan Event, 256),
	}
	m.register <- sub
	return sub
}

func (m *Manager) Unregister(sub *Subscriber) {
	m.unregister <- sub
}

func (m *Manager) Subscribe(sub *Subscriber, channels ...string) {
	m.subscribe <- struct {
		subscriber *Subscriber
		channels   []string
	}{subscriber: sub, channels: channels}
}

func (m *Manager) HandleChannel(channel string, handler func(Event)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlers[channel] = handler
}

func (m *Manager) SubscribeToMarketTicks(sub *Subscriber, ticker string) {
	channel := fmt.Sprintf("market:ticks:%s", ticker)
	m.Subscribe(sub, channel)
}

func (m *Manager) SubscribeToNewsArticles(sub *Subscriber) {
	m.Subscribe(sub, "news:articles")
}

func (m *Manager) SubscribeToNewsVideos(sub *Subscriber) {
	m.Subscribe(sub, "news:videos")
}