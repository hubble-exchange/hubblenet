package limitorders

import (
	"math/big"
	"sort"

	"github.com/ethereum/go-ethereum/common"
)

type LimitOrder struct {
	id                uint64
	PositionType      string
	UserAddress       string
	BaseAssetQuantity int
	Price             float64
	Status            string
	Salt              string
	Signature         []byte
	RawOrder          interface{}
	RawSignature      interface{}
	BlockNumber       uint64
}

// We might add more fields like openNotional etc
type Position struct {
	BaseAssetQuantity *big.Int
}

type InMemoryDatabase struct {
	orderMap    map[string]*LimitOrder
	positionMap map[common.Address]*Position
}

func NewInMemoryDatabase() *InMemoryDatabase {
	orderMap := map[string]*LimitOrder{}
	positionMap := map[common.Address]*Position{}
	return &InMemoryDatabase{orderMap, positionMap}
}

func (db *InMemoryDatabase) GetAllOrders() []*LimitOrder {
	allOrders := []*LimitOrder{}
	for _, order := range db.orderMap {
		allOrders = append(allOrders, order)
	}
	return allOrders
}

func (db *InMemoryDatabase) UpdateExecution(userAddress common.Address, baseAssetQuantity *big.Int) {
	db.positionMap[userAddress].BaseAssetQuantity.Add(db.positionMap[userAddress].BaseAssetQuantity, baseAssetQuantity)
}

func (db *InMemoryDatabase) Add(order *LimitOrder) {
	db.orderMap[string(order.Signature)] = order
}

// Deletes silently
func (db *InMemoryDatabase) Delete(signature []byte) {
	delete(db.orderMap, string(signature))
}

func (db *InMemoryDatabase) GetLongOrders() []*LimitOrder {
	var longOrders []*LimitOrder
	for _, order := range db.orderMap {
		if order.PositionType == "long" {
			longOrders = append(longOrders, order)
		}
	}
	sortLongOrders(longOrders)
	return longOrders
}

func (db *InMemoryDatabase) GetShortOrders() []*LimitOrder {
	var shortOrders []*LimitOrder
	for _, order := range db.orderMap {
		if order.PositionType == "short" {
			shortOrders = append(shortOrders, order)
		}
	}
	sortShortOrders(shortOrders)
	return shortOrders
}

func sortLongOrders(orders []*LimitOrder) []*LimitOrder {
	sort.SliceStable(orders, func(i, j int) bool {
		if orders[i].Price > orders[j].Price {
			return true
		}
		if orders[i].Price == orders[j].Price {
			if orders[i].BlockNumber < orders[j].BlockNumber {
				return true
			}
		}
		return false
	})
	return orders
}

func sortShortOrders(orders []*LimitOrder) []*LimitOrder {
	sort.SliceStable(orders, func(i, j int) bool {
		if orders[i].Price < orders[j].Price {
			return true
		}
		if orders[i].Price == orders[j].Price {
			if orders[i].BlockNumber < orders[j].BlockNumber {
				return true
			}
		}
		return false
	})
	return orders
}
