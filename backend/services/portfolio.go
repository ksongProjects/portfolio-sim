package services

import (
	"context"
	"fmt"
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
	TickerID    string
	Price       float64
	Bid         float64
	Ask         float64
	Volume      int64
	SourceID    string
	Timestamp   time.Time
}

type Position struct {
	ID          string
	PortfolioID string
	TickerID    string
	Symbol      string
	CompanyName string
	Quantity    float64
	AvgCost     float64
	CurrentPrice float64
	CurrentValue float64
	DayChange   float64
	DayChangePct float64
	TotalGain   float64
	TotalGainPct float64
	OpenedAt    time.Time
}

type PortfolioSummary struct {
	TotalValue    float64
	DayChange     float64
	DayChangePct  float64
	TotalInvested  float64
	TotalGain      float64
	TotalGainPct   float64
	CashBalance    float64
}

type MarketIndex struct {
	Symbol   string
	Name     string
	Price    float64
	Change   float64
	ChangePct float64
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
	ID          string
	Title       string
	Source      string
	URL         string
	Summary     string
	Sentiment   string
	PublishedAt time.Time
	TickerSymbols []string
}

type Strategy struct {
	ID       string
	Name     string
	Status   string
	Returns  float64
	Sharpe   float64
	MaxDD    float64
	Trades   int
	WinRate  float64
}

type Signal struct {
	ID        string
	Message   string
	Service   string
	Level     string
	Timestamp time.Time
}

type PortfolioService struct{}

func NewPortfolioService() *PortfolioService {
	return &PortfolioService{}
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
		SELECT np.ticker_id, np.price, np.bid, np.ask, np.volume, np.source_id, np.timestamp
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
		if err := rows.Scan(&p.TickerID, &p.Price, &p.Bid, &p.Ask, &p.Volume, &p.SourceID, &p.Timestamp); err != nil {
			continue
		}
		prices[p.TickerID] = p
	}
	return prices, nil
}

func (s *PortfolioService) GetPositions(ctx context.Context, db interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...interface{}) (int64, error)
}, portfolioID string) ([]Position, error) {
	query := `
		SELECT p.id, p.portfolio_id, p.ticker_id, t.symbol, t.company_name,
			   p.quantity, p.avg_cost, p.opened_at
		FROM positions p
		JOIN tickers t ON t.id = p.ticker_id
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
			&pos.CompanyName, &pos.Quantity, &pos.AvgCost, &pos.OpenedAt); err != nil {
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

	var totalValue, totalInvested, totalDayChange, totalDayChangePct float64
	for i := range positions {
		if price, ok := prices[positions[i].TickerID]; ok {
			positions[i].CurrentPrice = price.Price
		}
		if positions[i].CurrentPrice == 0 {
			positions[i].CurrentPrice = positions[i].AvgCost
		}
		positions[i].CurrentValue = positions[i].Quantity * positions[i].CurrentPrice
		positions[i].TotalGain = positions[i].CurrentValue - (positions[i].Quantity * positions[i].AvgCost)
		positions[i].TotalGainPct = (positions[i].TotalGain / (positions[i].Quantity * positions[i].AvgCost)) * 100
		positions[i].DayChange = positions[i].CurrentPrice * 0.01
		positions[i].DayChangePct = 1.0
		totalValue += positions[i].CurrentValue
		totalInvested += positions[i].Quantity * positions[i].AvgCost
		totalDayChange += positions[i].DayChange * positions[i].Quantity
	}

	totalDayChangePct = (totalDayChange / totalValue) * 100
	if totalInvested == 0 {
		totalInvested = 1
	}

	return &PortfolioSummary{
		TotalValue:    totalValue,
		DayChange:     totalDayChange,
		DayChangePct:  totalDayChangePct,
		TotalInvested: totalInvested,
		TotalGain:     totalValue - totalInvested,
		TotalGainPct:  ((totalValue - totalInvested) / totalInvested) * 100,
		CashBalance:   initialCash,
	}, nil
}

func (s *PortfolioService) GetMarketIndices(ctx context.Context, db interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}) ([]MarketIndex, error) {
	indexSymbols := []string{"SPY", "QQQ", "DIA", "IWM", "VIX", "DXY"}
	indexNames := map[string]string{
		"SPY": "S&P 500 ETF",
		"QQQ": "Nasdaq ETF",
		"DIA": "Dow Jones ETF",
		"IWM": "Russell 2000 ETF",
		"VIX": "Volatility Index",
		"DXY": "US Dollar Index",
	}

	tickers, err := s.GetTickers(ctx, db, indexSymbols)
	if err != nil || len(tickers) == 0 {
		return []MarketIndex{}, nil
	}

	var tickerIDs []string
	for _, t := range tickers {
		tickerIDs = append(tickerIDs, t.ID)
	}

	prices, _ := s.GetLatestPrices(ctx, db, tickerIDs)

	var indices []MarketIndex
	for _, sym := range indexSymbols {
		name := indexNames[sym]
		price := 0.0
		change := 0.0
		changePct := 0.0
		for _, t := range tickers {
			if t.Symbol == sym {
				if p, ok := prices[t.ID]; ok {
					price = p.Price
				}
				break
			}
		}
		indices = append(indices, MarketIndex{
			Symbol: sym, Name: name, Price: price, Change: change, ChangePct: changePct,
		})
	}
	return indices, nil
}

func (s *PortfolioService) GetNewsArticles(ctx context.Context, db interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
}) ([]NewsArticle, error) {
	query := `
		SELECT id, title, source, url, summary, sentiment, published_at
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
		if err := rows.Scan(&a.ID, &a.Title, &a.Source, &a.URL, &a.Summary, &a.Sentiment, &a.PublishedAt); err != nil {
			continue
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