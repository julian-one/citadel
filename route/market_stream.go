package route

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"citadel/internal/broker"

	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata/stream"
)

func StreamMarketData(logger *slog.Logger, b *broker.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		symbol := r.URL.Query().Get("symbol")
		if symbol == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "symbol is required"})
			return
		}

		// SSE headers
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "streaming unsupported"})
			return
		}

		tradeChan := make(chan stream.Trade, 50)

		handler := func(t stream.Trade) {
			select {
			case tradeChan <- t:
			default:
				// channel full, drop to prevent blocking the stream client
			}
		}

		err := b.Stream.SubscribeToTrades(handler, symbol)
		if err != nil {
			logger.Error("failed to subscribe to trades", "symbol", symbol, "error", err)
			return
		}

		// Note: we purposely omit UnsubscribeFromTrades on client disconnect here
		// because the Alpaca SDK manages a single handler map per symbol.
		// If a user refreshes the page, the new request subscribes, and the old request's
		// cancel context runs right after, which would accidentally unsubscribe the new request.
		// For a single-user system, leaving the subscription active is fine.

		ctx := r.Context()
		for {
			select {
			case <-ctx.Done():
				return // Client disconnected
			case t := <-tradeChan:
				data, _ := json.Marshal(t)
				w.Write([]byte("data: "))
				w.Write(data)
				w.Write([]byte("\n\n"))
				flusher.Flush()
			case <-time.After(15 * time.Second):
				w.Write([]byte(": ping\n\n"))
				flusher.Flush()
			}
		}
	}
}
