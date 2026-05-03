package quant

import (
	"time"
)

type TradeSide string

const (
	Buy  TradeSide = "buy"
	Sell TradeSide = "sell"
)

type Trade struct {
	Timestamp time.Time `json:"timestamp"`
	Symbol    string    `json:"symbol"`
	Side      TradeSide `json:"side"`
	Quantity  float64   `json:"quantity"`
	Price     float64   `json:"price"`
}

type Portfolio struct {
	Cash      float64            `json:"cash"`
	Positions map[string]float64 `json:"positions"`
	Trades    []Trade            `json:"trades"`
	EquityLog []EquitySnapshot   `json:"equity_log"`
	Metrics   Metrics            `json:"metrics"`
}

type EquitySnapshot struct {
	Timestamp time.Time `json:"timestamp"`
	Equity    float64   `json:"equity"`
}

func NewPortfolio(startingCash float64) *Portfolio {
	return &Portfolio{
		Cash:      startingCash,
		Positions: make(map[string]float64),
		Trades:    make([]Trade, 0),
		EquityLog: make([]EquitySnapshot, 0),
	}
}

func (p *Portfolio) Buy(symbol string, qty float64, price float64, ts time.Time) {
	cost := qty * price
	if p.Cash >= cost {
		p.Cash -= cost
		p.Positions[symbol] += qty
		p.Trades = append(p.Trades, Trade{
			Timestamp: ts,
			Symbol:    symbol,
			Side:      Buy,
			Quantity:  qty,
			Price:     price,
		})
	}
}

func (p *Portfolio) Sell(symbol string, qty float64, price float64, ts time.Time) {
	if p.Positions[symbol] >= qty {
		p.Positions[symbol] -= qty
		p.Cash += qty * price
		p.Trades = append(p.Trades, Trade{
			Timestamp: ts,
			Symbol:    symbol,
			Side:      Sell,
			Quantity:  qty,
			Price:     price,
		})
	}
}

func (p *Portfolio) CalculateEquity(prices map[string]float64) float64 {
	equity := p.Cash
	for symbol, qty := range p.Positions {
		if price, ok := prices[symbol]; ok {
			equity += qty * price
		}
	}
	return equity
}
