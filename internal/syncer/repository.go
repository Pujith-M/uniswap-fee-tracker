package syncer

import (
	"errors"

	"gorm.io/gorm"
)

type Repository interface {
	SaveTransaction(tx *Transaction) error
	SaveTransactions(txs []*Transaction) error
	GetTransaction(txHash string) (*Transaction, error)
	UpdateTransactionStatus(txHash string, status TransactionStatus) error
	CreateSyncProgress(sp *SyncProgress) error
	UpdateSyncProgress(sp *SyncProgress) error
	GetLatestSyncProgress() (*SyncProgress, error)
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
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &tx, nil
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

func (r *repository) GetLatestSyncProgress() (*SyncProgress, error) {
	var sp SyncProgress
	err := r.db.Order("created_at DESC").First(&sp).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &sp, nil
}

// AutoMigrate creates or updates database tables
func (r *repository) AutoMigrate() error {
	return r.db.AutoMigrate(&Transaction{}, &SyncProgress{})
}
