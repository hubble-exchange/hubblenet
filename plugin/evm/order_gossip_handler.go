package evm

import (
	"bytes"
	"encoding/gob"
	"sync"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/subnet-evm/core/txpool"
	"github.com/ava-labs/subnet-evm/plugin/evm/message"
	hu "github.com/ava-labs/subnet-evm/plugin/evm/orderbook/hubbleutils"
	"github.com/ethereum/go-ethereum/log"
)

// GossipHandler handles incoming gossip messages
type LegacyGossipHandler struct {
	mu     sync.RWMutex
	vm     *VM
	txPool *txpool.TxPool
	stats  GossipStats
}

func NewLegacyGossipHandler(vm *VM, stats GossipStats) *LegacyGossipHandler {
	return &LegacyGossipHandler{
		vm:     vm,
		txPool: vm.txPool,
		stats:  stats,
	}
}
func (h *LegacyGossipHandler) HandleSignedOrders(nodeID ids.NodeID, msg message.SignedOrdersGossip) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	log.Trace(
		"AppGossip called with SignedOrdersGossip",
		"peerID", nodeID,
		"bytes(orders)", len(msg.Orders),
	)

	if len(msg.Orders) == 0 {
		log.Warn(
			"AppGossip received empty SignedOrdersGossip Message",
			"peerID", nodeID,
		)
		return nil
	}

	orders := make([]*hu.SignedOrder, 0)
	buf := bytes.NewBuffer(msg.Orders)
	err := gob.NewDecoder(buf).Decode(&orders)
	if err != nil {
		log.Error("failed to decode signed orders", "err", err)
		return err
	}

	h.stats.IncSignedOrdersGossipReceived(int64(len(orders)))
	h.stats.IncSignedOrdersGossipBatchReceived()

	tradingAPI := h.vm.limitOrderProcesser.GetTradingAPI()

	// re-gossip orders, but not when we already knew the orders
	ordersToGossip := make([]*hu.SignedOrder, 0)
	for _, order := range orders {
		_, err := tradingAPI.PlaceOrder(order)
		if err == nil {
			h.stats.IncSignedOrdersGossipReceivedNew()
			ordersToGossip = append(ordersToGossip, order)
		} else if err == hu.ErrOrderAlreadyExists {
			h.stats.IncSignedOrdersGossipReceivedKnown()
		} else {
			h.stats.IncSignedOrdersGossipReceiveError()
			log.Error("failed to place order", "err", err)
		}
	}

	if len(ordersToGossip) > 0 {
		h.vm.legacyGossiper.GossipSignedOrders(ordersToGossip)
	}

	return nil
}
