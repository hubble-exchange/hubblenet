package bibliophile

import (
	"math/big"

	"github.com/ava-labs/subnet-evm/precompile/contract"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	IOC_ORDERBOOK_ADDRESS       = "0x0300000000000000000000000000000000000006"
	IOC_ORDER_INFO_SLOT   int64 = 53
)

// State Reader
func iocGetBlockPlaced(stateDB contract.StateDB, orderHash [32]byte) *big.Int {
	orderInfo := iocOrderInfoMappingStorageSlot(orderHash)
	return new(big.Int).SetBytes(stateDB.GetState(common.HexToAddress(IOC_ORDERBOOK_ADDRESS), common.BigToHash(orderInfo)).Bytes())
}

func iocGetOrderFilledAmount(stateDB contract.StateDB, orderHash [32]byte) *big.Int {
	orderInfo := iocOrderInfoMappingStorageSlot(orderHash)
	return new(big.Int).SetBytes(stateDB.GetState(common.HexToAddress(IOC_ORDERBOOK_ADDRESS), common.BigToHash(new(big.Int).Add(orderInfo, big.NewInt(1)))).Bytes())
}

func iocGetOrderStatus(stateDB contract.StateDB, orderHash [32]byte) int64 {
	orderInfo := iocOrderInfoMappingStorageSlot(orderHash)
	return new(big.Int).SetBytes(stateDB.GetState(common.HexToAddress(IOC_ORDERBOOK_ADDRESS), common.BigToHash(new(big.Int).Add(orderInfo, big.NewInt(2)))).Bytes()).Int64()
}

func iocOrderInfoMappingStorageSlot(orderHash [32]byte) *big.Int {
	return new(big.Int).SetBytes(crypto.Keccak256(append(orderHash[:], common.LeftPadBytes(big.NewInt(IOC_ORDER_INFO_SLOT).Bytes(), 32)...)))
}
