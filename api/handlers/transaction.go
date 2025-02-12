package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"uniswap-fee-tracker/api/models"
	"uniswap-fee-tracker/internal/syncer"
)

type TransactionHandler struct {
	syncService *syncer.Service
}

func NewTransactionHandler(syncService *syncer.Service) *TransactionHandler {
	return &TransactionHandler{
		syncService: syncService,
	}
}

// GetTransactionFee godoc
// @Summary Get transaction fee in USDT
// @Description Get the transaction fee in USDT for a specific Uniswap WETH-USDC transaction
// @Tags transactions
// @Accept json
// @Produce json
// @Param txHash path string true "Transaction Hash"
// @Success 200 {object} models.TransactionResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /api/v1/transactions/{txHash} [get]
func (h *TransactionHandler) GetTransactionFee(c *gin.Context) {
	txHash := c.Param("txHash")

	tx, err := h.syncService.GetTransaction(txHash)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "Transaction not found",
		})
		return
	}

	c.JSON(http.StatusOK, models.TransactionResponse{
		TxHash:      tx.TxHash,
		BlockNumber: tx.BlockNumber,
		Timestamp:   tx.Timestamp,
		GasUsed:     tx.GasUsed.String(),
		GasPrice:    tx.GasPrice.String(),
		FeeETH:      tx.FeeETH.Text('f', 18),
		FeeUSDT:     tx.FeeUSDT.Text('f', 6),
		ETHPrice:    tx.ETHPrice.Text('f', 6),
		Status:      tx.Status,
	})
}
