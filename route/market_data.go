package route

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"citadel/internal/broker"
)

func GetHistoricalBars(logger *slog.Logger, b *broker.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		symbol := strings.ToUpper(r.URL.Query().Get("symbol"))
		if symbol == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "symbol is required"})
			return
		}

		startStr := r.URL.Query().Get("start")
		endStr := r.URL.Query().Get("end")

		var start, end time.Time
		var err error

		if startStr != "" {
			start, err = time.Parse(time.RFC3339, startStr)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).
					Encode(map[string]string{"error": "invalid start date format, expected RFC3339"})
				return
			}
		} else {
			start = time.Now().AddDate(0, 0, -30) // Default to last 30 days
		}

		if endStr != "" {
			end, err = time.Parse(time.RFC3339, endStr)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).
					Encode(map[string]string{"error": "invalid end date format, expected RFC3339"})
				return
			}
		} else {
			end = time.Now()
		}

		bars, err := b.GetHistoricalBars(symbol, start, end)
		if err != nil {
			logger.Error("failed to get historical bars", "symbol", symbol, "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "failed to fetch market data"})
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(bars)
	}
}
