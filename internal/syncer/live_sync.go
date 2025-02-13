package syncer

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
)

const (
	blockBufferSize = 100
)

// BlockData represents a block and its transactions
type BlockData struct {
	Number    uint64
	Block     *types.Block
	Receipts  []*types.Receipt
	Timestamp time.Time
}

// StartLiveSync initiates real-time transaction monitoring starting from the given block
func (s *Service) StartLiveSync(ctx context.Context, startBlock uint64) {
	// Create a new context with cancellation
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Channel for blocks
	blockChan := make(chan *uint64, blockBufferSize)

	// Start workers
	go s.blockPoller(ctx, blockChan, startBlock)
	go s.blockProcessor(ctx, blockChan)

	// Wait for context cancellation
	<-ctx.Done()
	return
}

// blockPoller polls for new blocks at regular intervals
func (s *Service) blockPoller(ctx context.Context, blockChan chan *uint64, lastBlock uint64) {
	ticker := time.NewTicker(time.Second) // Poll every second
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Get latest block number
			latestBlock, err := s.nodeClient.GetLatestBlockNumber(ctx)
			if err != nil {
				log.Printf("Error getting latest block: %v", err)
				continue
			}

			log.Printf("Latest block: %d", latestBlock)

			// Process any new blocks
			if lastBlock < latestBlock {
				for i := lastBlock + 1; i <= latestBlock; i++ {
					blockChan <- &i
					lastBlock = i
				}
			}
		}
	}
}

// blockProcessor processes blocks from the channel
func (s *Service) blockProcessor(ctx context.Context, blockChan chan *uint64) {
	for {
		select {
		case <-ctx.Done():
			return
		case block := <-blockChan:
			// Process block transactions
			if err := s.processBlockTransactions(ctx, block); err != nil {
				log.Printf("Error processing block %d: %v", block, err)
			}
		}
	}
}

// processBlockTransactions processes transactions in a block
func (s *Service) processBlockTransactions(ctx context.Context, blockNum *uint64) error {
	// Get block and receipts
	block, err := s.nodeClient.GetBlockByNumber(ctx, *blockNum)
	if err != nil {
		return fmt.Errorf("failed to get block %d: %w", *blockNum, err)
	}

	receipts, err := s.nodeClient.GetBlockReceipts(ctx, *blockNum)
	if err != nil {
		return fmt.Errorf("failed to get receipts for block %d: %w", *blockNum, err)
	}

	// Create receipt map for quick lookup
	receiptMap := make(map[string]*types.Receipt)
	for _, receipt := range receipts {
		receiptMap[receipt.TxHash.Hex()] = receipt
	}

	blockTime := time.Unix(int64(block.Time()), 0)
	transactions := s.filterTransaction(block, receiptMap, blockNum, blockTime)

	// Save transactions to database
	if len(transactions) > 0 {
		// Get ETH/USDT price for this block
		kline, err := s.binanceClient.GetPrice(ctx, "ETHUSDT", blockTime)
		if err != nil {
			log.Printf("Error getting ETH price for block %d: %v", *blockNum, err)
			return fmt.Errorf("failed to get ETH price: %w", err)
		}
		for _, transaction := range transactions {
			transaction.UpdatePrices(kline.Close)
		}
		if err := s.repo.SaveTransactions(transactions); err != nil {
			return fmt.Errorf("failed to save transactions: %w", err)
		}
	}

	// Update tracker
	s.repo.UpdateLastTrackedBlock(*blockNum)

	return nil
}

func (s *Service) filterTransaction(block *types.Block, receiptMap map[string]*types.Receipt, blockNum *uint64, blockTime time.Time) []*Transaction {
	// Filter and process Uniswap WETH-USDC transactions
	var transactions []*Transaction

	for _, tx := range block.Transactions() {
		// Skip if no receipt
		txHash := tx.Hash().Hex()
		receipt := receiptMap[txHash]
		if receipt == nil {
			continue
		}

		// Check for Uniswap V3 WETH-USDC swap events
		isSwap := false
		for _, log := range receipt.Logs {
			// Check if log is from WETH-USDC pool and is a swap event
			if IsWethUsdcPool(log.Address.Hex()) && len(log.Topics) > 0 && IsSwapEvent(log.Topics[0].Hex()) {
				isSwap = true
				break
			}
		}

		// Skip if not a swap
		if !isSwap {
			continue
		}

		// Calculate fee in ETH
		gasUsed := big.NewInt(int64(receipt.GasUsed))
		gasPrice := tx.GasPrice()

		// Create transaction model
		txModel := &Transaction{
			TxHash:      txHash,
			BlockNumber: *blockNum,
			Timestamp:   blockTime,
			GasUsed:     NewBigInt(gasUsed),
			GasPrice:    NewBigInt(gasPrice),
			Status:      StatusPendingPrice,
		}

		transactions = append(transactions, txModel)
	}
	return transactions
}
