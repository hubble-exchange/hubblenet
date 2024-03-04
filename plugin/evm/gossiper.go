// (c) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package evm

import (
	"sync"
	"time"

	"github.com/ava-labs/avalanchego/codec"

	"github.com/ava-labs/subnet-evm/peer"

	"github.com/ava-labs/avalanchego/snow"

	"github.com/ava-labs/subnet-evm/core"
	"github.com/ava-labs/subnet-evm/core/types"
	"github.com/ava-labs/subnet-evm/plugin/evm/orderbook/hubbleutils"
)

const (
	// [ordersGossipInterval] is how often we attempt to gossip newly seen
	// signed orders to other nodes.
	ordersGossipInterval = 100 * time.Millisecond

	// [minGossipBatchInterval] is the minimum amount of time that must pass
	// before our last gossip to peers.
	minGossipBatchInterval = 50 * time.Millisecond

	// [minGossipOrdersBatchInterval] is the minimum amount of time that must pass
	// before our last gossip to peers.
	minGossipOrdersBatchInterval = 50 * time.Millisecond

	// [maxSignedOrdersGossipBatchSize] is the maximum number of orders we will
	// attempt to gossip at once.
	maxSignedOrdersGossipBatchSize = 100
)

type Gossiper interface {
	// GossipSignedOrders sends signed orders to the network
	GossipSignedOrders(orders []*hubbleutils.SignedOrder) error
}

// pushGossiper is used to gossip transactions to the network
type pushGossiper struct {
	ctx    *snow.Context
	config Config

	client     peer.NetworkClient
	blockchain *core.BlockChain

	shutdownChan chan struct{}
	shutdownWg   *sync.WaitGroup

	ordersToGossipChan chan []*hubbleutils.SignedOrder
	ordersToGossip     []*hubbleutils.SignedOrder
	lastOrdersGossiped time.Time

	codec  codec.Manager
	signer types.Signer
	stats  GossipSentStats
}

// createGossiper constructs and returns a pushGossiper or noopGossiper
// based on whether vm.chainConfig.SubnetEVMTimestamp is set
func (vm *VM) createGossiper(
	stats GossipStats,
) Gossiper {
	net := &pushGossiper{
		ctx:                vm.ctx,
		config:             vm.config,
		client:             vm.client,
		blockchain:         vm.blockChain,
		shutdownChan:       vm.shutdownChan,
		shutdownWg:         &vm.shutdownWg,
		codec:              vm.networkCodec,
		signer:             types.LatestSigner(vm.blockChain.Config()),
		stats:              stats,
		ordersToGossipChan: make(chan []*hubbleutils.SignedOrder),
		ordersToGossip:     []*hubbleutils.SignedOrder{},
	}

	net.awaitSignedOrderGossip()
	return net
}

// addrStatus used to track the metadata of addresses being queued for
// regossip.
type addrStatus struct {
	nonce    uint64
	txsAdded int
}

// GossipHandler handles incoming gossip messages
type SignedOrderGossipHandler struct {
	mu    sync.RWMutex
	vm    *VM
	stats GossipReceivedStats
}

func NewSignedOrderGossipHandler(vm *VM, stats GossipReceivedStats) *SignedOrderGossipHandler {
	return &SignedOrderGossipHandler{
		vm:    vm,
		stats: stats,
	}
}
