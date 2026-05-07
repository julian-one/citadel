package strategies

import (
	"citadel/internal/quant"
	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
)

type SMACrossover struct {
	Symbol      string
	ShortPeriod int
	LongPeriod  int

	history []float64
}

func NewSMACrossover(symbol string, shortPeriod, longPeriod int) *SMACrossover {
	return &SMACrossover{
		Symbol:      symbol,
		ShortPeriod: shortPeriod,
		LongPeriod:  longPeriod,
		history:     make([]float64, 0),
	}
}

func (s *SMACrossover) Name() string {
	return "SMA Crossover"
}

func (s *SMACrossover) Initialize(p *quant.Portfolio) {
	s.history = make([]float64, 0)
}

func (s *SMACrossover) OnBar(symbol string, bar marketdata.Bar, p *quant.Portfolio) {
	if symbol != s.Symbol {
		return // Only process bars for the symbol we care about
	}

	s.history = append(s.history, bar.Close)

	if len(s.history) <= s.LongPeriod {
		return // Not enough data to calculate previous long SMA
	}

	// Calculate SMAs
	shortSum := 0.0
	for _, p := range s.history[len(s.history)-s.ShortPeriod:] {
		shortSum += p
	}
	shortSMA := shortSum / float64(s.ShortPeriod)

	longSum := 0.0
	for _, p := range s.history[len(s.history)-s.LongPeriod:] {
		longSum += p
	}
	longSMA := longSum / float64(s.LongPeriod)

	// Check previous bar to detect a crossover
	prevShortSum := 0.0
	for _, p := range s.history[len(s.history)-s.ShortPeriod-1 : len(s.history)-1] {
		prevShortSum += p
	}
	prevShortSMA := prevShortSum / float64(s.ShortPeriod)

	prevLongSum := 0.0
	for _, p := range s.history[len(s.history)-s.LongPeriod-1 : len(s.history)-1] {
		prevLongSum += p
	}
	prevLongSMA := prevLongSum / float64(s.LongPeriod)

	crossoverUp := prevShortSMA <= prevLongSMA && shortSMA > longSMA
	crossoverDown := prevShortSMA >= prevLongSMA && shortSMA < longSMA

	currentPosition := p.Positions[s.Symbol]

	if crossoverUp && currentPosition == 0 {
		// Buy! Let's go all in
		qty := p.Cash / bar.Close
		if qty > 0 {
			p.Buy(s.Symbol, qty, bar.Close, bar.Timestamp)
		}
	} else if crossoverDown && currentPosition > 0 {
		// Sell! Liquidate position
		p.Sell(s.Symbol, currentPosition, bar.Close, bar.Timestamp)
	}
}
