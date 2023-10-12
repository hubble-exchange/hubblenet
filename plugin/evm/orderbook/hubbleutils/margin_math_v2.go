package hubbleutils

import (
	"math"
	"math/big"
)

func GetAvailableMargin(hState *HubbleState, userState *UserState) *big.Int {
	notionalPosition, margin := GetNotionalPositionAndMargin(hState, userState, Min_Allowable_Margin)
	return GetAvailableMargin_(notionalPosition, margin, userState.ReservedMargin, hState.MinAllowableMargin)
}

func GetMarginFraction(hState *HubbleState, userState *UserState) *big.Int {
	notionalPosition, margin := GetNotionalPositionAndMargin(hState, userState, Maintenance_Margin)
	if notionalPosition.Sign() == 0 {
		return big.NewInt(math.MaxInt64)
	}
	return Div(Mul1e6(margin), notionalPosition)
}

func GetNotionalPositionAndMargin(hState *HubbleState, userState *UserState, marginMode MarginMode) (*big.Int, *big.Int) {
	margin := Sub(GetNormalizedMargin(hState.Assets, userState.Margins), userState.PendingFunding)
	notionalPosition, unrealizedPnl := GetTotalNotionalPositionAndUnrealizedPnl(hState, userState, margin, marginMode)
	return notionalPosition, Add(margin, unrealizedPnl)
}

func GetTotalNotionalPositionAndUnrealizedPnl(hState *HubbleState, userState *UserState, margin *big.Int, marginMode MarginMode) (*big.Int, *big.Int) {
	notionalPosition := big.NewInt(0)
	unrealizedPnl := big.NewInt(0)
	for _, market := range hState.ActiveMarkets {
		_notionalPosition, _unrealizedPnl := GetOptimalPnl(hState, userState.Positions[market], margin, market, marginMode)
		notionalPosition.Add(notionalPosition, _notionalPosition)
		unrealizedPnl.Add(unrealizedPnl, _unrealizedPnl)
	}
	return notionalPosition, unrealizedPnl
}

func GetOptimalPnl(hState *HubbleState, position *Position, margin *big.Int, market Market, marginMode MarginMode) (notionalPosition *big.Int, uPnL *big.Int) {
	if position == nil || position.Size.Sign() == 0 {
		return big.NewInt(0), big.NewInt(0)
	}

	// based on last price
	notionalPosition, unrealizedPnl, _ := GetPositionMetadata(
		hState.MidPrices[market],
		position.OpenNotional,
		position.Size,
		margin,
	)

	// based on oracle price
	oracleBasedNotional, oracleBasedUnrealizedPnl, _ := GetPositionMetadata(
		hState.OraclePrices[market],
		position.OpenNotional,
		position.Size,
		margin,
	)

	if (marginMode == Maintenance_Margin && oracleBasedUnrealizedPnl.Cmp(unrealizedPnl) == 1) || // for liquidations
		(marginMode == Min_Allowable_Margin && oracleBasedUnrealizedPnl.Cmp(unrealizedPnl) == -1) { // for increasing leverage
		return oracleBasedNotional, oracleBasedUnrealizedPnl
	}
	return notionalPosition, unrealizedPnl
}
