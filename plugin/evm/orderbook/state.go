package orderbook

import (
	hu "github.com/ava-labs/subnet-evm/plugin/evm/orderbook/hubbleutils"
)

func GetHubbleState(configService IConfigService) *hu.HubbleState {
	count := configService.GetActiveMarketsCount()
	markets := make([]Market, count)
	for i := int64(0); i < count; i++ {
		markets[i] = Market(i)
	}
	hState := &hu.HubbleState{
		Assets:             configService.GetCollaterals(),
		ActiveMarkets:      markets,
		MinAllowableMargin: configService.GetMinAllowableMargin(),
		MaintenanceMargin:  configService.GetMaintenanceMargin(),
		TakerFee:           configService.GetTakerFee(),
		UpgradeVersion:     hu.V2,
	}

	return hState
}
