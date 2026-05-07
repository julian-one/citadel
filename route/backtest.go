package route

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"citadel/internal/broker"
	"citadel/internal/database"
	"citadel/internal/quant"
	"citadel/internal/quant/strategies"
	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type BacktestRequest struct {
	Symbols         []string               `json:"symbols"`
	Start           string                 `json:"start_date"`
	End             string                 `json:"end_date"`
	Strategy        string                 `json:"strategy"`
	StartingCapital float64                `json:"starting_capital"`
	Parameters      map[string]interface{} `json:"parameters"`
}

type BacktestResponse struct {
	Portfolio *quant.Portfolio `json:"portfolio"`
}

func RunBacktest(logger *slog.Logger, b *broker.Client, db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req BacktestRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
			return
		}

		start, err := time.Parse(time.RFC3339, req.Start)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid start_date format"})
			return
		}

		end, err := time.Parse(time.RFC3339, req.End)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid end_date format"})
			return
		}

		if len(req.Symbols) == 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "at least one symbol is required"})
			return
		}

		// Fetch historical data for all symbols
		barsMap := make(map[string][]marketdata.Bar)
		for i, sym := range req.Symbols {
			sym = strings.ToUpper(sym)
			req.Symbols[i] = sym

			bars, err := b.GetHistoricalBars(sym, start, end)
			if err != nil {
				logger.Error(
					"failed to get historical bars for backtest",
					"error",
					err,
					"symbol",
					sym,
				)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).
					Encode(map[string]string{"error": "failed to fetch market data for " + sym})
				return
			}
			barsMap[sym] = bars
		}

		logger.Info(
			"fetched backtest data",
			"symbols",
			strings.Join(req.Symbols, ","),
			"start",
			start,
			"end",
			end,
		)

		// Select strategy
		var strategy quant.Strategy
		switch req.Strategy {
		case "sma_crossover":
			shortPeriod := 10
			longPeriod := 50
			if req.Parameters != nil {
				if v, ok := req.Parameters["short_period"].(float64); ok {
					shortPeriod = int(v)
				}
				if v, ok := req.Parameters["long_period"].(float64); ok {
					longPeriod = int(v)
				}
			}
			strategy = strategies.NewSMACrossover(req.Symbols[0], shortPeriod, longPeriod)
		case "rsi_reversion":
			period := 14
			oversold := 30.0
			overbought := 70.0
			if req.Parameters != nil {
				if v, ok := req.Parameters["period"].(float64); ok {
					period = int(v)
				}
				if v, ok := req.Parameters["oversold"].(float64); ok {
					oversold = v
				}
				if v, ok := req.Parameters["overbought"].(float64); ok {
					overbought = v
				}
			}
			strategy = strategies.NewRSIReversion(req.Symbols[0], period, oversold, overbought)
		case "bollinger_bands":
			period := 20
			stdDev := 2.0
			if req.Parameters != nil {
				if v, ok := req.Parameters["period"].(float64); ok {
					period = int(v)
				}
				if v, ok := req.Parameters["std_dev"].(float64); ok {
					stdDev = v
				}
			}
			strategy = strategies.NewBollingerBands(req.Symbols[0], period, stdDev)
		case "pairs_trading":
			if len(req.Symbols) < 2 {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).
					Encode(map[string]string{"error": "pairs trading requires at least 2 symbols"})
				return
			}
			period := 20
			entryZ := 2.0
			exitZ := 0.0
			if req.Parameters != nil {
				if v, ok := req.Parameters["period"].(float64); ok {
					period = int(v)
				}
				if v, ok := req.Parameters["entry_z"].(float64); ok {
					entryZ = v
				}
				if v, ok := req.Parameters["exit_z"].(float64); ok {
					exitZ = v
				}
			}
			strategy = strategies.NewPairsTrading(
				req.Symbols[0],
				req.Symbols[1],
				period,
				entryZ,
				exitZ,
			)
		default:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "unknown strategy"})
			return
		}

		if req.StartingCapital <= 0 {
			req.StartingCapital = 100000.0 // Default if not provided
		}

		maxPosPct := 0.05
		dailyStopPct := 0.02
		if req.Parameters != nil {
			if v, ok := req.Parameters["max_position_size_pct"].(float64); ok {
				maxPosPct = v
			}
			if v, ok := req.Parameters["daily_stop_loss_pct"].(float64); ok {
				dailyStopPct = v
			}
		}

		rm := quant.NewDefaultRiskManager(maxPosPct, dailyStopPct)

		// Run simulation
		engine := quant.NewEngine(req.StartingCapital, strategy, rm)
		if err := engine.Run(barsMap); err != nil {
			logger.Error("backtest engine failed", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "engine failed"})
			return
		}

		engine.Portfolio.CalculateMetrics()

		// Save backtest results
		paramsJSON, _ := json.Marshal(req.Parameters)
		symbolsJSON, _ := json.Marshal(req.Symbols)
		metricsJSON, _ := json.Marshal(engine.Portfolio.Metrics)

		record := database.BacktestRecord{
			BacktestID:      uuid.NewString(),
			Strategy:        req.Strategy,
			Symbols:         string(symbolsJSON),
			StartDate:       req.Start,
			EndDate:         req.End,
			StartingCapital: req.StartingCapital,
			Parameters:      string(paramsJSON),
			Metrics:         string(metricsJSON),
		}

		if err := database.SaveBacktest(r.Context(), db, record); err != nil {
			logger.Error("failed to save backtest", "error", err)
			// we don't abort, just log it.
		}

		// Return portfolio state (trades and equity log)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(BacktestResponse{
			Portfolio: engine.Portfolio,
		})
	}
}

func ListBacktests(logger *slog.Logger, db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		records, err := database.ListBacktests(r.Context(), db)
		if err != nil {
			logger.Error("failed to list backtests", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to retrieve backtests"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(records)
	}
}
