package orderbook

import (
	"math/big"

	"github.com/ava-labs/subnet-evm/core"
	"github.com/ava-labs/subnet-evm/core/state"
	hu "github.com/ava-labs/subnet-evm/plugin/evm/orderbook/hubbleutils"
	"github.com/ava-labs/subnet-evm/precompile/contracts/bibliophile"
	"github.com/ethereum/go-ethereum/common"
)

type IConfigService interface {
	getMaxLiquidationRatio(market Market) *big.Int
	getLiquidationSpreadThreshold(market Market) *big.Int
	GetMinAllowableMargin() *big.Int
	GetMaintenanceMargin() *big.Int
	getMinSizeRequirement(market Market) *big.Int
	GetPriceMultiplier(market Market) *big.Int
	GetActiveMarketsCount() int64
	GetUnderlyingPrices() []*big.Int
	GetMidPrices() []*big.Int
	GetCollaterals() []hu.Collateral
	GetLastPremiumFraction(market Market, trader *common.Address) *big.Int
	GetCumulativePremiumFraction(market Market) *big.Int
	GetAcceptableBounds(market Market) (*big.Int, *big.Int)
	GetAcceptableBoundsForLiquidation(market Market) (*big.Int, *big.Int)
	GetTakerFee() *big.Int
	HasReferrer(trader common.Address) bool

	GetSignedOrderStatus(orderHash common.Hash) int64
	IsTradingAuthority(trader, signer common.Address) bool
	GetSignedOrderbookContract() common.Address

	GetMarketAddressFromMarketID(marketId int64) common.Address
	GetImpactMarginNotional(ammAddress common.Address) *big.Int
	GetReduceOnlyAmounts(trader common.Address) []*big.Int

	IsSettledAll() bool
}

type ConfigService struct {
	blockChain *core.BlockChain
	stateDB    *state.StateDB
}

func NewConfigService(blockChain *core.BlockChain) IConfigService {
	return &ConfigService{
		blockChain: blockChain,
	}
}

func NewConfigServiceFromStateDB(stateDB *state.StateDB) IConfigService {
	return &ConfigService{
		stateDB: stateDB,
	}
}

func (cs *ConfigService) getStateAtCurrentBlock() *state.StateDB {
	if cs.stateDB != nil {
		return cs.stateDB
	}
	stateDB, _ := cs.blockChain.StateAt(cs.blockChain.CurrentBlock().Root)
	return stateDB
}

func (cs *ConfigService) GetAcceptableBounds(market Market) (*big.Int, *big.Int) {
	return bibliophile.GetAcceptableBounds(cs.getStateAtCurrentBlock(), int64(market))
}

func (cs *ConfigService) GetAcceptableBoundsForLiquidation(market Market) (*big.Int, *big.Int) {
	return bibliophile.GetAcceptableBoundsForLiquidation(cs.getStateAtCurrentBlock(), int64(market))
}

func (cs *ConfigService) getLiquidationSpreadThreshold(market Market) *big.Int {
	return bibliophile.GetMaxLiquidationPriceSpread(cs.getStateAtCurrentBlock(), int64(market))
}

func (cs *ConfigService) getMaxLiquidationRatio(market Market) *big.Int {
	return bibliophile.GetMaxLiquidationRatio(cs.getStateAtCurrentBlock(), int64(market))
}

func (cs *ConfigService) GetMinAllowableMargin() *big.Int {
	return bibliophile.GetMinAllowableMargin(cs.getStateAtCurrentBlock())
}

func (cs *ConfigService) GetMaintenanceMargin() *big.Int {
	return bibliophile.GetMaintenanceMargin(cs.getStateAtCurrentBlock())
}

func (cs *ConfigService) getMinSizeRequirement(market Market) *big.Int {
	return bibliophile.GetMinSizeRequirement(cs.getStateAtCurrentBlock(), int64(market))
}

func (cs *ConfigService) GetPriceMultiplier(market Market) *big.Int {
	return bibliophile.GetMultiplier(cs.getStateAtCurrentBlock(), int64(market))
}

func (cs *ConfigService) GetActiveMarketsCount() int64 {
	return bibliophile.GetActiveMarketsCount(cs.getStateAtCurrentBlock())
}

func (cs *ConfigService) GetUnderlyingPrices() []*big.Int {
	return bibliophile.GetUnderlyingPrices(cs.getStateAtCurrentBlock())
}

func (cs *ConfigService) GetMidPrices() []*big.Int {
	return bibliophile.GetMidPrices(cs.getStateAtCurrentBlock())
}

func (cs *ConfigService) GetCollaterals() []hu.Collateral {
	return bibliophile.GetCollaterals(cs.getStateAtCurrentBlock())
}

func (cs *ConfigService) GetLastPremiumFraction(market Market, trader *common.Address) *big.Int {
	markets := bibliophile.GetMarketsIncludingSettled(cs.getStateAtCurrentBlock())
	return bibliophile.GetLastPremiumFraction(cs.getStateAtCurrentBlock(), markets[market], trader)
}

func (cs *ConfigService) GetCumulativePremiumFraction(market Market) *big.Int {
	markets := bibliophile.GetMarketsIncludingSettled(cs.getStateAtCurrentBlock())
	return bibliophile.GetCumulativePremiumFraction(cs.getStateAtCurrentBlock(), markets[market])
}

func (cs *ConfigService) GetTakerFee() *big.Int {
	takerFee := bibliophile.GetTakerFee(cs.getStateAtCurrentBlock())
	return hu.Div(hu.Mul(takerFee, big.NewInt(8)), big.NewInt(10)) // 20% discount, which is applied to everyone currently
}

func (cs *ConfigService) HasReferrer(trader common.Address) bool {
	return bibliophile.HasReferrer(cs.getStateAtCurrentBlock(), trader)
}

func (cs *ConfigService) GetSignedOrderStatus(orderHash common.Hash) int64 {
	return bibliophile.GetSignedOrderStatus(cs.getStateAtCurrentBlock(), orderHash)
}

func (cs *ConfigService) IsTradingAuthority(trader, signer common.Address) bool {
	return bibliophile.IsTradingAuthority(cs.getStateAtCurrentBlock(), trader, signer)
}

func (cs *ConfigService) GetSignedOrderbookContract() common.Address {
	return bibliophile.GetSignedOrderBookAddress(cs.getStateAtCurrentBlock())
}

func (cs *ConfigService) GetMarketAddressFromMarketID(marketId int64) common.Address {
	return bibliophile.GetMarketAddressFromMarketID(marketId, cs.getStateAtCurrentBlock())
}

func (cs *ConfigService) GetImpactMarginNotional(ammAddress common.Address) *big.Int {
	return bibliophile.GetImpactMarginNotional(cs.getStateAtCurrentBlock(), ammAddress)
}

func (cs *ConfigService) GetReduceOnlyAmounts(trader common.Address) []*big.Int {
	return bibliophile.GetReduceOnlyAmounts(cs.getStateAtCurrentBlock(), trader)
}

func (cs *ConfigService) IsSettledAll() bool {
	return bibliophile.IsSettledAll(cs.getStateAtCurrentBlock())
}
