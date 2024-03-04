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
	IncSignedOrdersGossipReceived(count int64)
	IncSignedOrdersGossipBatchReceived()
	IncSignedOrdersGossipReceivedKnown()
	IncSignedOrdersGossipReceivedNew()
	IncSignedOrdersGossipReceiveError()
}

// GossipSentStats groups functions for outgoing gossip stats.
type GossipSentStats interface {
	IncSignedOrdersGossipSent(count int64)
	IncSignedOrdersGossipBatchSent()
	IncSignedOrdersGossipSendError()
	IncSignedOrdersGossipOrderExpired()
}

// gossipStats implements stats for incoming and outgoing gossip stats.
type gossipStats struct {
	signedOrdersGossipSent          metrics.Counter
	signedOrdersGossipBatchSent     metrics.Counter
	signedOrdersGossipSendError     metrics.Counter
	signedOrdersGossipOrderExpired  metrics.Counter
	signedOrdersGossipReceived      metrics.Counter
	signedOrdersGossipBatchReceived metrics.Counter
	signedOrdersGossipReceivedKnown metrics.Counter
	signedOrdersGossipReceivedNew   metrics.Counter
	signedOrdersGossipReceiveError  metrics.Counter
}

func NewGossipStats() GossipStats {
	return &gossipStats{
		signedOrdersGossipSent:          metrics.GetOrRegisterCounter("gossip_signed_orders_sent", nil),
		signedOrdersGossipBatchSent:     metrics.GetOrRegisterCounter("gossip_signed_orders_batch_sent", nil),
		signedOrdersGossipSendError:     metrics.GetOrRegisterCounter("gossip_signed_orders_send_error", nil),
		signedOrdersGossipOrderExpired:  metrics.GetOrRegisterCounter("gossip_signed_orders_expired", nil),
		signedOrdersGossipReceived:      metrics.GetOrRegisterCounter("gossip_signed_orders_received", nil),
		signedOrdersGossipBatchReceived: metrics.GetOrRegisterCounter("gossip_signed_orders_batch_received", nil),
		signedOrdersGossipReceiveError:  metrics.GetOrRegisterCounter("gossip_signed_orders_received", nil),
		signedOrdersGossipReceivedKnown: metrics.GetOrRegisterCounter("gossip_signed_orders_received_known", nil),
		signedOrdersGossipReceivedNew:   metrics.GetOrRegisterCounter("gossip_signed_orders_received_new", nil),
	}
}

// outgoing messages
func (g *gossipStats) IncSignedOrdersGossipSent(count int64) { g.signedOrdersGossipSent.Inc(count) }
func (g *gossipStats) IncSignedOrdersGossipBatchSent()       { g.signedOrdersGossipBatchSent.Inc(1) }
func (g *gossipStats) IncSignedOrdersGossipSendError()       { g.signedOrdersGossipSendError.Inc(1) }
func (g *gossipStats) IncSignedOrdersGossipOrderExpired()    { g.signedOrdersGossipOrderExpired.Inc(1) }

// incoming messages
func (g *gossipStats) IncSignedOrdersGossipReceived(count int64) {
	g.signedOrdersGossipReceived.Inc(count)
}
func (g *gossipStats) IncSignedOrdersGossipBatchReceived() { g.signedOrdersGossipBatchReceived.Inc(1) }

// new vs. known txs received
func (g *gossipStats) IncSignedOrdersGossipReceiveError()  { g.signedOrdersGossipReceiveError.Inc(1) }
func (g *gossipStats) IncSignedOrdersGossipReceivedKnown() { g.signedOrdersGossipReceivedKnown.Inc(1) }
func (g *gossipStats) IncSignedOrdersGossipReceivedNew()   { g.signedOrdersGossipReceivedNew.Inc(1) }
