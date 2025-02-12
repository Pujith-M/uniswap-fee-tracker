package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"uniswap-fee-tracker/api"
	"uniswap-fee-tracker/api/handlers"
	"uniswap-fee-tracker/internal/binance"
	"uniswap-fee-tracker/internal/config"
	"uniswap-fee-tracker/internal/ethereum"
	"uniswap-fee-tracker/internal/etherscan"
	"uniswap-fee-tracker/internal/syncer"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Error,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	// Connect to PostgreSQL
	db, err := gorm.Open(postgres.Open(cfg.DBUri), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize clients and repository
	ethClient := etherscan.NewClient(&cfg.EtherscanConfig)
	binClient := binance.NewClient(&cfg.BinanceConfig)
	nodeClient, err := ethereum.NewClient(cfg.EthereumConfig.NodeURL)
	if err != nil {
		log.Fatalf("Failed to connect to Ethereum node: %v", err)
	}
	repo := syncer.NewRepository(db)

	// Auto migrate database schema
	if err := repo.AutoMigrate(); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	service := syncer.NewService(cfg, ethClient, binClient, nodeClient, repo)

	// Create context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start historical sync from Uniswap v3 deployment block
	log.Printf("Starting historical sync from block %d", cfg.UniswapStartBlock)
	if err := service.StartSync(ctx, cfg.UniswapStartBlock); err != nil {
		log.Fatalf("Failed to start historical sync: %v", err)
	}

	txHandler := handlers.NewTransactionHandler(service)

	// Create API server
	go func() {
		routes := api.NewServer(txHandler).RegisterRoutes()

		log.Println("Starting server on ", cfg.Port)
		if err := routes.Start(cfg.Port); err != nil {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("Shutting down gracefully...")
	cancel()
}
