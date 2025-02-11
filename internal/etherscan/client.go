package etherscan

import (
	"context"
	"fmt"
	"github.com/go-resty/resty/v2"
	"strconv"
	"uniswap-fee-tracker/internal/config"
)

// Client interface defines methods for interacting with Etherscan API
type Client interface {
	GetTokenTransfers(ctx context.Context, address string, startBlock, endBlock uint64) ([]TokenTransfer, error)
}

type client struct {
	cfg    *config.EtherscanConfig
	client *resty.Client
}

// NewClient creates a new Etherscan client with resty
func NewClient(cfg *config.EtherscanConfig) Client {
	restyClient := resty.New().
		SetBaseURL(cfg.BaseURL).
		SetTimeout(cfg.Timeout).
		SetRetryCount(cfg.RetryCount).
		SetRetryWaitTime(cfg.RetryWait).
		SetQueryParam("apikey", cfg.APIKey)

	return &client{
		cfg:    cfg,
		client: restyClient,
	}
}

func (c *client) GetTokenTransfers(ctx context.Context, address string, startBlock, endBlock uint64) ([]TokenTransfer, error) {
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
