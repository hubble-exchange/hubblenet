package limitorders

import (
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewInMemoryDatabase(t *testing.T) {
	inMemoryDatabase := NewInMemoryDatabase()
	assert.NotNil(t, inMemoryDatabase)
	assert.Equal(t, 0, len(inMemoryDatabase.orderMap))
}

func TestAdd(t *testing.T) {
	inMemoryDatabase := NewInMemoryDatabase()
	signature := []byte("Here is a string....")
	limitOrder := &LimitOrder{
		id:                123,
		PositionType:      "short",
		UserAddress:       "random-address",
		BaseAssetQuantity: -10,
		Price:             20.01,
		Status:            "UNFULFILLED",
		Salt:              "salt",
		Signature:         signature,
	}
	inMemoryDatabase.Add(*limitOrder)
	returnedOrder := inMemoryDatabase.orderMap[string(signature)]
	assert.Equal(t, limitOrder.id, returnedOrder.id)
	assert.Equal(t, limitOrder.PositionType, returnedOrder.PositionType)
	assert.Equal(t, limitOrder.UserAddress, returnedOrder.UserAddress)
	assert.Equal(t, limitOrder.BaseAssetQuantity, returnedOrder.BaseAssetQuantity)
	assert.Equal(t, limitOrder.Price, returnedOrder.Price)
	assert.Equal(t, limitOrder.Status, returnedOrder.Status)
	assert.Equal(t, limitOrder.Salt, returnedOrder.Salt)
}

func TestAllOrders(t *testing.T) {
	inMemoryDatabase := NewInMemoryDatabase()
	totalOrders := 5
	positionType := "short"
	userAddress := "random-address"
	baseAssetQuantity := -10
	price := 20.01
	status := "unfulfilled"
	salt := "salt"
	for i := 0; i < totalOrders; i++ {
		signature := []byte(fmt.Sprintf("Signature is %d", i))
		limitOrder := &LimitOrder{
			id:                int64(i),
			PositionType:      positionType,
			UserAddress:       userAddress,
			BaseAssetQuantity: baseAssetQuantity,
			Price:             price,
			Status:            status,
			Salt:              salt,
			Signature:         signature,
		}
		inMemoryDatabase.Add(*limitOrder)
	}
	returnedOrders := inMemoryDatabase.GetAllOrders()
	assert.Equal(t, totalOrders, len(returnedOrders))
	fmt.Println(returnedOrders)
	for _, returedOrder := range returnedOrders {
		assert.Equal(t, positionType, returedOrder.PositionType)
		assert.Equal(t, userAddress, returedOrder.UserAddress)
		assert.Equal(t, baseAssetQuantity, returedOrder.BaseAssetQuantity)
		assert.Equal(t, price, returedOrder.Price)
		assert.Equal(t, status, returedOrder.Status)
		assert.Equal(t, salt, returedOrder.Salt)
	}
}

func TestDelete(t *testing.T) {
	inMemoryDatabase := NewInMemoryDatabase()
	totalOrders := 5
	positionType := "short"
	userAddress := "random-address"
	baseAssetQuantity := -10
	price := 20.01
	status := "unfulfilled"
	salt := "salt"
	for i := 0; i < totalOrders; i++ {
		signature := []byte(fmt.Sprintf("Signature is %d", i))
		limitOrder := &LimitOrder{
			id:                int64(i),
			PositionType:      positionType,
			UserAddress:       userAddress,
			BaseAssetQuantity: baseAssetQuantity,
			Price:             price,
			Status:            status,
			Salt:              salt,
			Signature:         signature,
		}
		inMemoryDatabase.Add(*limitOrder)
	}

	deletedOrderId := 3
	inMemoryDatabase.Delete([]byte(fmt.Sprintf("Signature is %d", deletedOrderId)))
	expectedReturnedOrdersIds := []int{0, 1, 2, 4}

	returnedOrders := inMemoryDatabase.GetAllOrders()
	assert.Equal(t, totalOrders-1, len(returnedOrders))
	var returnedOrderIds []int
	for _, returnedOrder := range returnedOrders {
		returnedOrderIds = append(returnedOrderIds, int(returnedOrder.id))
	}
	sort.Ints(returnedOrderIds)
	assert.Equal(t, expectedReturnedOrdersIds, returnedOrderIds)
}

var userAddress = "random-address"
var baseAssetQuantity int
var salt = "salt"
var price = 20.01

func TestGetOrdersByPriceAndPositionTypeWhenNoLongOrderExists(t *testing.T) {
	inMemoryDatabase := NewInMemoryDatabase()
	addLimitOrderInDatabase(inMemoryDatabase, "short")
	orders1 := inMemoryDatabase.GetOrdersByPriceAndPositionType("long", 0)
	assert.Equal(t, 0, len(orders1))
	orders2 := inMemoryDatabase.GetOrdersByPriceAndPositionType("long", 20)
	assert.Equal(t, 0, len(orders2))
	orders3 := inMemoryDatabase.GetOrdersByPriceAndPositionType("long", 1000)
	assert.Equal(t, 0, len(orders3))
}
func TestGetOrdersByPriceAndPositionTypeWhenNoShortOrderExists(t *testing.T) {
	inMemoryDatabase := NewInMemoryDatabase()
	addLimitOrderInDatabase(inMemoryDatabase, "long")
	orders1 := inMemoryDatabase.GetOrdersByPriceAndPositionType("short", 0)
	assert.Equal(t, 0, len(orders1))
	orders2 := inMemoryDatabase.GetOrdersByPriceAndPositionType("short", 20)
	assert.Equal(t, 0, len(orders2))
	orders3 := inMemoryDatabase.GetOrdersByPriceAndPositionType("short", 100)
	assert.Equal(t, 0, len(orders3))
}

func TestGetOrdersByPriceAndPositionTypeWhenLongOrderExistsButPriceDontMatch(t *testing.T) {
	inMemoryDatabase := NewInMemoryDatabase()
	addLimitOrderInDatabase(inMemoryDatabase, "long")
	orders1 := inMemoryDatabase.GetOrdersByPriceAndPositionType("long", 21)
	orders2 := inMemoryDatabase.GetOrdersByPriceAndPositionType("long", 20.02)
	assert.Equal(t, 0, len(orders1))
	assert.Equal(t, 0, len(orders2))
}

func TestGetOrdersByPriceAndPositionTypeWhenShortOrderExistsButPriceDontMatch(t *testing.T) {
	inMemoryDatabase := NewInMemoryDatabase()
	addLimitOrderInDatabase(inMemoryDatabase, "short")
	orders1 := inMemoryDatabase.GetOrdersByPriceAndPositionType("short", 16)
	orders2 := inMemoryDatabase.GetOrdersByPriceAndPositionType("short", 15)
	assert.Equal(t, 0, len(orders1))
	assert.Equal(t, 0, len(orders2))
}

func TestGetOrdersByPriceAndPositionTypeWhenShortOrderExistsAndPriceMatch(t *testing.T) {
	inMemoryDatabase := NewInMemoryDatabase()
	addLimitOrderInDatabase(inMemoryDatabase, "short")
	orders1 := inMemoryDatabase.GetOrdersByPriceAndPositionType("short", 16.5)
	assert.Equal(t, 1, len(orders1))
	assert.Equal(t, 16.01, orders1[0].Price)

	limitOrder := &LimitOrder{
		id:                6,
		PositionType:      "short",
		UserAddress:       userAddress,
		BaseAssetQuantity: baseAssetQuantity,
		Price:             price,
		Status:            "fulfilled",
		Salt:              salt,
		Signature:         []byte("Signature is 10"),
	}
	inMemoryDatabase.Add(*limitOrder)
	orders2 := inMemoryDatabase.GetOrdersByPriceAndPositionType("short", 30)
	assert.Equal(t, 5, len(orders2))
	for i, order := range orders2 {
		assert.Equal(t, userAddress, order.UserAddress)
		assert.Equal(t, baseAssetQuantity, order.BaseAssetQuantity)
		assert.Equal(t, "unfulfilled", order.Status)
		assert.Equal(t, salt, order.Salt)
		expectedPrice := price - float64((4 - i))
		assert.Equal(t, expectedPrice, order.Price)
	}
}

func TestGetOrdersByPriceAndPositionTypeWhenLongOrderExistsAndPriceMatch(t *testing.T) {
	inMemoryDatabase := NewInMemoryDatabase()
	addLimitOrderInDatabase(inMemoryDatabase, "long")
	orders1 := inMemoryDatabase.GetOrdersByPriceAndPositionType("long", 20.00)
	assert.Equal(t, 1, len(orders1))
	assert.Equal(t, 20.01, orders1[0].Price)

	limitOrder := &LimitOrder{
		id:                6,
		PositionType:      "long",
		UserAddress:       userAddress,
		BaseAssetQuantity: baseAssetQuantity,
		Price:             price,
		Status:            "fulfilled",
		Salt:              salt,
		Signature:         []byte("Signature is 10"),
	}
	inMemoryDatabase.Add(*limitOrder)
	orders2 := inMemoryDatabase.GetOrdersByPriceAndPositionType("long", 10)
	assert.Equal(t, 5, len(orders2))
	for i, order := range orders2 {
		assert.Equal(t, userAddress, order.UserAddress)
		assert.Equal(t, baseAssetQuantity, order.BaseAssetQuantity)
		assert.Equal(t, "unfulfilled", order.Status)
		assert.Equal(t, salt, order.Salt)
		expectedPrice := price - float64(i)
		assert.Equal(t, expectedPrice, order.Price)
	}
}

func addLimitOrderInDatabase(database *inMemoryDatabase, positionType string) {
	totalOrders := 5
	if positionType == "short" {
		baseAssetQuantity = -10
	} else {
		baseAssetQuantity = 10
	}
	status := "unfulfilled"
	for i := 0; i < totalOrders; i++ {
		signature := []byte(fmt.Sprintf("Signature is %d", i))
		limitOrder := &LimitOrder{
			id:                int64(i),
			PositionType:      positionType,
			UserAddress:       userAddress,
			BaseAssetQuantity: baseAssetQuantity,
			Price:             price - float64(i),
			Status:            status,
			Salt:              salt,
			Signature:         signature,
		}
		database.Add(*limitOrder)
	}
}
