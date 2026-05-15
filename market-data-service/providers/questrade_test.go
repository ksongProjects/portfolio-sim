package providers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/portfolio-sim/market-data-service/config"
)

func TestQuestradeFetchPriceUsesSymbolIDAndJSONAccept(t *testing.T) {
	var searchCount int
	var quoteCount int

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Accept"); got != "application/json" {
			t.Fatalf("expected Accept application/json, got %q", got)
		}

		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/symbols/search":
			searchCount++
			if got := r.URL.Query().Get("prefix"); got != "NVDA" {
				t.Fatalf("expected search prefix NVDA, got %q", got)
			}
			_, _ = w.Write([]byte(`{"symbols":[{"symbol":"NVDA","symbolId":123,"description":"NVIDIA Corp.","securityType":"Stock","listingExchange":"NASDAQ","isQuotable":true,"isTradable":true}]}`))
		case "/v1/markets/quotes/123":
			quoteCount++
			_, _ = w.Write([]byte(`{"quotes":[{"symbol":"NVDA","lastTradePrice":905.25,"change":1.5,"changePercent":0.17,"bidPrice":905.1,"askPrice":905.4,"volume":1000,"lastTradeTime":"2026-05-14T09:30:00.000000-04:00"}]}`))
		case "/v1/markets/quotes/NVDA":
			t.Fatalf("quote endpoint used ticker instead of symbol ID")
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	provider := &QuestradeProvider{
		cfg:            config.QuestradeConfig{},
		client:         server.Client(),
		baseURL:        server.URL,
		token:          "token",
		tokenExpiresAt: time.Now().Add(time.Hour),
		rateLimiter:    NewRateLimiter(20, 15000),
	}

	price, err := provider.FetchPrice("NVDA")
	if err != nil {
		t.Fatalf("FetchPrice returned error: %v", err)
	}
	if price.Ticker != "NVDA" || price.Price != 905.25 || price.Change != 1.5 || price.ChangePct != 0.17 {
		t.Fatalf("unexpected price: %+v", price)
	}

	if _, err := provider.FetchPrice("NVDA"); err != nil {
		t.Fatalf("second FetchPrice returned error: %v", err)
	}
	if searchCount != 1 || quoteCount != 2 {
		t.Fatalf("expected one symbol search and two quotes, searchCount=%d quoteCount=%d", searchCount, quoteCount)
	}
}
