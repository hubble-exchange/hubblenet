// (c) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package evm

import (
	"bytes"
	"encoding/gob"
	"sync"

	"github.com/ava-labs/avalanchego/ids"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/ava-labs/subnet-evm/core/txpool"
	"github.com/ava-labs/subnet-evm/core/types"
	"github.com/ava-labs/subnet-evm/plugin/evm/message"
	hu "github.com/ava-labs/subnet-evm/plugin/evm/orderbook/hubbleutils"
)

// GossipHandler handles incoming gossip messages
type GossipHandler struct {
	mu     sync.RWMutex
	vm     *VM
	txPool *txpool.TxPool
	stats  GossipStats
}

func NewGossipHandler(vm *VM, stats GossipStats) *GossipHandler {
	return &GossipHandler{
		vm:     vm,
		txPool: vm.txPool,
		stats:  stats,
	}
}

func (h *GossipHandler) HandleEthTxs(nodeID ids.NodeID, msg message.EthTxsGossip) error {
	log.Trace(
		"AppGossip called with EthTxsGossip",
		"peerID", nodeID,
		"size(txs)", len(msg.Txs),
	)

	if len(msg.Txs) == 0 {
		log.Trace(
			"AppGossip received empty EthTxsGossip Message",
			"peerID", nodeID,
		)
		return nil
	}

	// The maximum size of this encoded object is enforced by the codec.
	txs := make([]*types.Transaction, 0)
	if err := rlp.DecodeBytes(msg.Txs, &txs); err != nil {
		log.Trace(
			"AppGossip provided invalid txs",
			"peerID", nodeID,
			"err", err,
		)
		return nil
	}
	h.stats.IncEthTxsGossipReceived()
	errs := h.txPool.AddRemotes(txs)
	for i, err := range errs {
		if err != nil {
			log.Trace(
				"AppGossip failed to add to mempool",
				"err", err,
				"tx", txs[i].Hash(),
			)
			if err == txpool.ErrAlreadyKnown {
				h.stats.IncEthTxsGossipReceivedKnown()
			} else {
				h.stats.IncEthTxsGossipReceivedError()
			}
			continue
		}
		h.stats.IncEthTxsGossipReceivedNew()
	}
	return nil
}

func (h *GossipHandler) HandleSignedOrders(nodeID ids.NodeID, msg message.SignedOrdersGossip) error {
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
		_, shouldTriggerMatching, err := tradingAPI.PlaceOrder(order)
		if err == nil {
			h.stats.IncSignedOrdersGossipReceivedNew()
			ordersToGossip = append(ordersToGossip, order)
			if shouldTriggerMatching {
				log.Info("received new match-able signed order, triggering matching pipeline...")
				h.vm.limitOrderProcesser.RunMatchingPipeline()
			}
		} else if err == hu.ErrOrderAlreadyExists {
			h.stats.IncSignedOrdersGossipReceivedKnown()
		} else {
			h.stats.IncSignedOrdersGossipReceiveError()
			log.Error("failed to place order", "err", err)
		}
	}

	if len(ordersToGossip) > 0 {
		h.vm.orderGossiper.GossipSignedOrders(ordersToGossip)
	}

	return nil
}
