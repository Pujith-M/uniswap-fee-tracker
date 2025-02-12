package binance

import (
	"context"
	"fmt"
	"strconv"
	"time"
	"uniswap-fee-tracker/internal/utils"

	"github.com/go-resty/resty/v2"
	"golang.org/x/time/rate"
)

const (
	baseURL = "https://api.binance.com/api/v3"
)

// Client defines methods for interacting with Binance API
type Client interface {
	GetPrice(ctx context.Context, symbol string, timestamp time.Time) (*KlineData, error)
}

type client struct {
	httpClient *resty.Client
	limiter    *rate.Limiter
}

// NewClient creates a new Binance API client
func NewClient() Client {
	httpClient := resty.New().
		SetBaseURL(baseURL).
		SetTimeout(10 * time.Second).
		SetRetryCount(2).
		SetRetryWaitTime(1 * time.Second)

	// Binance has a limit of 1200 requests per minute
	// Setting it to 1000 to be safe
	binanceAllowedRequestsPerMinute := 1200
	limiter := rate.NewLimiter(rate.Limit(binanceAllowedRequestsPerMinute/60.0), binanceAllowedRequestsPerMinute)

	return &client{
		httpClient: httpClient,
		limiter:    limiter,
	}
}

// GetPrice fetches the price data for a given symbol at a specific timestamp
func (c *client) GetPrice(ctx context.Context, symbol string, timestamp time.Time) (*KlineData, error) {
	// Wait for rate limiter
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter wait: %w", err)
	}

	var klines [][]interface{}

	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetQueryParams(map[string]string{
			"symbol":    symbol,
			"interval":  "1s",
			"startTime": strconv.FormatInt(timestamp.UnixMilli(), 10),
			"limit":     "1",
		}).
		SetResult(&klines).
		Get("/klines")

	if err != nil || !resp.IsSuccess() || len(klines) == 0 {
		return nil, fmt.Errorf("failed to fetch price data: %v", err)
	}

	// Parse the kline data into a struct
	k := klines[0]
	return &KlineData{
		OpenTime:                 time.UnixMilli(utils.MustParseInt64(k[0])),
		Open:                     utils.MustParseBigFloat(k[1]),
		High:                     utils.MustParseBigFloat(k[2]),
		Low:                      utils.MustParseBigFloat(k[3]),
		Close:                    utils.MustParseBigFloat(k[4]),
		Volume:                   utils.MustParseBigFloat(k[5]),
		CloseTime:                time.UnixMilli(utils.MustParseInt64(k[6])),
		QuoteAssetVolume:         utils.MustParseBigFloat(k[7]),
		NumberOfTrades:           utils.MustParseInt64(k[8]),
		TakerBuyBaseAssetVolume:  utils.MustParseBigFloat(k[9]),
		TakerBuyQuoteAssetVolume: utils.MustParseBigFloat(k[10]),
	}, nil
}
