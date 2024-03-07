// (c) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package evm

import (
	"sync"
	"time"

	"github.com/ava-labs/avalanchego/codec"

	"github.com/ava-labs/subnet-evm/peer"

	"github.com/ava-labs/avalanchego/cache"
	"github.com/ava-labs/avalanchego/snow"

	"github.com/ethereum/go-ethereum/common"

	"github.com/ava-labs/subnet-evm/core"
	"github.com/ava-labs/subnet-evm/core/txpool"
	"github.com/ava-labs/subnet-evm/core/types"
	"github.com/ava-labs/subnet-evm/plugin/evm/orderbook/hubbleutils"
)

const (
	// We allow [recentCacheSize] to be fairly large because we only store hashes
	// in the cache, not entire transactions.
	recentCacheSize = 512

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

// Gossiper handles outgoing gossip of transactions
type LegacyGossiper interface {
	// GossipSignedOrders sends signed orders to the network
	GossipSignedOrders(orders []*hubbleutils.SignedOrder) error
}

// pushGossiper is used to gossip transactions to the network
type legacyPushGossiper struct {
	ctx    *snow.Context
	config Config

	client     peer.NetworkClient
	blockchain *core.BlockChain
	txPool     *txpool.TxPool

	// We attempt to batch transactions we need to gossip to avoid runaway
	// amplification of mempol chatter.
	ethTxsToGossipChan chan []*types.Transaction
	ethTxsToGossip     map[common.Hash]*types.Transaction
	lastGossiped       time.Time
	shutdownChan       chan struct{}
	shutdownWg         *sync.WaitGroup

	ordersToGossipChan chan []*hubbleutils.SignedOrder
	ordersToGossip     []*hubbleutils.SignedOrder
	lastOrdersGossiped time.Time

	// [recentEthTxs] prevent us from over-gossiping the
	// same transaction in a short period of time.
	recentEthTxs *cache.LRU[common.Hash, interface{}]

	codec  codec.Manager
	signer types.Signer
	stats  GossipStats
}

// createGossiper constructs and returns a pushGossiper or noopGossiper
// based on whether vm.chainConfig.SubnetEVMTimestamp is set
func (vm *VM) createGossiper(
	stats GossipStats,
) LegacyGossiper {
	net := &legacyPushGossiper{
		ctx:                vm.ctx,
		config:             vm.config,
		client:             vm.client,
		blockchain:         vm.blockChain,
		txPool:             vm.txPool,
		ethTxsToGossipChan: make(chan []*types.Transaction),
		ethTxsToGossip:     make(map[common.Hash]*types.Transaction),
		shutdownChan:       vm.shutdownChan,
		shutdownWg:         &vm.shutdownWg,
		recentEthTxs:       &cache.LRU[common.Hash, interface{}]{Size: recentCacheSize},
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
