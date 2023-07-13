package orderbook

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ava-labs/subnet-evm/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

type ContractOrder interface {
	EncodeToABI() ([]byte, error)
	DecodeFromRawOrder(rawOrder interface{})
}

// LimitOrder type is copy of LimitOrder struct defined in Orderbook contract
type LimitOrder struct {
	AmmIndex          *big.Int       `json:"ammIndex"`
	Trader            common.Address `json:"trader"`
	BaseAssetQuantity *big.Int       `json:"baseAssetQuantity"`
	Price             *big.Int       `json:"price"`
	Salt              *big.Int       `json:"salt"`
	ReduceOnly        bool           `json:"reduceOnly"`
}

// IOCOrder type is copy of IOCOrder struct defined in Orderbook contract
type IOCOrder struct {
	LimitOrder
	OrderType uint8    `json:"orderType"`
	ExpireAt  *big.Int `json:"expireAt"`
}

// LimitOrder
func (order *LimitOrder) EncodeToABI() ([]byte, error) {
	limitOrderType, _ := abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{Name: "ammIndex", Type: "uint256"},
		{Name: "trader", Type: "address"},
		{Name: "baseAssetQuantity", Type: "int256"},
		{Name: "price", Type: "uint256"},
		{Name: "salt", Type: "uint256"},
		{Name: "reduceOnly", Type: "bool"},
	})

	encodedLimitOrder, err := abi.Arguments{{Type: limitOrderType}}.Pack(order)
	if err != nil {
		return nil, fmt.Errorf("limit order packing failed: %w", err)
	}

	orderType, _ := abi.NewType("uint8", "uint8", nil)
	orderBytesType, _ := abi.NewType("bytes", "bytes", nil)
	// 0 means ordertype = limit order
	encodedOrder, err := abi.Arguments{{Type: orderType}, {Type: orderBytesType}}.Pack(uint8(0), encodedLimitOrder)
	if err != nil {
		return nil, fmt.Errorf("order encoding failed: %w", err)
	}

	return encodedOrder, nil
}

func (order *LimitOrder) DecodeFromRawOrder(rawOrder interface{}) {
	marshalledOrder, _ := json.Marshal(rawOrder)
	json.Unmarshal(marshalledOrder, &order)
}

func DecodeLimitOrder(encodedOrder []byte) (*LimitOrder, error) {
	limitOrderType, _ := abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{Name: "ammIndex", Type: "uint256"},
		{Name: "trader", Type: "address"},
		{Name: "baseAssetQuantity", Type: "int256"},
		{Name: "price", Type: "uint256"},
		{Name: "salt", Type: "uint256"},
		{Name: "reduceOnly", Type: "bool"},
	})
	order, err := abi.Arguments{{Type: limitOrderType}}.Unpack(encodedOrder)
	if err != nil {
		return nil, err
	}
	source, ok := order[0].(struct {
		AmmIndex          *big.Int       `json:"ammIndex"`
		Trader            common.Address `json:"trader"`
		BaseAssetQuantity *big.Int       `json:"baseAssetQuantity"`
		Price             *big.Int       `json:"price"`
		Salt              *big.Int       `json:"salt"`
		ReduceOnly        bool           `json:"reduceOnly"`
	})
	if !ok {
		return nil, fmt.Errorf("order decoding failed: %w", err)
	}
	return &LimitOrder{
		AmmIndex:          source.AmmIndex,
		Trader:            source.Trader,
		BaseAssetQuantity: source.BaseAssetQuantity,
		Price:             source.Price,
		Salt:              source.Salt,
		ReduceOnly:        source.ReduceOnly,
	}, nil
}

// ----------------------------------------------------------------------------

// IOCOrder

func (order *IOCOrder) EncodeToABI() ([]byte, error) {
	iocOrderType, _ := abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{Name: "orderType", Type: "uint8"},
		{Name: "expireAt", Type: "uint256"},
		{Name: "ammIndex", Type: "uint256"},
		{Name: "trader", Type: "address"},
		{Name: "baseAssetQuantity", Type: "int256"},
		{Name: "price", Type: "uint256"},
		{Name: "salt", Type: "uint256"},
		{Name: "reduceOnly", Type: "bool"},
	})

	encodedIOCOrder, err := abi.Arguments{{Type: iocOrderType}}.Pack(order)
	if err != nil {
		return nil, fmt.Errorf("limit order packing failed: %w", err)
	}

	orderType, _ := abi.NewType("uint8", "uint8", nil)
	orderBytesType, _ := abi.NewType("bytes", "bytes", nil)
	// 1 means ordertype = IOC/market order
	encodedOrder, err := abi.Arguments{{Type: orderType}, {Type: orderBytesType}}.Pack(uint8(1), encodedIOCOrder)
	if err != nil {
		return nil, fmt.Errorf("order encoding failed: %w", err)
	}

	return encodedOrder, nil
}

func (order *IOCOrder) DecodeFromRawOrder(rawOrder interface{}) {
	marshalledOrder, _ := json.Marshal(rawOrder)
	json.Unmarshal(marshalledOrder, &order)
}

func DecodeIOCOrder(encodedOrder []byte) (*IOCOrder, error) {
	limitOrderType, _ := abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{Name: "orderType", Type: "uint8"},
		{Name: "expireAt", Type: "uint256"},
		{Name: "ammIndex", Type: "uint256"},
		{Name: "trader", Type: "address"},
		{Name: "baseAssetQuantity", Type: "int256"},
		{Name: "price", Type: "uint256"},
		{Name: "salt", Type: "uint256"},
		{Name: "reduceOnly", Type: "bool"},
	})
	order, err := abi.Arguments{{Type: limitOrderType}}.Unpack(encodedOrder)
	if err != nil {
		return nil, err
	}
	source, ok := order[0].(struct {
		OrderType         uint8          `json:"orderType"`
		ExpireAt          *big.Int       `json:"expireAt"`
		AmmIndex          *big.Int       `json:"ammIndex"`
		Trader            common.Address `json:"trader"`
		BaseAssetQuantity *big.Int       `json:"baseAssetQuantity"`
		Price             *big.Int       `json:"price"`
		Salt              *big.Int       `json:"salt"`
		ReduceOnly        bool           `json:"reduceOnly"`
	})
	if !ok {
		return nil, fmt.Errorf("order decoding failed: %w", err)
	}
	return &IOCOrder{
		OrderType: source.OrderType,
		ExpireAt:  source.ExpireAt,
		LimitOrder: LimitOrder{
			AmmIndex:          source.AmmIndex,
			Trader:            source.Trader,
			BaseAssetQuantity: source.BaseAssetQuantity,
			Price:             source.Price,
			Salt:              source.Salt,
			ReduceOnly:        source.ReduceOnly,
		},
	}, nil
}
