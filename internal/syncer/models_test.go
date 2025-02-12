package syncer

import (
	"math/big"
	"testing"
	"time"
)

func TestTransaction_UpdatePrices(t *testing.T) {
	tests := []struct {
		name        string
		tx          *Transaction
		ethPrice    string
		wantFeeETH  string
		wantFeeUSDT string
	}{
		{
			name: "simple gas calculation",
			tx: &Transaction{
				TxHash:    "0x123",
				GasUsed:   NewBigInt(big.NewInt(21000)),       // Standard ETH transfer gas
				GasPrice:  NewBigInt(big.NewInt(50000000000)), // 50 Gwei
				Status:    StatusPendingPrice,
				Timestamp: time.Now(),
			},
			ethPrice:    "2000.50",    // ETH price in USDT
			wantFeeETH:  "0.00105",    // 21000 * 50 Gwei = 0.00105 ETH
			wantFeeUSDT: "2.10052500", // 0.00105 * 2000.50 USDT
		},
		{
			name: "complex gas calculation",
			tx: &Transaction{
				TxHash:    "0x456",
				GasUsed:   NewBigInt(big.NewInt(300000)),       // Complex contract interaction
				GasPrice:  NewBigInt(big.NewInt(100000000000)), // 100 Gwei
				Status:    StatusPendingPrice,
				Timestamp: time.Now(),
			},
			ethPrice:    "1850.75",  // ETH price in USDT
			wantFeeETH:  "0.03",     // 300000 * 100 Gwei = 0.03 ETH
			wantFeeUSDT: "55.52250", // 0.03 * 1850.75 USDT
		},
		{
			name: "high gas price calculation",
			tx: &Transaction{
				TxHash:    "0x789",
				GasUsed:   NewBigInt(big.NewInt(21000)),        // Standard ETH transfer gas
				GasPrice:  NewBigInt(big.NewInt(500000000000)), // 500 Gwei (high gas price)
				Status:    StatusPendingPrice,
				Timestamp: time.Now(),
			},
			ethPrice:    "3000.25",   // ETH price in USDT
			wantFeeETH:  "0.0105",    // 21000 * 500 Gwei = 0.0105 ETH
			wantFeeUSDT: "31.502625", // 0.0105 * 3000.25 USDT
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert expected ETH price to big.Float
			ethPrice, _ := new(big.Float).SetString(tt.ethPrice)
			wantFeeETH, _ := new(big.Float).SetString(tt.wantFeeETH)
			wantFeeUSDT, _ := new(big.Float).SetString(tt.wantFeeUSDT)

			// Update prices
			tt.tx.UpdatePrices(ethPrice)

			// Check if status was updated
			if tt.tx.Status != StatusProcessed {
				t.Errorf("Status not updated, got %v, want %v", tt.tx.Status, StatusProcessed)
			}

			// Compare FeeETH with 5 decimal precision
			if tt.tx.FeeETH.Text('f', 5) != wantFeeETH.Text('f', 5) {
				t.Errorf("FeeETH = %v, want %v", tt.tx.FeeETH.Text('f', 5), wantFeeETH.Text('f', 5))
			}

			// Compare FeeUSDT with 6 decimal precision
			if tt.tx.FeeUSDT.Text('f', 6) != wantFeeUSDT.Text('f', 6) {
				t.Errorf("FeeUSDT = %v, want %v", tt.tx.FeeUSDT.Text('f', 6), wantFeeUSDT.Text('f', 6))
			}

			// Check if ETH price was stored correctly
			if tt.tx.ETHPrice.Cmp(ethPrice) != 0 {
				t.Errorf("ETHPrice not stored correctly, got %v, want %v", tt.tx.ETHPrice, ethPrice)
			}

			// Check if UpdatedAt was set
			if tt.tx.UpdatedAt.IsZero() {
				t.Error("UpdatedAt timestamp not set")
			}
		})
	}
}

func TestTransaction_UpdatePrices_ZeroValues(t *testing.T) {
	// Test with zero gas values
	tx := &Transaction{
		TxHash:    "0x000",
		GasUsed:   NewBigInt(big.NewInt(0)),
		GasPrice:  NewBigInt(big.NewInt(0)),
		Status:    StatusPendingPrice,
		Timestamp: time.Now(),
	}

	ethPrice, _ := new(big.Float).SetString("2000.50")
	tx.UpdatePrices(ethPrice)

	// Check if fees are zero
	zero := new(big.Float)
	if tx.FeeETH.Cmp(zero) != 0 {
		t.Errorf("Expected zero FeeETH, got %v", tx.FeeETH)
	}
	if tx.FeeUSDT.Cmp(zero) != 0 {
		t.Errorf("Expected zero FeeUSDT, got %v", tx.FeeUSDT)
	}
}
