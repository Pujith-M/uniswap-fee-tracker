package syncer

import (
	"strings"
)

// Uniswap V3 WETH-USDC pool constants
const (
	// WethUsdcPoolAddress is the address of the WETH-USDC pool
	WethUsdcPoolAddress = "0x88e6A0c2dDD26FEEb64F039a2c41296FcB3f5640"
	// SwapEventTopic is the topic0 for Uniswap V3 swap events
	SwapEventTopic = "0xc42079f94a6350d7e6235f29174924f928cc2ac818eb64fed8004e115fbcca67"
)

// IsWethUsdcPool checks if an address is the WETH-USDC pool
func IsWethUsdcPool(address string) bool {
	return strings.EqualFold(address, WethUsdcPoolAddress)
}

// IsSwapEvent checks if a log entry is a Uniswap V3 swap event
func IsSwapEvent(topic string) bool {
	return strings.EqualFold(topic, SwapEventTopic)
}
