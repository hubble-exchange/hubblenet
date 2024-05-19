package miner

import (
	"math/big"

	"github.com/ava-labs/subnet-evm/core/state"
	"github.com/ava-labs/subnet-evm/core/types"
	"github.com/ethereum/go-ethereum/common"
)

type OrderbookChecker interface {
	GetMatchingTxs(tx *types.Transaction, stateDB *state.StateDB, blockNumber *big.Int) map[common.Address]types.Transactions
	ResetMemoryDB()
}
