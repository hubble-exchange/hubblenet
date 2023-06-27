package rollupprecompile

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ava-labs/subnet-evm/accounts/abi"
	"github.com/ava-labs/subnet-evm/precompile/contract"
	"github.com/ava-labs/subnet-evm/vmerrs"

	_ "embed"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

var (
	OrderExpired    = errors.New("order is expired")
	InvalidSigner   = errors.New("invalid signer")
	OrderCancelled  = errors.New("order is cancelled")
	OrderOverfilled = errors.New("order_overfilled")
)

const (
	ORDER_STATUS_SLOT         = 2
	IS_TRADING_AUTHORITY_SLOT = 3
)

func ValidateOrdersAndDetermineFillPrice(stateDB contract.StateDB, inputStruct *ValidateOrdersAndDetermineFillPriceInput) (*ValidateOrdersAndDetermineFillPriceOutput, error) {
	longOrder := inputStruct.Orders[0]
	shortOrder := inputStruct.Orders[1]

	// @todo is validator check
	orderHash0, err := ValidateOrder(longOrder)
	if err != nil {
		return nil, err
	}
	output := &ValidateOrdersAndDetermineFillPriceOutput{}
	output.Instructions[0] = IClearingHouseInstruction{
		AmmIndex: longOrder.AmmIndex,
		Trader:  longOrder.Trader,
		OrderHash: orderHash0
	}

	output.Instructions[0].AmmIndex = longOrder.AmmIndex
	return output, nil

}

func ValidateOrder(stateDB contract.StateDB, input *ValidateOrderInput) ([32]byte, error) {
	timestamp := time.Now().Unix()

	if input.Order.ValidUntil.Cmp(big.NewInt(timestamp)) < 0 {
		return [32]byte{}, OrderExpired
	}

	signer, orderHash, err := verifySigner(input.Order, input.Signature)
	if err != nil {
		return [32]byte{}, err
	}

	if !strings.EqualFold(signer.String(), input.Order.Trader.String()) && !IsTradingAuthority(stateDB, signer, input.Order.Trader, input.Orderbook) {
		return [32]byte{}, InvalidSigner
	}

	filledAmount, isCancelled := GetOrderStatus(stateDB, orderHash, input.Orderbook)
	if isCancelled {
		return [32]byte{}, OrderCancelled
	}

	if input.OrderType == 0 {
		if input.FillAmount.Sign() < 0 {
			return [32]byte{}, errors.New("invalid_fill_amount")
		}
		if input.Order.BaseAssetQuantity.Sign() <= 0 {
			return [32]byte{}, errors.New("invalid_base_asset_quantity")
		}
		if new(big.Int).Add(filledAmount, input.FillAmount).Cmp(input.Order.BaseAssetQuantity) > 0 {
			return [32]byte{}, OrderOverfilled
		}
	} else {
		if input.FillAmount.Sign() > 0 {
			return [32]byte{}, errors.New("invalid_fill_amount")
		}
		if input.Order.BaseAssetQuantity.Sign() >= 0 {
			return [32]byte{}, errors.New("invalid_quote_asset_quantity")
		}
		if new(big.Int).Add(filledAmount, input.FillAmount).Cmp(input.Order.BaseAssetQuantity) < 0 {
			return [32]byte{}, OrderOverfilled
		}
	}
	return [32]byte{}, nil
}

func IsTradingAuthority(stateDB contract.StateDB, signer, trader, orderbook common.Address) bool {
	tradingAuthorityMappingSlot := crypto.Keccak256(append(common.LeftPadBytes(trader.Bytes(), 32), common.LeftPadBytes(big.NewInt(IS_TRADING_AUTHORITY_SLOT).Bytes(), 32)...))
	tradingAuthorityMappingSlot = crypto.Keccak256(append(common.LeftPadBytes(signer.Bytes(), 32), tradingAuthorityMappingSlot...))
	return stateDB.GetState(orderbook, common.BytesToHash(tradingAuthorityMappingSlot)).Big().Cmp(big.NewInt(1)) == 0
}

func GetOrderStatus(stateDB contract.StateDB, orderHash common.Hash, orderbook common.Address) (*big.Int, bool) {
	orderStatusMappingSlot := crypto.Keccak256(append(orderHash.Bytes(), common.LeftPadBytes(big.NewInt(ORDER_STATUS_SLOT).Bytes(), 32)...))
	filledAmount := stateDB.GetState(orderbook, common.BytesToHash(orderStatusMappingSlot)).Big()
	isCancelled := stateDB.GetState(orderbook, common.BigToHash(new(big.Int).Add(new(big.Int).SetBytes(orderStatusMappingSlot), big.NewInt(1)))).Big().Cmp(big.NewInt(1)) == 0
	return filledAmount, isCancelled
}

func verifySigner(order IOrderBookRollupOrder, signature []byte) (common.Address, common.Hash, error) {
	hash, err := GetOrderHash(order)
	if err != nil {
		return common.Address{}, common.Hash{}, err
	}
	signer, err := crypto.SigToPub(hash.Bytes(), signature)
	if err != nil {
		return common.Address{}, common.Hash{}, err
	}
	return crypto.PubkeyToAddress(*signer), hash, nil
}

func GetOrderHash(o IOrderBookRollupOrder) (hash common.Hash, err error) {
	message := map[string]interface{}{
		"ammIndex":          o.AmmIndex.String(),
		"trader":            o.Trader,
		"baseAssetQuantity": o.BaseAssetQuantity.String(),
		"price":             o.Price.String(),
		"salt":              o.Salt.String(),
		"reduceOnly":        o.ReduceOnly,
		"validUntil":        o.ValidUntil.String(),
	}
	domain := apitypes.TypedDataDomain{
		Name:              "Hubble",
		Version:           "v2",
		ChainId:           math.NewHexOrDecimal256(321123),
		VerifyingContract: ContractAddress.String(),
	}
	typedData := apitypes.TypedData{
		Types:       Eip712OrderTypes,
		PrimaryType: "Order",
		Domain:      domain,
		Message:     message,
	}
	hash, err = EncodeForSigning(typedData)
}
