package evm

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/ava-labs/subnet-evm/accounts/abi"
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

func newLimitOrderProcesser(vm *VM) LimitOrderProcesser {
	memoryDb := limitorders.NewInMemoryDatabase()
	jsonBytes, _ := ioutil.ReadFile(orderBookContractFileLocation)
	orderBookAbi, err := abi.FromSolidityJson(string(jsonBytes))
	if err != nil {
		panic(err)
	}
	lotp := limitorders.NewLimitOrderTxProcessor(vm.txPool, orderBookAbi, memoryDb, orderBookContractAddress)
	lop := NewLimitOrderProcesser(
		vm.ctx,
		vm.txPool,
		vm.shutdownChan,
		&vm.shutdownWg,
		vm.eth.APIBackend,
		vm.eth.BlockChain(),
		memoryDb,
		lotp,
	)
	return lop
}
func TestNewLimitOrderProcesser(t *testing.T) {
	vm := newVM(t)
	lop := newLimitOrderProcesser(vm)
	assert.NotNil(t, lop)
}