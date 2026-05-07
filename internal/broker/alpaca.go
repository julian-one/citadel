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

	loadAssets func() ([]alpaca.Asset, error)
}

func New(ctx context.Context, key, secret, endpoint string) *Client {
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

	c.loadAssets = sync.OnceValues(func() ([]alpaca.Asset, error) {
		return c.alpaca.GetAssets(alpaca.GetAssetsRequest{
			Status:     "active",
			AssetClass: "us_equity",
		})
	})

	err := c.Stream.Connect(ctx)
	if err != nil {
		slog.Error("failed to connect alpaca stream", "error", err)
	}

	return c
}

func (c *Client) GetAccount() (*alpaca.Account, error) {
	return c.alpaca.GetAccount()
}

func (c *Client) GetPositions() ([]alpaca.Position, error) {
	return c.alpaca.GetPositions()
}

func (c *Client) GetPortfolioHistory(
	period string,
	timeframe string,
) (*alpaca.PortfolioHistory, error) {
	req := alpaca.GetPortfolioHistoryRequest{
		Period:    period,
		TimeFrame: alpaca.TimeFrame(timeframe),
	}
	return c.alpaca.GetPortfolioHistory(req)
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
	assetsCache, err := c.loadAssets()
	if err != nil {
		return nil, err
	}

	var results []alpaca.Asset
	query = strings.ToLower(strings.TrimSpace(query))

	// If query is very short, just return the first few
	if len(query) == 0 {
		if len(assetsCache) > 20 {
			return assetsCache[:20], nil
		}
		return assetsCache, nil
	}

	for _, a := range assetsCache {
		if strings.Contains(strings.ToLower(a.Symbol), query) ||
			strings.Contains(strings.ToLower(a.Name), query) {
			results = append(results, a)
			if len(results) >= 20 {
				break
			}
		}
	}

	return results, nil
}

func (c *Client) PlaceOrder(req alpaca.PlaceOrderRequest) (*alpaca.Order, error) {
	return c.alpaca.PlaceOrder(req)
}

func (c *Client) CancelOrder(orderID string) error {
	return c.alpaca.CancelOrder(orderID)
}

func (c *Client) ConnectTradeUpdates(ctx context.Context, handler func(alpaca.TradeUpdate)) error {
	return c.alpaca.StreamTradeUpdates(ctx, handler, alpaca.StreamTradeUpdatesRequest{})
}
