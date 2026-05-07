package strategies

import (
	"math"

	"citadel/internal/quant"

	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
)

type PairsTrading struct {
	SymbolA string
	SymbolB string
	Period  int
	EntryZ  float64
	ExitZ   float64

	historyA []float64
	historyB []float64

	latestA float64
	latestB float64
	hasA    bool
	hasB    bool
}

func NewPairsTrading(symbolA, symbolB string, period int, entryZ, exitZ float64) *PairsTrading {
	return &PairsTrading{
		SymbolA:  symbolA,
		SymbolB:  symbolB,
		Period:   period,
		EntryZ:   entryZ,
		ExitZ:    exitZ,
		historyA: make([]float64, 0),
		historyB: make([]float64, 0),
	}
}

func (s *PairsTrading) Name() string {
	return "Pairs Trading"
}

func (s *PairsTrading) Initialize(p *quant.Portfolio) {
	s.historyA = make([]float64, 0)
	s.historyB = make([]float64, 0)
	s.hasA = false
	s.hasB = false
}

func (s *PairsTrading) OnBar(symbol string, bar marketdata.Bar, p *quant.Portfolio) {
	if symbol == s.SymbolA {
		s.latestA = bar.Close
		s.hasA = true
	} else if symbol == s.SymbolB {
		s.latestB = bar.Close
		s.hasB = true
	} else {
		return
	}

	if !s.hasA || !s.hasB {
		return
	}

	s.historyA = append(s.historyA, s.latestA)
	s.historyB = append(s.historyB, s.latestB)

	// Consume the flags so we sync updates
	s.hasA = false
	s.hasB = false

	if len(s.historyA) < s.Period {
		return
	}

	var spreads []float64
	for i := len(s.historyA) - s.Period; i < len(s.historyA); i++ {
		spreads = append(spreads, s.historyA[i]-s.historyB[i])
	}

	sum := 0.0
	for _, val := range spreads {
		sum += val
	}
	mean := sum / float64(s.Period)

	varianceSum := 0.0
	for _, val := range spreads {
		varianceSum += math.Pow(val-mean, 2)
	}
	stdDev := math.Sqrt(varianceSum / float64(s.Period))

	if stdDev == 0 {
		return
	}

	currentSpread := s.latestA - s.latestB
	zScore := (currentSpread - mean) / stdDev

	posA := p.Positions[s.SymbolA]
	posB := p.Positions[s.SymbolB]

	if zScore < -s.EntryZ {
		// A is underpriced. Sell B, Buy A.
		if posB > 0 {
			p.Sell(s.SymbolB, posB, s.latestB, bar.Timestamp)
		}
		if posA == 0 {
			qty := p.Cash / s.latestA
			if qty > 0 {
				p.Buy(s.SymbolA, qty, s.latestA, bar.Timestamp)
			}
		}
	} else if zScore > s.EntryZ {
		// B is underpriced. Sell A, Buy B.
		if posA > 0 {
			p.Sell(s.SymbolA, posA, s.latestA, bar.Timestamp)
		}
		if posB == 0 {
			qty := p.Cash / s.latestB
			if qty > 0 {
				p.Buy(s.SymbolB, qty, s.latestB, bar.Timestamp)
			}
		}
	} else if math.Abs(zScore) < s.ExitZ {
		// Mean reversion happened, exit positions
		if posA > 0 {
			p.Sell(s.SymbolA, posA, s.latestA, bar.Timestamp)
		}
		if posB > 0 {
			p.Sell(s.SymbolB, posB, s.latestB, bar.Timestamp)
		}
	}
}
