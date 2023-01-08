package limitorders

import (
	"sort"
)

type LimitOrder struct {
	id                      uint64
	PositionType            string
	UserAddress             string
	BaseAssetQuantity       int
	FilledBaseAssetQuantity int
	Price                   float64
	Status                  string
	Salt                    string
	Signature               []byte
	RawOrder                interface{}
	RawSignature            interface{}
	BlockNumber             uint64
}

type LimitOrderDatabase interface {
	GetAllOrders() []LimitOrder
	Add(order *LimitOrder)
	UpdateFilledBaseAssetQuantity(quantity int, signature []byte)
	Delete(signature []byte)
	GetLongOrders() []LimitOrder
	GetShortOrders() []LimitOrder
	GetOrder(signature []byte) *LimitOrder
}

type inMemoryDatabase struct {
	orderMap map[string]*LimitOrder
}

func NewInMemoryDatabase() LimitOrderDatabase {
	orderMap := map[string]*LimitOrder{}
	return &inMemoryDatabase{orderMap}
}

func (db *inMemoryDatabase) GetAllOrders() []LimitOrder {
	allOrders := []LimitOrder{}
	for _, order := range db.orderMap {
		allOrders = append(allOrders, *order)
	}
	return allOrders
}

func (db *inMemoryDatabase) Add(order *LimitOrder) {
	db.orderMap[string(order.Signature)] = order
}

func (db *inMemoryDatabase) UpdateFilledBaseAssetQuantity(quantity int, signature []byte) {
	limitOrder := db.orderMap[string(signature)]
	if limitOrder.BaseAssetQuantity == quantity {
		deleteOrder(db, signature)
		return
	}
	limitOrder.FilledBaseAssetQuantity = quantity
}

// Deletes silently
func (db *inMemoryDatabase) Delete(signature []byte) {
	deleteOrder(db, signature)
}

func (db *inMemoryDatabase) GetLongOrders() []LimitOrder {
	var longOrders []LimitOrder
	for _, order := range db.orderMap {
		if order.PositionType == "long" {
			longOrders = append(longOrders, *order)
		}
	}
	sortLongOrders(longOrders)
	return longOrders
}

func (db *inMemoryDatabase) GetShortOrders() []LimitOrder {
	var shortOrders []LimitOrder
	for _, order := range db.orderMap {
		if order.PositionType == "short" {
			shortOrders = append(shortOrders, *order)
		}
	}
	sortShortOrders(shortOrders)
	return shortOrders
}

func (db *inMemoryDatabase) GetOrder(signature []byte) *LimitOrder {
	return db.orderMap[string(signature)]
}

func sortLongOrders(orders []LimitOrder) []LimitOrder {
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

func sortShortOrders(orders []LimitOrder) []LimitOrder {
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

func deleteOrder(db *inMemoryDatabase, signature []byte) {
	delete(db.orderMap, string(signature))
}
