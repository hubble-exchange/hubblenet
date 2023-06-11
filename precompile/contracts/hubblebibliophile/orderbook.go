package hubblebibliophile

import (
	"fmt"
	"math/big"

	"github.com/ava-labs/subnet-evm/precompile/contract"
	"github.com/ava-labs/subnet-evm/utils"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	ORDERBOOK_GENESIS_ADDRESS       = "0x0300000000000000000000000000000000000000"
	ORDER_INFO_SLOT           int64 = 53
)

func ValidateOrdersAndDetermineFillPrice(stateDB contract.StateDB, inputStruct *ValidateOrdersAndDetermineFillPriceInput) (*ValidateOrdersAndDetermineFillPriceOutput, error) {
	longOrder := inputStruct.Orders[0]
	shortOrder := inputStruct.Orders[0]

	if longOrder.BaseAssetQuantity.Cmp(big.NewInt(0)) <= 0 {
		return nil, fmt.Errorf("OB_order_0_is_not_long")
	}

	if shortOrder.BaseAssetQuantity.Cmp(big.NewInt(0)) >= 0 {
		return nil, fmt.Errorf("OB_order_1_is_not_short")
	}

	if longOrder.AmmIndex.Cmp(shortOrder.AmmIndex) != 0 {
		return nil, fmt.Errorf("OB_orders_for_different_amms")
	}

	if longOrder.Price.Cmp(shortOrder.Price) == -1 {
		return nil, fmt.Errorf("OB_orders_do_not_match")
	}

	market := getMarketAddressFromMarketID(longOrder.AmmIndex.Int64(), stateDB)
	oraclePrice := getUnderlyingPrice(stateDB, market)
	spreadLimit := GetMaxOracleSpreadRatioForMarket(stateDB, longOrder.AmmIndex.Int64())
	upperbound := divide1e6(new(big.Int).Mul(oraclePrice, new(big.Int).Add(_1e6, spreadLimit)))
	lowerbound := big.NewInt(0)
	if spreadLimit.Cmp(_1e6) == -1 {
		lowerbound = divide1e6(new(big.Int).Mul(oraclePrice, new(big.Int).Sub(_1e6, spreadLimit)))
	}
	if longOrder.Price.Cmp(lowerbound) == -1 {
		return nil, fmt.Errorf("OB_long_order_price_too_low")
	}
	if shortOrder.Price.Cmp(upperbound) == 1 {
		return nil, fmt.Errorf("OB_short_order_price_too_high")
	}

	// CUSTOM CODE FOR AN OUTPUT
	output := ValidateOrdersAndDetermineFillPriceOutput{}

	blockPlaced0 := getBlockPlaced(stateDB, inputStruct.OrderHashes[0])
	blockPlaced1 := getBlockPlaced(stateDB, inputStruct.OrderHashes[1])

	if blockPlaced0.Cmp(blockPlaced1) == -1 {
		// long order is the maker order
		output.FillPrice = utils.BigIntMin(longOrder.Price, upperbound)
		output.Mode0 = 1 // Mode0 corresponds to the long order and `1` is maker
		output.Mode1 = 0 // Mode1 corresponds to the short order and `0` is taker
	} else { // if long order is placed after short order or in the same block as short
		// short order is the maker order
		output.FillPrice = utils.BigIntMax(shortOrder.Price, lowerbound)
		output.Mode0 = 0 // Mode0 corresponds to the long order and `0` is taker
		output.Mode1 = 1 // Mode1 corresponds to the short order and `1` is maker
	}

	// CUSTOM CODE ENDS HERE
	return &output, nil
}

func getBlockPlaced(stateDB contract.StateDB, orderHash [32]byte) *big.Int {
	orderInfo := orderInfoMappingStorageSlot(orderHash)
	return new(big.Int).SetBytes(stateDB.GetState(common.HexToAddress(ORDERBOOK_GENESIS_ADDRESS), common.BigToHash(orderInfo)).Bytes())
}

func getOrderStatus(stateDB contract.StateDB, orderHash [32]byte) *big.Int {
	orderInfo := orderInfoMappingStorageSlot(orderHash)
	return new(big.Int).SetBytes(stateDB.GetState(common.HexToAddress(ORDERBOOK_GENESIS_ADDRESS), common.BigToHash(new(big.Int).Add(orderInfo, big.NewInt(3)))).Bytes())
}

func orderInfoMappingStorageSlot(orderHash [32]byte) *big.Int {
	return new(big.Int).SetBytes(crypto.Keccak256(append(orderHash[:], common.LeftPadBytes(big.NewInt(ORDER_INFO_SLOT).Bytes(), 32)...)))
}
