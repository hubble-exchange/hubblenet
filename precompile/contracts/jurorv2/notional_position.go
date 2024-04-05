package jurorv2

import (
	hu "github.com/ava-labs/subnet-evm/hubbleutils"
	b "github.com/ava-labs/subnet-evm/precompile/contracts/bibliophile"
)

func GetNotionalPositionAndMargin(bibliophile b.BibliophileClient, input *GetNotionalPositionAndMarginInput) GetNotionalPositionAndMarginOutput {
	notionalPosition, margin := bibliophile.GetNotionalPositionAndMargin(input.Trader, input.IncludeFundingPayments, input.Mode, hu.V2)
	return GetNotionalPositionAndMarginOutput{
		NotionalPosition: notionalPosition,
		Margin:           margin,
	}
}
