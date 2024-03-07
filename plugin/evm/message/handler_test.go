// (c) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package message

import (
	"testing"

	"github.com/ava-labs/avalanchego/ids"

	"github.com/stretchr/testify/assert"
)

type CounterHandler struct {
	Orders int
	EthTxs int
}

func (h *CounterHandler) HandleEthTxs(ids.NodeID, EthTxsGossip) error {
	h.EthTxs++
	return nil
}

func (h *CounterHandler) HandleSignedOrders(ids.NodeID, SignedOrdersGossip) error {
	h.Orders++
	return nil
}

func TestHandleEthTxs(t *testing.T) {
	assert := assert.New(t)

	handler := CounterHandler{}
	msg := EthTxsGossip{}

	err := msg.Handle(&handler, ids.EmptyNodeID)
	assert.NoError(err)
	assert.Equal(1, handler.EthTxs)
}

func TestHandleSignedOrders(t *testing.T) {
	assert := assert.New(t)

	handler := CounterHandler{}
	msg := SignedOrdersGossip{}

	err := msg.Handle(&handler, ids.EmptyNodeID)
	assert.NoError(err)
	assert.Equal(1, handler.Orders)
}

func HandleEthTxs(t *testing.T) {
	assert := assert.New(t)

	handler := NoopMempoolGossipHandler{}

	err := handler.HandleEthTxs(ids.EmptyNodeID, EthTxsGossip{})
	assert.NoError(err)
}
func HandleSignedOrders(t *testing.T) {
	assert := assert.New(t)

	handler := NoopMempoolGossipHandler{}

	err := handler.HandleSignedOrders(ids.EmptyNodeID, SignedOrdersGossip{})
	assert.NoError(err)
}
