package strategies

import (
	"math"

	"citadel/internal/quant"

	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
)

type BollingerBands struct {
	Symbol  string
	Period  int
	StdDev  float64
	history []float64
}

func NewBollingerBands(symbol string, period int, stdDev float64) *BollingerBands {
	return &BollingerBands{
		Symbol:  symbol,
		Period:  period,
		StdDev:  stdDev,
		history: make([]float64, 0),
	}
}

func (s *BollingerBands) Name() string {
	return "Bollinger Bands"
}

func (s *BollingerBands) Initialize(p *quant.Portfolio) {
	s.history = make([]float64, 0)
}

func (s *BollingerBands) OnBar(symbol string, bar marketdata.Bar, p *quant.Portfolio) {
	if symbol != s.Symbol {
		return
	}

	s.history = append(s.history, bar.Close)

	if len(s.history) < s.Period {
		return
	}

	window := s.history[len(s.history)-s.Period:]
	sum := 0.0
	for _, val := range window {
		sum += val
	}
	sma := sum / float64(s.Period)

	varianceSum := 0.0
	for _, val := range window {
		varianceSum += math.Pow(val-sma, 2)
	}
	stdDev := math.Sqrt(varianceSum / float64(s.Period))

	lowerBand := sma - (s.StdDev * stdDev)
	upperBand := sma + (s.StdDev * stdDev)

	currentPosition := p.Positions[s.Symbol]

	if bar.Close < lowerBand && currentPosition == 0 {
		qty := p.Cash / bar.Close
		if qty > 0 {
			p.Buy(s.Symbol, qty, bar.Close, bar.Timestamp)
		}
	} else if bar.Close > upperBand && currentPosition > 0 {
		p.Sell(s.Symbol, currentPosition, bar.Close, bar.Timestamp)
	}
}
