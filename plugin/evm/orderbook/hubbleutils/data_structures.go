package hubbleutils

import (
	// "encoding/json"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type MarginMode = uint8

const (
	Maintenance_Margin MarginMode = iota
	Min_Allowable_Margin
)

type Collateral struct {
	Price    *big.Int // scaled by 1e6
	Weight   *big.Int // scaled by 1e6
	Decimals uint8
}

type Market = int

type Position struct {
	OpenNotional *big.Int `json:"open_notional"`
	Size         *big.Int `json:"size"`
}

type Trader struct {
	Positions map[Market]*Position `json:"positions"` // position for every market
	Margin    Margin               `json:"margin"`    // available margin/balance for every market
}

type Margin struct {
	Reserved  *big.Int                `json:"reserved"`
	Deposited map[Collateral]*big.Int `json:"deposited"`
}

type Side uint8

const (
	Long Side = iota
	Short
	Liquidation
)

type OrderStatus uint8

// has to be exact same as IOrderHandler
const (
	Invalid OrderStatus = iota
	Placed
	Filled
	Cancelled
)
