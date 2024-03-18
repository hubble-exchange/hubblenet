// (c) 2019-2022, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package peer

import (
	"context"
	"errors"

	"github.com/ava-labs/avalanchego/ids"

	"github.com/ava-labs/avalanchego/version"
)

var (
	_ NetworkClient = &client{}

	ErrRequestFailed = errors.New("request failed")
)

// NetworkClient defines ability to send request / response through the Network
type NetworkClient interface {
	// SendAppRequestAny synchronously sends request to an arbitrary peer with a
	// node version greater than or equal to minVersion.
	// Returns response bytes, the ID of the chosen peer, and ErrRequestFailed if
	// the request should be retried.
	SendAppRequestAny(ctx context.Context, minVersion *version.Application, request []byte) ([]byte, ids.NodeID, error)

	// SendAppRequest synchronously sends request to the selected nodeID
	// Returns response bytes, and ErrRequestFailed if the request should be retried.
	SendAppRequest(ctx context.Context, nodeID ids.NodeID, request []byte) ([]byte, error)

	// SendCrossChainRequest sends a request to a specific blockchain running on this node.
	// Returns response bytes, and ErrRequestFailed if the request failed.
	SendCrossChainRequest(ctx context.Context, chainID ids.ID, request []byte) ([]byte, error)

	// TrackBandwidth should be called for each valid request with the bandwidth
	// (length of response divided by request time), and with 0 if the response is invalid.
	TrackBandwidth(nodeID ids.NodeID, bandwidth float64)

	// FIXME this has been removed in subnet-evm
	LegacyGossip(gossip []byte, validators int, nonValidators int, peers int) error
}

// client implements NetworkClient interface
// provides ability to send request / responses through the Network and wait for a response
// so that the caller gets the result synchronously.
type client struct {
	network Network
}

// NewNetworkClient returns Client for a given network
func NewNetworkClient(network Network) NetworkClient {
	return &client{
		network: network,
	}
}

// SendAppRequestAny synchronously sends request to an arbitrary peer with a
// node version greater than or equal to minVersion.
// Returns response bytes, the ID of the chosen peer, and ErrRequestFailed if
// the request should be retried.
func (c *client) SendAppRequestAny(ctx context.Context, minVersion *version.Application, request []byte) ([]byte, ids.NodeID, error) {
	waitingHandler := newWaitingResponseHandler()
	nodeID, err := c.network.SendAppRequestAny(ctx, minVersion, request, waitingHandler)
	if err != nil {
		return nil, nodeID, err
	}
	response, err := waitingHandler.WaitForResult(ctx)
	return response, nodeID, err
}

// SendAppRequest synchronously sends request to the specified nodeID
// Returns response bytes and ErrRequestFailed if the request should be retried.
func (c *client) SendAppRequest(ctx context.Context, nodeID ids.NodeID, request []byte) ([]byte, error) {
	waitingHandler := newWaitingResponseHandler()
	if err := c.network.SendAppRequest(ctx, nodeID, request, waitingHandler); err != nil {
		return nil, err
	}
	return waitingHandler.WaitForResult(ctx)
}

// SendCrossChainRequest synchronously sends request to the specified chainID
// Returns response bytes and ErrRequestFailed if the request should be retried.
func (c *client) SendCrossChainRequest(ctx context.Context, chainID ids.ID, request []byte) ([]byte, error) {
	waitingHandler := newWaitingResponseHandler()
	if err := c.network.SendCrossChainRequest(ctx, chainID, request, waitingHandler); err != nil {
		return nil, err
	}
	return waitingHandler.WaitForResult(ctx)
}

// FIXME this has been removed in subnet-evm
func (c *client) LegacyGossip(gossip []byte, validators int, nonValidators int, peers int) error {
	return c.network.LegacyGossip(gossip, validators, nonValidators, peers)
}

func (c *client) TrackBandwidth(nodeID ids.NodeID, bandwidth float64) {
	c.network.TrackBandwidth(nodeID, bandwidth)
}
