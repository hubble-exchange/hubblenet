package jurorv2

import (
	b "github.com/ava-labs/subnet-evm/precompile/contracts/bibliophilev2"
)

func GetNotionalPositionAndMargin(bibliophile b.BibliophileClient, input *GetNotionalPositionAndMarginInput) GetNotionalPositionAndMarginOutput {
	notionalPosition, margin := bibliophile.GetNotionalPositionAndMargin(input.Trader, input.IncludeFundingPayments, input.Mode)
	return GetNotionalPositionAndMarginOutput{
		NotionalPosition: notionalPosition,
		Margin:           margin,
	}
}
