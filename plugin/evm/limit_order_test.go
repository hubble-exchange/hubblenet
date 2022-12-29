package evm

import (
	"fmt"
	"testing"

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
	return NewLimitOrderProcesser(
		vm.ctx,
		vm.chainConfig,
		vm.txPool,
		vm.shutdownChan,
		&vm.shutdownWg,
		vm.eth.APIBackend,
		vm.eth.BlockChain(),
	)

}
func TestNewLimitOrderProcesser(t *testing.T) {
	vm := newVM(t)
	lop := newLimitOrderProcesser(vm)
	assert.NotNil(t, lop)
}

func TestRunMatchingEngineWhenNoOrders(t *testing.T) {
	vm := newVM(t)
	lop := newLimitOrderProcesser(vm)
	pendingTxBeforeMatching := vm.txPool.Pending(true)
	assert.Equal(t, 0, len(pendingTxBeforeMatching))
	lop.RunMatchingEngine()
	pendingTxAfterMatching := vm.txPool.Pending(true)
	assert.Equal(t, 0, len(pendingTxAfterMatching))
}
