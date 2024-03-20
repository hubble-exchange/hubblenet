package orderbook

import (
	hu "github.com/ava-labs/subnet-evm/hubbleutils"
)

type ContractOrder interface {
	EncodeToABIWithoutType() ([]byte, error)
	EncodeToABI() ([]byte, error)
	DecodeFromRawOrder(rawOrder interface{})
	Map() map[string]interface{}
}

type LimitOrder = hu.LimitOrder
type IOCOrder = hu.IOCOrder
