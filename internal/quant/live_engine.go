package quant

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"citadel/internal/broker"
	"citadel/internal/database"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata/stream"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
)

type LiveEngine struct {
	Portfolio       *Portfolio
	Strategy        Strategy
	Broker          *broker.Client
	DB              *sqlx.DB
	SessionID       string
	Symbols         []string
	StartingCapital float64
	Parameters      string
	RiskManager     RiskManager
	latestPrices    map[string]float64
}

func NewLiveEngine(
	db *sqlx.DB,
	b *broker.Client,
	startingCash float64,
	strategy Strategy,
	symbols []string,
	parameters map[string]interface{},
	rm RiskManager,
) *LiveEngine {
	if rm == nil {
		rm = NewDefaultRiskManager(0, 0)
	}
	bParams, _ := json.Marshal(parameters)
	return &LiveEngine{
		Portfolio:       NewPortfolio(startingCash),
		Strategy:        strategy,
		Broker:          b,
		DB:              db,
		SessionID:       uuid.New().String(),
		Symbols:         symbols,
		StartingCapital: startingCash,
		Parameters:      string(bParams),
		RiskManager:     rm,
		latestPrices:    make(map[string]float64),
	}
}

// Ensure paper trading
func (e *LiveEngine) checkPaperTrading() error {
	account, err := e.Broker.GetAccount()
	if err != nil {
		return err
	}
	// Paper trading accounts usually have AccountNumber starting with "PA"
	if len(account.AccountNumber) < 2 || account.AccountNumber[:2] != "PA" {
		return fmt.Errorf(
			"LIVE TRADING NOT ALLOWED. Ensure you are using paper trading credentials",
		)
	}
	return nil
}

func (e *LiveEngine) Start(ctx context.Context, logger *slog.Logger) error {
	if err := e.checkPaperTrading(); err != nil {
		return err
	}

	bSymbols, _ := json.Marshal(e.Symbols)
	session := &database.TradingSession{
		SessionID:       e.SessionID,
		Strategy:        e.Strategy.Name(),
		Status:          "running",
		Symbols:         string(bSymbols),
		StartingCapital: e.StartingCapital,
		Parameters:      e.Parameters,
		StartedAt:       time.Now(),
	}
	if err := database.CreateTradingSession(ctx, e.DB, session); err != nil {
		return err
	}

	return e.startEngine(ctx, logger)
}

func (e *LiveEngine) Resume(ctx context.Context, logger *slog.Logger) error {
	if err := e.checkPaperTrading(); err != nil {
		return err
	}

	return e.startEngine(ctx, logger)
}

func (e *LiveEngine) startEngine(ctx context.Context, logger *slog.Logger) error {
	// Initialize Strategy
	e.Strategy.Initialize(e.Portfolio)

	// Subscribe to Market Data
	err := e.Broker.Stream.SubscribeToBars(func(bar stream.Bar) {
		logger.Info("received live bar", "symbol", bar.Symbol, "close", bar.Close)
		e.latestPrices[bar.Symbol] = bar.Close

		if err := e.RiskManager.EvaluatePortfolio(e.Portfolio, e.latestPrices); err != nil {
			logger.Warn("Risk Halt", "error", err)
			return // block new logic
		}

		e.Strategy.OnBar(bar.Symbol, marketdata.Bar{
			Timestamp:  bar.Timestamp,
			Open:       bar.Open,
			High:       bar.High,
			Low:        bar.Low,
			Close:      bar.Close,
			Volume:     bar.Volume,
			TradeCount: bar.TradeCount,
			VWAP:       bar.VWAP,
		}, e.Portfolio)
		e.processPendingOrders(ctx, logger)
	}, e.Symbols...)
	if err != nil {
		return err
	}

	// Subscribe to Trade Updates in a goroutine because it blocks
	go func() {
		err := e.Broker.ConnectTradeUpdates(ctx, func(update alpaca.TradeUpdate) {
			logger.Info("trade update", "event", update.Event, "order_id", update.Order.ID)
			e.handleTradeUpdate(ctx, logger, update)
		})
		if err != nil {
			logger.Error("failed to connect trade updates", "error", err, "session", e.SessionID)
		}
	}()

	logger.Info("Live Engine running", "session_id", e.SessionID)
	return nil
}

func (e *LiveEngine) Stop(ctx context.Context) {
	database.UpdateTradingSessionStatus(ctx, e.DB, e.SessionID, "stopped")
	if e.Broker != nil && e.Broker.Stream != nil {
		e.Broker.Stream.UnsubscribeFromBars(e.Symbols...)
	}
}

func (e *LiveEngine) processPendingOrders(ctx context.Context, logger *slog.Logger) {
	if len(e.Portfolio.PendingOrders) == 0 {
		return
	}

	var approvedOrders []OrderIntent
	for _, intent := range e.Portfolio.PendingOrders {
		approved, err := e.RiskManager.EvaluateOrder(intent, e.Portfolio, e.latestPrices)
		if err != nil {
			logger.Warn("order rejected by risk manager", "symbol", intent.Symbol, "error", err)
			continue
		}
		approvedOrders = append(approvedOrders, approved)
	}

	for _, intent := range approvedOrders {
		req := alpaca.PlaceOrderRequest{
			Symbol: intent.Symbol,
			Qty: func(q float64) *decimal.Decimal { d := decimal.NewFromFloat(q); return &d }(
				intent.Quantity,
			),
			Side:        alpaca.Side(intent.Side),
			Type:        alpaca.OrderType(intent.Type),
			TimeInForce: alpaca.Day,
		}

		if intent.LimitPrice != nil {
			d := decimal.NewFromFloat(*intent.LimitPrice)
			req.LimitPrice = &d
		}
		if intent.StopPrice != nil {
			d := decimal.NewFromFloat(*intent.StopPrice)
			req.StopPrice = &d
		}

		order, err := e.Broker.PlaceOrder(req)
		if err != nil {
			logger.Error("failed to place live order", "error", err, "symbol", intent.Symbol)
			continue
		}

		// Persist to DB
		dbOrder := &database.TradingOrder{
			OrderID:   order.ID,
			SessionID: e.SessionID,
			Symbol:    order.Symbol,
			Side:      string(order.Side),
			Type:      string(order.Type),
			Qty:       order.Qty.InexactFloat64(),
			Status:    "new",
			CreatedAt: order.CreatedAt,
			UpdatedAt: order.UpdatedAt,
		}
		if err := database.InsertTradingOrder(ctx, e.DB, dbOrder); err != nil {
			logger.Error("failed to save order to db", "error", err)
		}
	}

	// Clear pending
	e.Portfolio.PendingOrders = []OrderIntent{}
}

func (e *LiveEngine) handleTradeUpdate(
	ctx context.Context,
	logger *slog.Logger,
	update alpaca.TradeUpdate,
) {
	orderID := update.Order.ID
	status := string(update.Order.Status)

	filledQty := update.Order.FilledQty.InexactFloat64()
	avgPrice := 0.0

	if update.Order.FilledAvgPrice != nil {
		avgPrice = update.Order.FilledAvgPrice.InexactFloat64()
	}

	if err := database.UpdateTradingOrder(ctx, e.DB, orderID, filledQty, avgPrice, status); err != nil {
		logger.Error("failed to update order in db", "error", err)
	}

	// Log fills
	if update.Event == "fill" || update.Event == "partial_fill" {
		logger.Info(
			"order filled",
			"symbol",
			update.Order.Symbol,
			"qty",
			filledQty,
			"price",
			avgPrice,
		)
	}
}
