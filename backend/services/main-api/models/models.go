package models

import "time"

type Portfolio struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Name        string    `json:"name"`
	InitialCash float64   `json:"initial_cash"`
	CreatedAt   time.Time `json:"created_at"`
}

type Position struct {
	ID          string    `json:"id"`
	PortfolioID string    `json:"portfolio_id"`
	TickerID    string    `json:"ticker_id"`
	Quantity    float64   `json:"quantity"`
	AvgCost     float64   `json:"avg_cost"`
	OpenedAt    time.Time `json:"opened_at"`
}

type Watchlist struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type Ticker struct {
	ID        string    `json:"id"`
	Symbol    string    `json:"symbol"`
	Name      string    `json:"name"`
	Exchange  string    `json:"exchange"`
	CreatedAt time.Time `json:"created_at"`
}

type Job struct {
	ID          string    `json:"id"`
	JobType     string    `json:"job_type"`
	Status      string    `json:"status"`
	Payload     string    `json:"payload"`
	Result      string    `json:"result"`
	Error       string    `json:"error"`
	StartedAt   time.Time `json:"started_at"`
	CompletedAt time.Time `json:"completed_at"`
	CreatedAt   time.Time `json:"created_at"`
}
