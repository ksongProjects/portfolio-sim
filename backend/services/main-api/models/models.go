package models

import (
	"time"
)

type User struct {
	ID        string    `json:"id"`
	ClerkID   string    `json:"clerk_id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Portfolio struct {
	ID          string     `json:"id"`
	ClerkID     string     `json:"clerk_id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Positions   []Position `json:"positions,omitempty"`
}

type Position struct {
	ID          string    `json:"id"`
	PortfolioID string    `json:"portfolio_id"`
	TickerID    string    `json:"ticker_id"`
	Shares      float64   `json:"shares"`
	AvgCost     float64   `json:"avg_cost"`
	CreatedAt   time.Time `json:"created_at"`
	Ticker      *Ticker   `json:"ticker,omitempty"`
}

type Watchlist struct {
	ID        string    `json:"id"`
	ClerkID   string    `json:"clerk_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Tickers   []Ticker  `json:"tickers,omitempty"`
}

type Ticker struct {
	ID          string    `json:"id"`
	Symbol      string    `json:"symbol"`
	Name        string    `json:"name"`
	Exchange    string    `json:"exchange"`
	LastPrice   float64   `json:"last_price"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type WatchlistTicker struct {
	ID          string    `json:"id"`
	WatchlistID string    `json:"watchlist_id"`
	TickerID    string    `json:"ticker_id"`
	AddedAt     time.Time `json:"added_at"`
}

type Job struct {
	ID        string    `json:"id"`
	ClerkID   string    `json:"clerk_id"`
	Type      string    `json:"type"`
	Status    string    `json:"status"`
	Payload   string    `json:"payload"`
	Result    string    `json:"result"`
	Error     string    `json:"error"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}