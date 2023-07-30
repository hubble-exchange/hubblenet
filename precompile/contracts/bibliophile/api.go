package bibliophile

import (
	"math/big"

	"github.com/ava-labs/subnet-evm/precompile/contract"
	"github.com/ethereum/go-ethereum/common"
)

type VariablesReadFromClearingHouseSlots struct {
	MaintenanceMargin  *big.Int         `json:"maintenance_margin"`
	MinAllowableMargin *big.Int         `json:"min_allowable_margin"`
	Amms               []common.Address `json:"amms"`
}

func GetClearingHouseVariables(stateDB contract.StateDB) VariablesReadFromClearingHouseSlots {
	maintenanceMargin := GetMaintenanceMargin(stateDB)
	minAllowableMargin := GetMinAllowableMargin(stateDB)
	amms := GetMarkets(stateDB)
	return VariablesReadFromClearingHouseSlots{
		MaintenanceMargin:  maintenanceMargin,
		MinAllowableMargin: minAllowableMargin,
		Amms:               amms,
	}
}

type VariablesReadFromMarginAccountSlots struct {
	Margin *big.Int `json:"margin"`
}

func GetMarginAccountVariables(stateDB contract.StateDB, collateralIdx *big.Int, trader common.Address) VariablesReadFromMarginAccountSlots {
	margin := getMargin(stateDB, collateralIdx, trader)
	return VariablesReadFromMarginAccountSlots{
		Margin: margin,
	}
}

type VariablesReadFromAMMSlots struct {
	LastPrice                 *big.Int       `json:"last_price"`
	CumulativePremiumFraction *big.Int       `json:"cumulative_premium_fraction"`
	MaxOracleSpreadRatio      *big.Int       `json:"max_oracle_spread_ratio"`
	OracleAddress             common.Address `json:"oracle_address"`
	MaxLiquidationRatio       *big.Int       `json:"max_liquidation_ratio"`
	MinSizeRequirement        *big.Int       `json:"min_size_requirement"`
	UnderlyingAssetAddress    common.Address `json:"underlying_asset_address"`
	MaxLiquidationPriceSpread *big.Int       `json:"max_liquidation_price_spread"`
	RedStoneAdapterAddress    common.Address `json:"red_stone_adapter_address"`
	RedStoneFeedId            common.Hash    `json:"red_stone_feed_id"`
	Position                  Position       `json:"position"`
}

type Position struct {
	Size                 *big.Int `json:"size"`
	OpenNotional         *big.Int `json:"open_notional"`
	LastPremiumFraction  *big.Int `json:"last_premium_fraction"`
	LiquidationThreshold *big.Int `json:"liquidation_threshold"`
}

func GetAMMVariables(stateDB contract.StateDB, ammAddress common.Address, ammIndex int64, trader common.Address) VariablesReadFromAMMSlots {
	lastPrice := getLastPrice(stateDB, ammAddress)
	position := Position{
		Size:                getSize(stateDB, ammAddress, &trader),
		OpenNotional:        getOpenNotional(stateDB, ammAddress, &trader),
		LastPremiumFraction: GetLastPremiumFraction(stateDB, ammAddress, &trader),
	}
	cumulativePremiumFraction := GetCumulativePremiumFraction(stateDB, ammAddress)
	maxOracleSpreadRatio := GetMaxOraclePriceSpread(stateDB, ammIndex)
	maxLiquidationRatio := GetMaxLiquidationRatio(stateDB, ammIndex)
	minSizeRequirement := GetMinSizeRequirement(stateDB, ammIndex)
	oracleAddress := getOracleAddress(stateDB, ammAddress)
	underlyingAssetAddress := getUnderlyingAssetAddress(stateDB, ammAddress)
	maxLiquidationPriceSpread := GetMaxLiquidationPriceSpread(stateDB, ammIndex)
	redStoneAdapterAddress := getRedStoneAdapterAddress(stateDB, ammAddress)
	redStoneFeedId := getRedStoneFeedId(stateDB, ammAddress)
	return VariablesReadFromAMMSlots{
		LastPrice:                 lastPrice,
		CumulativePremiumFraction: cumulativePremiumFraction,
		MaxOracleSpreadRatio:      maxOracleSpreadRatio,
		OracleAddress:             oracleAddress,
		MaxLiquidationRatio:       maxLiquidationRatio,
		MinSizeRequirement:        minSizeRequirement,
		UnderlyingAssetAddress:    underlyingAssetAddress,
		MaxLiquidationPriceSpread: maxLiquidationPriceSpread,
		RedStoneAdapterAddress:    redStoneAdapterAddress,
		RedStoneFeedId:            redStoneFeedId,
		Position:                  position,
	}
}

type VariablesReadFromIOCOrdersSlots struct {
	OrderDetails     OrderDetails `json:"order_details"`
	IocExpirationCap *big.Int     `json:"ioc_expiration_cap"`
}

type OrderDetails struct {
	BlockPlaced  *big.Int `json:"block_placed"`
	FilledAmount *big.Int `json:"filled_amount"`
	OrderStatus  int64    `json:"order_status"`
}

func GetIOCOrdersVariables(stateDB contract.StateDB, orderHash common.Hash) VariablesReadFromIOCOrdersSlots {
	blockPlaced := iocGetBlockPlaced(stateDB, orderHash)
	filledAmount := iocGetOrderFilledAmount(stateDB, orderHash)
	orderStatus := iocGetOrderStatus(stateDB, orderHash)

	iocExpirationCap := iocGetExpirationCap(stateDB)
	return VariablesReadFromIOCOrdersSlots{
		OrderDetails: OrderDetails{
			BlockPlaced:  blockPlaced,
			FilledAmount: filledAmount,
			OrderStatus:  orderStatus,
		},
		IocExpirationCap: iocExpirationCap,
	}
}

type VariablesReadFromOrderbookSlots struct {
	OrderDetails      OrderDetails `json:"order_details"`
	IsTradingAuthoriy bool         `json:"is_trading_authority"`
}

func GetOrderBookVariables(stateDB contract.StateDB, traderAddress string, senderAddress string, orderHash common.Hash) VariablesReadFromOrderbookSlots {
	blockPlaced := getBlockPlaced(stateDB, orderHash)
	filledAmount := getOrderFilledAmount(stateDB, orderHash)
	orderStatus := getOrderStatus(stateDB, orderHash)
	isTradingAuthoriy := IsTradingAuthority(stateDB, common.HexToAddress(traderAddress), common.HexToAddress(senderAddress))
	return VariablesReadFromOrderbookSlots{
		OrderDetails: OrderDetails{
			BlockPlaced:  blockPlaced,
			FilledAmount: filledAmount,
			OrderStatus:  orderStatus,
		},
		IsTradingAuthoriy: isTradingAuthoriy,
	}
}
