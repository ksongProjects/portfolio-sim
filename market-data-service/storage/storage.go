package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	pool *pgxpool.Pool
}

func NewStorage(pool *pgxpool.Pool) *Storage {
	return &Storage{pool: pool}
}

func (s *Storage) GetTickerID(ctx context.Context, symbol string) (uuid.UUID, error) {
	var id uuid.UUID
	err := s.pool.QueryRow(ctx, "SELECT id FROM tickers WHERE symbol = $1", symbol).Scan(&id)
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func (s *Storage) UpsertNormalizedPrice(ctx context.Context, tickerID uuid.UUID, price, bid, ask float64, volume int64, sourceID string, timestamp time.Time) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO normalized_prices (ticker_id, price, bid, ask, volume, source_id, timestamp, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		ON CONFLICT (ticker_id, timestamp) DO UPDATE SET
			price = EXCLUDED.price,
			bid = EXCLUDED.bid,
			ask = EXCLUDED.ask,
			volume = EXCLUDED.volume,
			updated_at = NOW()
	`, tickerID, price, bid, ask, volume, sourceID, timestamp)
	return err
}

func (s *Storage) InsertIntradayBar(ctx context.Context, tickerID uuid.UUID, interval string, open, high, low, close float64, volume int64, timestamp time.Time) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO intraday_bars (ticker_id, interval, open, high, low, close, volume, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (ticker_id, interval, timestamp) DO UPDATE SET
			open = EXCLUDED.open,
			high = EXCLUDED.high,
			low = EXCLUDED.low,
			close = EXCLUDED.close,
			volume = EXCLUDED.volume
	`, tickerID, interval, open, high, low, close, volume, timestamp)
	return err
}

func (s *Storage) InsertFundamentalData(ctx context.Context, tickerID uuid.UUID, sourceID, dataType, period string, data interface{}, timestamp time.Time) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal fundamental data: %w", err)
	}

	_, err = s.pool.Exec(ctx, `
		INSERT INTO fundamental_data (ticker_id, source_id, data_type, period, json_data, timestamp, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (ticker_id, source_id, data_type, period, timestamp) DO UPDATE SET
			json_data = EXCLUDED.json_data,
			updated_at = NOW()
	`, tickerID, sourceID, dataType, period, jsonData, timestamp)
	return err
}

func (s *Storage) InsertOptionChain(ctx context.Context, tickerID uuid.UUID, sourceID string, chains []*OptionChainRecord) error {
	for _, c := range chains {
		_, err := s.pool.Exec(ctx, `
			INSERT INTO option_chains (underlying_ticker_id, source_id, expiration, strike, option_type, bid, ask, delta, gamma, theta, vega, implied_vol, volume, open_interest, timestamp, fetched_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, NOW())
			ON CONFLICT (underlying_ticker_id, expiration, strike, option_type, timestamp) DO UPDATE SET
				bid = EXCLUDED.bid,
				ask = EXCLUDED.ask,
				delta = EXCLUDED.delta,
				gamma = EXCLUDED.gamma,
				theta = EXCLUDED.theta,
				vega = EXCLUDED.vega,
				implied_vol = EXCLUDED.implied_vol,
				volume = EXCLUDED.volume,
				open_interest = EXCLUDED.open_interest
		`, tickerID, sourceID, c.Expiration, c.Strike, c.OptionType, c.Bid, c.Ask, c.Delta, c.Gamma, c.Theta, c.Vega, c.ImpliedVol, c.Volume, c.OpenInterest, c.Timestamp)
		if err != nil {
			return err
		}
	}
	return nil
}

type OptionChainRecord struct {
	Expiration   time.Time
	Strike       float64
	OptionType   string
	Bid          float64
	Ask          float64
	Delta        float64
	Gamma        float64
	Theta        float64
	Vega         float64
	ImpliedVol   float64
	Volume       int64
	OpenInterest int64
	Timestamp    time.Time
}

func (s *Storage) StoreRawPriceTick(ctx context.Context, tickerID uuid.UUID, sourceID string, rawJSON []byte, price, bid, ask float64, volume int64, receivedAt, timestamp time.Time) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO raw_price_ticks (ticker_id, source_id, raw_json, price, bid, ask, volume, received_at, timestamp, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
	`, tickerID, sourceID, rawJSON, price, bid, ask, volume, receivedAt, timestamp)
	return err
}

type QuestradeTokens struct {
	AccessToken  string
	RefreshToken string
	APIServer    string
	ExpiresAt    time.Time
}

func (s *Storage) GetQuestradeTokens(ctx context.Context) (*QuestradeTokens, error) {
	var accessToken, refreshToken, apiServer string
	var expiresAt *time.Time
	err := s.pool.QueryRow(ctx, `
		SELECT access_token, refresh_token, api_server, token_expires_at
		FROM provider_configurations
		WHERE provider_id = 'questrade'
	`).Scan(&accessToken, &refreshToken, &apiServer, &expiresAt)
	if err != nil {
		return nil, err
	}
	return &QuestradeTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		APIServer:    apiServer,
		ExpiresAt:    time.Now(),
	}, nil
}

func (s *Storage) UpdateQuestradeTokens(ctx context.Context, accessToken, refreshToken, apiServer string, expiresIn int) error {
	expiresAt := time.Now().Add(time.Duration(expiresIn) * time.Second)
	_, err := s.pool.Exec(ctx, `
		UPDATE provider_configurations
		SET access_token = $1, refresh_token = $2, api_server = $3, token_expires_at = $4, updated_at = NOW()
		WHERE provider_id = 'questrade'
	`, accessToken, refreshToken, apiServer, expiresAt)
	return err
}
