package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	EtherscanAPIKey string
	UniswapV3Pool   string
	EtherscanConfig EtherscanConfig
}

type EtherscanConfig struct {
	BaseURL     string
	APIKey      string
	RetryCount  int
	RetryWait   time.Duration
	MaxRequests int           // Rate limit: max requests per minute
	Timeout     time.Duration // HTTP client timeout
}

func LoadConfig() (*Config, error) {
	apiKey := os.Getenv("ETHERSCAN_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("ETHERSCAN_API_KEY environment variable is required")
	}

	return &Config{
		EtherscanAPIKey: apiKey,
		UniswapV3Pool:   "0x88e6a0c2ddd26feeb64f039a2c41296fcb3f5640", // Uniswap V3 USDC/WETH pool
		EtherscanConfig: EtherscanConfig{
			BaseURL:     "https://api.etherscan.io/api",
			APIKey:      apiKey,
			RetryCount:  3,
			RetryWait:   time.Second,
			MaxRequests: 5, // Etherscan free tier limit
			Timeout:     10 * time.Second,
		},
	}, nil
}
