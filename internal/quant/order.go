package quant

import "time"

type (
	OrderType   string
	OrderStatus string
)

const (
	Market       OrderType = "market"
	Limit        OrderType = "limit"
	Stop         OrderType = "stop"
	TrailingStop OrderType = "trailing_stop"

	StatusNew      OrderStatus = "new"
	StatusFilled   OrderStatus = "filled"
	StatusPartial  OrderStatus = "partially_filled"
	StatusRejected OrderStatus = "rejected"
	StatusCanceled OrderStatus = "canceled"
)

type OrderIntent struct {
	Symbol        string
	Side          TradeSide
	Type          OrderType
	Quantity      float64
	LimitPrice    *float64
	StopPrice     *float64
	TrailingPrice *float64
	TrailingPct   *float64
	TimeInForce   string // day, gtc, etc
}

type OrderRecord struct {
	ID        string      `json:"id"`
	ClientOID string      `json:"client_order_id"`
	Symbol    string      `json:"symbol"`
	Side      TradeSide   `json:"side"`
	Type      OrderType   `json:"type"`
	Qty       float64     `json:"qty"`
	FilledQty float64     `json:"filled_qty"`
	AvgPrice  float64     `json:"avg_price"`
	Status    OrderStatus `json:"status"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}
