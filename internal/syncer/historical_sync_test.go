package syncer

import (
	"math/big"
	"strconv"
	"testing"
	"time"
	"uniswap-fee-tracker/internal/etherscan"

	"github.com/stretchr/testify/assert"
)

// Helper function to create a mock transfer
func createMockTransfer(hash string, blockNumber uint64, gasUsed, gasPrice *big.Int) etherscan.TokenTransfer {
	timeStr := time.Now().Format("2006-01-02 15:04:05")
	return etherscan.TokenTransfer{
		Hash:        hash,
		BlockNumber: strconv.FormatUint(blockNumber, 10),
		TimeStamp:   timeStr,
		GasUsed:     gasUsed.String(),
		GasPrice:    gasPrice.String(),
	}
}

func TestFilterAndGroupTransactions(t *testing.T) {
	// Setup
	service := &Service{}

	// Create test transfers
	transfers := []etherscan.TokenTransfer{
		createMockTransfer("tx1", 100, big.NewInt(21000), big.NewInt(1000000000)),
		createMockTransfer("tx2", 100, big.NewInt(21000), big.NewInt(1000000000)),
		createMockTransfer("tx3", 101, big.NewInt(21000), big.NewInt(1000000000)),
	}

	// Execute
	result := service.filterAndGroupTransactions(transfers, true, 101)

	// Assert
	assert.Equal(t, 2, len(result), "Should have transactions grouped into 2 blocks")

	// Find block 100 transactions
	var block100Txs []*Transaction
	var block101Txs []*Transaction
	for _, blockTxs := range result {
		if len(blockTxs) > 0 {
			if blockTxs[0].BlockNumber == 100 {
				block100Txs = blockTxs
			} else if blockTxs[0].BlockNumber == 101 {
				block101Txs = blockTxs
			}
		}
	}

	assert.Equal(t, 2, len(block100Txs), "Block 100 should have 2 transactions")
	assert.Equal(t, 1, len(block101Txs), "Block 101 should have 1 transaction")
}

func TestToTransferModel(t *testing.T) {
	// Setup
	service := &Service{}
	gasUsed := big.NewInt(21000)
	gasPrice := big.NewInt(1000000000)

	transfer := createMockTransfer("tx1", 100, gasUsed, gasPrice)

	// Execute
	result := service.toTransferModel(transfer)

	// Assert
	assert.NotNil(t, result)
	assert.Equal(t, "tx1", result.TxHash)
	assert.Equal(t, uint64(100), result.BlockNumber)
	assert.Equal(t, gasUsed.String(), result.GasUsed.String())
	assert.Equal(t, gasPrice.String(), result.GasPrice.String())
	assert.Equal(t, StatusPendingPrice, result.Status)
}
