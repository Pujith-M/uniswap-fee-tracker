package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	Port                string
	DBUri               string
	UniswapV3Pool       string
	EtherscanConfig     EtherscanConfig
	BinanceConfig       BinanceConfig
	PriceFetchBatchSize int
}

// HTTPClientConfig contains common configuration for HTTP clients with rate limiting
type HTTPClientConfig struct {
	BaseURL    string
	RetryCount int
	RetryWait  time.Duration
	RateLimit  float64       // Rate limit per second
	RateBurst  int           // Maximum burst size
	Timeout    time.Duration // HTTP client timeout
}

type EtherscanConfig struct {
	HTTPClientConfig
	APIKey string
}

type BinanceConfig struct {
	HTTPClientConfig
}

func LoadConfig() (*Config, error) {
	// Required environment variables
	etherscanAPIKey := os.Getenv("ETHERSCAN_API_KEY")
	if etherscanAPIKey == "" {
		return nil, fmt.Errorf("ETHERSCAN_API_KEY environment variable is required")
	}

	dbURI := os.Getenv("DB_URI")
	if dbURI == "" {
		return nil, fmt.Errorf("DB_URI environment variable is required")
	}

	return &Config{
		Port:          ":8080",
		DBUri:         dbURI,
		UniswapV3Pool: "0x88e6a0c2ddd26feeb64f039a2c41296fcb3f5640", // Uniswap V3 USDC/WETH pool
		EtherscanConfig: EtherscanConfig{
			HTTPClientConfig: HTTPClientConfig{
				BaseURL:    "https://api.etherscan.io/api",
				RetryCount: 3,
				RetryWait:  time.Second,
				RateLimit:  5.0, // Etherscan limit: 5 calls per second
				RateBurst:  5,   // Allow burst of 5 requests
				Timeout:    10 * time.Second,
			},
			APIKey: etherscanAPIKey,
		},
		BinanceConfig: BinanceConfig{
			HTTPClientConfig: HTTPClientConfig{
				BaseURL:    "https://api.binance.com/api/v3",
				RetryCount: 3,
				RetryWait:  time.Second,
				RateLimit:  20.0, // Binance limit: 1200 requests per minute = 20 per second
				RateBurst:  50,   // Allow larger bursts for Binance
				Timeout:    10 * time.Second,
			},
		},
		PriceFetchBatchSize: 100,
	}, nil
}
