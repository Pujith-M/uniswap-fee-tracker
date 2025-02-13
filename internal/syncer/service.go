package syncer

import (
	"context"
	"errors"
	"fmt"
	"log"
	"uniswap-fee-tracker/internal/binance"
	"uniswap-fee-tracker/internal/config"
	"uniswap-fee-tracker/internal/ethereum"
	"uniswap-fee-tracker/internal/etherscan"

	"gorm.io/gorm"
)

type Service struct {
	config          *config.Config
	etherScanClient etherscan.Client
	binanceClient   binance.Client
	repo            Repository
	nodeClient      *ethereum.Client
}

func NewService(config *config.Config, ethClient etherscan.Client, binClient binance.Client, nodeClient *ethereum.Client, repo Repository) *Service {
	return &Service{
		config:          config,
		etherScanClient: ethClient,
		binanceClient:   binClient,
		nodeClient:      nodeClient,
		repo:            repo,
	}
}

// GetTransaction returns a transaction by its hash
func (s *Service) GetTransaction(txHash string) (*Transaction, error) {
	return s.repo.GetTransaction(txHash)
}

func (s *Service) StartSync(ctx context.Context, indexedStartBlock uint64) error {
	// Get last tracked block
	lastTrackedBlock, err := s.repo.GetLastTrackedBlock()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// If no blocks tracked yet, start from configured start block
			lastTrackedBlock = indexedStartBlock - 1
		} else {
			return fmt.Errorf("failed to get last tracked block: %w", err)
		}
	}

	// Get latest block from node
	latestBlock, err := s.nodeClient.GetLatestBlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("failed to get latest block: %w", err)
	}

	// Validate block numbers
	if lastTrackedBlock > latestBlock {
		return fmt.Errorf("last tracked block (%d) is greater than latest block (%d)",
			lastTrackedBlock, latestBlock)
	}

	// Start historical sync if needed
	if lastTrackedBlock < latestBlock {
		log.Printf("Starting historical sync from block %d to %d", lastTrackedBlock, latestBlock)
		if err := s.StartHistoricalSync(ctx, lastTrackedBlock, latestBlock); err != nil {
			return fmt.Errorf("failed to start historical sync: %w", err)
		}
	}

	// Start live sync from the latest block
	go func() {
		// Start live sync from latest block
		log.Printf("Starting live sync from block %d", latestBlock)
		s.StartLiveSync(ctx, latestBlock)
	}()

	return nil
}
