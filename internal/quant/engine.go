package quant

import (
	"fmt"

	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
)

type Engine struct {
	Portfolio *Portfolio
	Strategy  Strategy
}

func NewEngine(startingCash float64, strategy Strategy) *Engine {
	return &Engine{
		Portfolio: NewPortfolio(startingCash),
		Strategy:  strategy,
	}
}

func (e *Engine) Run(bars []marketdata.Bar, symbol string) error {
	if len(bars) == 0 {
		return fmt.Errorf("no bars provided for backtest")
	}

	e.Strategy.Initialize(e.Portfolio)

	prices := make(map[string]float64)

	for _, bar := range bars {
		// Update current price before making decisions
		prices[symbol] = bar.Close

		e.Strategy.OnBar(bar, e.Portfolio)

		// Record the equity at the end of the bar
		equity := e.Portfolio.CalculateEquity(prices)
		e.Portfolio.EquityLog = append(e.Portfolio.EquityLog, EquitySnapshot{
			Timestamp: bar.Timestamp,
			Equity:    equity,
		})
	}

	return nil
}
