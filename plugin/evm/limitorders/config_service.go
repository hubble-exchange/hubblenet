package limitorders

import (
	"math/big"

	"github.com/ava-labs/subnet-evm/core"
	"github.com/ava-labs/subnet-evm/precompile/contracts/hubbleconfigmanager"
)

type IConfigService interface {
	getSpreadRatioThreshold() *big.Int
	getMaxLiquidationRatio() *big.Int
	getMinAllowableMargin() *big.Int
	getMaintenanceMargin() *big.Int
	getMinSizeRequirement() *big.Int
}

type ConfigService struct {
	blockChain *core.BlockChain
}

func NewConfigService(blockChain *core.BlockChain) IConfigService {
	return &ConfigService{
		blockChain: blockChain,
	}
}

func (cs *ConfigService) getSpreadRatioThreshold() *big.Int {
	stateDB, _ := cs.blockChain.StateAt(cs.blockChain.CurrentBlock().Root())
	return hubbleconfigmanager.GetSpreadRatioThreshold(stateDB)
}

func (cs *ConfigService) getMaxLiquidationRatio() *big.Int {
	stateDB, _ := cs.blockChain.StateAt(cs.blockChain.CurrentBlock().Root())
	return hubbleconfigmanager.GetMaxLiquidationRatio(stateDB)
}

func (cs *ConfigService) getMinAllowableMargin() *big.Int {
	stateDB, _ := cs.blockChain.StateAt(cs.blockChain.CurrentBlock().Root())
	return hubbleconfigmanager.GetMinAllowableMargin(stateDB)
}

func (cs *ConfigService) getMaintenanceMargin() *big.Int {
	stateDB, _ := cs.blockChain.StateAt(cs.blockChain.CurrentBlock().Root())
	return hubbleconfigmanager.GetMaintenanceMargin(stateDB)
}

func (cs *ConfigService) getMinSizeRequirement() *big.Int {
	stateDB, _ := cs.blockChain.StateAt(cs.blockChain.CurrentBlock().Root())
	return hubbleconfigmanager.GetMinSizeRequirement(stateDB)
}
