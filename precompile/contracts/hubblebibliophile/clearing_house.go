package hubblebibliophile

import (
	"math/big"

	"github.com/ava-labs/subnet-evm/precompile/contract"

	"github.com/ethereum/go-ethereum/common"
)

type MarginMode uint8

const (
	Maintenance_Margin MarginMode = iota
	Min_Allowable_Margin
)

func GetMarginMode(marginMode uint8) MarginMode {
	if marginMode == 0 {
		return Maintenance_Margin
	}
	return Min_Allowable_Margin
}

func GetMarkets() []*common.Address {
	return []*common.Address{}
}

func GetNotionalPositionAndMargin(stateDB contract.StateDB, input *GetNotionalPositionAndMarginInput) GetNotionalPositionAndMarginOutput {
	margin := GetNormalizedMargin(stateDB, input.Trader)
	if input.IncludeFundingPayments {
		margin.Sub(margin, GetTotalFunding(stateDB, &input.Trader))
	}
	notionalPosition, unrealizedPnl := GetTotalNotionalPositionAndUnrealizedPnl(stateDB, &input.Trader, margin, GetMarginMode(input.Mode))
	return GetNotionalPositionAndMarginOutput{
		NotionalPosition: notionalPosition,
		Margin:           new(big.Int).Add(margin, unrealizedPnl),
	}
}

func GetTotalNotionalPositionAndUnrealizedPnl(stateDB contract.StateDB, trader *common.Address, margin *big.Int, marginMode MarginMode) (*big.Int, *big.Int) {
	notionalPosition := big.NewInt(0)
	unrealizedPnl := big.NewInt(0)
	for _, market := range GetMarkets() {
		lastPrice := getLastPrice(stateDB, market)
		// oraclePrice := getUnderlyingPrice(stateDB, market)
		oraclePrice := multiply1e6(big.NewInt(1800))
		_notionalPosition, _unrealizedPnl := getOptimalPnl(stateDB, market, oraclePrice, lastPrice, trader, margin, marginMode)
		notionalPosition.Add(notionalPosition, _notionalPosition)
		unrealizedPnl.Add(unrealizedPnl, _unrealizedPnl)
	}
	return notionalPosition, unrealizedPnl
}

func GetTotalFunding(stateDB contract.StateDB, trader *common.Address) *big.Int {
	totalFunding := big.NewInt(0)
	for _, market := range GetMarkets() {
		totalFunding.Add(totalFunding, getPendingFundingPayment(stateDB, market, trader))
	}
	return totalFunding
}
