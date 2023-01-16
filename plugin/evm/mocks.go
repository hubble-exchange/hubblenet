package evm

import (
	"github.com/ava-labs/subnet-evm/core/types"
	"github.com/ava-labs/subnet-evm/eth"
	"github.com/ava-labs/subnet-evm/plugin/evm/limitorders"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/mock"
)

type MockLimitOrderDatabase struct {
	mock.Mock
}

func NewMockLimitOrderDatabase() *MockLimitOrderDatabase {
	return &MockLimitOrderDatabase{}
}

func (db *MockLimitOrderDatabase) GetAllOrders() []limitorders.LimitOrder {
	args := db.Called()
	return args.Get(0).([]limitorders.LimitOrder)
}

func (db *MockLimitOrderDatabase) Add(order *limitorders.LimitOrder) {
}

func (db *MockLimitOrderDatabase) UpdateFilledBaseAssetQuantity(quantity uint, signature []byte) {
}

func (db *MockLimitOrderDatabase) Delete(signature []byte) {
}

func (db *MockLimitOrderDatabase) GetLongOrders(market limitorders.Market) []limitorders.LimitOrder {
	args := db.Called()
	return args.Get(0).([]limitorders.LimitOrder)
}

func (db *MockLimitOrderDatabase) GetShortOrders(market limitorders.Market) []limitorders.LimitOrder {
	args := db.Called()
	return args.Get(0).([]limitorders.LimitOrder)
}

func (db *MockLimitOrderDatabase) UpdatePosition(trader common.Address, market limitorders.Market, size float64, openNotional float64) {	
}

func (db *MockLimitOrderDatabase) UpdateMargin(trader common.Address, collateral limitorders.Collateral, addAmount float64) {
}

func (db *MockLimitOrderDatabase) UpdateUnrealisedFunding(market limitorders.Market, fundingRate float64) {
}

func (db *MockLimitOrderDatabase) ResetUnrealisedFunding(market limitorders.Market, trader common.Address) {
}

func (db *MockLimitOrderDatabase) UpdateNextFundingTime() {
}

func (db *MockLimitOrderDatabase) GetNextFundingTime() uint64 {
	return 0
}

func (db *MockLimitOrderDatabase) GetLiquidableTraders(market limitorders.Market, markPrice float64, oraclePrice float64) []limitorders.Liquidable {
	return nil
}

func (db *MockLimitOrderDatabase) UpdateLastPrice(market limitorders.Market,lastPrice float64) {
}

func (db *MockLimitOrderDatabase) GetLastPrice(market limitorders.Market) float64 {
	return 0
}

type MockLimitOrderTxProcessor struct {
	mock.Mock
}

func NewMockLimitOrderTxProcessor() *MockLimitOrderTxProcessor {
	return &MockLimitOrderTxProcessor{}
}

func (lotp *MockLimitOrderTxProcessor) ExecuteMatchedOrdersTx(incomingOrder limitorders.LimitOrder, matchedOrder limitorders.LimitOrder, fillAmount uint) error {
	args := lotp.Called(incomingOrder, matchedOrder, fillAmount)
	return args.Error(0)
}

func (lotp *MockLimitOrderTxProcessor) PurgeLocalTx() {
	lotp.Called()
}

func (lotp *MockLimitOrderTxProcessor) CheckIfOrderBookContractCall(tx *types.Transaction) bool {
	return true
}

func (lotp *MockLimitOrderTxProcessor) ExecuteFundingPaymentTx() error {
	return nil
}

func (lotp *MockLimitOrderTxProcessor) ExecuteLiquidation(trader common.Address, matchedOrder limitorders.LimitOrder) error {
	return nil
}

func (lotp *MockLimitOrderTxProcessor) HandleOrderBookEvent(event *types.Log) {
}

func (lotp *MockLimitOrderTxProcessor) HandleMarginAccountEvent(event *types.Log) {
}

func (lotp *MockLimitOrderTxProcessor) HandleClearingHouseEvent(event *types.Log) {
}
