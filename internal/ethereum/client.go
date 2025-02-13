package ethereum

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"time"
	"uniswap-fee-tracker/internal/config"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"golang.org/x/time/rate"
)

// Client represents an Ethereum node client
type Client struct {
	cfg     *config.EthereumConfig
	client  *ethclient.Client
	httpURL string
	limiter *rate.Limiter
}

// NewClient creates a new Ethereum client
func NewClient(cfg *config.EthereumConfig) (*Client, error) {
	if cfg.InfuraAPIKey == "" {
		return nil, fmt.Errorf("infura API key cannot be empty")
	}

	httpURL := fmt.Sprintf("%s/%s", cfg.BaseURL, cfg.InfuraAPIKey)
	log.Printf("Initializing Ethereum client with endpoint: %s", httpURL)

	client, err := ethclient.Dial(httpURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum node: %w", err)
	}

	log.Println("Successfully connected to Ethereum node")

	// Create rate limiter using configuration
	limiter := rate.NewLimiter(rate.Limit(cfg.RateLimit), cfg.RateBurst)

	return &Client{
		cfg:     cfg,
		client:  client,
		httpURL: httpURL,
		limiter: limiter,
	}, nil
}

// retry executes a function with exponential backoff retry logic
func retry(attempts int, delay time.Duration, fn func() error) error {
	var err error
	for i := 0; i < attempts; i++ {
		err = fn()
		if err == nil {
			return nil
		}
		log.Printf("Retry %d/%d failed: %v", i+1, attempts, err)
		time.Sleep(delay * time.Duration(1<<uint(i)))
	}
	return err
}

// GetLatestBlockNumber returns the latest block number from the Ethereum network
func (c *Client) GetLatestBlockNumber(ctx context.Context) (uint64, error) {
	// Wait for rate limiter
	if err := c.limiter.Wait(ctx); err != nil {
		return 0, fmt.Errorf("rate limiter wait: %w", err)
	}

	var blockNumber uint64
	err := retry(c.cfg.RetryCount, time.Second, func() error {
		withTimeout, cancel := context.WithTimeout(ctx, c.cfg.Timeout)
		defer cancel()
		var err error
		blockNumber, err = c.client.BlockNumber(withTimeout)
		return err
	})

	if err != nil {
		return 0, fmt.Errorf("failed to get latest block number: %w", err)
	}

	return blockNumber, nil
}

// GetBlockByNumber retrieves a block by its number
func (c *Client) GetBlockByNumber(ctx context.Context, number uint64) (*types.Block, error) {
	// Wait for rate limiter
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter wait: %w", err)
	}

	var block *types.Block
	err := retry(c.cfg.RetryCount, time.Second, func() error {
		withTimeout, cancel := context.WithTimeout(ctx, c.cfg.Timeout)
		defer cancel()
		var err error
		block, err = c.client.BlockByNumber(withTimeout, big.NewInt(int64(number)))
		return err
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get block %d: %w", number, err)
	}

	log.Printf("Successfully retrieved block %d with %d transactions", number, len(block.Transactions()))
	return block, nil
}

// GetBlockReceipts retrieves all transaction receipts for a block using eth_getBlockReceipts
func (c *Client) GetBlockReceipts(ctx context.Context, blockNumber uint64) ([]*types.Receipt, error) {
	// Wait for rate limiter
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter wait: %w", err)
	}

	var receipts []*types.Receipt
	err := retry(c.cfg.RetryCount, time.Second, func() error {
		withTimeout, cancel := context.WithTimeout(ctx, c.cfg.Timeout)
		defer cancel()
		blockHex := fmt.Sprintf("0x%x", blockNumber)
		return c.client.Client().CallContext(withTimeout, &receipts, "eth_getBlockReceipts", blockHex)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get receipts for block %d: %w", blockNumber, err)
	}

	log.Printf("Successfully retrieved %d receipts for block %d", len(receipts), blockNumber)
	return receipts, nil
}

// Close closes the client connection
func (c *Client) Close() {
	if c.client != nil {
		c.client.Close()
	}
}
