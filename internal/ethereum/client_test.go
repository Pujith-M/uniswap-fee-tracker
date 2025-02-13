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
	nodeURL := os.Getenv("INFURA_API_KEY")
	if nodeURL == "" {
		t.Skip("INFURA_API_KEY not set, skipping integration test")
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
