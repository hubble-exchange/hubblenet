package evm

import (
	"bytes"
	"context"
	"encoding/gob"
	"sync"
	"time"

	"github.com/ava-labs/avalanchego/codec"
	"github.com/ava-labs/avalanchego/snow"
	commonEng "github.com/ava-labs/avalanchego/snow/engine/common"
	"github.com/ava-labs/subnet-evm/hubbleutils"
	hu "github.com/ava-labs/subnet-evm/hubbleutils"
	"github.com/ava-labs/subnet-evm/orderbook"
	"github.com/ava-labs/subnet-evm/plugin/evm/message"
	"github.com/ethereum/go-ethereum/log"
)

const (
	// [ordersGossipInterval] is how often we attempt to gossip newly seen
	// signed orders to other nodes.
	ordersGossipInterval = 100 * time.Millisecond

	// [minGossipOrdersBatchInterval] is the minimum amount of time that must pass
	// before our last gossip to peers.
	minGossipOrdersBatchInterval = 50 * time.Millisecond

	// [maxSignedOrdersGossipBatchSize] is the maximum number of orders we will
	// attempt to gossip at once.
	maxSignedOrdersGossipBatchSize = 100
)

type OrderGossiper interface {
	// GossipSignedOrders sends signed orders to the network
	GossipSignedOrders(orders []*hubbleutils.SignedOrder) error
}

type orderPushGossiper struct {
	ctx    *snow.Context
	config Config

	shutdownChan chan struct{}
	shutdownWg   *sync.WaitGroup

	ordersToGossipChan chan []*hubbleutils.SignedOrder
	ordersToGossip     []*hubbleutils.SignedOrder
	lastOrdersGossiped time.Time

	codec codec.Manager
	stats GossipStats

	appSender commonEng.AppSender
}

// createOrderGossiper constructs and returns a orderPushGossiper or noopGossiper
func (vm *VM) createOrderGossiper(
	stats GossipStats,
) OrderGossiper {
	net := &orderPushGossiper{
		ctx:                vm.ctx,
		config:             vm.config,
		shutdownChan:       vm.shutdownChan,
		shutdownWg:         &vm.shutdownWg,
		codec:              vm.networkCodec,
		stats:              stats,
		ordersToGossipChan: make(chan []*hubbleutils.SignedOrder),
		ordersToGossip:     []*hubbleutils.SignedOrder{},
		appSender:          vm.p2pSender,
	}

	net.awaitSignedOrderGossip()
	return net
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
	msgBytes, err := message.BuildGossipMessage(n.codec, msg)
	if err != nil {
		return err
	}

	log.Trace(
		"gossiping signed orders",
		"len(orders)", len(orders),
		"size(orders)", len(msg.Orders),
	)

	validators := n.config.OrderGossipNumValidators
	nonValidators := n.config.OrderGossipNumNonValidators
	peers := n.config.OrderGossipNumPeers
	err = n.appSender.SendAppGossip(context.TODO(), msgBytes, validators, nonValidators, peers)
	if err != nil {
		log.Error("failed to gossip orders")
		return err
	}
	n.stats.IncSignedOrdersGossipSent(int64(len(orders)))
	n.stats.IncSignedOrdersGossipBatchSent()
	return nil
}
