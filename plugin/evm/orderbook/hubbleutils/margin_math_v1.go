package hubbleutils

import (
	"math/big"
)

func GetNotionalPositionAndMarginV1(hState *HubbleState, userState *UserState, marginMode MarginMode) (*big.Int, *big.Int) {
	margin := Sub(GetNormalizedMargin(hState.Assets, userState.Margins), userState.PendingFunding)
	notionalPosition, unrealizedPnl := GetTotalNotionalPositionAndUnrealizedPnlV1(hState, userState, margin, marginMode)
	return notionalPosition, Add(margin, unrealizedPnl)
}

func GetTotalNotionalPositionAndUnrealizedPnlV1(hState *HubbleState, userState *UserState, margin *big.Int, marginMode MarginMode) (*big.Int, *big.Int) {
	notionalPosition := big.NewInt(0)
	unrealizedPnl := big.NewInt(0)
	for _, market := range hState.ActiveMarkets {
		_notionalPosition, _unrealizedPnl := getOptimalPnlV1(hState, userState.Positions[market], margin, market, marginMode)
		notionalPosition.Add(notionalPosition, _notionalPosition)
		unrealizedPnl.Add(unrealizedPnl, _unrealizedPnl)
	}
	return notionalPosition, unrealizedPnl
}

func getOptimalPnlV1(hState *HubbleState, position *Position, margin *big.Int, market Market, marginMode MarginMode) (notionalPosition *big.Int, uPnL *big.Int) {
	if position == nil || position.Size.Sign() == 0 {
		return big.NewInt(0), big.NewInt(0)
	}

	// based on last price
	notionalPosition, unrealizedPnl, midPriceBasedMF := GetPositionMetadata(
		hState.MidPrices[market],
		position.OpenNotional,
		position.Size,
		margin,
	)

	// based on oracle price
	oracleBasedNotional, oracleBasedUnrealizedPnl, oracleBasedMF := GetPositionMetadata(
		hState.OraclePrices[market],
		position.OpenNotional,
		position.Size,
		margin,
	)

	if (marginMode == Maintenance_Margin && oracleBasedMF.Cmp(midPriceBasedMF) == 1) || // for liquidations
		(marginMode == Min_Allowable_Margin && oracleBasedMF.Cmp(midPriceBasedMF) == -1) { // for increasing leverage
		return oracleBasedNotional, oracleBasedUnrealizedPnl
	}
	return notionalPosition, unrealizedPnl
}
