package traderviewer

import (
	hu "github.com/ava-labs/subnet-evm/plugin/evm/orderbook/hubbleutils"
	b "github.com/ava-labs/subnet-evm/precompile/contracts/bibliophile"
)

func GetNotionalPositionAndMargin(bibliophile b.BibliophileClient, input *GetNotionalPositionAndMarginInput) GetNotionalPositionAndMarginOutput {
	notionalPosition, margin, requiredMargin := bibliophile.GetNotionalPositionAndRequiredMargin(input.Trader, input.IncludeFundingPayments, input.Mode, hu.V2) // @todo check if this is the right upgrade version
	return GetNotionalPositionAndMarginOutput{
		NotionalPosition: notionalPosition,
		Margin:           margin,
		RequiredMargin:   requiredMargin,
	}
}
