package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/portfolio-sim/shared/secrets"
)

type Storage struct {
	pool  *pgxpool.Pool
	codec *secrets.Codec
}

func NewStorage(pool *pgxpool.Pool, codec *secrets.Codec) *Storage {
	return &Storage{pool: pool, codec: codec}
}

func (s *Storage) GetTickerID(ctx context.Context, symbol string) (uuid.UUID, error) {
	var id uuid.UUID
	err := s.pool.QueryRow(ctx, "SELECT id FROM tickers WHERE symbol = $1", symbol).Scan(&id)
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func (s *Storage) UpsertNormalizedPrice(ctx context.Context, tickerID uuid.UUID, price, change, changePct, bid, ask float64, volume int64, sourceID string, timestamp time.Time) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO normalized_prices (ticker_id, price, change, change_pct, bid, ask, volume, source_id, timestamp, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
		ON CONFLICT (ticker_id, timestamp) DO UPDATE SET
			price = EXCLUDED.price,
			change = EXCLUDED.change,
			change_pct = EXCLUDED.change_pct,
			bid = EXCLUDED.bid,
			ask = EXCLUDED.ask,
			volume = EXCLUDED.volume,
			updated_at = NOW()
	`, tickerID, price, change, changePct, bid, ask, volume, sourceID, timestamp)
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

func (s *Storage) GetProviderAPIKey(ctx context.Context, providerID string) (string, error) {
	var encryptedKey string
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(encrypted_key, '')
		FROM provider_configurations
		WHERE provider_id = $1
	`, providerID).Scan(&encryptedKey)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	return s.codec.DecryptString(encryptedKey)
}

func (s *Storage) IsProviderValidated(ctx context.Context, providerID string) (bool, error) {
	var isValidated bool
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(is_validated, false)
		FROM provider_configurations
		WHERE provider_id = $1
	`, providerID).Scan(&isValidated)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return isValidated, nil
}

func (s *Storage) GetQuestradeTokens(ctx context.Context) (*QuestradeTokens, error) {
	var encryptedAccessToken, encryptedRefreshToken, encryptedAPIServer string
	var expiresAt *time.Time
	err := s.pool.QueryRow(ctx, `
		SELECT access_token, refresh_token, api_server, token_expires_at
		FROM provider_configurations
		WHERE provider_id = 'questrade'
	`).Scan(&encryptedAccessToken, &encryptedRefreshToken, &encryptedAPIServer, &expiresAt)
	if err != nil {
		return nil, err
	}
	if encryptedAccessToken == "" || encryptedRefreshToken == "" || encryptedAPIServer == "" {
		return nil, fmt.Errorf("questrade tokens not fully configured: access=%s refresh=%s api=%s",
			encryptedAccessToken, encryptedRefreshToken, encryptedAPIServer)
	}
	accessToken, err := s.codec.DecryptString(encryptedAccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt access token: %w", err)
	}
	refreshToken, err := s.codec.DecryptString(encryptedRefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt refresh token: %w", err)
	}
	apiServer, err := s.codec.DecryptString(encryptedAPIServer)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt api server: %w", err)
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
	encryptedAccessToken, err := s.codec.EncryptString(accessToken)
	if err != nil {
		return err
	}
	encryptedRefreshToken, err := s.codec.EncryptString(refreshToken)
	if err != nil {
		return err
	}
	encryptedAPIServer, err := s.codec.EncryptString(apiServer)
	if err != nil {
		return err
	}
	expiresAt := time.Now().Add(time.Duration(expiresIn) * time.Second)
_, err = s.pool.Exec(ctx, `
		INSERT INTO provider_configurations (id, provider_id, encrypted_key, access_token, refresh_token, api_server, token_expires_at, is_validated, validated_at, validation_error, created_at, updated_at)
		VALUES (gen_random_uuid(), 'questrade', $1, $1, $2, $3, $4, true, NOW(), NULL, NOW(), NOW())
		ON CONFLICT (provider_id) DO UPDATE SET
			encrypted_key = EXCLUDED.encrypted_key,
			access_token = EXCLUDED.access_token,
			refresh_token = EXCLUDED.refresh_token,
			api_server = EXCLUDED.api_server,
			token_expires_at = EXCLUDED.token_expires_at,
			is_validated = true,
			validated_at = NOW(),
			validation_error = NULL,
			updated_at = NOW()
	`, encryptedAccessToken, encryptedRefreshToken, encryptedAPIServer, expiresAt)
	return err
}

func (s *Storage) SearchTickers(ctx context.Context, query string) ([]TickerSearchResult, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT t.symbol, t.company_name, t.exchange,
			COALESCE(np.price, 0) as price,
			COALESCE(np.change, 0) as change,
			COALESCE(np.change_pct, 0) as change_pct
		FROM tickers t
		LEFT JOIN LATERAL (
			SELECT price, change, change_pct
			FROM normalized_prices np
			WHERE np.ticker_id = t.id
			ORDER BY np.timestamp DESC
			LIMIT 1
		) np ON true
		WHERE t.is_active = true
			AND (t.symbol ILIKE $1 OR t.company_name ILIKE $1)
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
		if err := rows.Scan(&r.Symbol, &r.Name, &r.Exchange, &r.Price, &r.Change, &r.ChangePct); err != nil {
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
	Type      string  `json:"type"`
	SymbolID  int     `json:"symbolId"`
	Price     float64 `json:"price"`
	Change    float64 `json:"change"`
	ChangePct float64 `json:"changePct"`
}

func (s *Storage) UpsertTickerFromSearch(ctx context.Context, symbol, name, exchange, assetType string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO tickers (symbol, company_name, exchange, asset_type, is_active)
		VALUES ($1, $2, $3, $4, true)
		ON CONFLICT (symbol) DO UPDATE SET
			company_name = COALESCE($2, tickers.company_name),
			exchange = COALESCE($3, tickers.exchange),
			asset_type = COALESCE($4, tickers.asset_type)
	`, symbol, name, exchange, assetType)
	return err
}

func (s *Storage) EnsureTickerExists(ctx context.Context, symbol string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO tickers (symbol, is_active)
		VALUES ($1, true)
		ON CONFLICT (symbol) DO NOTHING
	`, symbol)
	return err
}

func (s *Storage) GetTickerDetails(ctx context.Context, symbol string) (*TickerDetails, error) {
	var d TickerDetails
	err := s.pool.QueryRow(ctx, `
		SELECT t.symbol, t.company_name, t.exchange,
			COALESCE(np.price, 0), COALESCE(np.change, 0), COALESCE(np.change_pct, 0),
			COALESCE(np.volume, 0), COALESCE(np.market_cap, 0)
		FROM tickers t
		LEFT JOIN LATERAL (
			SELECT price, change, change_pct, volume, market_cap
			FROM normalized_prices np
			WHERE np.ticker_id = t.id
			ORDER BY np.timestamp DESC
			LIMIT 1
		) np ON true
		WHERE t.symbol = $1
	`, symbol).Scan(
		&d.Symbol, &d.Name, &d.Exchange,
		&d.Price, &d.Change, &d.ChangePct,
		&d.Volume, &d.MarketCap,
	)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

type TickerDetails struct {
	Symbol    string  `json:"symbol"`
	Name      string  `json:"name"`
	Exchange  string  `json:"exchange"`
	Price     float64 `json:"price"`
	Change    float64 `json:"change"`
	ChangePct float64 `json:"changePct"`
	DayOpen   float64 `json:"dayOpen"`
	Volume    int64   `json:"volume"`
	MarketCap float64 `json:"marketCap"`
}

func (s *Storage) IsTickerDataStale(ctx context.Context, symbol string, maxAge time.Duration) (bool, error) {
	var updatedAt *time.Time
	err := s.pool.QueryRow(ctx, `
		SELECT MAX(np.timestamp)
		FROM normalized_prices np
		JOIN tickers t ON t.id = np.ticker_id
		WHERE t.symbol = $1
	`, symbol).Scan(&updatedAt)
	if err != nil {
		return true, err
	}
	if updatedAt == nil {
		return true, nil
	}
	return time.Since(*updatedAt) > maxAge, nil
}

func (s *Storage) UpdateTickerPrice(ctx context.Context, symbol string, price, change, changePct float64, volume int64, marketCap float64, sourceID string) error {
	tickerID, err := s.GetTickerID(ctx, symbol)
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO normalized_prices (ticker_id, price, change, change_pct, volume, market_cap, source_id, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
	`, tickerID, price, change, changePct, volume, marketCap, sourceID)
	return err
}

func (s *Storage) UpdateTickerProfile(ctx context.Context, symbol, sector, industry string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE tickers SET sector = $2, industry = $3 WHERE symbol = $1
	`, symbol, sector, industry)
	return err
}

func (s *Storage) GetIntradayBars(ctx context.Context, symbol string, interval string, from time.Time, to time.Time) ([]IntradayBarRecord, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT ib.timestamp, ib.open, ib.high, ib.low, ib.close, ib.volume
		FROM intraday_bars ib
		JOIN tickers t ON t.id = ib.ticker_id
		WHERE t.symbol = $1 AND ib.interval = $2 AND ib.timestamp >= $3 AND ib.timestamp <= $4
		ORDER BY ib.timestamp ASC
	`, symbol, interval, from, to)
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

func (s *Storage) GetIntradayBarsRange(ctx context.Context, symbol string, interval string, from time.Time, to time.Time) (openPrice float64, closePrice float64, change float64, changePct float64, err error) {
	bars, err := s.GetIntradayBars(ctx, symbol, interval, from, to)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	if len(bars) == 0 {
		return 0, 0, 0, 0, nil
	}
	openPrice = bars[0].Open
	closePrice = bars[len(bars)-1].Close
	if openPrice > 0 {
		change = closePrice - openPrice
		changePct = (change / openPrice) * 100
	}
	return openPrice, closePrice, change, changePct, nil
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
