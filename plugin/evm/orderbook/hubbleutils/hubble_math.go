package hubbleutils

import (
	"math/big"
)

var (
	ONE_E_6  = big.NewInt(1e6)
	ONE_E_12 = big.NewInt(1e12)
	ONE_E_18 = big.NewInt(1e18)
)

func Mul1e6(number *big.Int) *big.Int {
	return new(big.Int).Mul(number, ONE_E_6)
}

func Div1e6(number *big.Int) *big.Int {
	return big.NewInt(0).Div(number, ONE_E_6)
}

func Mul1e18(number *big.Int) *big.Int {
	return big.NewInt(0).Mul(number, ONE_E_18)
}

func Div1e18(number *big.Int) *big.Int {
	return big.NewInt(0).Div(number, ONE_E_18)
}
