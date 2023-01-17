package evm

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/ava-labs/subnet-evm/plugin/evm/limitorders"
	"github.com/ava-labs/subnet-evm/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSetOrderBookContractFileLocation(t *testing.T) {
	newFileLocation := "new/location"
	limitorders.SetOrderBookContractFileLocation(newFileLocation)
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
			longOrders := make([]limitorders.LimitOrder, 0)
			shortOrders := make([]limitorders.LimitOrder, 0)
			db.On("GetLongOrders").Return(longOrders)
			db.On("GetShortOrders").Return(shortOrders)
			lotp.On("PurgeLocalTx").Return(nil)
			lop.RunLiquidationsAndMatching()
			lotp.AssertNotCalled(t, "ExecuteMatchedOrdersTx", mock.Anything, mock.Anything, mock.Anything)
		})
		t.Run("Matching engine does not make call ExecuteMatchedOrders when short orders are present", func(t *testing.T) {
			db, lotp, lop := setupDependencies(t)
			longOrders := make([]limitorders.LimitOrder, 0)
			shortOrders := make([]limitorders.LimitOrder, 0)
			shortOrders = append(shortOrders, getShortOrder())
			db.On("GetLongOrders").Return(longOrders)
			db.On("GetShortOrders").Return(shortOrders)
			lotp.On("PurgeLocalTx").Return(nil)
			lop.RunLiquidationsAndMatching()
			lotp.AssertNotCalled(t, "ExecuteMatchedOrdersTx", mock.Anything, mock.Anything, mock.Anything)
		})
	})
	t.Run("Matching engine does not make call ExecuteMatchedOrders when no short orders are present in memorydb", func(t *testing.T) {
		t.Run("Matching engine does not make call ExecuteMatchedOrders when long orders are present", func(t *testing.T) {
			db, lotp, lop := setupDependencies(t)
			longOrders := make([]limitorders.LimitOrder, 0)
			shortOrders := make([]limitorders.LimitOrder, 0)
			longOrder := getLongOrder()
			longOrders = append(longOrders, longOrder)
			db.On("GetLongOrders").Return(longOrders)
			db.On("GetShortOrders").Return(shortOrders)
			lotp.On("PurgeLocalTx").Return(nil)
			lop.RunLiquidationsAndMatching()
			lotp.AssertNotCalled(t, "ExecuteMatchedOrdersTx", mock.Anything, mock.Anything, mock.Anything)
		})
	})
	t.Run("When long and short orders are present in db", func(t *testing.T) {
		t.Run("Matching engine does not make call ExecuteMatchedOrders when price is not same", func(t *testing.T) {
			db, lotp, lop := setupDependencies(t)
			longOrders := make([]limitorders.LimitOrder, 0)
			shortOrders := make([]limitorders.LimitOrder, 0)
			longOrder := getLongOrder()
			longOrders = append(longOrders, longOrder)
			shortOrder := getShortOrder()
			shortOrder.Price = big.NewInt(0).Sub(shortOrder.Price, big.NewInt(2))
			shortOrders = append(shortOrders, shortOrder)
			db.On("GetLongOrders").Return(longOrders)
			db.On("GetShortOrders").Return(shortOrders)
			lotp.On("PurgeLocalTx").Return(nil)
			lop.RunLiquidationsAndMatching()
			lotp.AssertNotCalled(t, "ExecuteMatchedOrdersTx", mock.Anything, mock.Anything, mock.Anything)
		})
		t.Run("When price is same", func(t *testing.T) {
			t.Run("When long order and short order's unfulfilled quantity is same", func(t *testing.T) {
				t.Run("When long order and short order's base asset quantity is same", func(t *testing.T) {
					//Add 2 long orders
					db, lotp, lop := setupDependencies(t)
					longOrders := make([]limitorders.LimitOrder, 0)
					longOrder1 := getLongOrder()
					longOrders = append(longOrders, longOrder1)
					longOrder2 := getLongOrder()
					longOrder2.Signature = []byte("Here is a 2nd long order")
					longOrders = append(longOrders, longOrder2)

					// Add 2 short orders
					shortOrder1 := getShortOrder()
					shortOrders := make([]limitorders.LimitOrder, 0)
					shortOrders = append(shortOrders, shortOrder1)
					shortOrder2 := getShortOrder()
					shortOrder2.Signature = []byte("Here is a 2nd short order")
					shortOrders = append(shortOrders, shortOrder2)

					db.On("GetLongOrders").Return(longOrders)
					db.On("GetShortOrders").Return(shortOrders)
					lotp.On("PurgeLocalTx").Return(nil)
					fillAmount1 := longOrder1.BaseAssetQuantity
					fillAmount2 := longOrder2.BaseAssetQuantity
					lotp.On("ExecuteMatchedOrdersTx", longOrder1, shortOrder1, fillAmount1).Return(nil)
					lotp.On("ExecuteMatchedOrdersTx", longOrder2, shortOrder2, fillAmount2).Return(nil)
					lop.RunLiquidationsAndMatching()
					lotp.AssertCalled(t, "ExecuteMatchedOrdersTx", longOrder1, shortOrder1, fillAmount1)
					lotp.AssertCalled(t, "ExecuteMatchedOrdersTx", longOrder2, shortOrder2, fillAmount2)
				})
				t.Run("When long order and short order's base asset quantity is different", func(t *testing.T) {
					db, lotp, lop := setupDependencies(t)
					//Add 2 long orders with half base asset quantity of short order
					longOrders := make([]limitorders.LimitOrder, 0)
					longOrder := getLongOrder()
					longOrder.BaseAssetQuantity = big.NewInt(20)
					longOrder.FilledBaseAssetQuantity = big.NewInt(5)
					longOrders = append(longOrders, longOrder)

					// Add 2 short orders
					shortOrder := getShortOrder()
					shortOrder.BaseAssetQuantity = big.NewInt(-30)
					shortOrder.FilledBaseAssetQuantity = big.NewInt(-15)
					shortOrders := make([]limitorders.LimitOrder, 0)
					shortOrders = append(shortOrders, shortOrder)

					fillAmount := big.NewInt(0).Sub(longOrder.BaseAssetQuantity, longOrder.FilledBaseAssetQuantity)
					db.On("GetLongOrders").Return(longOrders)
					db.On("GetShortOrders").Return(shortOrders)
					lotp.On("PurgeLocalTx").Return(nil)
					lotp.On("ExecuteMatchedOrdersTx", longOrder, shortOrder, fillAmount).Return(nil)
					lop.RunLiquidationsAndMatching()
					lotp.AssertCalled(t, "ExecuteMatchedOrdersTx", longOrder, shortOrder, fillAmount)
				})
			})
			t.Run("When long order and short order's unfulfilled quantity is not same", func(t *testing.T) {
				db, lotp, lop := setupDependencies(t)
				longOrders := make([]limitorders.LimitOrder, 0)
				longOrder1 := getLongOrder()
				longOrder1.BaseAssetQuantity = big.NewInt(20)
				longOrder1.FilledBaseAssetQuantity = big.NewInt(5)
				longOrder2 := getLongOrder()
				longOrder2.BaseAssetQuantity = big.NewInt(40)
				longOrder2.FilledBaseAssetQuantity = big.NewInt(0)
				longOrder2.Signature = []byte("Here is a 2nd long order")
				longOrder3 := getLongOrder()
				longOrder3.BaseAssetQuantity = big.NewInt(10)
				longOrder3.FilledBaseAssetQuantity = big.NewInt(3)
				longOrder3.Signature = []byte("Here is a 3rd long order")
				longOrders = append(longOrders, longOrder1, longOrder2, longOrder3)

				// Add 2 short orders
				shortOrders := make([]limitorders.LimitOrder, 0)
				shortOrder1 := getShortOrder()
				shortOrder1.BaseAssetQuantity = big.NewInt(-30)
				shortOrder1.FilledBaseAssetQuantity = big.NewInt(-2)
				shortOrder2 := getShortOrder()
				shortOrder2.BaseAssetQuantity = big.NewInt(-50)
				shortOrder2.FilledBaseAssetQuantity = big.NewInt(-20)
				shortOrder2.Signature = []byte("Here is a 2nd short order")
				shortOrder3 := getShortOrder()
				shortOrder3.BaseAssetQuantity = big.NewInt(-20)
				shortOrder3.FilledBaseAssetQuantity = big.NewInt(-10)
				shortOrder3.Signature = []byte("Here is a 3rd short order")
				shortOrders = append(shortOrders, shortOrder1, shortOrder2, shortOrder3)

				lotp.On("ExecuteMatchedOrdersTx", mock.Anything, mock.Anything, mock.Anything).Return(nil).Times(5)

				db.On("GetLongOrders").Return(longOrders)
				db.On("GetShortOrders").Return(shortOrders)
				lotp.On("PurgeLocalTx").Return(nil)
				lop.RunLiquidationsAndMatching()

				//During 1st  matching iteration
				longOrder1UnfulfilledQuantity := longOrder1.GetUnFilledBaseAssetQuantity()
				shortOrder1UnfulfilledQuantity := shortOrder1.GetUnFilledBaseAssetQuantity()
				fillAmount := utils.BigIntMinAbs(longOrder1UnfulfilledQuantity, shortOrder1UnfulfilledQuantity)
				lotp.AssertCalled(t, "ExecuteMatchedOrdersTx", longOrder1, shortOrder1, fillAmount)
				//After 1st matching iteration longOrder1 has been matched fully but shortOrder1 has not
				longOrder1.FilledBaseAssetQuantity.Add(longOrder1.FilledBaseAssetQuantity, fillAmount)
				shortOrder1.FilledBaseAssetQuantity.Sub(shortOrder1.FilledBaseAssetQuantity, fillAmount)

				//During 2nd iteration longOrder2 with be matched with shortOrder1
				longOrder2UnfulfilledQuantity := longOrder2.GetUnFilledBaseAssetQuantity()
				shortOrder1UnfulfilledQuantity = shortOrder1.GetUnFilledBaseAssetQuantity()
				fillAmount = utils.BigIntMinAbs(longOrder2UnfulfilledQuantity, shortOrder1UnfulfilledQuantity)
				lotp.AssertCalled(t, "ExecuteMatchedOrdersTx", longOrder2, shortOrder1, fillAmount)
				//After 2nd matching iteration shortOrder1 has been matched fully but longOrder2 has not
				longOrder2.FilledBaseAssetQuantity.Add(longOrder2.FilledBaseAssetQuantity, fillAmount)
				shortOrder1.FilledBaseAssetQuantity.Sub(longOrder2.FilledBaseAssetQuantity, fillAmount)

				//During 3rd iteration longOrder2 with be matched with shortOrder2
				longOrder2UnfulfilledQuantity = longOrder2.GetUnFilledBaseAssetQuantity()
				shortOrder2UnfulfilledQuantity := shortOrder2.GetUnFilledBaseAssetQuantity()
				fillAmount = utils.BigIntMinAbs(longOrder2UnfulfilledQuantity, shortOrder2UnfulfilledQuantity)
				lotp.AssertCalled(t, "ExecuteMatchedOrdersTx", longOrder2, shortOrder2, fillAmount)
				//After 3rd matching iteration longOrder2 has been matched fully but shortOrder2 has not
				longOrder2.FilledBaseAssetQuantity.Add(longOrder2.FilledBaseAssetQuantity, fillAmount)
				shortOrder2.FilledBaseAssetQuantity.Sub(shortOrder2.FilledBaseAssetQuantity, fillAmount)

				//So during 4th iteration longOrder3 with be matched with shortOrder2
				longOrder3UnfulfilledQuantity := longOrder3.GetUnFilledBaseAssetQuantity()
				shortOrder2UnfulfilledQuantity = shortOrder2.GetUnFilledBaseAssetQuantity()
				fillAmount = utils.BigIntMinAbs(longOrder3UnfulfilledQuantity, shortOrder2UnfulfilledQuantity)
				lotp.AssertCalled(t, "ExecuteMatchedOrdersTx", longOrder3, shortOrder2, fillAmount)
				//After 4rd matching iteration shortOrder2 has been matched fully but longOrder3 has not
				longOrder3.FilledBaseAssetQuantity.Add(longOrder3.FilledBaseAssetQuantity, fillAmount)
				shortOrder2.FilledBaseAssetQuantity.Sub(shortOrder2.FilledBaseAssetQuantity, fillAmount)

				//So during 5th iteration longOrder3 with be matched with shortOrder3
				longOrder3UnfulfilledQuantity = longOrder3.GetUnFilledBaseAssetQuantity()
				shortOrder3UnfulfilledQuantity := shortOrder3.GetUnFilledBaseAssetQuantity()
				fillAmount = utils.BigIntMinAbs(longOrder3UnfulfilledQuantity, shortOrder3UnfulfilledQuantity)
				lotp.AssertCalled(t, "ExecuteMatchedOrdersTx", longOrder3, shortOrder3, fillAmount)
			})
		})
	})
}

func getShortOrder() limitorders.LimitOrder {
	signature := []byte("Here is a short order")
	shortOrder := createLimitOrder("short", "0x22Bb736b64A0b4D4081E103f83bccF864F0404aa", big.NewInt( -10), big.NewInt(20.0), "unfulfilled", signature, big.NewInt(2))
	return shortOrder
}

func getLongOrder() limitorders.LimitOrder {
	signature := []byte("Here is a long order")
	longOrder := createLimitOrder("long", "0x22Bb736b64A0b4D4081E103f83bccF864F0404aa", big.NewInt(10), big.NewInt(20.0), "unfulfilled", signature, big.NewInt(2))
	return longOrder
}

func createLimitOrder(positionType string, userAddress string, baseAssetQuantity *big.Int, price *big.Int, status string, signature []byte, blockNumber *big.Int) limitorders.LimitOrder {
	return limitorders.LimitOrder{
		PositionType:      positionType,
		UserAddress:       userAddress,
		BaseAssetQuantity: baseAssetQuantity,
		Price:             price,
		Status:            limitorders.Status(status),
		Signature:         signature,
		BlockNumber:       blockNumber,
	}
}
