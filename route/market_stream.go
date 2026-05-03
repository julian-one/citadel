package route

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"citadel/internal/broker"

	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata/stream"
)

func StreamMarketData(l *slog.Logger, b *broker.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		symbol := r.URL.Query().Get("symbol")
		if symbol == "" {
			http.Error(w, "symbol is required", http.StatusBadRequest)
			return
		}

		// SSE headers
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
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
			l.Error("failed to subscribe to trades", "symbol", symbol, "error", err)
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
