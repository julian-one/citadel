package quant

import "fmt"

type RiskManager interface {
	EvaluateOrder(
		intent OrderIntent,
		p *Portfolio,
		currentPrices map[string]float64,
	) (OrderIntent, error)
	EvaluatePortfolio(p *Portfolio, currentPrices map[string]float64) error
}

type DefaultRiskManager struct {
	MaxPositionSizePct float64
	DailyStopLossPct   float64

	peakEquity float64
	halted     bool
}

func NewDefaultRiskManager(maxPositionSizePct, dailyStopLossPct float64) *DefaultRiskManager {
	return &DefaultRiskManager{
		MaxPositionSizePct: maxPositionSizePct,
		DailyStopLossPct:   dailyStopLossPct,
	}
}

func (rm *DefaultRiskManager) EvaluateOrder(
	intent OrderIntent,
	p *Portfolio,
	currentPrices map[string]float64,
) (OrderIntent, error) {
	if rm.halted && intent.Side == Buy {
		return intent, fmt.Errorf("risk halt triggered: blocking new entries")
	}

	if intent.Side != Buy {
		return intent, nil
	}

	if rm.MaxPositionSizePct <= 0 {
		return intent, nil // No limit
	}

	equity := p.CalculateEquity(currentPrices)
	maxPositionDollar := equity * rm.MaxPositionSizePct

	currentPrice := currentPrices[intent.Symbol]
	if currentPrice <= 0 {
		return intent, nil // fallback if price unknown
	}

	currentPositionQty := p.Positions[intent.Symbol]
	currentPositionDollar := currentPositionQty * currentPrice

	requestedDollar := intent.Quantity * currentPrice
	newPositionDollar := currentPositionDollar + requestedDollar

	if newPositionDollar > maxPositionDollar {
		allowedDollar := maxPositionDollar - currentPositionDollar
		if allowedDollar <= 0 {
			return intent, fmt.Errorf("max position size reached for %s", intent.Symbol)
		}
		// Modify intent quantity to fit the limit
		allowedQty := allowedDollar / currentPrice
		intent.Quantity = allowedQty
	}

	return intent, nil
}

func (rm *DefaultRiskManager) EvaluatePortfolio(
	p *Portfolio,
	currentPrices map[string]float64,
) error {
	if rm.DailyStopLossPct <= 0 {
		return nil
	}

	equity := p.CalculateEquity(currentPrices)

	// Initialize peak equity
	if rm.peakEquity == 0 || equity > rm.peakEquity {
		rm.peakEquity = equity
	}

	drawdown := (rm.peakEquity - equity) / rm.peakEquity
	if drawdown >= rm.DailyStopLossPct {
		rm.halted = true
		return fmt.Errorf(
			"daily stop-loss triggered: drawdown %.2f%% >= %.2f%%",
			drawdown*100,
			rm.DailyStopLossPct*100,
		)
	}

	return nil
}
