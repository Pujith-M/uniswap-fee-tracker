package etherscan

import (
	"context"
	"fmt"
	"strconv"
	"uniswap-fee-tracker/internal/config"

	"github.com/go-resty/resty/v2"
	"golang.org/x/time/rate"
)

// Client interface defines methods for interacting with Etherscan API
type Client interface {
	GetTokenTransfers(ctx context.Context, address string, startBlock, endBlock uint64) ([]TokenTransfer, error)
}

type client struct {
	cfg     *config.EtherscanConfig
	client  *resty.Client
	limiter *rate.Limiter
}

// NewClient creates a new Etherscan client with resty
func NewClient(cfg *config.EtherscanConfig) Client {
	restyClient := resty.New().
		SetBaseURL(cfg.HTTPClientConfig.BaseURL).
		SetTimeout(cfg.HTTPClientConfig.Timeout).
		SetRetryCount(cfg.HTTPClientConfig.RetryCount).
		SetRetryWaitTime(cfg.HTTPClientConfig.RetryWait).
		SetQueryParam("apikey", cfg.APIKey)

	// Create rate limiter using configured limit and burst
	limiter := rate.NewLimiter(rate.Limit(cfg.HTTPClientConfig.RateLimit), cfg.HTTPClientConfig.RateBurst)

	return &client{
		cfg:     cfg,
		client:  restyClient,
		limiter: limiter,
	}
}

func (c *client) GetTokenTransfers(ctx context.Context, address string, startBlock, endBlock uint64) ([]TokenTransfer, error) {
	// Wait for rate limiter
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter wait: %w", err)
	}

	var response EtherscanResponse[TokenTransfer]

	_, err := c.client.R().
		SetContext(ctx).
		SetQueryParams(map[string]string{
			"module":     "account",
			"action":     "tokentx",
			"address":    address,
			"sort":       "asc",
			"startblock": strconv.FormatUint(startBlock, 10),
			"endblock":   strconv.FormatUint(endBlock, 10),
		}).
		SetResult(&response).
		Get("")

	if err != nil {
		return nil, fmt.Errorf("failed to get token transfers: %w", err)
	}

	if response.Status != "1" {
		return nil, fmt.Errorf("api error: %s", response.Message)
	}

	return response.Result, nil
}
