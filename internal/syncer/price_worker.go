package syncer

import (
	"context"
	"log"
	"sync"
)

// processBatch handles a batch of transactions, updating their prices concurrently and saving to DB
func (s *Service) processBatch(ctx context.Context, txs [][]*Transaction, batchSize int) []*Transaction {
	total := len(txs)

	// Process transactions in batches
	for i := 0; i < total; i += batchSize {
		end := i + batchSize
		if end > total {
			end = total
		}

		var wg sync.WaitGroup
		currentBatch := txs[i:end]
		log.Printf("Starting price fetch batch %d to %d out of %d total transactions", i+1, end, total)

		for _, tx := range currentBatch {
			wg.Add(1)
			go func(txs []*Transaction) {
				defer wg.Done()

				// Fetch ETH/USDT price for the transaction
				kline, err := s.binClient.GetPrice(ctx, "ETHUSDT", txs[0].Timestamp)
				if err != nil {
					log.Printf("failed to get ETH price: %v", err)
					for _, tx := range txs {
						tx.Status = StatusFailed
					}
					return
				}
				for _, tx := range txs {
					tx.UpdatePrices(kline.Close)
				}
			}(tx)
		}

		// Wait for current batch to finish before processing next batch
		wg.Wait()
		log.Printf("Completed price fetch batch %d to %d out of %d total transactions", i+1, end, total)
	}
	results := make([]*Transaction, 0)
	for _, tx := range txs {
		results = append(results, tx...)
	}
	return results
}
