package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"uniswap-fee-tracker/internal/binance"
	"uniswap-fee-tracker/internal/config"
	"uniswap-fee-tracker/internal/etherscan"
	"uniswap-fee-tracker/internal/syncer"
)

func setupRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.String(200, "Service is healthy")
	})

	return r
}

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
	binClient := binance.NewClient()
	repo := syncer.NewRepository(db)

	// Auto migrate database schema
	if err := repo.AutoMigrate(); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Initialize syncer service
	syncerConfig := &syncer.Config{
		PoolAddress:        cfg.UniswapV3Pool,
		BatchSize:          100,
		MaxRetries:         3,
		RetryDelay:         time.Second * 5,
		PriceUpdateWorkers: 5,
	}

	service := syncer.NewService(syncerConfig, ethClient, binClient, repo)

	// Create context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start historical sync from Uniswap v3 deployment block
	startBlock := uint64(12376729)  // Uniswap v3 deployment block
	latestBlock := uint64(21824598) // You can adjust this or get it dynamically

	log.Printf("Starting historical sync from block %d to %d", startBlock, latestBlock)
	if err := service.StartHistoricalSync(ctx, startBlock, latestBlock); err != nil {
		log.Fatalf("Failed to start historical sync: %v", err)
	}

	// Start HTTP server
	gin.SetMode(gin.ReleaseMode)
	router := setupRouter()
	go func() {
		log.Println("Starting server on :8080")
		if err := router.Run(":8080"); err != nil {
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
