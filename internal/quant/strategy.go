package quant

import (
	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
)

type Strategy interface {
	Name() string
	Initialize(p *Portfolio)
	OnBar(bar marketdata.Bar, p *Portfolio)
}
