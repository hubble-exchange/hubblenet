package evm

import (
	"bytes"
	"encoding/gob"
	"time"

	"github.com/ava-labs/subnet-evm/plugin/evm/message"
	"github.com/ava-labs/subnet-evm/plugin/evm/orderbook"
	hu "github.com/ava-labs/subnet-evm/plugin/evm/orderbook/hubbleutils"
	"github.com/ethereum/go-ethereum/log"
)

func (n *legacyPushGossiper) GossipSignedOrders(orders []*hu.SignedOrder) error {
	select {
	case n.ordersToGossipChan <- orders:
	case <-n.shutdownChan:
	}
	return nil
}

func (n *legacyPushGossiper) awaitSignedOrderGossip() {
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

func (n *legacyPushGossiper) gossipSignedOrders() (int, error) {
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

func (n *legacyPushGossiper) sendSignedOrders(orders []*hu.SignedOrder) error {
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
	n.stats.IncSignedOrdersGossipSent(int64(len(orders)))
	n.stats.IncSignedOrdersGossipBatchSent()
	return n.client.LegacyGossip(msgBytes)
}
