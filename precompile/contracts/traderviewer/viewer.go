package traderviewer

import (
	"math/big"
	b "github.com/ava-labs/subnet-evm/precompile/contracts/bibliophile"
	"github.com/ethereum/go-ethereum/common"
)

func GetNotionalPositionAndMargin(bibliophile b.BibliophileClient, input *GetNotionalPositionAndMarginInput) GetNotionalPositionAndMarginOutput {
	notionalPosition, margin, requiredMargin := bibliophile.GetNotionalPositionAndRequiredMargin(input.Trader, input.IncludeFundingPayments, input.Mode)
	return GetNotionalPositionAndMarginOutput{
		NotionalPosition: notionalPosition,
		Margin:           margin,
		RequiredMargin:   requiredMargin,
	}
}

func GetCrossMarginAccountData(bibliophile b.BibliophileClient, input *GetCrossMarginAccountDataInput) GetCrossMarginAccountDataOutput {
	notionalPosition, requiredMargin, unrealizedPnl, pendingFunding := bibliophile.GetCrossMarginAccountData(input.Trader, input.Mode)
	return GetCrossMarginAccountDataOutput{
		NotionalPosition: notionalPosition,
		RequiredMargin:   requiredMargin,
		UnrealizedPnl:    unrealizedPnl,
		PendingFunding:   pendingFunding,
	}
}

func GetTotalFundingForCrossMarginPositions(bibliophile b.BibliophileClient, trader *common.Address) *big.Int {
	return bibliophile.GetTotalFundingForCrossMarginPositions(trader)
}

func GetTraderDataForMarket(bibliophile b.BibliophileClient, input *GetTraderDataForMarketInput) GetTraderDataForMarketOutput {
	isIsolated, notionalPosition, unrealizedPnl, requiredMargin, pendingFunding := bibliophile.GetTraderDataForMarket(input.Trader, input.AmmIndex.Int64(), input.Mode)
	return GetTraderDataForMarketOutput{
		IsIsolated:      isIsolated,
		NotionalPosition: notionalPosition,
		UnrealizedPnl:   unrealizedPnl,
		RequiredMargin:  requiredMargin,
		PendingFunding:  pendingFunding,
	}
}
