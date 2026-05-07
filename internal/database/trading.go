package database

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

type TradingSession struct {
	SessionID       string     `db:"session_id"       json:"session_id"`
	Strategy        string     `db:"strategy"         json:"strategy"`
	Status          string     `db:"status"           json:"status"`
	Symbols         string     `db:"symbols"          json:"symbols"`
	StartingCapital float64    `db:"starting_capital" json:"starting_capital"`
	Parameters      string     `db:"parameters"       json:"parameters"`
	StartedAt       time.Time  `db:"started_at"       json:"started_at"`
	EndedAt         *time.Time `db:"ended_at"         json:"ended_at"`
}

type TradingOrder struct {
	OrderID       string    `db:"order_id"        json:"order_id"`
	SessionID     string    `db:"session_id"      json:"session_id"`
	ClientOrderID *string   `db:"client_order_id" json:"client_order_id"`
	Symbol        string    `db:"symbol"          json:"symbol"`
	Side          string    `db:"side"            json:"side"`
	Type          string    `db:"type"            json:"type"`
	Qty           float64   `db:"qty"             json:"qty"`
	FilledQty     float64   `db:"filled_qty"      json:"filled_qty"`
	AvgPrice      *float64  `db:"avg_price"       json:"avg_price"`
	Status        string    `db:"status"          json:"status"`
	CreatedAt     time.Time `db:"created_at"      json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"      json:"updated_at"`
}

func CreateTradingSession(ctx context.Context, db *sqlx.DB, session *TradingSession) error {
	query, args, err := QB.Insert("trading_sessions").
		Columns("session_id", "strategy", "status", "symbols", "starting_capital", "parameters", "started_at").
		Values(session.SessionID, session.Strategy, session.Status, session.Symbols, session.StartingCapital, session.Parameters, session.StartedAt).
		ToSql()
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, query, args...)
	return err
}

func UpdateTradingSessionStatus(ctx context.Context, db *sqlx.DB, sessionID, status string) error {
	query, args, err := QB.Update("trading_sessions").
		Set("status", status).
		Set("ended_at", time.Now()).
		Where(sq.Eq{"session_id": sessionID}).
		ToSql()
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, query, args...)
	return err
}

func InsertTradingOrder(ctx context.Context, db *sqlx.DB, order *TradingOrder) error {
	query, args, err := QB.Insert("trading_orders").
		Columns(
			"order_id", "session_id", "client_order_id", "symbol", "side", "type",
			"qty", "filled_qty", "avg_price", "status", "created_at", "updated_at",
		).
		Values(
			order.OrderID, order.SessionID, order.ClientOrderID, order.Symbol, order.Side, order.Type,
			order.Qty, order.FilledQty, order.AvgPrice, order.Status, order.CreatedAt, order.UpdatedAt,
		).
		ToSql()
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, query, args...)
	return err
}

func UpdateTradingOrder(
	ctx context.Context,
	db *sqlx.DB,
	orderID string,
	filledQty float64,
	avgPrice float64,
	status string,
) error {
	query, args, err := QB.Update("trading_orders").
		Set("filled_qty", filledQty).
		Set("avg_price", avgPrice).
		Set("status", status).
		Set("updated_at", time.Now()).
		Where(sq.Eq{"order_id": orderID}).
		ToSql()
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, query, args...)
	return err
}

func ListTradingSessions(ctx context.Context, db *sqlx.DB) ([]TradingSession, error) {
	query, args, err := QB.Select("*").
		From("trading_sessions").
		OrderBy("started_at DESC").
		ToSql()
	if err != nil {
		return nil, err
	}

	var sessions []TradingSession
	err = db.SelectContext(ctx, &sessions, query, args...)
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

func GetTradingSession(
	ctx context.Context,
	db *sqlx.DB,
	sessionID string,
) (*TradingSession, error) {
	query, args, err := QB.Select("*").
		From("trading_sessions").
		Where(sq.Eq{"session_id": sessionID}).
		ToSql()
	if err != nil {
		return nil, err
	}

	var session TradingSession
	err = db.GetContext(ctx, &session, query, args...)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func GetTradingSessionOrders(
	ctx context.Context,
	db *sqlx.DB,
	sessionID string,
) ([]TradingOrder, error) {
	query, args, err := QB.Select("*").
		From("trading_orders").
		Where(sq.Eq{"session_id": sessionID}).
		OrderBy("created_at DESC").
		ToSql()
	if err != nil {
		return nil, err
	}

	var orders []TradingOrder
	err = db.SelectContext(ctx, &orders, query, args...)
	if err != nil {
		return nil, err
	}
	return orders, nil
}
