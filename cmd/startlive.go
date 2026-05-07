package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var startLiveCmd = &cobra.Command{
	Use:   "startlive",
	Short: "Start 10 live trading sessions via the API",
	Long:  `The startlive command logs into the local Citadel server and deploys the 10 recommended strategies.`,
	RunE:  runStartLive,
}

func init() {
	rootCmd.AddCommand(startLiveCmd)

	startLiveCmd.Flags().String("api-url", "http://localhost:8080", "URL of the Citadel server")
	startLiveCmd.Flags().String("username", "admin", "Admin username")
	startLiveCmd.Flags().String("password", "password", "Admin password")

	_ = viper.BindPFlag("startlive.api_url", startLiveCmd.Flags().Lookup("api-url"))
	_ = viper.BindPFlag("startlive.username", startLiveCmd.Flags().Lookup("username"))
	_ = viper.BindPFlag("startlive.password", startLiveCmd.Flags().Lookup("password"))
}

func runStartLive(cmd *cobra.Command, args []string) error {
	apiUrl := viper.GetString("startlive.api_url")
	username := viper.GetString("startlive.username")
	password := viper.GetString("startlive.password")

	jar, err := cookiejar.New(nil)
	if err != nil {
		return fmt.Errorf("failed to create cookie jar: %w", err)
	}

	client := &http.Client{
		Jar: jar,
	}

	// 1. Login
	slog.Info("logging into Citadel API", "url", apiUrl, "username", username)
	req, err := http.NewRequest("POST", apiUrl+"/login", nil)
	if err != nil {
		return fmt.Errorf("failed to create login request: %w", err)
	}
	req.SetBasicAuth(username, password)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(b))
	}
	slog.Info("login successful")

	// 2. Start Live Sessions
	type startReq struct {
		Symbols         []string               `json:"symbols"`
		Strategy        string                 `json:"strategy"`
		StartingCapital float64                `json:"starting_capital"`
		Parameters      map[string]interface{} `json:"parameters"`
	}

	strategies := []startReq{
		{
			Symbols:         []string{"AAPL"},
			Strategy:        "rsi_reversion",
			StartingCapital: 1000,
			Parameters:      map[string]interface{}{"period": 21, "oversold": 25, "overbought": 75},
		},
		{
			Symbols:         []string{"QQQ"},
			Strategy:        "rsi_reversion",
			StartingCapital: 1000,
			Parameters:      map[string]interface{}{"period": 21, "oversold": 30, "overbought": 75},
		},
		{
			Symbols:         []string{"SPY"},
			Strategy:        "rsi_reversion",
			StartingCapital: 1000,
			Parameters:      map[string]interface{}{"period": 21, "oversold": 30, "overbought": 75},
		},
		{
			Symbols:         []string{"MSFT"},
			Strategy:        "rsi_reversion",
			StartingCapital: 1000,
			Parameters:      map[string]interface{}{"period": 14, "oversold": 30, "overbought": 75},
		},
		{
			Symbols:         []string{"TSLA"},
			Strategy:        "bollinger_bands",
			StartingCapital: 1000,
			Parameters:      map[string]interface{}{"period": 20, "std_dev": 2.5},
		},
		{
			Symbols:         []string{"NVDA"},
			Strategy:        "sma_crossover",
			StartingCapital: 1000,
			Parameters:      map[string]interface{}{"short_period": 10, "long_period": 50},
		},
		{
			Symbols:         []string{"JPM"},
			Strategy:        "rsi_reversion",
			StartingCapital: 1000,
			Parameters:      map[string]interface{}{"period": 14, "oversold": 30, "overbought": 70},
		},
		{
			Symbols:         []string{"GLD"},
			Strategy:        "sma_crossover",
			StartingCapital: 1000,
			Parameters:      map[string]interface{}{"short_period": 20, "long_period": 100},
		},
		{
			Symbols:         []string{"AMD", "INTC"},
			Strategy:        "pairs_trading",
			StartingCapital: 1000,
			Parameters:      map[string]interface{}{"period": 50, "entry_z": 2.0, "exit_z": 0.0},
		},
		{
			Symbols:         []string{"TLT"},
			Strategy:        "bollinger_bands",
			StartingCapital: 1000,
			Parameters:      map[string]interface{}{"period": 50, "std_dev": 2.0},
		},
	}

	for _, s := range strategies {
		slog.Info("starting strategy", "symbols", s.Symbols, "strategy", s.Strategy)
		b, err := json.Marshal(s)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}

		req, err := http.NewRequest("POST", apiUrl+"/trading/live/start", bytes.NewBuffer(b))
		if err != nil {
			return fmt.Errorf("failed to create start request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			slog.Error("failed to start live session", "error", err, "symbols", s.Symbols)
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			slog.Error(
				"failed to start live session",
				"status",
				resp.StatusCode,
				"body",
				string(body),
				"symbols",
				s.Symbols,
			)
		} else {
			slog.Info("successfully started live session", "symbols", s.Symbols)
		}
	}

	slog.Info("all live sessions started")
	return nil
}
