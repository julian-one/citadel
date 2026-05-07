package route

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"sync"

	"citadel/internal/broker"
	"citadel/internal/database"
	"citadel/internal/quant"
	"citadel/internal/quant/strategies"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
)

func GetTradingAccount(logger *slog.Logger, b *broker.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, err := b.GetAccount()
		if err != nil {
			logger.Error("failed to get trading account", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).
				Encode(map[string]string{"error": "Failed to retrieve trading account"})
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(account)
	}
}

var (
	activeEngines = make(map[string]*quant.LiveEngine)
	enginesMutex  sync.RWMutex
)

type StartLiveRequest struct {
	Symbols         []string               `json:"symbols"`
	Strategy        string                 `json:"strategy"`
	StartingCapital float64                `json:"starting_capital"`
	Parameters      map[string]interface{} `json:"parameters"`
}

func StartLiveEngine(logger *slog.Logger, b *broker.Client, db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req StartLiveRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
			return
		}

		if len(req.Symbols) == 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "symbols required"})
			return
		}

		for i, sym := range req.Symbols {
			req.Symbols[i] = strings.ToUpper(sym)
		}

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
				json.NewEncoder(w).Encode(map[string]string{"error": "requires 2 symbols"})
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
		engine := quant.NewLiveEngine(db, b, req.StartingCapital, strategy, req.Symbols, req.Parameters, rm)
		if err := engine.Start(r.Context(), logger); err != nil {
			logger.Error("failed to start live engine", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		enginesMutex.Lock()
		activeEngines[engine.SessionID] = engine
		enginesMutex.Unlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message":    "Live engine started",
			"session_id": engine.SessionID,
		})
	}
}

func StopLiveEngine(logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.URL.Query().Get("session_id")
		if sessionID == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "session_id required"})
			return
		}

		enginesMutex.Lock()
		engine, ok := activeEngines[sessionID]
		if ok {
			delete(activeEngines, sessionID)
		}
		enginesMutex.Unlock()

		if !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).
				Encode(map[string]string{"error": "session not found or already stopped"})
			return
		}

		engine.Stop(r.Context())

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Live engine stopped"})
	}
}

func ListTradingSessions(logger *slog.Logger, db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessions, err := database.ListTradingSessions(r.Context(), db)
		if err != nil {
			logger.Error("failed to list trading sessions", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).
				Encode(map[string]string{"error": "Failed to retrieve trading sessions"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(sessions)
	}
}

func GetTradingSessionDetails(logger *slog.Logger, db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.PathValue("id")
		if sessionID == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "session ID required"})
			return
		}

		session, err := database.GetTradingSession(r.Context(), db, sessionID)
		if err != nil {
			logger.Error("failed to get trading session", "error", err, "session_id", sessionID)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Trading session not found"})
			return
		}

		orders, err := database.GetTradingSessionOrders(r.Context(), db, sessionID)
		if err != nil {
			logger.Error(
				"failed to get trading session orders",
				"error",
				err,
				"session_id",
				sessionID,
			)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).
				Encode(map[string]string{"error": "Failed to retrieve session orders"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"session": session,
			"orders":  orders,
		})
	}
}

func GetPositions(logger *slog.Logger, b *broker.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		positions, err := b.GetPositions()
		if err != nil {
			logger.Error("failed to get positions", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to retrieve positions"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(positions)
	}
}

func GetPortfolioHistory(logger *slog.Logger, b *broker.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		period := r.URL.Query().Get("period")
		if period == "" {
			period = "1M"
		}
		timeframe := r.URL.Query().Get("timeframe")
		if timeframe == "" {
			timeframe = "1D"
		}

		history, err := b.GetPortfolioHistory(period, timeframe)
		if err != nil {
			logger.Error("failed to get portfolio history", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).
				Encode(map[string]string{"error": "Failed to retrieve portfolio history"})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(history)
	}
}

type PlaceOrderRequestPayload struct {
	Symbol   string  `json:"symbol"`
	Quantity float64 `json:"quantity"`
	Side     string  `json:"side"` // "buy" or "sell"
	Type     string  `json:"type"` // "market" or "limit"
	Limit    float64 `json:"limit,omitempty"`
}

func PlaceManualOrder(logger *slog.Logger, b *broker.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req PlaceOrderRequestPayload
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
			return
		}

		if req.Symbol == "" || req.Quantity <= 0 || req.Side == "" || req.Type == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "missing required fields"})
			return
		}

		qty := decimal.NewFromFloat(req.Quantity)
		alpacaReq := alpaca.PlaceOrderRequest{
			Symbol:      strings.ToUpper(req.Symbol),
			Qty:         &qty,
			Side:        alpaca.Side(strings.ToLower(req.Side)),
			Type:        alpaca.OrderType(strings.ToLower(req.Type)),
			TimeInForce: alpaca.Day,
		}

		if req.Type == "limit" && req.Limit > 0 {
			limitPrice := decimal.NewFromFloat(req.Limit)
			alpacaReq.LimitPrice = &limitPrice
		}

		order, err := b.PlaceOrder(alpacaReq)
		if err != nil {
			logger.Error("failed to place manual order", "error", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(order)
	}
}

func ResumeLiveEngines(ctx context.Context, logger *slog.Logger, b *broker.Client, db *sqlx.DB) {
	sessions, err := database.ListTradingSessions(ctx, db)
	if err != nil {
		logger.Error("failed to list trading sessions for resume", "error", err)
		return
	}

	for _, session := range sessions {
		if session.Status != "running" {
			continue
		}

		var symbols []string
		if err := json.Unmarshal([]byte(session.Symbols), &symbols); err != nil {
			logger.Error("failed to parse symbols for session", "error", err, "session", session.SessionID)
			continue
		}

		var params map[string]interface{}
		if session.Parameters != "" {
			if err := json.Unmarshal([]byte(session.Parameters), &params); err != nil {
				logger.Error("failed to parse parameters for session", "error", err, "session", session.SessionID)
				continue
			}
		}

		var strategy quant.Strategy
		strategyID := strings.ToLower(strings.ReplaceAll(session.Strategy, " ", "_"))
		switch strategyID {
		case "sma_crossover":
			shortPeriod := 10
			longPeriod := 50
			if params != nil {
				if v, ok := params["short_period"].(float64); ok {
					shortPeriod = int(v)
				}
				if v, ok := params["long_period"].(float64); ok {
					longPeriod = int(v)
				}
			}
			strategy = strategies.NewSMACrossover(symbols[0], shortPeriod, longPeriod)
		case "rsi_reversion":
			period := 14
			oversold := 30.0
			overbought := 70.0
			if params != nil {
				if v, ok := params["period"].(float64); ok {
					period = int(v)
				}
				if v, ok := params["oversold"].(float64); ok {
					oversold = v
				}
				if v, ok := params["overbought"].(float64); ok {
					overbought = v
				}
			}
			strategy = strategies.NewRSIReversion(symbols[0], period, oversold, overbought)
		case "bollinger_bands":
			period := 20
			stdDev := 2.0
			if params != nil {
				if v, ok := params["period"].(float64); ok {
					period = int(v)
				}
				if v, ok := params["std_dev"].(float64); ok {
					stdDev = v
				}
			}
			strategy = strategies.NewBollingerBands(symbols[0], period, stdDev)
		case "pairs_trading":
			period := 20
			entryZ := 2.0
			exitZ := 0.0
			if params != nil {
				if v, ok := params["period"].(float64); ok {
					period = int(v)
				}
				if v, ok := params["entry_z"].(float64); ok {
					entryZ = v
				}
				if v, ok := params["exit_z"].(float64); ok {
					exitZ = v
				}
			}
			if len(symbols) >= 2 {
				strategy = strategies.NewPairsTrading(symbols[0], symbols[1], period, entryZ, exitZ)
			} else {
				logger.Error("pairs_trading requires 2 symbols", "session", session.SessionID)
				continue
			}
		default:
			logger.Error("unknown strategy", "strategy", session.Strategy)
			continue
		}

		maxPosPct := 0.05
		dailyStopPct := 0.02
		if params != nil {
			if v, ok := params["max_position_size_pct"].(float64); ok {
				maxPosPct = v
			}
			if v, ok := params["daily_stop_loss_pct"].(float64); ok {
				dailyStopPct = v
			}
		}

		rm := quant.NewDefaultRiskManager(maxPosPct, dailyStopPct)
		engine := quant.NewLiveEngine(db, b, session.StartingCapital, strategy, symbols, params, rm)
		engine.SessionID = session.SessionID

		if err := engine.Resume(ctx, logger); err != nil {
			logger.Error("failed to resume live engine", "error", err, "session", session.SessionID)
			continue
		}

		enginesMutex.Lock()
		activeEngines[engine.SessionID] = engine
		enginesMutex.Unlock()

		logger.Info("successfully resumed live engine", "session", session.SessionID)
	}
}
