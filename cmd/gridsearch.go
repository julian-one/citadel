package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"citadel/internal/broker"
	"citadel/internal/database"
	"citadel/internal/quant"
	"citadel/internal/quant/strategies"
	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var gridSearchCmd = &cobra.Command{
	Use:   "gridsearch",
	Short: "Run grid search for backtest strategies",
	Long:  `The gridsearch command runs various strategy permutations and saves them to the backtests table.`,
	RunE:  runGridSearch,
}

func init() {
	rootCmd.AddCommand(gridSearchCmd)

	gridSearchCmd.Flags().String("db-path", "./citadel.db", "path to the SQLite database")
	gridSearchCmd.Flags().
		String("db-schema", "./schema/model.sql", "path to the database schema file")

	_ = viper.BindPFlag("database.path", gridSearchCmd.Flags().Lookup("db-path"))
	_ = viper.BindPFlag("database.schema", gridSearchCmd.Flags().Lookup("db-schema"))
}

func runGridSearch(cmd *cobra.Command, args []string) error {
	dbPath := viper.GetString("database.path")
	dbSchema := viper.GetString("database.schema")

	db, err := database.New(dbPath, dbSchema)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	b := broker.New(
		context.Background(),
		viper.GetString("alpaca.key"),
		viper.GetString("alpaca.secret"),
		viper.GetString("alpaca.endpoint"),
	)

	symbols := []string{"AAPL", "MSFT", "SPY", "QQQ"}
	end := time.Now()
	start := end.AddDate(-2, 0, 0)
	startingCapital := 100000.0

	slog.Info("fetching historical data for grid search", "symbols", symbols)

	barsMap := make(map[string][]marketdata.Bar)
	for _, sym := range symbols {
		bars, err := b.GetHistoricalBars(sym, start, end)
		if err != nil {
			return fmt.Errorf("failed to fetch market data for %s: %w", sym, err)
		}
		barsMap[sym] = bars
	}

	ctx := context.Background()

	// SMA Crossover
	smaShorts := []int{10, 20}
	smaLongs := []int{50, 100, 200}
	for _, short := range smaShorts {
		for _, long := range smaLongs {
			for _, sym := range symbols {
				runAndSave(
					ctx,
					db,
					barsMap,
					sym,
					start,
					end,
					startingCapital,
					"sma_crossover",
					map[string]interface{}{
						"short_period": short,
						"long_period":  long,
					},
				)
			}
		}
	}

	// RSI Reversion
	rsiPeriods := []int{14, 21}
	rsiOversolds := []float64{25, 30}
	rsiOverboughts := []float64{70, 75}
	for _, period := range rsiPeriods {
		for _, oversold := range rsiOversolds {
			for _, overbought := range rsiOverboughts {
				for _, sym := range symbols {
					runAndSave(
						ctx,
						db,
						barsMap,
						sym,
						start,
						end,
						startingCapital,
						"rsi_reversion",
						map[string]interface{}{
							"period":     period,
							"oversold":   oversold,
							"overbought": overbought,
						},
					)
				}
			}
		}
	}

	// Bollinger Bands
	bbPeriods := []int{20, 50}
	bbStdDevs := []float64{2.0, 2.5}
	for _, period := range bbPeriods {
		for _, stdDev := range bbStdDevs {
			for _, sym := range symbols {
				runAndSave(
					ctx,
					db,
					barsMap,
					sym,
					start,
					end,
					startingCapital,
					"bollinger_bands",
					map[string]interface{}{
						"period":  period,
						"std_dev": stdDev,
					},
				)
			}
		}
	}

	slog.Info("grid search completed successfully")
	return nil
}

func runAndSave(
	ctx context.Context,
	db *sqlx.DB,
	barsMap map[string][]marketdata.Bar,
	sym string,
	start, end time.Time,
	capital float64,
	strategyName string,
	params map[string]interface{},
) {
	var strategy quant.Strategy
	switch strategyName {
	case "sma_crossover":
		strategy = strategies.NewSMACrossover(
			sym,
			params["short_period"].(int),
			params["long_period"].(int),
		)
	case "rsi_reversion":
		strategy = strategies.NewRSIReversion(
			sym,
			params["period"].(int),
			params["oversold"].(float64),
			params["overbought"].(float64),
		)
	case "bollinger_bands":
		strategy = strategies.NewBollingerBands(
			sym,
			params["period"].(int),
			params["std_dev"].(float64),
		)
	}

	rm := quant.NewDefaultRiskManager(0.05, 0.02)
	engine := quant.NewEngine(capital, strategy, rm)

	// Only pass the needed symbol
	symMap := map[string][]marketdata.Bar{
		sym: barsMap[sym],
	}

	if err := engine.Run(symMap); err != nil {
		slog.Error("failed to run strategy", "strategy", strategyName, "symbol", sym, "error", err)
		return
	}

	engine.Portfolio.CalculateMetrics()

	paramsJSON, _ := json.Marshal(params)
	symbolsJSON, _ := json.Marshal([]string{sym})
	metricsJSON, _ := json.Marshal(engine.Portfolio.Metrics)

	record := database.BacktestRecord{
		BacktestID:      uuid.NewString(),
		Strategy:        strategyName,
		Symbols:         string(symbolsJSON),
		StartDate:       start.Format(time.RFC3339),
		EndDate:         end.Format(time.RFC3339),
		StartingCapital: capital,
		Parameters:      string(paramsJSON),
		Metrics:         string(metricsJSON),
	}

	err := database.SaveBacktest(ctx, db, record)
	if err != nil {
		slog.Error("failed to save backtest", "strategy", strategyName, "symbol", sym, "error", err)
	} else {
		slog.Info("saved backtest", "strategy", strategyName, "symbol", sym)
	}
}
