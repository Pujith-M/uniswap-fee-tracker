package binance_test

import (
	"context"
	"math/big"
	"testing"
	"time"
	"uniswap-fee-tracker/internal/binance"

	"github.com/stretchr/testify/assert"
)

func TestGetPrice(t *testing.T) {
	// Create a new Binance client
	client := binance.NewClient()

	// Define test data
	ctx := context.Background()
	timestamp := time.Unix(1739280301000, 0) // Convert UNIX timestamp
	expectedPrice := big.NewFloat(2680.98)

	// Perform the actual API call
	priceData, err := client.GetPrice(ctx, "ETHUSDT", timestamp)

	// Assertions
	assert.NoError(t, err, "Unexpected error occurred while fetching price")
	assert.NotNil(t, priceData, "Price data should not be nil")
	assert.Equal(t, expectedPrice.String(), priceData.Close.String(), "Close price does not match expected value")
}
