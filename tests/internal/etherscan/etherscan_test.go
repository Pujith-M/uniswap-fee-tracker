package etherscan_test

import (
	"context"
	"testing"
	"uniswap-fee-tracker/internal/config"
	"uniswap-fee-tracker/internal/etherscan"

	"github.com/stretchr/testify/assert"
)

func TestGetTokenTransfers(t *testing.T) {
	// Set up configurations for the Etherscan client
	testConfig, err := config.LoadConfig()
	assert.NoError(t, err, "should not return errors")

	// Create a new Etherscan client
	client := etherscan.NewClient(&testConfig.EtherscanConfig)

	// Define test parameters
	address := testConfig.UniswapV3Pool
	startBlock := uint64(0)
	endBlock := uint64(21823108)

	// Perform the API request
	ctx := context.Background()
	transfers, err := client.GetTokenTransfers(ctx, address, startBlock, endBlock)

	// Validate the results
	assert.NoError(t, err, "should not return errors")
	assert.NotNil(t, transfers, "transfers should not be nil")
	assert.Len(t, transfers, 10000, "should return exactly 10,000 token transfers")
}
