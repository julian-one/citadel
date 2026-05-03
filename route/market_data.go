package route

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"citadel/internal/broker"
)

func GetHistoricalBars(l *slog.Logger, b *broker.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		symbol := strings.ToUpper(r.URL.Query().Get("symbol"))
		if symbol == "" {
			http.Error(w, "symbol is required", http.StatusBadRequest)
			return
		}

		startStr := r.URL.Query().Get("start")
		endStr := r.URL.Query().Get("end")

		var start, end time.Time
		var err error

		if startStr != "" {
			start, err = time.Parse(time.RFC3339, startStr)
			if err != nil {
				http.Error(w, "invalid start date format, expected RFC3339", http.StatusBadRequest)
				return
			}
		} else {
			start = time.Now().AddDate(0, 0, -30) // Default to last 30 days
		}

		if endStr != "" {
			end, err = time.Parse(time.RFC3339, endStr)
			if err != nil {
				http.Error(w, "invalid end date format, expected RFC3339", http.StatusBadRequest)
				return
			}
		} else {
			end = time.Now()
		}

		bars, err := b.GetHistoricalBars(symbol, start, end)
		if err != nil {
			l.Error("failed to get historical bars", "symbol", symbol, "error", err)
			http.Error(w, "failed to fetch market data", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(bars)
	}
}
