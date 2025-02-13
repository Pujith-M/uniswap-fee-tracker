package ethereum

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Client represents an Ethereum node client
type Client struct {
	client     *ethclient.Client
	httpURL    string
	retryDelay time.Duration
	mu         sync.RWMutex
}

// NewClient creates a new Ethereum client
func NewClient(infuraApiKey string) (*Client, error) {
	httpURL := fmt.Sprintf("https://mainnet.infura.io/v3/%s", infuraApiKey)

	// Create HTTP client
	client, err := ethclient.Dial(httpURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to HTTP endpoint: %w", err)
	}

	return &Client{
		client:     client,
		httpURL:    httpURL,
		retryDelay: 5 * time.Second,
	}, nil
}

// GetLatestBlockNumber returns the latest block number from the Ethereum network
func (c *Client) GetLatestBlockNumber(ctx context.Context) (uint64, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	blockNumber, err := c.client.BlockNumber(ctx)
	if err != nil {
		// Try to reconnect if the connection is lost
		if err := c.reconnect(); err != nil {
			return 0, fmt.Errorf("failed to get latest block number and reconnect: %w", err)
		}
		// Retry once after reconnecting
		blockNumber, err = c.client.BlockNumber(ctx)
		if err != nil {
			return 0, fmt.Errorf("failed to get latest block number after reconnect: %w", err)
		}
	}

	return blockNumber, nil
}

// reconnect attempts to reconnect to the Ethereum node
func (c *Client) reconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Close existing connection
	if c.client != nil {
		c.client.Close()
	}

	// Try to reconnect
	client, err := ethclient.Dial(c.httpURL)
	if err != nil {
		return fmt.Errorf("failed to reconnect to HTTP endpoint: %w", err)
	}

	c.client = client
	return nil
}

// Close closes the client connection
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.client != nil {
		c.client.Close()
	}
}

// GetBlockByNumber retrieves a block by its number
func (c *Client) GetBlockByNumber(ctx context.Context, number uint64) (*types.Block, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	block, err := c.client.BlockByNumber(ctx, big.NewInt(int64(number)))
	if err != nil {
		// Try to reconnect if the connection is lost
		if err := c.reconnect(); err != nil {
			return nil, fmt.Errorf("failed to get block and reconnect: %w", err)
		}
		// Retry once after reconnecting
		block, err = c.client.BlockByNumber(ctx, big.NewInt(int64(number)))
		if err != nil {
			return nil, fmt.Errorf("failed to get block after reconnect: %w", err)
		}
	}

	return block, nil
}

// GetBlockReceipts retrieves all transaction receipts for a block using eth_getBlockReceipts
func (c *Client) GetBlockReceipts(ctx context.Context, blockNumber uint64) ([]*types.Receipt, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var receipts []*types.Receipt

	// Call eth_getBlockReceipts RPC method
	err := c.client.Client().CallContext(ctx, &receipts, "eth_getBlockReceipts", fmt.Sprintf("0x%x", blockNumber))
	if err != nil {
		// Try to reconnect if the connection is lost
		if err := c.reconnect(); err != nil {
			return nil, fmt.Errorf("failed to get block receipts and reconnect: %w", err)
		}
		// Retry once after reconnecting
		err = c.client.Client().CallContext(ctx, &receipts, "eth_getBlockReceipts", fmt.Sprintf("0x%x", blockNumber))
		if err != nil {
			return nil, fmt.Errorf("failed to get block receipts after reconnect: %w", err)
		}
	}

	return receipts, nil
}
