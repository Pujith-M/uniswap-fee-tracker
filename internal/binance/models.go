package binance

import (
	"math/big"
	"time"
)

// KlineData represents a single Kline/Candlestick data point
type KlineData struct {
	OpenTime                 time.Time
	Open                     *big.Float
	High                     *big.Float
	Low                      *big.Float
	Close                    *big.Float
	Volume                   *big.Float
	CloseTime                time.Time
	QuoteAssetVolume         *big.Float
	NumberOfTrades           int64
	TakerBuyBaseAssetVolume  *big.Float
	TakerBuyQuoteAssetVolume *big.Float
}
