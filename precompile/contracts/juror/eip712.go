package juror

import (
	"fmt"

	"github.com/ava-labs/subnet-evm/precompile/contracts/bibliophile"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

func GetLimitOrderHash(o *LimitOrder) (hash common.Hash, err error) {
	message := map[string]interface{}{
		"ammIndex":          o.AmmIndex.String(),
		"trader":            o.Trader.String(),
		"baseAssetQuantity": o.BaseAssetQuantity.String(),
		"price":             o.Price.String(),
		"salt":              o.Salt.String(),
		"reduceOnly":        o.ReduceOnly,
	}
	domain := apitypes.TypedDataDomain{
		Name:              "Hubble",
		Version:           "2.0",
		ChainId:           math.NewHexOrDecimal256(321123), // @todo chain id
		VerifyingContract: common.HexToAddress(bibliophile.ORDERBOOK_GENESIS_ADDRESS).String(),
	}
	typedData := apitypes.TypedData{
		Types:       Eip712OrderTypes,
		PrimaryType: "LimitOrder",
		Domain:      domain,
		Message:     message,
	}
	return EncodeForSigning(typedData)
}

func getIOCOrderHash(o *IOCOrder) (hash common.Hash, err error) {
	message := map[string]interface{}{
		"orderType":         uint8(o.OrderType),
		"ammIndex":          o.AmmIndex.String(),
		"trader":            o.Trader.String(),
		"baseAssetQuantity": o.BaseAssetQuantity.String(),
		"price":             o.Price.String(),
		"salt":              o.Salt.String(),
		"reduceOnly":        o.ReduceOnly,
	}
	domain := apitypes.TypedDataDomain{
		Name:              "Hubble",
		Version:           "2.0",
		ChainId:           math.NewHexOrDecimal256(321123), // @todo chain id
		VerifyingContract: common.HexToAddress(bibliophile.ORDERBOOK_GENESIS_ADDRESS).String(),
	}
	typedData := apitypes.TypedData{
		Types:       Eip712OrderTypes,
		PrimaryType: "IOCOrder",
		Domain:      domain,
		Message:     message,
	}
	return EncodeForSigning(typedData)
}

// EncodeForSigning - Encoding the typed data
func EncodeForSigning(typedData apitypes.TypedData) (hash common.Hash, err error) {
	domainSeparator, err := typedData.HashStruct("EIP712Domain", typedData.Domain.Map())
	if err != nil {
		return
	}
	typedDataHash, err := typedData.HashStruct(typedData.PrimaryType, typedData.Message)
	if err != nil {
		return
	}
	rawData := []byte(fmt.Sprintf("\x19\x01%s%s", string(domainSeparator), string(typedDataHash)))
	hash = common.BytesToHash(crypto.Keccak256(rawData))
	return
}

var Eip712OrderTypes = apitypes.Types{
	"EIP712Domain": {
		{
			Name: "name",
			Type: "string",
		},
		{
			Name: "version",
			Type: "string",
		},
		{
			Name: "chainId",
			Type: "uint256",
		},
		{
			Name: "verifyingContract",
			Type: "address",
		},
	},
	"LimitOrder": {
		{
			Name: "ammIndex",
			Type: "uint256",
		},
		{
			Name: "trader",
			Type: "address",
		},
		{
			Name: "baseAssetQuantity",
			Type: "int256",
		},
		{
			Name: "price",
			Type: "uint256",
		},
		{
			Name: "salt",
			Type: "uint256",
		},
		{
			Name: "reduceOnly",
			Type: "bool",
		},
	},
	"IOCOrder": {
		{
			Name: "orderType",
			Type: "uint8",
		},
		{
			Name: "ammIndex",
			Type: "uint256",
		},
		{
			Name: "trader",
			Type: "address",
		},
		{
			Name: "baseAssetQuantity",
			Type: "int256",
		},
		{
			Name: "price",
			Type: "uint256",
		},
		{
			Name: "salt",
			Type: "uint256",
		},
		{
			Name: "reduceOnly",
			Type: "bool",
		},
	},
}
