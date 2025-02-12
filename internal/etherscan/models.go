package etherscan

import (
	"fmt"
	"math/big"
	"strconv"
	"time"
)

// TokenTransfer represents a token transfer event from Etherscan
type TokenTransfer struct {
	BlockNumber       string `json:"blockNumber"`
	TimeStamp         string `json:"timeStamp"`
	Hash              string `json:"hash"`
	Nonce             string `json:"nonce"`
	BlockHash         string `json:"blockHash"`
	From              string `json:"from"`
	ContractAddress   string `json:"contractAddress"`
	To                string `json:"to"`
	Value             string `json:"value"`
	TokenName         string `json:"tokenName"`
	TokenSymbol       string `json:"tokenSymbol"`
	TokenDecimal      string `json:"tokenDecimal"`
	TransactionIndex  string `json:"transactionIndex"`
	Gas               string `json:"gas"`
	GasPrice          string `json:"gasPrice"`
	GasUsed           string `json:"gasUsed"`
	CumulativeGasUsed string `json:"cumulativeGasUsed"`
	Input             string `json:"input"`
	Confirmations     string `json:"confirmations"`
}

// GetBlockNumber converts the BlockNumber field from string to uint64
func (t TokenTransfer) GetBlockNumber() uint64 {
	blockNumber, err := strconv.ParseUint(t.BlockNumber, 10, 64)
	if err != nil {
		fmt.Printf("Failed to parse BlockNumber '%s' to uint64: %v\n", t.BlockNumber, err)
		return 0 // Return 0 on error
	}
	return blockNumber
}

// GetTimeStamp converts the TimeStamp field from string (Unix timestamp) to time.Time
func (t TokenTransfer) GetTimeStamp() time.Time {
	timestampInt, err := strconv.ParseInt(t.TimeStamp, 10, 64)
	if err != nil {
		fmt.Printf("Failed to parse TimeStamp '%s' to int64: %v\n", t.TimeStamp, err)
		return time.Time{} // Return zero time on error
	}
	return time.Unix(timestampInt, 0) // Convert Unix timestamp to time.Time
}

// GetGasUsed converts the GasUsed field from string to uint64
func (t TokenTransfer) GetGasUsed() *big.Int {
	gasUsed, ok := big.NewInt(0).SetString(t.GasUsed, 10)
	if !ok {
		fmt.Printf("Error: Unable to parse 'GasUsed' field '%s' into *big.Int\n", t.GasUsed)
		return nil // Return nil on error
	}
	return gasUsed
}

// GetGasUsed converts the GasUsed field from string to uint64
func (t TokenTransfer) GetGasPrice() *big.Int {
	gasPrice, ok := big.NewInt(0).SetString(t.GasPrice, 10)
	if !ok {
		fmt.Printf("Error: Unable to parse 'GasUsed' field '%s' into *big.Int\n", t.GasUsed)
		return nil // Return nil on error
	}
	return gasPrice
}

// EtherscanResponse represents the response from Etherscan API
type EtherscanResponse[T any] struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Result  []T    `json:"result"`
}

// Error represents an error response from Etherscan
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *Error) Error() string {
	return e.Message
}
