package bibliophile

import (
	"math/big"

	"github.com/ava-labs/subnet-evm/precompile/contract"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// @todo change this to the correct values
const (
	SIGNED_ORDERBOOK_ADDRESS       = "0x03000000000000000000000000000000000000b4"
	SIGNED_ORDER_INFO_SLOT   int64 = 1
)

// State Reader
func GetSignedOrderFilledAmount(stateDB contract.StateDB, orderHash [32]byte) *big.Int {
	orderInfo := signedOrderInfoMappingStorageSlot(orderHash)
	num := stateDB.GetState(common.HexToAddress(SIGNED_ORDERBOOK_ADDRESS), common.BigToHash(orderInfo)).Bytes()
	return fromTwosComplement(num)
}

func GetSignedOrderStatus(stateDB contract.StateDB, orderHash [32]byte) int64 {
	orderInfo := signedOrderInfoMappingStorageSlot(orderHash)
	return new(big.Int).SetBytes(stateDB.GetState(common.HexToAddress(SIGNED_ORDERBOOK_ADDRESS), common.BigToHash(new(big.Int).Add(orderInfo, big.NewInt(1)))).Bytes()).Int64()
}

func signedOrderInfoMappingStorageSlot(orderHash [32]byte) *big.Int {
	return new(big.Int).SetBytes(crypto.Keccak256(append(orderHash[:], common.LeftPadBytes(big.NewInt(SIGNED_ORDER_INFO_SLOT).Bytes(), 32)...)))
}
