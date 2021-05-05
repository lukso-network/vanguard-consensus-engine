package api

import (
	"context"
	"github.com/ethereum/go-ethereum/event"
	eth2Types "github.com/prysmaticlabs/eth2-types"
	eth "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"github.com/prysmaticlabs/prysm/beacon-chain/rpc/beacon"
	"github.com/prysmaticlabs/prysm/beacon-chain/rpc/subscriber/api/events"
)

type APIBackend struct {
	BeaconChain       beacon.Server
	consensusInfoFeed event.Feed
	scope             event.SubscriptionScope
}

func (backend *APIBackend) SubscribeNewEpochEvent(ch chan<- *events.MinimalEpochConsensusInfo) event.Subscription {
	return backend.scope.Track(backend.consensusInfoFeed.Subscribe(ch))
}

func (backend *APIBackend) GetProposerListForEpoch(ctx context.Context, epoch eth2Types.Epoch) (*eth.ValidatorAssignments, error) {
	return backend.BeaconChain.GetProposerListForEpoch(ctx, epoch)
}

func (backend *APIBackend) GetMinimalConsensusInfo(ctx context.Context, epoch eth2Types.Epoch) (*events.MinimalEpochConsensusInfo, error) {
	return backend.BeaconChain.GetMinimalConsensusInfo(ctx, epoch)
}

func (backend *APIBackend) GetMinimalConsensusInfoRange(
	ctx context.Context,
	epoch eth2Types.Epoch,
) ([]*events.MinimalEpochConsensusInfo, error) {
	return backend.BeaconChain.GetMinimalConsensusInfoRange(ctx, epoch)
}

func (backend *APIBackend) GetBeaconChain() (beaconChain beacon.Server) {
	return backend.BeaconChain
}
