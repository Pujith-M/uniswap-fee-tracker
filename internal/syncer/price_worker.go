package syncer

import (
	"context"
	"log"
	"sync"
	"uniswap-fee-tracker/internal/binance"
)

type priceUpdateWorker struct {
	id        int
	binClient binance.Client
	repo      Repository
	tasks     chan *Transaction
	wg        *sync.WaitGroup
}

func newPriceUpdateWorker(id int, binClient binance.Client, repo Repository, wg *sync.WaitGroup) *priceUpdateWorker {
	return &priceUpdateWorker{
		id:        id,
		binClient: binClient,
		repo:      repo,
		tasks:     make(chan *Transaction, 100), // Add buffer to prevent blocking
		wg:        wg,
	}
}

func (w *priceUpdateWorker) start(ctx context.Context) {
	log.Printf("Starting price update worker %d", w.id)
	for {
		select {
		case tx := <-w.tasks:
			if err := w.updateTransactionPrice(ctx, tx); err != nil {
				log.Printf("Worker %d: Error updating price for tx %s: %v", w.id, tx.TxHash, err)
				err := w.repo.UpdateTransactionStatus(tx.TxHash, StatusFailed)
				if err != nil {
					log.Printf("Worker %d: Failed to update transaction status: %v", w.id, err)
					return
				}
			}
			w.wg.Done()
		case <-ctx.Done():
			log.Printf("Stopping price update worker %d", w.id)
			return
		}
	}
}

func (w *priceUpdateWorker) updateTransactionPrice(ctx context.Context, tx *Transaction) error {
	// Get ETH/USDT price at transaction timestamp
	kline, err := w.binClient.GetPrice(ctx, "ETHUSDT", tx.Timestamp)
	if err != nil {
		return err
	}

	// Update transaction with price data
	tx.UpdatePrices(kline.Close)

	// Log successful price update
	log.Printf("Worker %d: Updating price for tx %s: ETH price %s, fee %s ETH (%s USDT)",
		w.id, tx.TxHash, tx.ETHPrice.String(), tx.FeeETH.String(), tx.FeeUSDT.String())

	return w.repo.SaveTransaction(tx)
}

type PriceUpdatePool struct {
	workers []*priceUpdateWorker
	wg      sync.WaitGroup
}

func NewPriceUpdatePool(numWorkers int, binClient binance.Client, repo Repository) *PriceUpdatePool {
	pool := &PriceUpdatePool{
		workers: make([]*priceUpdateWorker, numWorkers),
	}

	for i := 0; i < numWorkers; i++ {
		pool.workers[i] = newPriceUpdateWorker(i, binClient, repo, &pool.wg)
	}

	return pool
}

func (p *PriceUpdatePool) Start(ctx context.Context) {
	for _, worker := range p.workers {
		go worker.start(ctx)
	}
}

func (p *PriceUpdatePool) Stop() {
	p.wg.Wait()
}

func (p *PriceUpdatePool) SubmitTransaction(tx *Transaction) {
	p.wg.Add(1)
	// Simple round-robin distribution
	workerIndex := tx.BlockNumber % uint64(len(p.workers))
	select {
	case p.workers[workerIndex].tasks <- tx:
		// Transaction submitted successfully
	default:
		// Channel is full, log warning and try next worker
		log.Printf("Warning: Worker %d channel full, trying next worker", workerIndex)
		nextWorker := (workerIndex + 1) % uint64(len(p.workers))
		p.workers[nextWorker].tasks <- tx
	}
}
