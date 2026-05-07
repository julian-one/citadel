package database

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
)

type BacktestRecord struct {
	BacktestID      string    `json:"backtest_id"      db:"backtest_id"`
	Strategy        string    `json:"strategy"         db:"strategy"`
	Symbols         string    `json:"symbols"          db:"symbols"` // JSON string array
	StartDate       string    `json:"start_date"       db:"start_date"`
	EndDate         string    `json:"end_date"         db:"end_date"`
	StartingCapital float64   `json:"starting_capital" db:"starting_capital"`
	Parameters      string    `json:"parameters"       db:"parameters"` // JSON string
	Metrics         string    `json:"metrics"          db:"metrics"`    // JSON string
	CreatedAt       time.Time `json:"created_at"       db:"created_at"`
}

func SaveBacktest(ctx context.Context, db *sqlx.DB, record BacktestRecord) error {
	query := `
		INSERT INTO trading_backtests (
			backtest_id, strategy, symbols, start_date, end_date, starting_capital, parameters, metrics
		) VALUES (
			:backtest_id, :strategy, :symbols, :start_date, :end_date, :starting_capital, :parameters, :metrics
		)
	`
	_, err := db.NamedExecContext(ctx, query, record)
	return err
}

func ListBacktests(ctx context.Context, db *sqlx.DB) ([]BacktestRecord, error) {
	var records []BacktestRecord
	query := `SELECT * FROM trading_backtests ORDER BY created_at DESC`
	err := db.SelectContext(ctx, &records, query)
	return records, err
}
