package evm

import (
	"fmt"
	"testing"

	"github.com/ava-labs/subnet-evm/plugin/evm/limitorders"
	"github.com/stretchr/testify/assert"
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
	db := NewMockLimitOrderDatabase()
	lotp := NewMockLimitOrderTxProcessor()
	lop := newLimitOrderProcesser(t, db, lotp)
	assert.NotNil(t, lop)
}

func TestRunMatchingEngine(t *testing.T) {
	db := NewMockLimitOrderDatabase()
	lotp := NewMockLimitOrderTxProcessor()
	lop := newLimitOrderProcesser(t, db, lotp)
	t.Run("Matching engine does not make call ExecuteMatchedOrders when no long orders are present in memorydb", func(t *testing.T) {
		t.Run("Matching engine does not make call ExecuteMatchedOrders when no short orders are present", func(t *testing.T) {
			longOrders := make([]*limitorders.LimitOrder, 0)
			shortOrders := make([]*limitorders.LimitOrder, 0)
			db.On("GetLongOrders").Return(longOrders)
			db.On("GetShortOrders").Return(shortOrders)
			lotp.AssertNotCalled(t, "ExecuteMatchedOrdersTx")
			lop.RunMatchingEngine()
		})
		t.Run("Matching engine does not make call ExecuteMatchedOrders when short orders are present", func(t *testing.T) {
			longOrders := make([]*limitorders.LimitOrder, 0)
			shortOrders := make([]*limitorders.LimitOrder, 0)
			shortOrders = append(shortOrders, getShortOrder())
			db.On("GetLongOrders").Return(longOrders)
			db.On("GetShortOrders").Return(shortOrders)
			lotp.AssertNotCalled(t, "ExecuteMatchedOrdersTx")
			lop.RunMatchingEngine()
		})
	})
	t.Run("Matching engine does not make call ExecuteMatchedOrders when no short orders are present in memorydb", func(t *testing.T) {
		t.Run("Matching engine does not make call ExecuteMatchedOrders when long orders are present", func(t *testing.T) {
			longOrders := make([]*limitorders.LimitOrder, 0)
			longOrders = append(longOrders, getLongOrder())
			shortOrders := make([]*limitorders.LimitOrder, 0)
			db.On("GetLongOrders").Return(longOrders)
			db.On("GetShortOrders").Return(shortOrders)
			lotp.AssertNotCalled(t, "ExecuteMatchedOrdersTx")
			lop.RunMatchingEngine()
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
