package hubbleutils

import (
	"math/big"
)

type Collateral struct {
	Price    *big.Int // scaled by 1e6
	Weight   *big.Int // scaled by 1e6
	Decimals uint8
}

func GetNormalizedMargin(assets []Collateral, margin []*big.Int) *big.Int {
	weighted, _ := WeightedAndSpotCollateral(assets, margin)
	return weighted
}

func WeightedAndSpotCollateral(assets []Collateral, margin []*big.Int) (weighted, spot *big.Int) {
	weighted = big.NewInt(0)
	spot = big.NewInt(0)
	for i, asset := range assets {
		if margin[i].Sign() == 0 {
			continue
		}
		numerator := Mul(margin[i], asset.Price) // margin[i] is scaled by asset.Decimal
		spot.Add(spot, Unscale(numerator, asset.Decimals))
		weighted.Add(weighted, Unscale(Mul(numerator, asset.Weight), asset.Decimals+6))
	}
	return weighted, spot
}
