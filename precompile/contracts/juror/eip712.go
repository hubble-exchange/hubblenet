package juror

import (
	"fmt"
	"strconv"

	"github.com/ava-labs/subnet-evm/accounts/abi"
	"github.com/ava-labs/subnet-evm/plugin/evm/orderbook"
	"github.com/ava-labs/subnet-evm/precompile/contracts/bibliophile"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

func GetLimitOrderHash(o *orderbook.LimitOrder) (hash common.Hash, err error) {
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
		ChainId:           math.NewHexOrDecimal256(321123), // @todo chain id from config
		VerifyingContract: common.HexToAddress(bibliophile.ORDERBOOK_GENESIS_ADDRESS).String(),
	}
	typedData := apitypes.TypedData{
		Types:       Eip712OrderTypes,
		PrimaryType: "Order",
		Domain:      domain,
		Message:     message,
	}
	return EncodeForSigning(typedData)
}

func GetLimitOrderV2Hash(o *ILimitOrderBookOrderV2) (common.Hash, error) {
	// bytes32 ORDERV2_TYPEHASH = crypto.keccak256("OrderV2(uint256 ammIndex,address trader,int256 baseAssetQuantity,uint256 price,uint256 salt,bool reduceOnly,bool postOnly)");
	// return keccak256(abi.encode(ORDERV2_TYPEHASH, order));
	limitOrderType, err := abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{Name: "ammIndex", Type: "uint256"},
		{Name: "trader", Type: "address"},
		{Name: "baseAssetQuantity", Type: "int256"},
		{Name: "price", Type: "uint256"},
		{Name: "salt", Type: "uint256"},
		{Name: "reduceOnly", Type: "bool"},
		{Name: "postOnly", Type: "bool"},
	})
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed getting abi type: %w", err)
	}
	typeHash, _ := abi.NewType("bytes32", "", nil)
	args := abi.Arguments{
		{Type: typeHash, Name: "TypeHash"},
		{Type: limitOrderType, Name: "LimitOrderType"},
	}
	var hash [32]byte
	copy(hash[:], crypto.Keccak256([]byte("OrderV2(uint256 ammIndex,address trader,int256 baseAssetQuantity,uint256 price,uint256 salt,bool reduceOnly,bool postOnly)")))
	encodedData, err := args.Pack(hash, o)
	if err != nil {
		return common.Hash{}, fmt.Errorf("limit order v2 packing failed: %w", err)
	}
	return crypto.Keccak256Hash(encodedData), nil
}

func getIOCOrderHash(o *orderbook.IOCOrder) (hash common.Hash, err error) {
	message := map[string]interface{}{
		"orderType":         strconv.FormatUint(uint64(o.OrderType), 10),
		"expireAt":          o.ExpireAt.String(),
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
		ChainId:           math.NewHexOrDecimal256(321123), // @todo chain id from config
		VerifyingContract: common.HexToAddress(bibliophile.IOC_ORDERBOOK_ADDRESS).String(),
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
	"Order": { // has to be same as the struct name or whatever was passed when building the typed hash
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
	"OrderV2": {
		{
			Name: "ammIndex",
			Type: "uint256",
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
		{
			Name: "postOnly",
			Type: "bool",
		},
		{
			Name: "trader",
			Type: "address",
		},
	},
	"IOCOrder": {
		{
			Name: "orderType",
			Type: "uint8",
		},
		{
			Name: "expireAt",
			Type: "uint256",
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
