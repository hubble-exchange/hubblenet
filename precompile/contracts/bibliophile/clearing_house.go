package bibliophile

import (
	"math/big"

	hu "github.com/ava-labs/subnet-evm/plugin/evm/orderbook/hubbleutils"
	"github.com/ava-labs/subnet-evm/precompile/contract"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	CLEARING_HOUSE_GENESIS_ADDRESS       = "0x03000000000000000000000000000000000000b2"
	MAINTENANCE_MARGIN_SLOT        int64 = 1
	MIN_ALLOWABLE_MARGIN_SLOT      int64 = 2
	TAKER_FEE_SLOT                 int64 = 3
	AMMS_SLOT                      int64 = 12
	REFERRAL_SLOT                  int64 = 13
	SETTLED_ALL_SLOT               int64 = 19
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
	if isSettledAll(stateDB) {
		return 0
	}
	return GetMarketsCountRaw(stateDB)
}

func GetMarketsCountRaw(stateDB contract.StateDB) int64 {
	rawVal := stateDB.GetState(common.HexToAddress(CLEARING_HOUSE_GENESIS_ADDRESS), common.BytesToHash(common.LeftPadBytes(big.NewInt(AMMS_SLOT).Bytes(), 32)))
	return new(big.Int).SetBytes(rawVal.Bytes()).Int64()
}

func isSettledAll(stateDB contract.StateDB) bool {
	return stateDB.GetState(common.HexToAddress(CLEARING_HOUSE_GENESIS_ADDRESS), common.BigToHash(big.NewInt(SETTLED_ALL_SLOT))).Big().Sign() == 1
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

func GetMarketsIncludingSettled(stateDB contract.StateDB) []common.Address {
	numMarkets := GetMarketsCountRaw(stateDB)
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

func getNotionalPositionAndMargin(stateDB contract.StateDB, input *GetNotionalPositionAndMarginInput, upgradeVersion hu.UpgradeVersion) GetNotionalPositionAndMarginOutput {
	markets := GetMarketsIncludingSettled(stateDB)
	numMarkets := len(markets)
	positions := make(map[int]*hu.Position, numMarkets)
	underlyingPrices := make(map[int]*big.Int, numMarkets)
	midPrices := make(map[int]*big.Int, numMarkets)
	settlementPrices := make(map[int]*big.Int, numMarkets)
	var activeMarketIds []int
	for i, market := range markets {
		positions[i] = getPosition(stateDB, GetMarketAddressFromMarketID(int64(i), stateDB), &input.Trader)
		underlyingPrices[i] = getUnderlyingPrice(stateDB, market)
		midPrices[i] = getMidPrice(stateDB, market)
		settlementPrices[i] = getSettlementPrice(stateDB, market)
		if settlementPrices[i] == nil || settlementPrices[i].Sign() == 0 {
			activeMarketIds = append(activeMarketIds, i)
		}
	}
	pendingFunding := big.NewInt(0)
	if input.IncludeFundingPayments {
		pendingFunding = GetTotalFunding(stateDB, &input.Trader)
	}
	notionalPosition, margin := hu.GetNotionalPositionAndMargin(
		&hu.HubbleState{
			Assets:           GetCollaterals(stateDB),
			OraclePrices:     underlyingPrices,
			MidPrices:        midPrices,
			SettlementPrices: settlementPrices,
			ActiveMarkets:    activeMarketIds,
			UpgradeVersion:   upgradeVersion,
		},
		&hu.UserState{
			Positions:      positions,
			Margins:        getMargins(stateDB, input.Trader),
			PendingFunding: pendingFunding,
		},
		input.Mode,
	)
	return GetNotionalPositionAndMarginOutput{
		NotionalPosition: notionalPosition,
		Margin:           margin,
	}
}

func GetTotalFunding(stateDB contract.StateDB, trader *common.Address) *big.Int {
	totalFunding := big.NewInt(0)
	for _, market := range GetMarketsIncludingSettled(stateDB) {
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

func GetMidPrices(stateDB contract.StateDB) []*big.Int {
	underlyingPrices := make([]*big.Int, 0)
	for _, market := range GetMarkets(stateDB) {
		underlyingPrices = append(underlyingPrices, getMidPrice(stateDB, market))
	}
	return underlyingPrices
}

func GetReduceOnlyAmounts(stateDB contract.StateDB, trader common.Address) []*big.Int {
	numMarkets := GetActiveMarketsCount(stateDB)
	sizes := make([]*big.Int, numMarkets)
	for i := int64(0); i < numMarkets; i++ {
		sizes[i] = getReduceOnlyAmount(stateDB, trader, big.NewInt(i))
	}
	return sizes
}

func getPosSizes(stateDB contract.StateDB, trader *common.Address) []*big.Int {
	positionSizes := make([]*big.Int, 0)
	for _, market := range GetMarketsIncludingSettled(stateDB) {
		positionSizes = append(positionSizes, getSize(stateDB, market, trader))
	}
	return positionSizes
}

// GetMarketAddressFromMarketID returns the market address for a given marketID
func GetMarketAddressFromMarketID(marketID int64, stateDB contract.StateDB) common.Address {
	baseStorageSlot := marketsStorageSlot()
	amm := stateDB.GetState(common.HexToAddress(CLEARING_HOUSE_GENESIS_ADDRESS), common.BigToHash(new(big.Int).Add(baseStorageSlot, big.NewInt(marketID))))
	return common.BytesToAddress(amm.Bytes())
}

func getReferralAddress(stateDB contract.StateDB) common.Address {
	referral := stateDB.GetState(common.HexToAddress(CLEARING_HOUSE_GENESIS_ADDRESS), common.BigToHash(big.NewInt(REFERRAL_SLOT)))
	return common.BytesToAddress(referral.Bytes())
}
