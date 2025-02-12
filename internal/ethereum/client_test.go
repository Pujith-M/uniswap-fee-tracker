package ethereum

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestClientIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Get node URL from environment variable
	nodeURL := os.Getenv("ETH_NODE_URL")
	if nodeURL == "" {
		t.Skip("ETH_NODE_URL not set, skipping integration test")
	}

	// Create client
	client, err := NewClient(nodeURL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	t.Run("GetLatestBlockNumber", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		blockNumber, err := client.GetLatestBlockNumber(ctx)
		if err != nil {
			t.Fatalf("Failed to get latest block number: %v", err)
		}

		if blockNumber == 0 {
			t.Error("Got block number 0, expected non-zero block number")
		}

		t.Logf("Latest block number: %d", blockNumber)
	})

	t.Run("GetBlockTimestamp", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Get latest block number first
		blockNumber, err := client.GetLatestBlockNumber(ctx)
		if err != nil {
			t.Fatalf("Failed to get latest block number: %v", err)
		}

		// Get timestamp for that block
		timestamp, err := client.GetBlockTimestamp(ctx, blockNumber)
		if err != nil {
			t.Fatalf("Failed to get block timestamp: %v", err)
		}

		// Block timestamp should be recent (within last hour)
		if time.Since(timestamp) > time.Hour {
			t.Errorf("Block timestamp too old: %v", timestamp)
		}

		t.Logf("Block %d timestamp: %v", blockNumber, timestamp)
	})

	t.Run("Reconnection", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Get block number to ensure initial connection works
		_, err := client.GetLatestBlockNumber(ctx)
		if err != nil {
			t.Fatalf("Failed to get initial block number: %v", err)
		}

		// Force a reconnection
		if err := client.reconnect(); err != nil {
			t.Fatalf("Failed to reconnect: %v", err)
		}

		// Verify we can still get block number after reconnection
		blockNumber, err := client.GetLatestBlockNumber(ctx)
		if err != nil {
			t.Fatalf("Failed to get block number after reconnect: %v", err)
		}

		if blockNumber == 0 {
			t.Error("Got block number 0 after reconnect, expected non-zero block number")
		}
	})
}

func TestClientUnit(t *testing.T) {
	t.Run("InvalidURL", func(t *testing.T) {
		_, err := NewClient("invalid-url")
		if err == nil {
			t.Error("Expected error for invalid URL, got nil")
		}
	})

	t.Run("EmptyURL", func(t *testing.T) {
		_, err := NewClient("")
		if err == nil {
			t.Error("Expected error for empty URL, got nil")
		}
	})
}
