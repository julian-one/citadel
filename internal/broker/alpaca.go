package broker

import (
	"context"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata/stream"
)

type Client struct {
	alpaca     *alpaca.Client
	marketdata *marketdata.Client
	Stream     *stream.StocksClient

	assetsCache []alpaca.Asset
	cacheMutex  sync.RWMutex
}

func New(key, secret, endpoint string) *Client {
	endpoint = strings.TrimSuffix(endpoint, "/v2")
	streamClient := stream.NewStocksClient(
		marketdata.IEX,
		stream.WithCredentials(key, secret),
	)

	c := &Client{
		alpaca: alpaca.NewClient(alpaca.ClientOpts{
			APIKey:    key,
			APISecret: secret,
			BaseURL:   endpoint,
		}),
		marketdata: marketdata.NewClient(marketdata.ClientOpts{
			APIKey:    key,
			APISecret: secret,
		}),
		Stream: streamClient,
	}

	go func() {
		err := c.Stream.Connect(context.Background())
		if err != nil {
			slog.Error("failed to connect alpaca stream", "error", err)
		}
	}()

	return c
}

func (c *Client) GetAccount() (*alpaca.Account, error) {
	return c.alpaca.GetAccount()
}

func (c *Client) GetHistoricalBars(symbol string, start, end time.Time) ([]marketdata.Bar, error) {
	req := marketdata.GetBarsRequest{
		TimeFrame: marketdata.OneDay,
		Start:     start,
		End:       end,
		Feed:      marketdata.IEX,
	}
	return c.marketdata.GetBars(symbol, req)
}

// SearchAssets returns up to 20 active US equities matching the query
func (c *Client) SearchAssets(query string) ([]alpaca.Asset, error) {
	c.cacheMutex.RLock()
	needsLoad := len(c.assetsCache) == 0
	c.cacheMutex.RUnlock()

	if needsLoad {
		c.cacheMutex.Lock()
		if len(c.assetsCache) == 0 { // double check pattern
			assets, err := c.alpaca.GetAssets(alpaca.GetAssetsRequest{
				Status:     "active",
				AssetClass: "us_equity",
			})
			if err != nil {
				c.cacheMutex.Unlock()
				return nil, err
			}
			c.assetsCache = assets
		}
		c.cacheMutex.Unlock()
	}

	c.cacheMutex.RLock()
	defer c.cacheMutex.RUnlock()

	var results []alpaca.Asset
	query = strings.ToLower(strings.TrimSpace(query))

	// If query is very short, just return the first few
	if len(query) == 0 {
		if len(c.assetsCache) > 20 {
			return c.assetsCache[:20], nil
		}
		return c.assetsCache, nil
	}

	for _, a := range c.assetsCache {
		if strings.Contains(strings.ToLower(a.Symbol), query) || strings.Contains(strings.ToLower(a.Name), query) {
			results = append(results, a)
			if len(results) >= 20 {
				break
			}
		}
	}

	return results, nil
}
