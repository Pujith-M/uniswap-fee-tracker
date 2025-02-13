package syncer

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"
	"uniswap-fee-tracker/internal/etherscan"
)

// StartHistoricalSync starts a historical sync from startBlock to latestBlock
func (s *Service) StartHistoricalSync(ctx context.Context, startBlock, latestBlock uint64) error {
	if os.Getenv("DISABLE_HISTORICAL_SYNC") == "true" {
		return nil
	}
	// Create sync progress record
	progress := &SyncProgress{
		StartBlock:            startBlock,
		EndBlock:              latestBlock,
		LastProcessedBlock:    startBlock,
		TransactionsProcessed: 0,
		Status:                SyncStatusRunning,
		ErrorMessage:          "",
		CompletedAt:           nil,
	}

	if err := s.repo.CreateSyncProgress(progress); err != nil {
		return fmt.Errorf("failed to create sync progress: %w", err)
	}
	err := s.repo.UpdateLastTrackedBlock(latestBlock)
	if err != nil {
		log.Printf("failed to update last tracked block: %v", err)
		return err
	}
	// Start sync in a goroutine
	syncProgress, err := s.repo.GetIncompleteSyncProgress()
	if err != nil {
		log.Printf("failed to get incomplete sync progress: %v", err)
		return err
	}
	for _, sync := range syncProgress {
		go s.runHistoricalSync(ctx, &sync)
	}

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

	currentBlock := progress.LastProcessedBlock + 1
	for currentBlock <= progress.EndBlock {
		select {
		case <-ctx.Done():
			progress.Status = SyncStatusPaused
			progress.ErrorMessage = "context cancelled"
			s.repo.UpdateSyncProgress(progress)
			return
		default:
		}

		// Get transactions for current batch
		transfers, err := s.etherScanClient.GetTokenTransfers(ctx, s.config.UniswapV3Pool, currentBlock, progress.EndBlock)
		if err != nil {
			log.Printf("Failed to get token transfers for block %d: %v", currentBlock, err)
			time.Sleep(10 * time.Second)
			return
		}

		if len(transfers) == 0 {
			break
		}
		lastBlockInBatch := transfers[len(transfers)-1].GetBlockNumber()
		isFinalIteration := lastBlockInBatch == progress.EndBlock || lastBlockInBatch == currentBlock
		txBatch := s.filterAndGroupTransactions(transfers, isFinalIteration, lastBlockInBatch)

		// Log batch processing
		log.Printf("Fetching historic price for batch of %d transactions from block %d to %d",
			len(txBatch), currentBlock, lastBlockInBatch)

		// Fetch prices for batch transactions
		txsWithPrice := s.processBatch(ctx, txBatch, s.config.PriceFetchBatchSize)

		// Save batch to database and submit for price updates
		if err := s.repo.SaveTransactions(txsWithPrice); err != nil {
			progress.Status = SyncStatusFailed
			progress.ErrorMessage = fmt.Sprintf("failed to save transactions: %v", err)
			s.repo.UpdateSyncProgress(progress)
			return
		}

		log.Printf("Successfully fetched batch. Total transactions so far: %d",
			progress.TransactionsProcessed+uint64(len(txsWithPrice)))

		// Update progress
		progress.LastProcessedBlock = lastBlockInBatch
		if !isFinalIteration {
			progress.LastProcessedBlock = lastBlockInBatch - 1
		}
		progress.TransactionsProcessed += uint64(len(txsWithPrice))

		if err := s.repo.UpdateSyncProgress(progress); err != nil {
			log.Printf("Failed to update sync progress: %v", err)
		}

		// Set next start block
		currentBlock = lastBlockInBatch
		if isFinalIteration {
			break
		}
	}

	// Mark sync as completed
	progress.Status = SyncStatusCompleted
	s.repo.UpdateSyncProgress(progress)
}

func (s *Service) filterAndGroupTransactions(transfers []etherscan.TokenTransfer, isFinalIteration bool, lastBlockInBatch uint64) [][]*Transaction {
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

		// Exclude transactions from the last block if it's not the final iteration to avoid processing incomplete data
		if lastBlockInBatch == blockNum && !isFinalIteration {
			continue
		}

		txMap[transfer.Hash] = s.toTransferModel(transfer)
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

func (s *Service) toTransferModel(transfer etherscan.TokenTransfer) *Transaction {
	// Convert transfer to transaction
	gasUsed := transfer.GetGasUsed()
	gasPrice := transfer.GetGasPrice()

	tx := &Transaction{
		TxHash:      transfer.Hash,
		BlockNumber: transfer.GetBlockNumber(),
		Timestamp:   transfer.GetTimeStamp(),
		GasUsed:     NewBigInt(gasUsed),
		GasPrice:    NewBigInt(gasPrice),
		Status:      StatusPendingPrice,
	}
	return tx
}
