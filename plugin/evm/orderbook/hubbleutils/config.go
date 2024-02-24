package hubbleutils

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
	return &HubbleState{
		Assets:             hState.Assets,
		ActiveMarkets:      hState.ActiveMarkets,
		MinAllowableMargin: hState.MinAllowableMargin,
		MaintenanceMargin:  hState.MaintenanceMargin,
		TakerFee:           hState.TakerFee,
		UpgradeVersion:     hState.UpgradeVersion,
	}
}
