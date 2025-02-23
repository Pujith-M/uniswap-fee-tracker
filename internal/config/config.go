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
	UniswapStartBlock   uint64
	EtherscanConfig     EtherscanConfig
	BinanceConfig       BinanceConfig
	EthereumConfig      EthereumConfig
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

type EthereumConfig struct {
	InfuraAPIKey string
	HTTPClientConfig
}

func LoadConfig() (*Config, error) {
	// Required environment variables
	etherscanAPIKey := os.Getenv("ETHERSCAN_API_KEY")
	if etherscanAPIKey == "" {
		return nil, fmt.Errorf("ETHERSCAN_API_KEY environment variable is required")
	}

	infuraAPIKey := os.Getenv("INFURA_API_KEY")
	if infuraAPIKey == "" {
		return nil, fmt.Errorf("INFURA_API_KEY environment variable is required")
	}

	dbURI := os.Getenv("DB_URI")
	if dbURI == "" {
		return nil, fmt.Errorf("DB_URI environment variable is required")
	}

	return &Config{
		Port:              ":8080",
		DBUri:             dbURI,
		UniswapV3Pool:     "0x88e6a0c2ddd26feeb64f039a2c41296fcb3f5640", // Uniswap V3 USDC/WETH pool
		UniswapStartBlock: 12376729,                                     // Uniswap V3 deployment block
		EthereumConfig: EthereumConfig{
			InfuraAPIKey: infuraAPIKey,
			HTTPClientConfig: HTTPClientConfig{
				BaseURL:    "https://mainnet.infura.io/v3",
				RetryCount: 3,
				RetryWait:  time.Second,
				RateLimit:  10.0, // Example limit: 10 requests per second
				RateBurst:  5,    // Allow burst of 5 requests
				Timeout:    10 * time.Second,
			},
		},
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
