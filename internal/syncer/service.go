package syncer

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"uniswap-fee-tracker/internal/binance"
	"uniswap-fee-tracker/internal/etherscan"
)

type Config struct {
	PoolAddress        string
	BatchSize          int
	MaxRetries         int
	RetryDelay         time.Duration
	PriceUpdateWorkers int
}

type Service struct {
	config    *Config
	ethClient etherscan.Client
	binClient binance.Client
	repo      Repository
	mu        sync.Mutex
	pricePool *PriceUpdatePool
}

func NewService(config *Config, ethClient etherscan.Client, binClient binance.Client, repo Repository) *Service {
	s := &Service{
		config:    config,
		ethClient: ethClient,
		binClient: binClient,
		repo:      repo,
		pricePool: NewPriceUpdatePool(config.PriceUpdateWorkers, binClient, repo),
	}
	return s
}

// StartHistoricalSync starts a historical sync from startBlock to latestBlock
func (s *Service) StartHistoricalSync(ctx context.Context, startBlock, latestBlock uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

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

	// Start price update pool
	s.pricePool.Start(ctx)

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
		// Wait for all price updates to complete
		s.pricePool.Stop()
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
		transfers, err := s.ethClient.GetTokenTransfers(ctx, s.config.PoolAddress, currentBlock, progress.LatestBlock)
		if err != nil {
			progress.Status = SyncStatusFailed
			progress.ErrorMessage = fmt.Sprintf("failed to get token transfers: %v", err)
			s.repo.UpdateSyncProgress(progress)
			return
		}

		if len(transfers) == 0 {
			break
		}

		// Process transfers in batches
		txBatch := make([]*Transaction, 0, len(transfers))
		lastBlockInBatch := currentBlock

		// Map to track transactions by block number to handle duplicates
		txByBlock := make(map[uint64]map[string]*Transaction)

		for _, transfer := range transfers {
			blockNum := transfer.GetBlockNumber()

			// Skip transactions from the last block as they might be incomplete
			if lastBlockInBatch < blockNum {
				lastBlockInBatch = blockNum
			}

			// Only process transfers TO the pool address (incoming transfers)
			if transfer.To != s.config.PoolAddress {
				continue
			}

			// Initialize block map if not exists
			if _, exists := txByBlock[blockNum]; !exists {
				txByBlock[blockNum] = make(map[string]*Transaction)
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
				GasUsed:     gasUsed,
				GasPrice:    gasPrice,
				Status:      StatusPendingPrice,
			}

			// Store only one transaction per hash per block
			txByBlock[blockNum][transfer.Hash] = tx
		}

		// Flatten the deduplicated transactions into a batch
		for _, blockTxs := range txByBlock {
			for _, tx := range blockTxs {
				txBatch = append(txBatch, tx)
			}
		}

		// Skip if no valid transactions found after deduplication
		if len(txBatch) == 0 {
			currentBlock = lastBlockInBatch + 1
			continue
		}

		// Log batch processing
		log.Printf("Processing batch of %d transactions from block %d to %d",
			len(txBatch), currentBlock, lastBlockInBatch)

		// Save batch to database and submit for price updates
		if err := s.repo.SaveTransactions(txBatch); err != nil {
			progress.Status = SyncStatusFailed
			progress.ErrorMessage = fmt.Sprintf("failed to save transactions: %v", err)
			s.repo.UpdateSyncProgress(progress)
			return
		}

		// Submit transactions for price updates
		for _, tx := range txBatch {
			s.pricePool.SubmitTransaction(tx)
		}

		log.Printf("Successfully processed batch. Total transactions so far: %d",
			progress.TransactionsProcessed+uint64(len(txBatch)))

		// Update progress
		progress.LastProcessedBlock = lastBlockInBatch
		progress.TransactionsProcessed += uint64(len(txBatch))

		if err := s.repo.UpdateSyncProgress(progress); err != nil {
			log.Printf("Failed to update sync progress: %v", err)
		}

		// Set next start block
		currentBlock = lastBlockInBatch + 1
	}

	// Mark sync as completed
	progress.Status = SyncStatusCompleted
	s.repo.UpdateSyncProgress(progress)
}

// GetSyncProgress returns the current sync progress
func (s *Service) GetSyncProgress() (*SyncProgress, error) {
	return s.repo.GetLatestSyncProgress()
}

// GetTransaction returns a transaction by its hash
func (s *Service) GetTransaction(txHash string) (*Transaction, error) {
	return s.repo.GetTransaction(txHash)
}
