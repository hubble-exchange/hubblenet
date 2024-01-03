// (c) 2022, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package evm

import "github.com/ava-labs/subnet-evm/metrics"

var _ GossipStats = &gossipStats{}

// GossipStats contains methods for updating incoming and outgoing gossip stats.
type GossipStats interface {
	GossipReceivedStats
	GossipSentStats
}

// GossipReceivedStats groups functions for incoming gossip stats.
type GossipReceivedStats interface {
	IncEthTxsGossipReceived()

	// new vs. known txs received
	IncEthTxsGossipReceivedKnown()
	IncEthTxsGossipReceivedNew()

	IncSignedOrdersGossipReceived()

	// new vs. known txs received
	IncSignedOrdersGossipReceivedKnown()
	IncSignedOrdersGossipReceivedNew()
	IncSignedOrdersGossipReceiveError()
}

// GossipSentStats groups functions for outgoing gossip stats.
type GossipSentStats interface {
	IncEthTxsGossipSent()

	// regossip
	IncEthTxsRegossipQueued()
	IncEthTxsRegossipQueuedLocal(count int)
	IncEthTxsRegossipQueuedRemote(count int)

	IncSignedOrdersGossipSent()
	IncSignedOrdersGossipSendError()

	// regossip
	IncSignedOrdersRegossipQueued()
	IncSignedOrdersRegossipQueuedLocal(count int)
	IncSignedOrdersRegossipQueuedRemote(count int)
}

// gossipStats implements stats for incoming and outgoing gossip stats.
type gossipStats struct {
	// messages
	ethTxsGossipSent     metrics.Counter
	ethTxsGossipReceived metrics.Counter

	// regossip
	ethTxsRegossipQueued       metrics.Counter
	ethTxsRegossipQueuedLocal  metrics.Counter
	ethTxsRegossipQueuedRemote metrics.Counter

	// new vs. known txs received
	ethTxsGossipReceivedKnown metrics.Counter
	ethTxsGossipReceivedNew   metrics.Counter

	// messages
	signedOrdersGossipSent      metrics.Counter
	signedOrdersGossipSendError metrics.Counter
	signedOrdersGossipReceived  metrics.Counter

	// regossip
	signedOrdersRegossipQueued       metrics.Counter
	signedOrdersRegossipQueuedLocal  metrics.Counter
	signedOrdersRegossipQueuedRemote metrics.Counter

	// new vs. known txs received
	signedOrdersGossipReceivedKnown metrics.Counter
	signedOrdersGossipReceivedNew   metrics.Counter
	signedOrdersGossipReceiveError  metrics.Counter
}

func NewGossipStats() GossipStats {
	return &gossipStats{
		ethTxsGossipSent:     metrics.GetOrRegisterCounter("gossip_eth_txs_sent", nil),
		ethTxsGossipReceived: metrics.GetOrRegisterCounter("gossip_eth_txs_received", nil),

		ethTxsRegossipQueued:       metrics.GetOrRegisterCounter("regossip_eth_txs_queued_attempts", nil),
		ethTxsRegossipQueuedLocal:  metrics.GetOrRegisterCounter("regossip_eth_txs_queued_local_tx_count", nil),
		ethTxsRegossipQueuedRemote: metrics.GetOrRegisterCounter("regossip_eth_txs_queued_remote_tx_count", nil),

		ethTxsGossipReceivedKnown: metrics.GetOrRegisterCounter("gossip_eth_txs_received_known", nil),
		ethTxsGossipReceivedNew:   metrics.GetOrRegisterCounter("gossip_eth_txs_received_new", nil),

		signedOrdersGossipSent:         metrics.GetOrRegisterCounter("gossip_signed_orders_sent", nil),
		signedOrdersGossipSendError:    metrics.GetOrRegisterCounter("gossip_signed_orders_send_error", nil),
		signedOrdersGossipReceived:     metrics.GetOrRegisterCounter("gossip_signed_orders_received", nil),
		signedOrdersGossipReceiveError: metrics.GetOrRegisterCounter("gossip_signed_orders_received", nil),

		signedOrdersRegossipQueued:       metrics.GetOrRegisterCounter("regossip_signed_orders_queued_attempts", nil),
		signedOrdersRegossipQueuedLocal:  metrics.GetOrRegisterCounter("regossip_signed_orders_queued_local_tx_count", nil),
		signedOrdersRegossipQueuedRemote: metrics.GetOrRegisterCounter("regossip_signed_orders_queued_remote_tx_count", nil),

		signedOrdersGossipReceivedKnown: metrics.GetOrRegisterCounter("gossip_signed_orders_received_known", nil),
		signedOrdersGossipReceivedNew:   metrics.GetOrRegisterCounter("gossip_signed_orders_received_new", nil),
	}
}

// incoming messages
func (g *gossipStats) IncEthTxsGossipReceived() { g.ethTxsGossipReceived.Inc(1) }

// new vs. known txs received
func (g *gossipStats) IncEthTxsGossipReceivedKnown() { g.ethTxsGossipReceivedKnown.Inc(1) }
func (g *gossipStats) IncEthTxsGossipReceivedNew()   { g.ethTxsGossipReceivedNew.Inc(1) }

// outgoing messages
func (g *gossipStats) IncEthTxsGossipSent() { g.ethTxsGossipSent.Inc(1) }

// regossip
func (g *gossipStats) IncEthTxsRegossipQueued() { g.ethTxsRegossipQueued.Inc(1) }
func (g *gossipStats) IncEthTxsRegossipQueuedLocal(count int) {
	g.ethTxsRegossipQueuedLocal.Inc(int64(count))
}
func (g *gossipStats) IncEthTxsRegossipQueuedRemote(count int) {
	g.ethTxsRegossipQueuedRemote.Inc(int64(count))
}

// incoming messages
func (g *gossipStats) IncSignedOrdersGossipReceived() { g.signedOrdersGossipReceived.Inc(1) }

// new vs. known txs received
func (g *gossipStats) IncSignedOrdersGossipReceivedKnown() { g.signedOrdersGossipReceivedKnown.Inc(1) }
func (g *gossipStats) IncSignedOrdersGossipReceivedNew()   { g.signedOrdersGossipReceivedNew.Inc(1) }
func (g *gossipStats) IncSignedOrdersGossipReceiveError()  { g.signedOrdersGossipReceiveError.Inc(1) }

// outgoing messages
func (g *gossipStats) IncSignedOrdersGossipSent()      { g.signedOrdersGossipSent.Inc(1) }
func (g *gossipStats) IncSignedOrdersGossipSendError() { g.signedOrdersGossipSendError.Inc(1) }

// regossip
func (g *gossipStats) IncSignedOrdersRegossipQueued() { g.signedOrdersRegossipQueued.Inc(1) }
func (g *gossipStats) IncSignedOrdersRegossipQueuedLocal(count int) {
	g.signedOrdersRegossipQueuedLocal.Inc(int64(count))
}
func (g *gossipStats) IncSignedOrdersRegossipQueuedRemote(count int) {
	g.signedOrdersRegossipQueuedRemote.Inc(int64(count))
}
