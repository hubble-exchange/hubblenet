package limitorders

import "math/big"

var (
	maintenanceMargin             = big.NewInt(1e5)
	spreadRatioThreshold          = big.NewInt(1e6)
	maxLiquidationRatio  *big.Int = big.NewInt(25 * 1e4) // 25%
	minSizeRequirement   *big.Int = big.NewInt(0).Mul(big.NewInt(5), _1e18)
)
