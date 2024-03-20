package hubbleutils

import "math/big"

var (
	ChainId           int64
	VerifyingContract string
	hState            *HubbleState
)

func SetChainIdAndVerifyingSignedOrdersContract(chainId int64, verifyingContract string) {
	ChainId = chainId
	VerifyingContract = verifyingContract
}

func SetHubbleState(_hState *HubbleState) {
	hState = _hState
}

func GetHubbleState() *HubbleState {
	assets := make([]Collateral, len(hState.Assets))
	copy(assets, hState.Assets)

	activeMarkets := make([]Market, len(hState.ActiveMarkets))
	copy(activeMarkets, hState.ActiveMarkets)

	return &HubbleState{
		Assets:             assets,
		ActiveMarkets:      activeMarkets,
		MinAllowableMargin: new(big.Int).Set(hState.MinAllowableMargin),
		MaintenanceMargin:  new(big.Int).Set(hState.MaintenanceMargin),
		TakerFee:           new(big.Int).Set(hState.TakerFee),
		UpgradeVersion:     hState.UpgradeVersion,
	}
}
