package models

import (
	"time"
	"uniswap-fee-tracker/internal/syncer"
)

// @title Uniswap Fee Tracker API
// @version 1.0
// @description API for tracking Uniswap WETH-USDC transaction fees in USDT
// @host localhost:8080
// @BasePath /api/v1

// TransactionResponse represents the API response for transaction details
// @Description Response containing transaction details including gas fees
type TransactionResponse struct {
	// Transaction hash
	// @Description Unique identifier of the transaction
	TxHash string `json:"tx_hash"`

	// Block number
	// @Description The block number in which this transaction was included
	BlockNumber uint64 `json:"block_number"`

	// Transaction timestamp
	// @Description When this transaction was processed
	Timestamp time.Time `json:"timestamp"`

	// Gas used
	// @Description Amount of gas used by this transaction
	GasUsed string `json:"gas_used"`

	// Gas price
	// @Description Price per unit of gas in Wei
	GasPrice string `json:"gas_price"`

	// Fee in ETH
	// @Description Transaction fee in ETH
	FeeETH string `json:"fee_eth"`

	// Fee in USDT
	// @Description Transaction fee converted to USDT
	FeeUSDT string `json:"fee_usdt"`

	// ETH price in USDT
	// @Description ETH/USDT price at transaction time
	ETHPrice string `json:"eth_price"`

	// Transaction status
	// @Description Current processing status of the transaction
	Status syncer.TransactionStatus `json:"status"`
}

// ErrorResponse represents the API error response
// @Description Error response when the API request fails
type ErrorResponse struct {
	// Error message
	// @Description Description of what went wrong
	Error string `json:"error"`
}
