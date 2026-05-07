package strategies

import (
	"citadel/internal/quant"
	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
)

type RSIReversion struct {
	Symbol     string
	Period     int
	Oversold   float64
	Overbought float64

	history []float64
	avgGain float64
	avgLoss float64
}

func NewRSIReversion(symbol string, period int, oversold, overbought float64) *RSIReversion {
	return &RSIReversion{
		Symbol:     symbol,
		Period:     period,
		Oversold:   oversold,
		Overbought: overbought,
		history:    make([]float64, 0),
	}
}

func (s *RSIReversion) Name() string {
	return "RSI Reversion"
}

func (s *RSIReversion) Initialize(p *quant.Portfolio) {
	s.history = make([]float64, 0)
	s.avgGain = 0
	s.avgLoss = 0
}

func (s *RSIReversion) OnBar(symbol string, bar marketdata.Bar, p *quant.Portfolio) {
	if symbol != s.Symbol {
		return
	}

	s.history = append(s.history, bar.Close)

	if len(s.history) <= s.Period {
		if len(s.history) > 1 {
			diff := bar.Close - s.history[len(s.history)-2]
			if diff > 0 {
				s.avgGain += diff
			} else {
				s.avgLoss -= diff
			}
		}
		if len(s.history) == s.Period+1 {
			s.avgGain /= float64(s.Period)
			s.avgLoss /= float64(s.Period)
		}
		return
	}

	diff := bar.Close - s.history[len(s.history)-2]
	gain := 0.0
	loss := 0.0
	if diff > 0 {
		gain = diff
	} else {
		loss = -diff
	}

	s.avgGain = (s.avgGain*float64(s.Period-1) + gain) / float64(s.Period)
	s.avgLoss = (s.avgLoss*float64(s.Period-1) + loss) / float64(s.Period)

	var rsi float64
	if s.avgLoss == 0 {
		rsi = 100
	} else {
		rs := s.avgGain / s.avgLoss
		rsi = 100 - (100 / (1 + rs))
	}

	currentPosition := p.Positions[s.Symbol]

	if rsi < s.Oversold && currentPosition == 0 {
		qty := p.Cash / bar.Close
		if qty > 0 {
			p.Buy(s.Symbol, qty, bar.Close, bar.Timestamp)
		}
	} else if rsi > s.Overbought && currentPosition > 0 {
		p.Sell(s.Symbol, currentPosition, bar.Close, bar.Timestamp)
	}
}
