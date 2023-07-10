// Code generated by MockGen. DO NOT EDIT.
// Source: client.go

// Package mock_bibliophile is a generated GoMock package.
package bibliophile

import (
	big "math/big"
	reflect "reflect"

	contract "github.com/ava-labs/subnet-evm/precompile/contract"
	common "github.com/ethereum/go-ethereum/common"
	gomock "github.com/golang/mock/gomock"
)

// MockBibliophileClient is a mock of BibliophileClient interface.
type MockBibliophileClient struct {
	ctrl     *gomock.Controller
	recorder *MockBibliophileClientMockRecorder
}

// MockBibliophileClientMockRecorder is the mock recorder for MockBibliophileClient.
type MockBibliophileClientMockRecorder struct {
	mock *MockBibliophileClient
}

// NewMockBibliophileClient creates a new mock instance.
func NewMockBibliophileClient(ctrl *gomock.Controller) *MockBibliophileClient {
	mock := &MockBibliophileClient{ctrl: ctrl}
	mock.recorder = &MockBibliophileClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockBibliophileClient) EXPECT() *MockBibliophileClientMockRecorder {
	return m.recorder
}

// DetermineFillPrice mocks base method.
func (m *MockBibliophileClient) DetermineFillPrice(marketId int64, longOrderPrice, shortOrderPrice, blockPlaced0, blockPlaced1 *big.Int) (*ValidateOrdersAndDetermineFillPriceOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DetermineFillPrice", marketId, longOrderPrice, shortOrderPrice, blockPlaced0, blockPlaced1)
	ret0, _ := ret[0].(*ValidateOrdersAndDetermineFillPriceOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DetermineFillPrice indicates an expected call of DetermineFillPrice.
func (mr *MockBibliophileClientMockRecorder) DetermineFillPrice(marketId, longOrderPrice, shortOrderPrice, blockPlaced0, blockPlaced1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DetermineFillPrice", reflect.TypeOf((*MockBibliophileClient)(nil).DetermineFillPrice), marketId, longOrderPrice, shortOrderPrice, blockPlaced0, blockPlaced1)
}

// DetermineLiquidationFillPrice mocks base method.
func (m *MockBibliophileClient) DetermineLiquidationFillPrice(marketId int64, baseAssetQuantity, price *big.Int) (*big.Int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DetermineLiquidationFillPrice", marketId, baseAssetQuantity, price)
	ret0, _ := ret[0].(*big.Int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DetermineLiquidationFillPrice indicates an expected call of DetermineLiquidationFillPrice.
func (mr *MockBibliophileClientMockRecorder) DetermineLiquidationFillPrice(marketId, baseAssetQuantity, price interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DetermineLiquidationFillPrice", reflect.TypeOf((*MockBibliophileClient)(nil).DetermineLiquidationFillPrice), marketId, baseAssetQuantity, price)
}

// GetAccessibleState mocks base method.
func (m *MockBibliophileClient) GetAccessibleState() contract.AccessibleState {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAccessibleState")
	ret0, _ := ret[0].(contract.AccessibleState)
	return ret0
}

// GetAccessibleState indicates an expected call of GetAccessibleState.
func (mr *MockBibliophileClientMockRecorder) GetAccessibleState() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAccessibleState", reflect.TypeOf((*MockBibliophileClient)(nil).GetAccessibleState))
}

// GetBlockPlaced mocks base method.
func (m *MockBibliophileClient) GetBlockPlaced(orderHash [32]byte) *big.Int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBlockPlaced", orderHash)
	ret0, _ := ret[0].(*big.Int)
	return ret0
}

// GetBlockPlaced indicates an expected call of GetBlockPlaced.
func (mr *MockBibliophileClientMockRecorder) GetBlockPlaced(orderHash interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBlockPlaced", reflect.TypeOf((*MockBibliophileClient)(nil).GetBlockPlaced), orderHash)
}

// GetMarketAddressFromMarketID mocks base method.
func (m *MockBibliophileClient) GetMarketAddressFromMarketID(marketId int64) common.Address {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMarketAddressFromMarketID", marketId)
	ret0, _ := ret[0].(common.Address)
	return ret0
}

// GetMarketAddressFromMarketID indicates an expected call of GetMarketAddressFromMarketID.
func (mr *MockBibliophileClientMockRecorder) GetMarketAddressFromMarketID(marketId interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMarketAddressFromMarketID", reflect.TypeOf((*MockBibliophileClient)(nil).GetMarketAddressFromMarketID), marketId)
}

// GetMinSizeRequirement mocks base method.
func (m *MockBibliophileClient) GetMinSizeRequirement(marketId int64) *big.Int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMinSizeRequirement", marketId)
	ret0, _ := ret[0].(*big.Int)
	return ret0
}

// GetMinSizeRequirement indicates an expected call of GetMinSizeRequirement.
func (mr *MockBibliophileClientMockRecorder) GetMinSizeRequirement(marketId interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMinSizeRequirement", reflect.TypeOf((*MockBibliophileClient)(nil).GetMinSizeRequirement), marketId)
}

// GetOrderFilledAmount mocks base method.
func (m *MockBibliophileClient) GetOrderFilledAmount(orderHash [32]byte) *big.Int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetOrderFilledAmount", orderHash)
	ret0, _ := ret[0].(*big.Int)
	return ret0
}

// GetOrderFilledAmount indicates an expected call of GetOrderFilledAmount.
func (mr *MockBibliophileClientMockRecorder) GetOrderFilledAmount(orderHash interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOrderFilledAmount", reflect.TypeOf((*MockBibliophileClient)(nil).GetOrderFilledAmount), orderHash)
}

// GetOrderStatus mocks base method.
func (m *MockBibliophileClient) GetOrderStatus(orderHash [32]byte) int64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetOrderStatus", orderHash)
	ret0, _ := ret[0].(int64)
	return ret0
}

// GetOrderStatus indicates an expected call of GetOrderStatus.
func (mr *MockBibliophileClientMockRecorder) GetOrderStatus(orderHash interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOrderStatus", reflect.TypeOf((*MockBibliophileClient)(nil).GetOrderStatus), orderHash)
}

// GetSize mocks base method.
func (m *MockBibliophileClient) GetSize(market common.Address, trader *common.Address) *big.Int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSize", market, trader)
	ret0, _ := ret[0].(*big.Int)
	return ret0
}

// GetSize indicates an expected call of GetSize.
func (mr *MockBibliophileClientMockRecorder) GetSize(market, trader interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSize", reflect.TypeOf((*MockBibliophileClient)(nil).GetSize), market, trader)
}

// IOC_GetBlockPlaced mocks base method.
func (m *MockBibliophileClient) IOC_GetBlockPlaced(orderHash [32]byte) *big.Int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IOC_GetBlockPlaced", orderHash)
	ret0, _ := ret[0].(*big.Int)
	return ret0
}

// IOC_GetBlockPlaced indicates an expected call of IOC_GetBlockPlaced.
func (mr *MockBibliophileClientMockRecorder) IOC_GetBlockPlaced(orderHash interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IOC_GetBlockPlaced", reflect.TypeOf((*MockBibliophileClient)(nil).IOC_GetBlockPlaced), orderHash)
}

// IOC_GetOrderFilledAmount mocks base method.
func (m *MockBibliophileClient) IOC_GetOrderFilledAmount(orderHash [32]byte) *big.Int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IOC_GetOrderFilledAmount", orderHash)
	ret0, _ := ret[0].(*big.Int)
	return ret0
}

// IOC_GetOrderFilledAmount indicates an expected call of IOC_GetOrderFilledAmount.
func (mr *MockBibliophileClientMockRecorder) IOC_GetOrderFilledAmount(orderHash interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IOC_GetOrderFilledAmount", reflect.TypeOf((*MockBibliophileClient)(nil).IOC_GetOrderFilledAmount), orderHash)
}

// IOC_GetOrderStatus mocks base method.
func (m *MockBibliophileClient) IOC_GetOrderStatus(orderHash [32]byte) int64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IOC_GetOrderStatus", orderHash)
	ret0, _ := ret[0].(int64)
	return ret0
}

// IOC_GetOrderStatus indicates an expected call of IOC_GetOrderStatus.
func (mr *MockBibliophileClientMockRecorder) IOC_GetOrderStatus(orderHash interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IOC_GetOrderStatus", reflect.TypeOf((*MockBibliophileClient)(nil).IOC_GetOrderStatus), orderHash)
}