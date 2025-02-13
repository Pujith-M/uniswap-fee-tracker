package syncer

import (
	"errors"

	"gorm.io/gorm"
)

type Repository interface {
	// Transaction operations
	SaveTransaction(tx *Transaction) error
	SaveTransactions(txs []*Transaction) error
	GetTransaction(txHash string) (*Transaction, error)
	UpdateTransactionStatus(txHash string, status TransactionStatus) error

	// Sync progress operations
	CreateSyncProgress(sp *SyncProgress) error
	UpdateSyncProgress(sp *SyncProgress) error
	GetIncompleteSyncProgress() ([]SyncProgress, error)

	// Block tracking operations
	UpdateLastTrackedBlock(blockNumber uint64) error
	GetLastTrackedBlock() (uint64, error)

	// Database operations
	AutoMigrate() error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) SaveTransaction(tx *Transaction) error {
	return r.db.Save(tx).Error
}

func (r *repository) SaveTransactions(txs []*Transaction) error {
	return r.db.CreateInBatches(txs, 100).Error
}

func (r *repository) GetTransaction(txHash string) (*Transaction, error) {
	var tx Transaction
	err := r.db.Where("tx_hash = ?", txHash).First(&tx).Error
	return &tx, err
}

func (r *repository) UpdateTransactionStatus(txHash string, status TransactionStatus) error {
	return r.db.Model(&Transaction{}).
		Where("tx_hash = ?", txHash).
		Update("status", status).
		Error
}

func (r *repository) CreateSyncProgress(sp *SyncProgress) error {
	return r.db.Create(sp).Error
}

func (r *repository) UpdateSyncProgress(sp *SyncProgress) error {
	return r.db.Save(sp).Error
}

func (r *repository) GetIncompleteSyncProgress() ([]SyncProgress, error) {
	var syncProgresses []SyncProgress
	err := r.db.Where("status != ?", SyncStatusCompleted).Order("created_at DESC").Find(&syncProgresses).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return syncProgresses, nil
}

// UpdateLastTrackedBlock updates the last processed block number
func (r *repository) UpdateLastTrackedBlock(blockNumber uint64) error {
	tracker := BlockTracker{Model: gorm.Model{ID: 1}, BlockNumber: blockNumber}
	return r.db.Save(&tracker).Error
}

// GetLastTrackedBlock returns the last processed block number
func (r *repository) GetLastTrackedBlock() (uint64, error) {
	var tracker BlockTracker
	return tracker.BlockNumber, r.db.First(&tracker).Error
}

// AutoMigrate creates or updates database tables
func (r *repository) AutoMigrate() error {
	return r.db.AutoMigrate(&Transaction{}, &SyncProgress{}, &BlockTracker{})
}
