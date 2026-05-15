package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetTickerQuoteRequestsProfileFalse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/tickers/DIA/details" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("profile"); got != "false" {
			t.Fatalf("expected profile=false, got %q", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"symbol":"DIA","name":"Dow Jones","price":450.25,"change":1.1,"changePct":0.24}`))
	}))
	defer server.Close()

	service := NewTickerService(server.URL, nil)
	details, err := service.GetTickerQuote(context.Background(), "DIA")
	if err != nil {
		t.Fatalf("GetTickerQuote returned error: %v", err)
	}
	if details.Symbol != "DIA" || details.Price != 450.25 {
		t.Fatalf("unexpected details: %+v", details)
	}
}
