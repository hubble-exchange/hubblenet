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
