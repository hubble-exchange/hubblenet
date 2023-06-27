package rollupprecompile

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	// "github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

// func SignEip712Order(o IOrderBookRollupOrder) (*CounterOrder, error) {
// 	message := map[string]interface{}{
// 		"ammIndex":          o.AmmIndex.String(),
// 		"trader":            o.Trader,
// 		"baseAssetQuantity": o.BaseAssetQuantity.String(),
// 		"price":             o.Price.String(),
// 		"salt":              o.Salt.String(),
// 		"reduceOnly":        o.ReduceOnly,
// 		"validUntil":        o.ValidUntil.String(),
// 	}
// 	domain := apitypes.TypedDataDomain{
// 		Name:              "Gnosis Protocol",
// 		Version:           "v2",
// 		ChainId:           math.NewHexOrDecimal256(321123),
// 		VerifyingContract: "0x0300000000000000000000000000000000000005",
// 	}
// 	typedData := apitypes.TypedData{
// 		Types:       Eip712OrderTypes,
// 		PrimaryType: "Order",
// 		Domain:      domain,
// 		Message:     message,
// 	}

// 	sigBytes, err := SignTypedData(typedData, c.TransactionSigner.PrivateKey)
// 	if err != nil {
// 		return nil, err
// 	}
// 	signature := fmt.Sprintf("0x%s", common.Bytes2Hex(sigBytes))
// 	o.Signature = signature
// 	return o, nil
// }

// SignTypedData - Sign typed data
func SignTypedData(typedData apitypes.TypedData, privateKey *ecdsa.PrivateKey) (sig []byte, err error) {
	hash, err := EncodeForSigning(typedData)
	if err != nil {
		return
	}
	sig, err = crypto.Sign(hash.Bytes(), privateKey)
	if err != nil {
		return
	}
	sig[64] += 27
	return
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

// VerifySig - Verify signature with recovered address
func VerifySig(from, sigHex string, msg []byte) bool {
	sig := hexutil.MustDecode(sigHex)
	//msg = accounts.TextHash(msg)
	if sig[crypto.RecoveryIDOffset] == 27 || sig[crypto.RecoveryIDOffset] == 28 {
		sig[crypto.RecoveryIDOffset] -= 27 // Transform yellow paper V from 27/28 to 0/1
	}
	recovered, err := crypto.SigToPub(msg, sig)
	recoveredAddr1 := crypto.PubkeyToAddress(*recovered)
	fmt.Printf("the recovered address: %v \n", recoveredAddr1)
	if err != nil {
		return false
	}
	recoveredAddr := crypto.PubkeyToAddress(*recovered)
	return from == recoveredAddr.Hex()
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
	"Order": {
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
		{
			Name: "validUntil",
			Type: "uint64",
		},
	},
}
