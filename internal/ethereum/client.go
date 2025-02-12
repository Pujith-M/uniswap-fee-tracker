package ethereum

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
)

// Client represents an Ethereum node client
type Client struct {
	client     *ethclient.Client
	nodeURL    string
	retryDelay time.Duration
	mu         sync.RWMutex
}

// NewClient creates a new Ethereum client
func NewClient(nodeURL string) (*Client, error) {
	client, err := ethclient.Dial(nodeURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum node: %w", err)
	}

	return &Client{
		client:     client,
		nodeURL:    nodeURL,
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

// GetBlockTimestamp returns the timestamp of a specific block
func (c *Client) GetBlockTimestamp(ctx context.Context, blockNumber uint64) (time.Time, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	block, err := c.client.BlockByNumber(ctx, big.NewInt(int64(blockNumber)))
	if err != nil {
		// Try to reconnect if the connection is lost
		if err := c.reconnect(); err != nil {
			return time.Time{}, fmt.Errorf("failed to get block and reconnect: %w", err)
		}
		// Retry once after reconnecting
		block, err = c.client.BlockByNumber(ctx, big.NewInt(int64(blockNumber)))
		if err != nil {
			return time.Time{}, fmt.Errorf("failed to get block after reconnect: %w", err)
		}
	}

	return time.Unix(int64(block.Time()), 0), nil
}

// reconnect attempts to reconnect to the Ethereum node
func (c *Client) reconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	client, err := ethclient.Dial(c.nodeURL)
	if err != nil {
		return fmt.Errorf("failed to reconnect to Ethereum node: %w", err)
	}

	c.client = client
	return nil
}

// Close closes the Ethereum client connection
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.client != nil {
		c.client.Close()
	}
}
