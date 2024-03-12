package evm

import (
	"bytes"
	"context"
	"encoding/gob"
	"sync"
	"time"

	"github.com/ava-labs/avalanchego/cache"
	"github.com/ava-labs/avalanchego/codec"
	"github.com/ava-labs/avalanchego/snow"
	commonEng "github.com/ava-labs/avalanchego/snow/engine/common"
	"github.com/ava-labs/subnet-evm/core"
	"github.com/ava-labs/subnet-evm/core/txpool"
	"github.com/ava-labs/subnet-evm/core/types"
	"github.com/ava-labs/subnet-evm/peer"
	"github.com/ava-labs/subnet-evm/plugin/evm/message"
	"github.com/ava-labs/subnet-evm/plugin/evm/orderbook"
	"github.com/ava-labs/subnet-evm/plugin/evm/orderbook/hubbleutils"
	hu "github.com/ava-labs/subnet-evm/plugin/evm/orderbook/hubbleutils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
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

var _ OrderGossiper = &orderPushGossiper{}

// Gossiper handles outgoing gossip of transactions
type OrderGossiper interface {
	// GossipSignedOrders sends signed orders to the network
	GossipSignedOrders(orders []*hubbleutils.SignedOrder) error
}

// pushGossiper is used to gossip transactions to the network
type orderPushGossiper struct {
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

	appSender commonEng.AppSender
}

// createGossiper constructs and returns a pushGossiper or noopGossiper
// based on whether vm.chainConfig.SubnetEVMTimestamp is set
func (vm *VM) createGossiper(
	stats GossipStats,
) OrderGossiper {
	net := &orderPushGossiper{
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
		appSender:          vm.p2pSender,
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

func (n *orderPushGossiper) GossipSignedOrders(orders []*hu.SignedOrder) error {
	select {
	case n.ordersToGossipChan <- orders:
	case <-n.shutdownChan:
	}
	return nil
}

func (n *orderPushGossiper) awaitSignedOrderGossip() {
	n.shutdownWg.Add(1)
	go executeFuncAndRecoverPanic(func() {
		var (
			gossipTicker = time.NewTicker(ordersGossipInterval)
		)
		defer func() {
			gossipTicker.Stop()
			n.shutdownWg.Done()
		}()

		for {
			select {
			case <-gossipTicker.C:
				if attempted, err := n.gossipSignedOrders(); err != nil {
					log.Warn(
						"failed to send signed orders",
						"len(orders)", attempted,
						"err", err,
					)
				}
			case orders := <-n.ordersToGossipChan:
				for _, order := range orders {
					n.ordersToGossip = append(n.ordersToGossip, order)
				}
				if attempted, err := n.gossipSignedOrders(); err != nil {
					log.Warn(
						"failed to send signed orders",
						"len(orders)", attempted,
						"err", err,
					)
				}
			case <-n.shutdownChan:
				return
			}
		}
	}, "panic in awaitSignedOrderGossip", orderbook.AwaitSignedOrdersGossipPanicsCounter)
}

func (n *orderPushGossiper) gossipSignedOrders() (int, error) {
	if (time.Since(n.lastOrdersGossiped) < minGossipOrdersBatchInterval) || len(n.ordersToGossip) == 0 {
		return 0, nil
	}
	n.lastOrdersGossiped = time.Now()
	now := time.Now().Unix()
	selectedOrders := []*hu.SignedOrder{}
	numConsumed := 0
	for _, order := range n.ordersToGossip {
		if len(selectedOrders) >= maxSignedOrdersGossipBatchSize {
			break
		}
		numConsumed++
		if order.ExpireAt.Int64() < now {
			n.stats.IncSignedOrdersGossipOrderExpired()
			log.Warn("signed order expired before gossip", "order", order, "now", now)
			continue
		}
		selectedOrders = append(selectedOrders, order)
	}
	// delete all selected orders from n.ordersToGossip
	n.ordersToGossip = n.ordersToGossip[numConsumed:]

	if len(selectedOrders) == 0 {
		return 0, nil
	}

	err := n.sendSignedOrders(selectedOrders)
	if err != nil {
		n.stats.IncSignedOrdersGossipSendError()
	}
	return len(selectedOrders), err
}

func (n *orderPushGossiper) sendSignedOrders(orders []*hu.SignedOrder) error {
	if len(orders) == 0 {
		return nil
	}

	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(&orders)
	if err != nil {
		return err
	}
	ordersBytes := buf.Bytes()
	msg := message.SignedOrdersGossip{
		Orders: ordersBytes,
	}
	msgBytes, err := message.BuildOrderGossipMessage(n.codec, msg)
	if err != nil {
		return err
	}

	log.Trace(
		"gossiping signed orders",
		"len(orders)", len(orders),
		"size(orders)", len(msg.Orders),
	)
	n.stats.IncSignedOrdersGossipSent(int64(len(orders)))
	n.stats.IncSignedOrdersGossipBatchSent()

	// TODO: fix the number of validators, non-validators, and peers
	return n.appSender.SendAppGossip(context.TODO(), msgBytes, 0, 0, 0)
}
