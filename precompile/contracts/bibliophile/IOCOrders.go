package bibliophile

import (
	"math/big"

	"github.com/ava-labs/subnet-evm/precompile/contract"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	IOC_ORDERBOOK_ADDRESS          = "" // todo
	IOC_ORDER_INFO_SLOT      int64 = 2
	EXECUTION_THRESHOLD_SLOT int64 = 3
)

// State Reader
func IOC_GetBlockPlaced(stateDB contract.StateDB, orderHash [32]byte) *big.Int {
	orderInfo := iocOrderInfoMappingStorageSlot(orderHash)
	return new(big.Int).SetBytes(stateDB.GetState(common.HexToAddress(IOC_ORDERBOOK_ADDRESS), common.BigToHash(orderInfo)).Bytes())
}

func IOC_GetOrderFilledAmount(stateDB contract.StateDB, orderHash [32]byte) *big.Int {
	orderInfo := iocOrderInfoMappingStorageSlot(orderHash)
	return new(big.Int).SetBytes(stateDB.GetState(common.HexToAddress(IOC_ORDERBOOK_ADDRESS), common.BigToHash(new(big.Int).Add(orderInfo, big.NewInt(1)))).Bytes())
}

func IOC_GetTimestamp(stateDB contract.StateDB, orderHash [32]byte) *big.Int {
	orderInfo := iocOrderInfoMappingStorageSlot(orderHash)
	return new(big.Int).SetBytes(stateDB.GetState(common.HexToAddress(IOC_ORDERBOOK_ADDRESS), common.BigToHash(new(big.Int).Add(orderInfo, big.NewInt(2)))).Bytes())
}

func IOC_GetOrderStatus(stateDB contract.StateDB, orderHash [32]byte) int64 {
	orderInfo := iocOrderInfoMappingStorageSlot(orderHash)
	return new(big.Int).SetBytes(stateDB.GetState(common.HexToAddress(IOC_ORDERBOOK_ADDRESS), common.BigToHash(new(big.Int).Add(orderInfo, big.NewInt(3)))).Bytes()).Int64()
}

func iocOrderInfoMappingStorageSlot(orderHash [32]byte) *big.Int {
	return new(big.Int).SetBytes(crypto.Keccak256(append(orderHash[:], common.LeftPadBytes(big.NewInt(IOC_ORDER_INFO_SLOT).Bytes(), 32)...)))
}

func IOC_ExecutionThreshold(stateDB contract.StateDB) *big.Int {
	return new(big.Int).SetBytes(stateDB.GetState(common.HexToAddress(IOC_ORDERBOOK_ADDRESS), common.BigToHash(big.NewInt(EXECUTION_THRESHOLD_SLOT))).Bytes())
}
