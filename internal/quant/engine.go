package quant

import (
	"fmt"
	"sort"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
)

type Engine struct {
	Portfolio   *Portfolio
	Strategy    Strategy
	RiskManager RiskManager
}

func NewEngine(startingCash float64, strategy Strategy, rm RiskManager) *Engine {
	if rm == nil {
		rm = NewDefaultRiskManager(0, 0)
	}
	return &Engine{
		Portfolio:   NewPortfolio(startingCash),
		Strategy:    strategy,
		RiskManager: rm,
	}
}

type symbolBar struct {
	Symbol string
	Bar    marketdata.Bar
}

func (e *Engine) Run(barsMap map[string][]marketdata.Bar) error {
	if len(barsMap) == 0 {
		return fmt.Errorf("no bars provided for backtest")
	}

	e.Strategy.Initialize(e.Portfolio)

	var allBars []symbolBar
	for symbol, bars := range barsMap {
		for _, bar := range bars {
			allBars = append(allBars, symbolBar{Symbol: symbol, Bar: bar})
		}
	}

	sort.Slice(allBars, func(i, j int) bool {
		return allBars[i].Bar.Timestamp.Before(allBars[j].Bar.Timestamp)
	})

	prices := make(map[string]float64)

	var lastTimestamp *time.Time

	for _, sb := range allBars {
		// Update current price before making decisions
		prices[sb.Symbol] = sb.Bar.Close

		e.Strategy.OnBar(sb.Symbol, sb.Bar, e.Portfolio)

		e.RiskManager.EvaluatePortfolio(e.Portfolio, prices)

		var approvedOrders []OrderIntent
		for _, intent := range e.Portfolio.PendingOrders {
			approved, err := e.RiskManager.EvaluateOrder(intent, e.Portfolio, prices)
			if err == nil {
				approvedOrders = append(approvedOrders, approved)
			}
		}

		for _, intent := range approvedOrders {
			if intent.Side == Buy {
				e.Portfolio.Buy(
					intent.Symbol,
					intent.Quantity,
					prices[intent.Symbol],
					sb.Bar.Timestamp,
				)
			} else if intent.Side == Sell {
				e.Portfolio.Sell(intent.Symbol, intent.Quantity, prices[intent.Symbol], sb.Bar.Timestamp)
			}
		}
		e.Portfolio.PendingOrders = []OrderIntent{}

		// Record equity at most once per timestamp to avoid duplication
		if lastTimestamp == nil || !lastTimestamp.Equal(sb.Bar.Timestamp) {
			equity := e.Portfolio.CalculateEquity(prices)
			e.Portfolio.EquityLog = append(e.Portfolio.EquityLog, EquitySnapshot{
				Timestamp: sb.Bar.Timestamp,
				Equity:    equity,
			})
			t := sb.Bar.Timestamp
			lastTimestamp = &t
		}
	}

	return nil
}
