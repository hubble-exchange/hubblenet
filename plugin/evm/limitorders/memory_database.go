package limitorders

import "sort"

type LimitOrder struct {
	id                int64
	PositionType      string
	UserAddress       string
	BaseAssetQuantity int
	Price             float64
	Status            string
	Salt              string
	Signature         []byte
	RawOrder          interface{}
	RawSignature      interface{}
}

type InMemoryDatabase interface {
	GetAllOrders() []*LimitOrder
	GetOrdersByPriceAndPositionType(positionType string, price float64) []*LimitOrder
	Add(order LimitOrder)
	Delete(signature []byte)
}

type inMemoryDatabase struct {
	orderMap map[string]*LimitOrder
}

func NewInMemoryDatabase() *inMemoryDatabase {
	orderMap := map[string]*LimitOrder{}
	return &inMemoryDatabase{orderMap}
}

func (db *inMemoryDatabase) GetAllOrders() []*LimitOrder {
	allOrders := []*LimitOrder{}
	for _, order := range db.orderMap {
		allOrders = append(allOrders, order)
	}
	return allOrders
}

func (db *inMemoryDatabase) Add(order LimitOrder) {
	db.orderMap[string(order.Signature)] = &order
}

// Deletes silently
func (db *inMemoryDatabase) Delete(signature []byte) {
	delete(db.orderMap, string(signature))
}

func (db *inMemoryDatabase) GetOrdersByPriceAndPositionType(positionType string, price float64) []*LimitOrder {
	if positionType == "long" {
		return getLongOrdersByPrice(db.orderMap, price)
	}
	if positionType == "short" {
		return getShortOrdersByPrice(db.orderMap, price)
	}
	return nil
}

func getLongOrdersByPrice(orderMap map[string]*LimitOrder, price float64) []*LimitOrder {
	matchingLongOrders := []*LimitOrder{}
	for _, order := range orderMap {
		if order.PositionType == "long" && order.Status == "unfulfilled" && price <= order.Price {
			matchingLongOrders = append(matchingLongOrders, order)
		}
	}
	sort.SliceStable(matchingLongOrders, func(i, j int) bool {
		return matchingLongOrders[i].Price > matchingLongOrders[j].Price
	})
	return matchingLongOrders
}

func getShortOrdersByPrice(orderMap map[string]*LimitOrder, price float64) []*LimitOrder {
	matchingShortOrders := []*LimitOrder{}
	for _, order := range orderMap {
		if order.PositionType == "short" && order.Status == "unfulfilled" && price >= order.Price {
			matchingShortOrders = append(matchingShortOrders, order)
		}
	}
	sort.SliceStable(matchingShortOrders, func(i, j int) bool {
		return matchingShortOrders[i].Price < matchingShortOrders[j].Price
	})
	return matchingShortOrders
}
