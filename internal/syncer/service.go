package syncer

import (
	"context"
	"fmt"
	"log"
	"time"
	"uniswap-fee-tracker/internal/binance"
	"uniswap-fee-tracker/internal/config"
	"uniswap-fee-tracker/internal/etherscan"
)

type Service struct {
	config    *config.Config
	ethClient etherscan.Client
	binClient binance.Client
	repo      Repository
}

func NewService(config *config.Config, ethClient etherscan.Client, binClient binance.Client, repo Repository) *Service {
	return &Service{
		config:    config,
		ethClient: ethClient,
		binClient: binClient,
		repo:      repo,
	}
}

// StartHistoricalSync starts a historical sync from startBlock to latestBlock
func (s *Service) StartHistoricalSync(ctx context.Context, startBlock, latestBlock uint64) error {
	// Create sync progress record
	progress := &SyncProgress{
		StartBlock:         startBlock,
		LastProcessedBlock: startBlock,
		LatestBlock:        latestBlock,
		Status:             SyncStatusRunning,
	}

	if err := s.repo.CreateSyncProgress(progress); err != nil {
		return fmt.Errorf("failed to create sync progress: %w", err)
	}

	// Start sync in a goroutine
	go s.runHistoricalSync(ctx, progress)

	return nil
}

func (s *Service) runHistoricalSync(ctx context.Context, progress *SyncProgress) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in historical sync: %v", r)
			progress.Status = SyncStatusFailed
			progress.ErrorMessage = fmt.Sprintf("panic: %v", r)
			s.repo.UpdateSyncProgress(progress)
		}
	}()

	currentBlock := progress.StartBlock
	for currentBlock <= progress.LatestBlock {
		select {
		case <-ctx.Done():
			progress.Status = SyncStatusFailed
			progress.ErrorMessage = "context cancelled"
			s.repo.UpdateSyncProgress(progress)
			return
		default:
		}

		// Get transactions for current batch
		transfers, err := s.ethClient.GetTokenTransfers(ctx, s.config.UniswapV3Pool, currentBlock, progress.LatestBlock)
		if err != nil {
			log.Printf("Failed to get token transfers for block %d: %v", currentBlock, err)
			time.Sleep(10 * time.Second)
			return
		}

		if len(transfers) == 0 {
			break
		}
		lastBlockInBatch := transfers[len(transfers)-1].GetBlockNumber()

		txBatch := s.filterAndGroupTransactions(transfers, lastBlockInBatch)

		// Log batch processing
		log.Printf("Fetching historic price for batch of %d transactions from block %d to %d",
			len(txBatch), currentBlock, lastBlockInBatch-1)

		// Fetch prices for batch transactions
		txsWithPrice := s.processBatch(ctx, txBatch, s.config.PriceFetchBatchSize)

		// Save batch to database and submit for price updates
		if err := s.repo.SaveTransactions(txsWithPrice); err != nil {
			progress.Status = SyncStatusFailed
			progress.ErrorMessage = fmt.Sprintf("failed to save transactions: %v", err)
			s.repo.UpdateSyncProgress(progress)
			return
		}

		log.Printf("Successfully processed batch. Total transactions so far: %d",
			progress.TransactionsProcessed+uint64(len(txsWithPrice)))

		// Update progress
		progress.LastProcessedBlock = lastBlockInBatch - 1
		progress.TransactionsProcessed += uint64(len(txsWithPrice))

		if err := s.repo.UpdateSyncProgress(progress); err != nil {
			log.Printf("Failed to update sync progress: %v", err)
		}

		// Set next start block
		currentBlock = lastBlockInBatch
	}

	// Mark sync as completed
	progress.Status = SyncStatusCompleted
	s.repo.UpdateSyncProgress(progress)
}

func (s *Service) filterAndGroupTransactions(transfers []etherscan.TokenTransfer, lastBlockInBatch uint64) [][]*Transaction {
	// Process transfers in batches using a map to track transactions
	txMap := make(map[string]*Transaction)

	// First pass: collect all transactions
	for _, transfer := range transfers {
		// Check if transaction already exists in the map to avoid duplicates
		_, found := txMap[transfer.Hash]
		if found {
			continue
		}
		blockNum := transfer.GetBlockNumber()

		// Skip transactions from the last block as they might be incomplete
		if lastBlockInBatch == blockNum {
			continue
		}

		// Convert transfer to transaction
		gasUsed := transfer.GetGasUsed()
		gasPrice := transfer.GetGasPrice()
		if gasUsed == nil || gasPrice == nil {
			log.Printf("Warning: Invalid gas values for tx %s, skipping", transfer.Hash)
			continue
		}

		tx := &Transaction{
			TxHash:      transfer.Hash,
			BlockNumber: blockNum,
			Timestamp:   transfer.GetTimeStamp(),
			GasUsed:     NewBigInt(gasUsed),
			GasPrice:    NewBigInt(gasPrice),
			Status:      StatusPendingPrice,
		}

		txMap[transfer.Hash] = tx
	}

	// Build final batch from the tracked transactions
	groupedTxMap := make(map[uint64][]*Transaction)
	for _, tx := range txMap {
		blockTxs, found := groupedTxMap[tx.BlockNumber]
		if !found {
			blockTxs = make([]*Transaction, 0)
		}
		blockTxs = append(blockTxs, tx)
		groupedTxMap[tx.BlockNumber] = blockTxs
	}
	txBatch := make([][]*Transaction, 0)
	for _, transactions := range groupedTxMap {
		txBatch = append(txBatch, transactions)
	}
	return txBatch
}

// GetSyncProgress returns the current sync progress
func (s *Service) GetSyncProgress() (*SyncProgress, error) {
	return s.repo.GetLatestSyncProgress()
}

// GetTransaction returns a transaction by its hash
func (s *Service) GetTransaction(txHash string) (*Transaction, error) {
	return s.repo.GetTransaction(txHash)
}
