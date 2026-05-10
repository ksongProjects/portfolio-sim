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

func (s *Storage) GetActiveTickers(ctx context.Context) []string {
	rows, err := s.pool.Query(ctx, `
		SELECT DISTINCT t.symbol
		FROM tickers t
		WHERE t.is_active = true
		AND (
			EXISTS (SELECT 1 FROM positions p WHERE p.ticker_id = t.id)
			OR EXISTS (SELECT 1 FROM watchlist_tickers wt WHERE wt.ticker_id = t.id)
		)
	`)
	if err != nil {
		return []string{}
	}
	defer rows.Close()

	var symbols []string
	for rows.Next() {
		var symbol string
		if err := rows.Scan(&symbol); err == nil {
			symbols = append(symbols, symbol)
		}
	}
	return symbols
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
	expAt := time.Now()
	if expiresAt != nil {
		expAt = *expiresAt
	}
	return &QuestradeTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		APIServer:    apiServer,
		ExpiresAt:    expAt,
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

func (s *Storage) SearchTickers(ctx context.Context, query string) ([]TickerSearchResult, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT t.symbol, t.name, t.exchange, t.sector,
			COALESCE(np.price, 0) as price,
			COALESCE(np.change, 0) as change,
			COALESCE(np.change_pct, 0) as change_pct
		FROM tickers t
		LEFT JOIN normalized_prices np ON t.id = np.ticker_id
		WHERE t.is_active = true
			AND (t.symbol ILIKE $1 OR t.name ILIKE $1)
		ORDER BY t.symbol
		LIMIT 20
	`, "%"+query+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []TickerSearchResult
	for rows.Next() {
		var r TickerSearchResult
		if err := rows.Scan(&r.Symbol, &r.Name, &r.Exchange, &r.Sector, &r.Price, &r.Change, &r.ChangePct); err != nil {
			continue
		}
		results = append(results, r)
	}
	return results, nil
}

type TickerSearchResult struct {
	Symbol    string  `json:"symbol"`
	Name      string  `json:"name"`
	Exchange  string  `json:"exchange"`
	Sector    string  `json:"sector"`
	Price     float64 `json:"price"`
	Change    float64 `json:"change"`
	ChangePct float64 `json:"changePct"`
}

func (s *Storage) GetTickerDetails(ctx context.Context, symbol string) (*TickerDetails, error) {
	var d TickerDetails
	err := s.pool.QueryRow(ctx, `
		SELECT t.symbol, t.name, t.exchange, t.sector, t.industry,
			COALESCE(np.price, 0), COALESCE(np.change, 0), COALESCE(np.change_pct, 0),
			COALESCE(np.volume, 0), COALESCE(np.avg_volume, 0),
			COALESCE(np.market_cap, 0), COALESCE(np.pe_ratio, 0), COALESCE(np.eps, 0),
			COALESCE(np.dividend_yield, 0), COALESCE(np.week52_high, 0), COALESCE(np.week52_low, 0)
		FROM tickers t
		LEFT JOIN normalized_prices np ON t.id = np.ticker_id
		WHERE t.symbol = $1
	`, symbol).Scan(
		&d.Symbol, &d.Name, &d.Exchange, &d.Sector, &d.Industry,
		&d.Price, &d.Change, &d.ChangePct,
		&d.Volume, &d.AvgVolume,
		&d.MarketCap, &d.PeRatio, &d.Eps,
		&d.DividendYield, &d.Week52High, &d.Week52Low,
	)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

type TickerDetails struct {
	Symbol        string  `json:"symbol"`
	Name          string  `json:"name"`
	Exchange      string  `json:"exchange"`
	Sector        string  `json:"sector"`
	Industry      string  `json:"industry"`
	Price         float64 `json:"price"`
	Change        float64 `json:"change"`
	ChangePct     float64 `json:"changePct"`
	Volume        int64   `json:"volume"`
	AvgVolume     int64   `json:"avgVolume"`
	MarketCap     float64 `json:"marketCap"`
	PeRatio       float64 `json:"peRatio"`
	Eps           float64 `json:"eps"`
	DividendYield float64 `json:"dividendYield"`
	Week52High    float64 `json:"week52High"`
	Week52Low     float64 `json:"week52Low"`
}

func (s *Storage) GetIntradayBars(ctx context.Context, symbol string, limit int) ([]IntradayBarRecord, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT ib.timestamp, ib.open, ib.high, ib.low, ib.close, ib.volume
		FROM intraday_bars ib
		JOIN tickers t ON t.id = ib.ticker_id
		WHERE t.symbol = $1
		ORDER BY ib.timestamp DESC
		LIMIT $2
	`, symbol, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bars []IntradayBarRecord
	for rows.Next() {
		var b IntradayBarRecord
		if err := rows.Scan(&b.Timestamp, &b.Open, &b.High, &b.Low, &b.Close, &b.Volume); err != nil {
			continue
		}
		bars = append(bars, b)
	}
	return bars, nil
}

type IntradayBarRecord struct {
	Timestamp time.Time `json:"timestamp"`
	Open      float64   `json:"open"`
	High      float64   `json:"high"`
	Low       float64   `json:"low"`
	Close     float64   `json:"close"`
	Volume    int64     `json:"volume"`
}

func (s *Storage) GetFinancialRatios(ctx context.Context, symbol string) ([]FinancialRatioRecord, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT fd.data_type, fd.json_data, fd.timestamp
		FROM fundamental_data fd
		JOIN tickers t ON t.id = fd.ticker_id
		WHERE t.symbol = $1 AND fd.data_type = 'ratios'
		ORDER BY fd.timestamp DESC
		LIMIT 1
	`, symbol)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		var dataType string
		var jsonData []byte
		var ts time.Time
		if err := rows.Scan(&dataType, &jsonData, &ts); err != nil {
			return nil, err
		}
		var ratios []FinancialRatioRecord
		if err := json.Unmarshal(jsonData, &ratios); err == nil {
			return ratios, nil
		}
	}
	return nil, nil
}

type FinancialRatioRecord struct {
	Label       string `json:"label"`
	Value       string `json:"value"`
	Description string `json:"description"`
}
