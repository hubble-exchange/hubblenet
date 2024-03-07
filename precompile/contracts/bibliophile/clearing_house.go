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
	// @todo when we ever settle a market, here it needs to be taken care of
	// because currently the following assumes that all markets are active
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
	RequiredMargin   *big.Int
}

type GetTraderDataForMarketOutput struct {
	NotionalPosition *big.Int
	RequiredMargin   *big.Int
	UnrealizedPnl    *big.Int
	PendingFunding   *big.Int
	IsIsolated 	 	 bool
}

func getNotionalPositionAndMargin(stateDB contract.StateDB, input *GetNotionalPositionAndMarginInput, upgradeVersion hu.UpgradeVersion) GetNotionalPositionAndMarginOutput {
	markets := GetMarkets(stateDB)
	numMarkets := len(markets)
	positions := make(map[int]*hu.Position, numMarkets)
	underlyingPrices := make(map[int]*big.Int, numMarkets)
	midPrices := make(map[int]*big.Int, numMarkets)
	var activeMarketIds []int
	for i, market := range markets {
		positions[i] = getPosition(stateDB, GetMarketAddressFromMarketID(int64(i), stateDB), &input.Trader)
		underlyingPrices[i] = getUnderlyingPrice(stateDB, market)
		midPrices[i] = getMidPrice(stateDB, market)
		activeMarketIds = append(activeMarketIds, i)
	}
	pendingFunding := big.NewInt(0)
	if input.IncludeFundingPayments {
		pendingFunding = GetTotalFunding(stateDB, &input.Trader)
	}
	notionalPosition, margin := hu.GetNotionalPositionAndMargin(
		&hu.HubbleState{
			Assets:         GetCollaterals(stateDB),
			OraclePrices:   underlyingPrices,
			MidPrices:      midPrices,
			ActiveMarkets:  activeMarketIds,
			UpgradeVersion: upgradeVersion,
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

func getNotionalPositionAndRequiredMargin(stateDB contract.StateDB, input *GetNotionalPositionAndMarginInput, upgradeVersion hu.UpgradeVersion) GetNotionalPositionAndMarginOutput {
	positions, underlyingPrices, accountPreferences, activeMarketIds := getMarketsDataFromDB(stateDB, &input.Trader, input.Mode)
	pendingFunding := big.NewInt(0)
	if input.IncludeFundingPayments {
		pendingFunding = getTotalFundingForCrossMarginPositions(stateDB, &input.Trader)
	}
	notionalPosition, margin, requiredMargin := hu.GetNotionalPositionAndRequiredMargin(
		&hu.HubbleState{
			Assets:         GetCollaterals(stateDB),
			OraclePrices:   underlyingPrices,
			ActiveMarkets:  activeMarketIds,
			UpgradeVersion: upgradeVersion,
		},
		&hu.UserState{
			Positions:      positions,
			Margins:        getMargins(stateDB, input.Trader),
			PendingFunding: pendingFunding,
			AccountPreferences: accountPreferences,
		},
	)
	return GetNotionalPositionAndMarginOutput{
		NotionalPosition: notionalPosition,
		Margin:           margin,
		RequiredMargin:   requiredMargin,
	}
}

func getCrossMarginAccountData(stateDB contract.StateDB, trader *common.Address, mode uint8, upgradeVersion hu.UpgradeVersion) GetTraderDataForMarketOutput {
	positions, underlyingPrices, accountPreferences, activeMarketIds := getMarketsDataFromDB(stateDB, trader, mode)
	notionalPosition, requiredMargin, unrealizedPnl := hu.GetCrossMarginAccountData(
		&hu.HubbleState{
			ActiveMarkets:  activeMarketIds,
			OraclePrices:   underlyingPrices,
			UpgradeVersion: upgradeVersion,
		},
		&hu.UserState{
			Positions:          positions,
			AccountPreferences: accountPreferences,
		},
	)
	pendingFunding := getTotalFundingForCrossMarginPositions(stateDB, trader)
	return GetTraderDataForMarketOutput {
		NotionalPosition: notionalPosition,
		RequiredMargin:   requiredMargin,
		UnrealizedPnl:    unrealizedPnl,
		PendingFunding:   pendingFunding,
	}
}

func getMarketsDataFromDB(stateDB contract.StateDB, trader *common.Address, mode uint8) (positions map[int]*hu.Position, underlyingPrices map[int]*big.Int, accountPreferences map[int]*hu.AccountPreferences, activeMarketIds []int) {
	markets := GetMarkets(stateDB)
	numMarkets := len(markets)
	positions = make(map[int]*hu.Position, numMarkets)
	underlyingPrices = make(map[int]*big.Int, numMarkets)
	accountPreferences = make(map[int]*hu.AccountPreferences, numMarkets)
	activeMarketIds = make([]int, numMarkets)
	for i, market := range markets {
		// @todo can use `market` instead of `GetMarketAddressFromMarketID`?
		positions[i] = getPosition(stateDB, GetMarketAddressFromMarketID(int64(i), stateDB), trader)
		underlyingPrices[i] = getUnderlyingPrice(stateDB, market)
		activeMarketIds[i] = i
		accountPreferences[i].MarginType = getMarginType(stateDB, market, trader)
		accountPreferences[i].MarginFraction = getMarginFractionByMode(stateDB, market, trader, mode)
	}
	return positions, underlyingPrices, accountPreferences, activeMarketIds
}

func getTotalFundingForCrossMarginPositions(stateDB contract.StateDB, trader *common.Address) *big.Int {
	totalFunding := big.NewInt(0)
	for _, market := range GetMarkets(stateDB) {
		if getMarginType(stateDB, market, trader) == hu.Cross_Margin {
			totalFunding.Add(totalFunding, getPendingFundingPayment(stateDB, market, trader))
		}
	}
	return totalFunding
}

func getTraderDataForMarket(stateDB contract.StateDB, trader *common.Address, marketId int64, mode uint8) GetTraderDataForMarketOutput {
	market := GetMarketAddressFromMarketID(marketId, stateDB)
	position := getPosition(stateDB, market, trader)
	marginFraction := getMarginFractionByMode(stateDB, market, trader, mode)
	underlyingPrice := getUnderlyingPrice(stateDB, market)
	notionalPosition, unrealizedPnl, requiredMargin := hu.GetTraderPositionDetails(position, underlyingPrice, marginFraction)
	pendingFunding := getPendingFundingPayment(stateDB, market, trader)
	isIsolated := getMarginType(stateDB, market, trader) == hu.Isolated_Margin
	return GetTraderDataForMarketOutput{
		IsIsolated:       isIsolated,
		NotionalPosition: notionalPosition,
		RequiredMargin:   requiredMargin,
		UnrealizedPnl:    unrealizedPnl,
		PendingFunding:   pendingFunding,
	}
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
	for _, market := range GetMarkets(stateDB) {
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
