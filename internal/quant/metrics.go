package quant

import (
	"math"
)

type Metrics struct {
	TotalReturn  float64 `json:"total_return"`
	MaxDrawdown  float64 `json:"max_drawdown"`
	SharpeRatio  float64 `json:"sharpe_ratio"`
	WinRate      float64 `json:"win_rate"`
	TotalTrades  int     `json:"total_trades"`
	ProfitFactor float64 `json:"profit_factor"`
}

func (p *Portfolio) CalculateMetrics() {
	if len(p.EquityLog) < 2 {
		p.Metrics = Metrics{}
		return
	}

	startingEquity := p.EquityLog[0].Equity
	endingEquity := p.EquityLog[len(p.EquityLog)-1].Equity

	// 1. Total Return
	totalReturn := (endingEquity - startingEquity) / startingEquity

	// 2. Max Drawdown
	peak := startingEquity
	maxDrawdown := 0.0

	var returns []float64

	for i := 1; i < len(p.EquityLog); i++ {
		current := p.EquityLog[i].Equity
		prev := p.EquityLog[i-1].Equity

		// Update peak
		if current > peak {
			peak = current
		}

		// Calculate drawdown from peak
		drawdown := (peak - current) / peak
		if drawdown > maxDrawdown {
			maxDrawdown = drawdown
		}

		// Calculate daily return for Sharpe
		dailyReturn := (current - prev) / prev
		returns = append(returns, dailyReturn)
	}

	// 3. Sharpe Ratio (Annualized, assuming 252 trading days and 0% risk-free rate)
	meanReturn := 0.0
	for _, r := range returns {
		meanReturn += r
	}
	meanReturn /= float64(len(returns))

	variance := 0.0
	for _, r := range returns {
		variance += math.Pow(r-meanReturn, 2)
	}
	variance /= float64(len(returns))

	stdDev := math.Sqrt(variance)

	sharpeRatio := 0.0
	if stdDev > 0 {
		sharpeRatio = (meanReturn / stdDev) * math.Sqrt(252.0)
	}

	// 4. Additional Metrics (Win Rate, Total Trades, Profit Factor based on daily returns)
	winningDays := 0
	grossProfit := 0.0
	grossLoss := 0.0

	for _, r := range returns {
		if r > 0 {
			winningDays++
			grossProfit += r
		} else if r < 0 {
			grossLoss += math.Abs(r)
		}
	}

	winRate := 0.0
	if len(returns) > 0 {
		winRate = float64(winningDays) / float64(len(returns))
	}

	profitFactor := 0.0
	if grossLoss > 0 {
		profitFactor = grossProfit / grossLoss
	} else if grossProfit > 0 {
		profitFactor = 999.0 // arbitrarily high
	}

	totalTrades := len(p.Trades)

	p.Metrics = Metrics{
		TotalReturn:  totalReturn,
		MaxDrawdown:  maxDrawdown,
		SharpeRatio:  sharpeRatio,
		WinRate:      winRate,
		TotalTrades:  totalTrades,
		ProfitFactor: profitFactor,
	}
}
