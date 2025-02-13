package ethereum

import (
	"context"
	"os"
	"testing"
	"time"
	"uniswap-fee-tracker/internal/config"

	"github.com/stretchr/testify/assert"
)

func TestClientIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg, err2 := config.LoadConfig()
	assert.NoError(t, err2, "Unexpected error occurred while loading config")
	// Get node URL from environment variable
	nodeURL := os.Getenv("INFURA_API_KEY")
	if nodeURL == "" {
		t.Skip("INFURA_API_KEY not set, skipping integration test")
	}

	// Create client
	client, err := NewClient(&cfg.EthereumConfig)
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

	t.Run("GetBlockByNumber", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		blockNumber := uint64(1) // Example block number
		block, err := client.GetBlockByNumber(ctx, blockNumber)
		if err != nil {
			t.Fatalf("Failed to get block by number: %v", err)
		}

		if block == nil {
			t.Error("Expected block, got nil")
		}

		// Add more assertions based on block contents if needed
	})

	t.Run("GetBlockReceipts", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		blockNumber := uint64(21837889) // Example block number
		receipts, err := client.GetBlockReceipts(ctx, blockNumber)
		if err != nil {
			t.Fatalf("Failed to get block receipts: %v", err)
		}

		if len(receipts) == 0 {
			t.Error("Expected receipts, got empty slice")
		}
	})

	t.Run("GetBlockReceipts_NonExistentBlock", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		blockNumber := uint64(99999999) // Non-existent block number
		receipts, err := client.GetBlockReceipts(ctx, blockNumber)
		if err != nil {
			t.Fatalf("Failed to get block receipts: %v", err)
		}

		if len(receipts) != 0 {
			t.Errorf("Expected empty receipts, got %+v", receipts)
		}
		time.Sleep(time.Second) // Implement rate limiting
	})

	t.Run("NewClientError", func(t *testing.T) {
		_, err := NewClient(&config.EthereumConfig{
			InfuraAPIKey: "",
		}) // Missing InfuraAPIKey
		if err == nil {
			t.Fatal("Expected error when creating client without InfuraAPIKey")
		}
	})

	t.Run("NewClientError_InvalidConfig", func(t *testing.T) {
		_, err := NewClient(&config.EthereumConfig{
			InfuraAPIKey: "invalid-api-key", // Invalid API key
		})
		if err == nil {
			t.Fatal("Expected error when creating client with invalid InfuraAPIKey")
		}
	})
}
