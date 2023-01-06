package evm

import (
	"fmt"
	"testing"

	"github.com/ava-labs/subnet-evm/plugin/evm/limitorders"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSetOrderBookContractFileLocation(t *testing.T) {
	newFileLocation := "new/location"
	SetOrderBookContractFileLocation(newFileLocation)
	assert.Equal(t, newFileLocation, orderBookContractFileLocation)
}

func newVM(t *testing.T) *VM {
	txFeeCap := float64(11)
	enabledEthAPIs := []string{"debug"}
	configJSON := fmt.Sprintf("{\"rpc-tx-fee-cap\": %g,\"eth-apis\": %s}", txFeeCap, fmt.Sprintf("[%q]", enabledEthAPIs[0]))
	_, vm, _, _ := GenesisVM(t, false, "", configJSON, "")
	return vm
}

func newLimitOrderProcesser(t *testing.T, db limitorders.LimitOrderDatabase, lotp limitorders.LimitOrderTxProcessor) LimitOrderProcesser {
	vm := newVM(t)
	lop := NewLimitOrderProcesser(
		vm.ctx,
		vm.txPool,
		vm.shutdownChan,
		&vm.shutdownWg,
		vm.eth.APIBackend,
		vm.eth.BlockChain(),
		db,
		lotp,
	)
	return lop
}
func TestNewLimitOrderProcesser(t *testing.T) {
	_, _, lop := setupDependencies(t)
	assert.NotNil(t, lop)
}

func setupDependencies(t *testing.T) (*MockLimitOrderDatabase, *MockLimitOrderTxProcessor, LimitOrderProcesser) {
	db := NewMockLimitOrderDatabase()
	lotp := NewMockLimitOrderTxProcessor()
	lop := newLimitOrderProcesser(t, db, lotp)
	return db, lotp, lop
}

func TestRunMatchingEngine(t *testing.T) {
	t.Run("Matching engine does not make call ExecuteMatchedOrders when no long orders are present in memorydb", func(t *testing.T) {
		t.Run("Matching engine does not make call ExecuteMatchedOrders when no short orders are present", func(t *testing.T) {
			db, lotp, lop := setupDependencies(t)
			longOrders := make([]*limitorders.LimitOrder, 0)
			shortOrders := make([]*limitorders.LimitOrder, 0)
			db.On("GetLongOrders").Return(longOrders)
			db.On("GetShortOrders").Return(shortOrders)
			lotp.On("PurgeLocalTx").Return(nil)
			lop.RunMatchingEngine()
			lotp.AssertNotCalled(t, "ExecuteMatchedOrdersTx", mock.Anything, mock.Anything)
		})
		t.Run("Matching engine does not make call ExecuteMatchedOrders when short orders are present", func(t *testing.T) {
			db, lotp, lop := setupDependencies(t)
			longOrders := make([]*limitorders.LimitOrder, 0)
			shortOrders := make([]*limitorders.LimitOrder, 0)
			shortOrders = append(shortOrders, getShortOrder())
			db.On("GetLongOrders").Return(longOrders)
			db.On("GetShortOrders").Return(shortOrders)
			lotp.On("PurgeLocalTx").Return(nil)
			lop.RunMatchingEngine()
			lotp.AssertNotCalled(t, "ExecuteMatchedOrdersTx", mock.Anything, mock.Anything)
		})
	})
	t.Run("Matching engine does not make call ExecuteMatchedOrders when no short orders are present in memorydb", func(t *testing.T) {
		t.Run("Matching engine does not make call ExecuteMatchedOrders when long orders are present", func(t *testing.T) {
			db, lotp, lop := setupDependencies(t)
			longOrders := make([]*limitorders.LimitOrder, 0)
			longOrder := getLongOrder()
			longOrders = append(longOrders, longOrder)
			shortOrders := make([]*limitorders.LimitOrder, 0)
			db.On("GetLongOrders").Return(longOrders)
			db.On("GetShortOrders").Return(shortOrders)
			lotp.On("PurgeLocalTx").Return(nil)
			lop.RunMatchingEngine()
			lotp.AssertNotCalled(t, "ExecuteMatchedOrdersTx", mock.Anything, mock.Anything)
		})
	})
	t.Run("When long and short orders are present in db", func(t *testing.T) {
		t.Run("Matching engine does not make call ExecuteMatchedOrders when price is not same", func(t *testing.T) {
			db, lotp, lop := setupDependencies(t)
			longOrders := make([]*limitorders.LimitOrder, 0)
			shortOrders := make([]*limitorders.LimitOrder, 0)
			longOrder := getLongOrder()
			longOrders = append(longOrders, longOrder)
			shortOrder := getShortOrder()
			shortOrder.Price = longOrder.Price - 2
			shortOrders = append(shortOrders, shortOrder)
			db.On("GetLongOrders").Return(longOrders)
			db.On("GetShortOrders").Return(shortOrders)
			lop.RunMatchingEngine()
			lotp.AssertNotCalled(t, "ExecuteMatchedOrdersTx", mock.Anything, mock.Anything)
		})
		t.Run("When price is same", func(t *testing.T) {
			t.Run("When mod of baseAssetQuantity is not same", func(t *testing.T) {
				db, lotp, lop := setupDependencies(t)
				longOrders := make([]*limitorders.LimitOrder, 0)
				shortOrders := make([]*limitorders.LimitOrder, 0)
				longOrder := getLongOrder()
				longOrders = append(longOrders, longOrder)
				shortOrder := getShortOrder()
				shortOrder.BaseAssetQuantity = shortOrder.BaseAssetQuantity + 1
				shortOrders = append(shortOrders, shortOrder)
				db.On("GetLongOrders").Return(longOrders)
				db.On("GetShortOrders").Return(shortOrders)
				lop.RunMatchingEngine()
				lotp.AssertNotCalled(t, "ExecuteMatchedOrdersTx", mock.Anything, mock.Anything)
			})
			t.Run("When mod of baseAssetQuantity is same", func(t *testing.T) {
				t.Run("When ExecuteMatchedOrderTx return error it tries again", func(t *testing.T) {
					//Write test when we handle error in a better way
				})
				t.Run("When ExecuteMatchedOrderTx does not return error", func(t *testing.T) {
					db, lotp, lop := setupDependencies(t)
					//Write test when we handle error in a better way
					longOrders := make([]*limitorders.LimitOrder, 0)
					shortOrders := make([]*limitorders.LimitOrder, 0)
					longOrder := getLongOrder()
					longOrders = append(longOrders, longOrder)
					shortOrder := getShortOrder()
					shortOrders = append(shortOrders, shortOrder)
					db.On("GetLongOrders").Return(longOrders)
					db.On("GetShortOrders").Return(shortOrders)
					lotp.On("ExecuteMatchedOrdersTx", *longOrder, *shortOrder).Return(nil)
					lop.RunMatchingEngine()
				})
			})
		})
	})
}

func getShortOrder() *limitorders.LimitOrder {
	signature := []byte("Here is a short order")
	shortOrder := createLimitOrder("short", "0x22Bb736b64A0b4D4081E103f83bccF864F0404aa", -10, 20.01, "unfulfilled", "salt", signature, 2)
	return shortOrder
}

func getLongOrder() *limitorders.LimitOrder {
	signature := []byte("Here is a long order")
	longOrder := createLimitOrder("long", "0x22Bb736b64A0b4D4081E103f83bccF864F0404aa", 10, 20.01, "unfulfilled", "salt", signature, 2)
	return longOrder
}

func createLimitOrder(positionType string, userAddress string, baseAssetQuantity int, price float64, status string, salt string, signature []byte, blockNumber uint64) *limitorders.LimitOrder {
	return &limitorders.LimitOrder{
		PositionType:      positionType,
		UserAddress:       userAddress,
		BaseAssetQuantity: baseAssetQuantity,
		Price:             price,
		Status:            status,
		Salt:              salt,
		Signature:         signature,
		BlockNumber:       blockNumber,
	}
}
