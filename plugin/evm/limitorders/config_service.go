package limitorders

import (
	"math/big"

	"github.com/ava-labs/subnet-evm/core"
	"github.com/ava-labs/subnet-evm/precompile/contracts/hubbleconfigmanager"
)

var (
	minAllowableMargin  = big.NewInt(2 * 1e5) // 5x
	maintenanceMargin   = big.NewInt(1e5)
	maxLiquidationRatio = big.NewInt(25 * 1e4) // 25%
	minSizeRequirement  = big.NewInt(0).Mul(big.NewInt(5), _1e18)
)

type IConfigService interface {
	getSpreadRatioThreshold() *big.Int
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
