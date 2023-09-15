package bibliophile

import (
	"math/big"

	"github.com/ava-labs/subnet-evm/precompile/contract"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	CLEARING_HOUSE_GENESIS_ADDRESS       = "0x0300000000000000000000000000000000000002"
	MAINTENANCE_MARGIN_SLOT        int64 = 1
	MIN_ALLOWABLE_MARGIN_SLOT      int64 = 2
	TAKER_FEE_SLOT                 int64 = 3
	AMMS_SLOT                      int64 = 12
)

type MarginMode uint8

const (
	Maintenance_Margin MarginMode = iota
	Min_Allowable_Margin
)

func GetMarginMode(marginMode uint8) MarginMode {
	if marginMode == 0 {
		return Maintenance_Margin
	}
	return Min_Allowable_Margin
}

func marketsStorageSlot() *big.Int {
	return new(big.Int).SetBytes(crypto.Keccak256(common.LeftPadBytes(big.NewInt(AMMS_SLOT).Bytes(), 32)))
}

func GetActiveMarketsCount(stateDB contract.StateDB) int64 {
	rawVal := stateDB.GetState(common.HexToAddress(CLEARING_HOUSE_GENESIS_ADDRESS), common.BytesToHash(common.LeftPadBytes(big.NewInt(AMMS_SLOT).Bytes(), 32)))
	return new(big.Int).SetBytes(rawVal.Bytes()).Int64()
}

func GetMarkets(stateDB contract.StateDB) []common.Address {
	numMarkets := GetActiveMarketsCount(stateDB)
	markets := make([]common.Address, numMarkets)
	baseStorageSlot := marketsStorageSlot()
	for i := int64(0); i < numMarkets; i++ {
		amm := stateDB.GetState(common.HexToAddress(CLEARING_HOUSE_GENESIS_ADDRESS), common.BigToHash(new(big.Int).Add(baseStorageSlot, big.NewInt(i))))
		markets[i] = common.BytesToAddress(amm.Bytes())

	}
	return markets
}

type GetNotionalPositionAndMarginInput struct {
	Trader                 common.Address
	IncludeFundingPayments bool
	Mode                   uint8
}

type GetNotionalPositionAndMarginOutput struct {
	NotionalPosition *big.Int
	Margin           *big.Int
}

func GetNotionalPositionAndMargin(stateDB contract.StateDB, input *GetNotionalPositionAndMarginInput) GetNotionalPositionAndMarginOutput {
	margin := GetNormalizedMargin(stateDB, input.Trader)
	if input.IncludeFundingPayments {
		margin.Sub(margin, GetTotalFunding(stateDB, &input.Trader))
	}
	notionalPosition, unrealizedPnl := GetTotalNotionalPositionAndUnrealizedPnl(stateDB, &input.Trader, margin, GetMarginMode(input.Mode))
	return GetNotionalPositionAndMarginOutput{
		NotionalPosition: notionalPosition,
		Margin:           new(big.Int).Add(margin, unrealizedPnl),
	}
}

func GetTotalNotionalPositionAndUnrealizedPnl(stateDB contract.StateDB, trader *common.Address, margin *big.Int, marginMode MarginMode) (*big.Int, *big.Int) {
	notionalPosition := big.NewInt(0)
	unrealizedPnl := big.NewInt(0)
	for _, market := range GetMarkets(stateDB) {
		lastPrice := getLastPrice(stateDB, market)
		oraclePrice := getUnderlyingPrice(stateDB, market)
		_notionalPosition, _unrealizedPnl := getOptimalPnl(stateDB, market, oraclePrice, lastPrice, trader, margin, marginMode)
		notionalPosition.Add(notionalPosition, _notionalPosition)
		unrealizedPnl.Add(unrealizedPnl, _unrealizedPnl)
	}
	return notionalPosition, unrealizedPnl
}

func GetTotalFunding(stateDB contract.StateDB, trader *common.Address) *big.Int {
	totalFunding := big.NewInt(0)
	for _, market := range GetMarkets(stateDB) {
		totalFunding.Add(totalFunding, getPendingFundingPayment(stateDB, market, trader))
	}
	return totalFunding
}

// GetMaintenanceMargin returns the maintenance margin for a trader
func GetMaintenanceMargin(stateDB contract.StateDB) *big.Int {
	return new(big.Int).SetBytes(stateDB.GetState(common.HexToAddress(CLEARING_HOUSE_GENESIS_ADDRESS), common.BytesToHash(common.LeftPadBytes(big.NewInt(MAINTENANCE_MARGIN_SLOT).Bytes(), 32))).Bytes())
}

// GetMinAllowableMargin returns the minimum allowable margin for a trader
func GetMinAllowableMargin(stateDB contract.StateDB) *big.Int {
	return new(big.Int).SetBytes(stateDB.GetState(common.HexToAddress(CLEARING_HOUSE_GENESIS_ADDRESS), common.BytesToHash(common.LeftPadBytes(big.NewInt(MIN_ALLOWABLE_MARGIN_SLOT).Bytes(), 32))).Bytes())
}

// GetTakerFee returns the taker fee for a trader
func GetTakerFee(stateDB contract.StateDB) *big.Int {
	return fromTwosComplement(stateDB.GetState(common.HexToAddress(CLEARING_HOUSE_GENESIS_ADDRESS), common.BigToHash(big.NewInt(TAKER_FEE_SLOT))).Bytes())
}

func GetUnderlyingPrices(stateDB contract.StateDB) []*big.Int {
	underlyingPrices := make([]*big.Int, 0)
	for _, market := range GetMarkets(stateDB) {
		underlyingPrices = append(underlyingPrices, getUnderlyingPrice(stateDB, market))
	}
	return underlyingPrices
}

func getPosSizes(stateDB contract.StateDB, trader *common.Address) []*big.Int {
	positionSizes := make([]*big.Int, 0)
	for _, market := range GetMarkets(stateDB) {
		positionSizes = append(positionSizes, getSize(stateDB, market, trader))
	}
	return positionSizes
}

// func _getPositionSizesAndUpperBoundsForMarkets(stateDB contract.StateDB, trader *common.Address) GetPositionSizesAndUpperBoundsForMarketsOutput {
// 	markets := GetMarkets(stateDB)
// 	positionSizes := make([]*big.Int, len(markets))
// 	upperBounds := make([]*big.Int, len(markets))
// 	for i, market := range markets {
// 		positionSizes[i] = getSize(stateDB, market, trader)
// 		oraclePrice := getUnderlyingPrice(stateDB, market)
// 		spreadLimit := GetMaxOraclePriceSpread(stateDB, int64(i))
// 		upperBounds[i], _ = calculateBounds(spreadLimit, oraclePrice)
// 	}
// 	return GetPositionSizesAndUpperBoundsForMarketsOutput{
// 		PosSizes:    positionSizes,
// 		UpperBounds: upperBounds,
// 	}
// }

// getMarketAddressFromMarketID returns the market address for a given marketID
func getMarketAddressFromMarketID(marketID int64, stateDB contract.StateDB) common.Address {
	baseStorageSlot := marketsStorageSlot()
	amm := stateDB.GetState(common.HexToAddress(CLEARING_HOUSE_GENESIS_ADDRESS), common.BigToHash(new(big.Int).Add(baseStorageSlot, big.NewInt(marketID))))
	return common.BytesToAddress(amm.Bytes())
}
