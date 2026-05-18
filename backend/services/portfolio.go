package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

type Ticker struct {
	ID          string
	Symbol      string
	CompanyName string
	Exchange    string
	AssetType   string
}

type PriceData struct {
	TickerID  string
	Price     float64
	Change    float64
	ChangePct float64
	Bid       float64
	Ask       float64
	Volume    int64
	SourceID  string
	Timestamp time.Time
}

type Position struct {
	ID           string
	PortfolioID  string
	TickerID     string
	Symbol       string
	CompanyName  string
	Sector       string
	Quantity     float64
	AvgCost      float64
	CurrentPrice float64
	CurrentValue float64
	DayChange    float64
	DayChangePct float64
	TotalGain    float64
	TotalGainPct float64
	OpenedAt     time.Time
}

type PortfolioSummary struct {
	TotalValue    float64
	DayChange     float64
	DayChangePct  float64
	TotalInvested float64
	TotalGain     float64
	TotalGainPct  float64
	CashBalance   float64
}

type MarketIndex struct {
	Symbol    string
	Name      string
	Price     float64
	Change    float64
	ChangePct float64
}

type MarketIndexSetting struct {
	Symbol string `json:"symbol"`
	Name   string `json:"name"`
}

type Trade struct {
	ID        string
	Type      string
	Symbol    string
	Shares    float64
	Price     float64
	Total     float64
	Timestamp time.Time
}

type NewsArticle struct {
	ID             string
	Title          string
	Source         string
	SourceType     string
	URL            string
	Summary        string
	Content        string
	Sentiment      string
	SentimentValue string
	PublishedAt    time.Time
	TickerSymbols  []string
	Channel        string
}

type Strategy struct {
	ID      string
	Name    string
	Status  string
	Returns float64
	Sharpe  float64
	MaxDD   float64
	Trades  int
	WinRate float64
}

type Signal struct {
	ID        string
	Message   string
	Service   string
	Level     string
	Timestamp time.Time
}

type PortfolioService struct {
	logger *slog.Logger
}

func NewPortfolioService() *PortfolioService {
	return &PortfolioService{
		logger: slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})),
	}
}

func safePercent(numerator, denominator float64) float64 {
	if denominator == 0 {
		return 0
	}

	return (numerator / denominator) * 100
}

func defaultMarketIndexSettings() []MarketIndexSetting {
	return []MarketIndexSetting{
		{Symbol: "SPY", Name: "S&P 500"},
		{Symbol: "QQQ", Name: "Nasdaq-100"},
		{Symbol: "DIA", Name: "Dow Jones"},
		{Symbol: "IWM", Name: "Russell 2000"},
		{Symbol: "VIX", Name: "Volatility Index"},
		{Symbol: "DXY", Name: "US Dollar Index"},
	}
}

func normalizeMarketIndexSettings(settings []MarketIndexSetting) []MarketIndexSetting {
	normalized := make([]MarketIndexSetting, 0, len(settings))
	seen := make(map[string]bool, len(settings))

	for _, setting := range settings {
		symbol := strings.ToUpper(strings.TrimSpace(setting.Symbol))
		if symbol == "" || seen[symbol] {
			continue
		}

		name := strings.TrimSpace(setting.Name)
		if name == "" {
			name = symbol
		}

		normalized = append(normalized, MarketIndexSetting{
			Symbol: symbol,
			Name:   name,
		})
		seen[symbol] = true
	}

	return normalized
}

func (s *PortfolioService) GetTickers(ctx context.Context, db interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}, symbols []string) ([]Ticker, error) {
	query := `SELECT id, symbol, company_name, exchange, asset_type FROM tickers WHERE is_active = true`
	args := []interface{}{}
	if len(symbols) > 0 {
		query += " AND symbol = ANY($1)"
		args = append(args, symbols)
	}
	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tickers []Ticker
	for rows.Next() {
		var t Ticker
		if err := rows.Scan(&t.ID, &t.Symbol, &t.CompanyName, &t.Exchange, &t.AssetType); err != nil {
			continue
		}
		tickers = append(tickers, t)
	}
	return tickers, nil
}

func (s *PortfolioService) GetLatestPrices(ctx context.Context, db interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}, tickerIDs []string) (map[string]PriceData, error) {
	if len(tickerIDs) == 0 {
		return make(map[string]PriceData), nil
	}
	query := `
		SELECT np.ticker_id, np.price, np.change, np.change_pct, np.bid, np.ask, np.volume, np.source_id, np.timestamp
		FROM normalized_prices np
		INNER JOIN (
			SELECT ticker_id, MAX(timestamp) as max_ts
			FROM normalized_prices
			WHERE ticker_id = ANY($1)
			GROUP BY ticker_id
		) latest ON np.ticker_id = latest.ticker_id AND np.timestamp = latest.max_ts
	`
	rows, err := db.Query(ctx, query, tickerIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	prices := make(map[string]PriceData)
	for rows.Next() {
		var p PriceData
		if err := rows.Scan(&p.TickerID, &p.Price, &p.Change, &p.ChangePct, &p.Bid, &p.Ask, &p.Volume, &p.SourceID, &p.Timestamp); err != nil {
			continue
		}
		prices[p.TickerID] = p
	}
	return prices, nil
}

func (s *PortfolioService) GetLatestPriceSnapshots(ctx context.Context, db interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}, tickerIDs []string) (map[string][2]float64, error) {
	if len(tickerIDs) == 0 {
		return make(map[string][2]float64), nil
	}

	query := `
		WITH ranked_prices AS (
			SELECT ticker_id, price,
			       ROW_NUMBER() OVER (PARTITION BY ticker_id ORDER BY timestamp DESC) AS rank
			FROM normalized_prices
			WHERE ticker_id = ANY($1)
		)
		SELECT ticker_id,
		       COALESCE(MAX(CASE WHEN rank = 1 THEN price END), 0),
		       COALESCE(MAX(CASE WHEN rank = 2 THEN price END), 0)
		FROM ranked_prices
		WHERE rank <= 2
		GROUP BY ticker_id
	`
	rows, err := db.Query(ctx, query, tickerIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	snapshots := make(map[string][2]float64)
	for rows.Next() {
		var tickerID string
		var latestPrice, previousPrice float64
		if err := rows.Scan(&tickerID, &latestPrice, &previousPrice); err != nil {
			continue
		}
		snapshots[tickerID] = [2]float64{latestPrice, previousPrice}
	}

	return snapshots, nil
}

type TickerPriceBar struct {
	Timestamp string  `json:"timestamp"`
	Price     float64 `json:"price"`
	Volume    int64   `json:"volume"`
}

type TickerQuote struct {
	Symbol      string
	Name        string
	Exchange    string
	Price       float64
	Change      float64
	ChangePct   float64
	Volume      int64
	AvgVolume   int64
	MarketCap   float64
	Sector      string
	Industry    string
	Timestamp   time.Time
}

func (s *PortfolioService) GetTickerQuote(ctx context.Context, db interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}, symbol string) (*TickerQuote, error) {
	query := `
		SELECT t.symbol, t.company_name, t.exchange,
		       np.price, np.change, np.change_pct, np.volume, np.timestamp
		FROM tickers t
		LEFT JOIN normalized_prices np ON np.ticker_id = t.id
		WHERE t.symbol = $1 AND t.is_active = true
		  AND (np.timestamp IS NULL OR np.timestamp = (
			SELECT MAX(timestamp) FROM normalized_prices WHERE ticker_id = t.id
		))
		ORDER BY np.timestamp DESC NULLS LAST
		LIMIT 1
	`
	rows, err := db.Query(ctx, query, symbol)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		var q TickerQuote
		var price, change, changePct *float64
		var volume *int64
		var ts *time.Time
		if err := rows.Scan(&q.Symbol, &q.Name, &q.Exchange, &price, &change, &changePct, &volume, &ts); err != nil {
			return nil, err
		}
		if price != nil {
			q.Price = *price
		}
		if change != nil {
			q.Change = *change
		}
		if changePct != nil {
			q.ChangePct = *changePct
		}
		if volume != nil {
			q.Volume = *volume
		}
		if ts != nil {
			q.Timestamp = *ts
		}
		return &q, nil
	}
	return nil, nil
}

func (s *PortfolioService) GetPriceBars(ctx context.Context, db interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}, tickerID string, hours int) ([]TickerPriceBar, error) {
	query := `
		SELECT timestamp, price, volume
		FROM normalized_prices
		WHERE ticker_id = $1 AND timestamp >= NOW() - make_interval(hours => $2)
		ORDER BY timestamp ASC
	`
	rows, err := db.Query(ctx, query, tickerID, hours)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bars []TickerPriceBar
	for rows.Next() {
		var bar TickerPriceBar
		var ts time.Time
		if err := rows.Scan(&ts, &bar.Price, &bar.Volume); err != nil {
			continue
		}
		bar.Timestamp = ts.UTC().Format(time.RFC3339)
		bars = append(bars, bar)
	}
	return bars, nil
}

type IntradayBarData struct {
	Timestamp string  `json:"timestamp"`
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
	Volume    int64   `json:"volume"`
}

func (s *PortfolioService) GetIntradayBarsFromDB(ctx context.Context, db interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}, tickerID string, interval string, hours int) ([]IntradayBarData, error) {
	query := `
		SELECT timestamp, open, high, low, close, volume
		FROM intraday_bars
		WHERE ticker_id = $1 AND interval = $2 AND timestamp >= NOW() - ($3 || ' hours')::INTERVAL
		ORDER BY timestamp ASC
	`
	rows, err := db.Query(ctx, query, tickerID, interval, hours)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bars []IntradayBarData
	for rows.Next() {
		var bar IntradayBarData
		var ts time.Time
		if err := rows.Scan(&ts, &bar.Open, &bar.High, &bar.Low, &bar.Close, &bar.Volume); err != nil {
			continue
		}
		bar.Timestamp = ts.UTC().Format(time.RFC3339)
		bars = append(bars, bar)
	}
	return bars, nil
}

func (s *PortfolioService) GetPositions(ctx context.Context, db interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...interface{}) (int64, error)
}, portfolioID string) ([]Position, error) {
	query := `
		SELECT p.id, p.portfolio_id, p.ticker_id, t.symbol, t.company_name,
			   COALESCE(fd.json_data->>'sector', '') as sector,
			   p.quantity, p.avg_cost, p.opened_at
		FROM positions p
		JOIN tickers t ON t.id = p.ticker_id
		LEFT JOIN fundamental_data fd ON fd.ticker_id = t.id AND fd.data_type = 'company_profile'
		WHERE p.portfolio_id = $1
		ORDER BY p.quantity * p.avg_cost DESC
	`
	rows, err := db.Query(ctx, query, portfolioID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var positions []Position
	for rows.Next() {
		var pos Position
		if err := rows.Scan(&pos.ID, &pos.PortfolioID, &pos.TickerID, &pos.Symbol,
			&pos.CompanyName, &pos.Sector, &pos.Quantity, &pos.AvgCost, &pos.OpenedAt); err != nil {
			continue
		}
		positions = append(positions, pos)
	}
	return positions, nil
}

func (s *PortfolioService) AddPosition(ctx context.Context, db interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...interface{}) (int64, error)
}, portfolioID, symbol string, quantity, avgCost float64) error {
	var tickerID string
	rows, err := db.Query(ctx, `SELECT id FROM tickers WHERE symbol = $1 AND is_active = true`, symbol)
	if err != nil {
		return fmt.Errorf("ticker query failed: %s", symbol)
	}
	defer rows.Close()
	if rows.Next() {
		rows.Scan(&tickerID)
	} else {
		return fmt.Errorf("ticker not found: %s", symbol)
	}
	_, err = db.Exec(ctx, `
		INSERT INTO positions (portfolio_id, ticker_id, quantity, avg_cost, opened_at)
		VALUES ($1, $2, $3, $4, NOW())
	`, portfolioID, tickerID, quantity, avgCost)
	return err
}

func (s *PortfolioService) RemovePosition(ctx context.Context, db interface {
	Exec(ctx context.Context, sql string, args ...interface{}) (int64, error)
}, portfolioID, positionID string) error {
	_, err := db.Exec(ctx, `DELETE FROM positions WHERE id = $1 AND portfolio_id = $2`, positionID, portfolioID)
	return err
}

func (s *PortfolioService) GetPortfolioSummary(ctx context.Context, db interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...interface{}) (int64, error)
}, portfolioID string) (*PortfolioSummary, error) {
	rows, err := db.Query(ctx, `
		SELECT id, initial_cash FROM portfolios WHERE id = $1
	`, portfolioID)
	initialCash := 0.0
	if err == nil {
		defer rows.Close()
		if rows.Next() {
			var id string
			rows.Scan(&id, &initialCash)
		}
	}

	positions, err := s.GetPositions(ctx, db, portfolioID)
	if err != nil {
		return nil, err
	}

	var tickerIDs []string
	for _, p := range positions {
		tickerIDs = append(tickerIDs, p.TickerID)
	}

	prices, _ := s.GetLatestPrices(ctx, db, tickerIDs)
	snapshots, _ := s.GetLatestPriceSnapshots(ctx, db, tickerIDs)

	var totalValue, totalInvested, totalDayChange float64
	for i := range positions {
		if price, ok := prices[positions[i].TickerID]; ok {
			positions[i].CurrentPrice = price.Price
		}
		if positions[i].CurrentPrice == 0 {
			positions[i].CurrentPrice = positions[i].AvgCost
		}
		if snap, ok := snapshots[positions[i].TickerID]; ok && snap[1] > 0 {
			dayChange := snap[0] - snap[1]
			positions[i].DayChange = dayChange
			positions[i].DayChangePct = safePercent(dayChange, snap[1])
		}
		positions[i].CurrentValue = positions[i].Quantity * positions[i].CurrentPrice
		positions[i].TotalGain = positions[i].CurrentValue - (positions[i].Quantity * positions[i].AvgCost)
		positions[i].TotalGainPct = safePercent(positions[i].TotalGain, positions[i].Quantity*positions[i].AvgCost)
		totalValue += positions[i].CurrentValue
		totalInvested += positions[i].Quantity * positions[i].AvgCost
		totalDayChange += positions[i].DayChange * positions[i].Quantity
	}

	totalDayChangePct := safePercent(totalDayChange, totalValue)
	totalGain := totalValue - totalInvested
	totalGainPct := safePercent(totalGain, totalInvested)

	return &PortfolioSummary{
		TotalValue:    totalValue,
		DayChange:     totalDayChange,
		DayChangePct:  totalDayChangePct,
		TotalInvested: totalInvested,
		TotalGain:     totalGain,
		TotalGainPct:  totalGainPct,
		CashBalance:   initialCash,
	}, nil
}

type CompanyProfile struct {
	Symbol        string
	Name          string
	Exchange      string
	Sector        string
	Industry      string
	MarketCap     int64
	PeRatio       float64
	Eps           float64
	DividendYield float64
	Week52High    float64
	Week52Low     float64
	AvgVolume     int64
}

func (s *PortfolioService) GetCompanyProfile(ctx context.Context, db interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}, tickerID string) (*CompanyProfile, error) {
	query := `
		SELECT json_data
		FROM fundamental_data
		WHERE ticker_id = $1 AND source_id = 'company_profile' AND data_type = 'profile'
		ORDER BY timestamp DESC
		LIMIT 1
	`
	rows, err := db.Query(ctx, query, tickerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		var raw []byte
		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}
		var profile CompanyProfile
		if err := json.Unmarshal(raw, &profile); err != nil {
			return nil, err
		}
		return &profile, nil
	}
	return nil, nil
}

func (s *PortfolioService) SaveCompanyProfile(ctx context.Context, db interface {
	Exec(ctx context.Context, sql string, args ...interface{}) (int64, error)
}, tickerID string, profile *CompanyProfile) error {
	data, err := json.Marshal(profile)
	if err != nil {
		return err
	}
	query := `
		INSERT INTO fundamental_data (ticker_id, source_id, data_type, period, json_data, timestamp)
		VALUES ($1, 'company_profile', 'profile', '', $2, NOW())
	`
	_, err = db.Exec(ctx, query, tickerID, data)
	return err
}

func (s *PortfolioService) SaveIntradayBars(ctx context.Context, db interface {
	Exec(ctx context.Context, sql string, args ...interface{}) (int64, error)
}, tickerID string, bars []IntradayBarData) error {
	if len(bars) == 0 {
		return nil
	}
	for _, bar := range bars {
		timestamp, err := time.Parse(time.RFC3339, bar.Timestamp)
		if err != nil {
			continue
		}
		query := `
			INSERT INTO intraday_bars (ticker_id, interval, open, high, low, close, volume, timestamp)
			VALUES ($1, '1min', $2, $3, $4, $5, $6, $7)
			ON CONFLICT (ticker_id, interval, timestamp) DO UPDATE SET
				open = EXCLUDED.open,
				high = EXCLUDED.high,
				low = EXCLUDED.low,
				close = EXCLUDED.close,
				volume = EXCLUDED.volume
		`
		_, err = db.Exec(ctx, query, tickerID, bar.Open, bar.High, bar.Low, bar.Close, bar.Volume, timestamp)
		if err != nil {
			s.logger.Warn("failed to save intraday bar", "ticker_id", tickerID, "error", err)
		}
	}
	return nil
}

func (s *PortfolioService) GetTickerBySymbol(ctx context.Context, db interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}, symbol string) (string, string, error) {
	query := `SELECT id, company_name FROM tickers WHERE symbol = $1 AND is_active = true`
	rows, err := db.Query(ctx, query, symbol)
	if err != nil {
		return "", "", err
	}
	defer rows.Close()

	if rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			return "", "", err
		}
		return id, name, nil
	}
	return "", "", nil
}

func (s *PortfolioService) GetMarketIndexSettings(ctx context.Context, db interface {
	QueryRow(ctx context.Context, sql string, args ...interface{}) (pgx.Row, error)
}) ([]MarketIndexSetting, error) {
	row, _ := db.QueryRow(ctx, `
		SELECT value
		FROM app_settings
		WHERE setting_key = $1
	`, "market_indices")

	var raw []byte
	if err := row.Scan(&raw); err != nil {
		if err == pgx.ErrNoRows {
			return defaultMarketIndexSettings(), nil
		}

		return nil, err
	}

	var settings []MarketIndexSetting
	if err := json.Unmarshal(raw, &settings); err != nil {
		return defaultMarketIndexSettings(), nil
	}

	return normalizeMarketIndexSettings(settings), nil
}

func (s *PortfolioService) SaveMarketIndexSettings(ctx context.Context, db interface {
	Exec(ctx context.Context, sql string, args ...interface{}) (int64, error)
}, settings []MarketIndexSetting) error {
	normalized := normalizeMarketIndexSettings(settings)
	payload, err := json.Marshal(normalized)
	if err != nil {
		return err
	}

	_, err = db.Exec(ctx, `
		INSERT INTO app_settings (setting_key, value, updated_at)
		VALUES ($1, $2::jsonb, NOW())
		ON CONFLICT (setting_key) DO UPDATE
		SET value = EXCLUDED.value,
		    updated_at = NOW()
	`, "market_indices", payload)
	return err
}

func (s *PortfolioService) GetMarketIndices(ctx context.Context, db interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) (pgx.Row, error)
}) ([]MarketIndex, error) {
	indexSettings, err := s.GetMarketIndexSettings(ctx, db)
	if err != nil {
		return nil, err
	}

	indexSymbols := make([]string, 0, len(indexSettings))
	for _, setting := range indexSettings {
		indexSymbols = append(indexSymbols, setting.Symbol)
	}

	tickers, err := s.GetTickers(ctx, db, indexSymbols)
	if err != nil {
		return nil, err
	}

	var tickerIDs []string
	tickersBySymbol := make(map[string]Ticker, len(tickers))
	for _, t := range tickers {
		tickerIDs = append(tickerIDs, t.ID)
		tickersBySymbol[t.Symbol] = t
	}

	priceSnapshots, _ := s.GetLatestPriceSnapshots(ctx, db, tickerIDs)

	indices := make([]MarketIndex, 0, len(indexSettings))
	for _, setting := range indexSettings {
		price := 0.0
		change := 0.0
		changePct := 0.0

		if ticker, ok := tickersBySymbol[setting.Symbol]; ok {
			if snapshot, hasSnapshot := priceSnapshots[ticker.ID]; hasSnapshot {
				price = snapshot[0]
				if snapshot[1] != 0 {
					change = snapshot[0] - snapshot[1]
					changePct = safePercent(change, snapshot[1])
				}
			}
		}

		indices = append(indices, MarketIndex{
			Symbol:    setting.Symbol,
			Name:      setting.Name,
			Price:     price,
			Change:    change,
			ChangePct: changePct,
		})
	}
	return indices, nil
}

func (s *PortfolioService) GetNewsArticles(ctx context.Context, db interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}) ([]NewsArticle, error) {
	query := `
		SELECT id, title, source, source_type, url, summary, content,
		       sentiment, sentiment_value, published_at, channel
		FROM news_articles
		ORDER BY published_at DESC
	`
	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articles []NewsArticle
	for rows.Next() {
		var a NewsArticle
		var content, sentimentValue, channel *string
		if err := rows.Scan(&a.ID, &a.Title, &a.Source, &a.SourceType, &a.URL,
			&a.Summary, content, &a.Sentiment, sentimentValue, &a.PublishedAt, channel); err != nil {
			continue
		}
		if content != nil {
			a.Content = *content
		}
		if sentimentValue != nil {
			a.SentimentValue = *sentimentValue
		}
		if channel != nil {
			a.Channel = *channel
		}
		articles = append(articles, a)
	}
	return articles, nil
}

func (s *PortfolioService) GetStrategies(ctx context.Context, db interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}) ([]Strategy, error) {
	query := `
		SELECT COALESCE(sim.metrics->>'total_return', '0') as returns,
			   COALESCE(sim.metrics->>'sharpe_ratio', '0') as sharpe,
			   COALESCE(sim.metrics->>'max_drawdown', '0') as maxdd,
			   COALESCE(sim.metrics->>'num_trades', '0') as trades,
			   COALESCE(sim.metrics->>'win_rate', '0') as winrate,
			   sim.created_at
		FROM simulation_runs sim
		ORDER BY sim.created_at DESC
		LIMIT 10
	`
	rows, err := db.Query(ctx, query)
	if err != nil {
		return []Strategy{}, nil
	}
	defer rows.Close()

	var strategies []Strategy
	strategyNames := []string{"Momentum Growth", "Value Scanner", "Mean Reversion", "Sector Rotation"}
	statuses := []string{"active", "active", "paused", "active"}

	i := 0
	for rows.Next() {
		var returns, sharpe, maxDD, winRate float64
		var trades int
		var createdAt time.Time
		if err := rows.Scan(&returns, &sharpe, &maxDD, &trades, &winRate, &createdAt); err != nil {
			continue
		}
		name := "Strategy"
		status := "active"
		if i < len(strategyNames) {
			name = strategyNames[i]
			status = statuses[i]
		}
		strategies = append(strategies, Strategy{
			ID:      createdAt.Format("20060102150405"),
			Name:    name,
			Status:  status,
			Returns: returns,
			Sharpe:  sharpe,
			MaxDD:   maxDD,
			Trades:  trades,
			WinRate: winRate,
		})
		i++
	}

	return strategies, nil
}

func (s *PortfolioService) GetSignals(ctx context.Context, db interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}, limit int) ([]Signal, error) {
	if limit <= 0 {
		limit = 10
	}
	query := `
		SELECT id, message, service, level, timestamp
		FROM logs
		WHERE service LIKE 'signal-%'
		ORDER BY timestamp DESC
		LIMIT $1
	`
	rows, err := db.Query(ctx, query, limit)
	if err != nil {
		return []Signal{}, nil
	}
	defer rows.Close()

	var signals []Signal
	for rows.Next() {
		var id, message, service, level string
		var timestamp time.Time
		if err := rows.Scan(&id, &message, &service, &level, &timestamp); err != nil {
			continue
		}
		signal := Signal{
			ID:        id,
			Message:   message,
			Service:   service,
			Level:     level,
			Timestamp: timestamp,
		}
		signals = append(signals, signal)
	}

	return signals, nil
}
