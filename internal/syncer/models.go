package syncer

import (
	"math/big"
	"time"

	"gorm.io/gorm"
)

// TransactionStatus represents the processing status of a transaction
type TransactionStatus string

const (
	StatusProcessed    TransactionStatus = "PROCESSED"
	StatusPendingPrice TransactionStatus = "PENDING_PRICE"
	StatusFailed       TransactionStatus = "FAILED"
)

// SyncStatus represents the status of a sync operation
type SyncStatus string

const (
	SyncStatusRunning   SyncStatus = "RUNNING"
	SyncStatusCompleted SyncStatus = "COMPLETED"
	SyncStatusFailed    SyncStatus = "FAILED"
)

// Transaction represents a processed Ethereum transaction with its fee in USDT
type Transaction struct {
	TxHash      string            `gorm:"primaryKey;type:varchar(66)" json:"tx_hash"`
	BlockNumber uint64            `gorm:"index" json:"block_number"`
	Timestamp   time.Time         `gorm:"index" json:"timestamp"`
	GasUsed     *big.Int          `gorm:"type:numeric" json:"gas_used"`
	GasPrice    *big.Int          `gorm:"type:numeric" json:"gas_price"` // Stored as string for big.Int
	FeeETH      *big.Float        `gorm:"type:numeric" json:"fee_eth"`   // Stored as string for big.Float
	FeeUSDT     *big.Float        `gorm:"type:numeric" json:"fee_usdt"`  // Stored as string for big.Float
	ETHPrice    *big.Float        `gorm:"type:numeric" json:"eth_price"` // ETH price in USDT at transaction time
	Status      TransactionStatus `gorm:"type:varchar(20)" json:"status"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// UpdatePrices calculates transaction fees based on ETH price
func (tx *Transaction) UpdatePrices(ethPrice *big.Float) {
	// Calculate fee in Wei (gas_used * gas_price)
	feeWei := new(big.Int).Mul(tx.GasUsed, tx.GasPrice)

	// Convert Wei to ETH (divide by 10^18)
	denominator := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	tx.FeeETH = new(big.Float).Quo(
		new(big.Float).SetInt(feeWei),
		new(big.Float).SetInt(denominator),
	)

	// Store ETH price
	tx.ETHPrice = new(big.Float).Set(ethPrice)

	// Calculate fee in USDT
	tx.FeeUSDT = new(big.Float).Mul(tx.FeeETH, ethPrice)

	// Update status
	tx.Status = StatusProcessed
	tx.UpdatedAt = time.Now()
}

// SyncProgress tracks the progress of historical sync
type SyncProgress struct {
	gorm.Model
	StartBlock            uint64     `gorm:"not null" json:"start_block"`
	LastProcessedBlock    uint64     `gorm:"not null" json:"last_processed_block"`
	LatestBlock           uint64     `gorm:"not null" json:"latest_block"`
	TransactionsProcessed uint64     `gorm:"not null" json:"transactions_processed"`
	Status                SyncStatus `gorm:"type:varchar(20);not null" json:"status"`
	ErrorMessage          string     `json:"error_message,omitempty"`
}

// TableName specifies the table name for Transaction
func (Transaction) TableName() string {
	return "transactions"
}

// TableName specifies the table name for SyncProgress
func (SyncProgress) TableName() string {
	return "sync_progress"
}
