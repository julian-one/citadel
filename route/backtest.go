package route

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"citadel/internal/broker"
	"citadel/internal/quant"
	"citadel/internal/quant/strategies"
)

type BacktestRequest struct {
	Symbol          string  `json:"symbol"`
	Start           string  `json:"start_date"`
	End             string  `json:"end_date"`
	Strategy        string  `json:"strategy"`
	StartingCapital float64 `json:"starting_capital"`
}

type BacktestResponse struct {
	Portfolio *quant.Portfolio `json:"portfolio"`
}

func RunBacktest(l *slog.Logger, b *broker.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req BacktestRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		start, err := time.Parse(time.RFC3339, req.Start)
		if err != nil {
			http.Error(w, "invalid start_date format", http.StatusBadRequest)
			return
		}

		end, err := time.Parse(time.RFC3339, req.End)
		if err != nil {
			http.Error(w, "invalid end_date format", http.StatusBadRequest)
			return
		}

		req.Symbol = strings.ToUpper(req.Symbol)

		// Fetch historical data
		bars, err := b.GetHistoricalBars(req.Symbol, start, end)
		if err != nil {
			l.Error("failed to get historical bars for backtest", "error", err)
			http.Error(w, "failed to fetch market data", http.StatusInternalServerError)
			return
		}

		l.Info("fetched backtest data", "symbol", req.Symbol, "start", start, "end", end, "bars", len(bars))

		// Select strategy
		var strategy quant.Strategy
		switch req.Strategy {
		case "sma_crossover":
			strategy = strategies.NewSMACrossover(req.Symbol, 10, 50)
		default:
			http.Error(w, "unknown strategy", http.StatusBadRequest)
			return
		}

		if req.StartingCapital <= 0 {
			req.StartingCapital = 100000.0 // Default if not provided
		}

		// Run simulation
		engine := quant.NewEngine(req.StartingCapital, strategy)
		if err := engine.Run(bars, req.Symbol); err != nil {
			l.Error("backtest engine failed", "error", err)
			http.Error(w, "engine failed", http.StatusInternalServerError)
			return
		}

		engine.Portfolio.CalculateMetrics()

		// Return portfolio state (trades and equity log)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(BacktestResponse{
			Portfolio: engine.Portfolio,
		})
	}
}
