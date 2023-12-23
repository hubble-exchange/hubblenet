package hubbleutils

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"

	"github.com/ava-labs/subnet-evm/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

// func getSignedOrderBookAddress() common.Address {
// 	if
// }

type SignedOrder struct {
	LimitOrder
	OrderType uint8    `json:"orderType"`
	ExpireAt  *big.Int `json:"expireAt"`
	Sig       []byte   `json:"sig"`
}

func (order *SignedOrder) EncodeToABIWithoutType() ([]byte, error) {
	signedOrderType, err := getOrderType("signed")
	if err != nil {
		return nil, err
	}
	encodedOrder, err := abi.Arguments{{Type: signedOrderType}}.Pack(order)
	if err != nil {
		return nil, err
	}
	return encodedOrder, nil
}

func (order *SignedOrder) EncodeToABI() ([]byte, error) {
	encodedSignedOrder, err := order.EncodeToABIWithoutType()
	if err != nil {
		return nil, fmt.Errorf("failed getting abi type: %w", err)
	}

	orderType, _ := abi.NewType("uint8", "uint8", nil)
	orderBytesType, _ := abi.NewType("bytes", "bytes", nil)
	encodedOrder, err := abi.Arguments{{Type: orderType}, {Type: orderBytesType}}.Pack(uint8(Signed), encodedSignedOrder)
	if err != nil {
		return nil, fmt.Errorf("order encoding failed: %w", err)
	}

	return encodedOrder, nil
}

func (order *SignedOrder) DecodeFromRawOrder(rawOrder interface{}) {
	marshalledOrder, _ := json.Marshal(rawOrder)
	fmt.Println("marshalledOrder", string(marshalledOrder))
	err := json.Unmarshal(marshalledOrder, &order)
	if err != nil {
		fmt.Println("err in DecodeFromRawOrder")
		fmt.Println(err)
	}
}

// func (order *SignedOrder) Map() map[string]interface{} {
// 	return map[string]interface{}{
// 		"ammIndex":          order.AmmIndex,
// 		"trader":            order.Trader,
// 		"baseAssetQuantity": utils.BigIntToFloat(order.BaseAssetQuantity, 18),
// 		"price":             utils.BigIntToFloat(order.Price, 6),
// 		"reduceOnly":        order.ReduceOnly,
// 		"postOnly":          order.PostOnly,
// 		"salt":              order.Salt,
// 		"orderType":         order.OrderType,
// 		"expireAt":          order.ExpireAt,
// 		"sig":               order.Sig,
// 	}
// }

func DecodeSignedOrder(encodedOrder []byte) (*SignedOrder, error) {
	orderType, err := getOrderType("signed")
	if err != nil {
		return nil, fmt.Errorf("failed getting abi type: %w", err)
	}
	order, err := abi.Arguments{{Type: orderType}}.Unpack(encodedOrder)
	if err != nil {
		return nil, err
	}
	signedOrder := &SignedOrder{}
	signedOrder.DecodeFromRawOrder(order[0])
	return signedOrder, nil
}

func (o *SignedOrder) Hash(verifyingContract common.Address) (hash common.Hash, err error) {
	message := map[string]interface{}{
		"orderType":         strconv.FormatUint(uint64(o.OrderType), 10),
		"expireAt":          o.ExpireAt.String(),
		"ammIndex":          o.AmmIndex.String(),
		"trader":            o.Trader.String(),
		"baseAssetQuantity": o.BaseAssetQuantity.String(),
		"price":             o.Price.String(),
		"salt":              o.Salt.String(),
		"reduceOnly":        o.ReduceOnly,
		"postOnly":          o.PostOnly,
	}
	domain := apitypes.TypedDataDomain{
		Name:              "Hubble",
		Version:           "2.0",
		ChainId:           math.NewHexOrDecimal256(321123), // @todo chain id from config
		VerifyingContract: common.HexToAddress("0x809d550fca64d94Bd9F66E60752A544199cfAC3D").String(),
	}
	typedData := apitypes.TypedData{
		Types:       Eip712OrderTypes,
		PrimaryType: "Order",
		Domain:      domain,
		Message:     message,
	}
	return EncodeForSigning(typedData)
}

// Trading API methods

func (o *SignedOrder) UnmarshalJSON(data []byte) error {
	// Redefine the structs with simple types for JSON unmarshalling
	aux := &struct {
		AmmIndex          uint64         `json:"ammIndex"`
		Trader            common.Address `json:"trader"`
		BaseAssetQuantity string         `json:"baseAssetQuantity"`
		Price             string         `json:"price"`
		Salt              string         `json:"salt"`
		ReduceOnly        bool           `json:"reduceOnly"`
		PostOnly          bool           `json:"postOnly"`
		OrderType         uint8          `json:"orderType"`
		ExpireAt          uint64         `json:"expireAt"`
		Sig               string         `json:"sig"`
	}{}

	// Perform the unmarshalling
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	// Convert and assign the values to the original struct
	o.AmmIndex = new(big.Int).SetUint64(aux.AmmIndex)

	o.Trader = aux.Trader

	o.BaseAssetQuantity = new(big.Int)
	o.BaseAssetQuantity.SetString(aux.BaseAssetQuantity, 10)

	o.Price = new(big.Int)
	o.Price.SetString(aux.Price, 10)

	o.Salt = new(big.Int)
	o.Salt.SetBytes(common.FromHex(aux.Salt))

	o.ReduceOnly = aux.ReduceOnly
	o.PostOnly = aux.PostOnly
	o.OrderType = aux.OrderType

	o.ExpireAt = new(big.Int).SetUint64(aux.ExpireAt)
	o.Sig = common.FromHex(aux.Sig)
	return nil
}

func (order *SignedOrder) DecodeAPIOrder(rawOrder interface{}) error {
	order_, ok := rawOrder.(string)
	if !ok {
		fmt.Println("invalid data format")
	}

	orderJson := []byte(order_)
	err := json.Unmarshal(orderJson, &order)
	if err != nil {
		return err
	}
	return nil
}
