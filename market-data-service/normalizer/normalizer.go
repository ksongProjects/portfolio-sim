package normalizer

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/portfolio-sim/backend/services/market-data-service/providers"
)

type NormalizedPrice struct {
	TickerID uuid.UUID
	Price    float64
	Bid      float64
	Ask      float64
	Volume   int64
	SourceID string
	Timestamp time.Time
}

type NormalizedIntradayBar struct {
	TickerID uuid.UUID
	Interval  string
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    int64
	Timestamp time.Time
}

type NormalizedOptionChain struct {
	UnderlyingTickerID uuid.UUID
	Expiration         time.Time
	Strike             float64
	OptionType         string
	Bid                float64
	Ask                float64
	Delta              float64
	Gamma              float64
	Theta              float64
	Vega               float64
	ImpliedVol         float64
	Volume             int64
	OpenInterest       int64
	SourceID           string
	Timestamp          time.Time
}

type Normalizer struct{}

func NewNormalizer() *Normalizer {
	return &Normalizer{}
}

func (n *Normalizer) NormalizePrice(price *providers.Price, tickerID uuid.UUID) (*NormalizedPrice, error) {
	return &NormalizedPrice{
		TickerID: tickerID,
		Price:    price.Price,
		Bid:      price.Bid,
		Ask:      price.Ask,
		Volume:   price.Volume,
		SourceID: price.Source,
		Timestamp: price.Timestamp,
	}, nil
}

func (n *Normalizer) NormalizeIntradayBar(bar *providers.IntradayBar, tickerID uuid.UUID) (*NormalizedIntradayBar, error) {
	return &NormalizedIntradayBar{
		TickerID: tickerID,
		Interval:  bar.Interval,
		Open:      bar.Open,
		High:      bar.High,
		Low:       bar.Low,
		Close:     bar.Close,
		Volume:    bar.Volume,
		Timestamp: bar.Timestamp,
	}, nil
}

func (n *Normalizer) NormalizeOptionChain(chain *providers.OptionChain, tickerID uuid.UUID) (*NormalizedOptionChain, error) {
	return &NormalizedOptionChain{
		UnderlyingTickerID: tickerID,
		Expiration:        chain.Expiration,
		Strike:            chain.Strike,
		OptionType:        chain.OptionType,
		Bid:               chain.Bid,
		Ask:               chain.Ask,
		Delta:             chain.Delta,
		Gamma:             chain.Gamma,
		Theta:             chain.Theta,
		Vega:              chain.Vega,
		ImpliedVol:        chain.ImpliedVol,
		Volume:            chain.Volume,
		OpenInterest:      chain.OpenInterest,
		SourceID:          chain.Source,
		Timestamp:         chain.Timestamp,
	}, nil
}

func (n *Normalizer) PriceToJSON(np *NormalizedPrice) ([]byte, error) {
	return json.Marshal(np)
}

func (n *Normalizer) BarToJSON(nb *NormalizedIntradayBar) ([]byte, error) {
	return json.Marshal(nb)
}

func (n *Normalizer) ChainToJSON(nc *NormalizedOptionChain) ([]byte, error) {
	return json.Marshal(nc)
}
