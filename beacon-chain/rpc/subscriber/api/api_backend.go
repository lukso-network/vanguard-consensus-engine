package api

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/event"
	eth2Types "github.com/prysmaticlabs/eth2-types"
	eth "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"github.com/prysmaticlabs/prysm/beacon-chain/blockchain"
	"github.com/prysmaticlabs/prysm/beacon-chain/core/feed"
	statefeed "github.com/prysmaticlabs/prysm/beacon-chain/core/feed/state"
	"github.com/prysmaticlabs/prysm/beacon-chain/rpc/beacon"
	"github.com/prysmaticlabs/prysm/beacon-chain/rpc/subscriber/api/events"
	"github.com/prysmaticlabs/prysm/beacon-chain/state"
	"github.com/prysmaticlabs/prysm/shared/params"
	log "github.com/sirupsen/logrus"
)

type APIBackend struct {
	BeaconChain       beacon.Server
	consensusInfoFeed event.Feed
	scope             event.SubscriptionScope
}

func (backend *APIBackend) SubscribeNewEpochEvent(
	ctx context.Context,
	epoch eth2Types.Epoch,
	consensusChannel chan interface{},
) {
	beaconChain := backend.BeaconChain
	headFetcher := beaconChain.HeadFetcher
	headState, err := headFetcher.HeadState(ctx)

	if err != nil {
		log.Warn("Could not access head state")

		return
	}

	if headState == nil {
		err := fmt.Errorf("we are not ready to serve information")
		log.WithField("fromEpoch", epoch).Error(err.Error())

		return
	}

	go handleMinimalConsensusSubscription(
		ctx,
		headFetcher,
		headState,
		beaconChain,
		consensusChannel,
	)

	return
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

func handleMinimalConsensusSubscription(
	ctx context.Context,
	headFetcher blockchain.HeadFetcher,
	headState *state.BeaconState,
	beaconChain beacon.Server,
	consensusInfoChannel chan interface{},
) {
	subscriptionStartEpoch := eth2Types.Epoch(headState.Slot() / params.BeaconConfig().SlotsPerEpoch)
	stateChannel := make(chan *feed.Event, params.BeaconConfig().SlotsPerEpoch)
	stateNotifier := beaconChain.StateNotifier
	stateFeed := stateNotifier.StateFeed()
	stateSubscription := stateFeed.Subscribe(stateChannel)

	defer stateSubscription.Unsubscribe()

	if nil == headState {
		panic("head state cannot be nil")
	}

	log.WithField("fromEpoch", subscriptionStartEpoch).Info("registered new subscriber for consensus info")

	for {
		stateEvent := <-stateChannel

		if statefeed.BlockProcessed != stateEvent.Type {
			continue
		}

		currentHeadState, currentErr := headFetcher.HeadState(ctx)

		if nil != currentErr {
			log.Error("could not fetch state during minimalConsensusInfoCheck")
			continue
		}

		blockEpoch := eth2Types.Epoch(currentHeadState.Slot() / params.BeaconConfig().SlotsPerEpoch)

		// Epoch did not progress
		if blockEpoch == subscriptionStartEpoch {
			continue
		}

		subscriptionStartEpoch = blockEpoch
		consensusInfo, currentErr := beaconChain.GetMinimalConsensusInfo(ctx, blockEpoch)

		if nil != currentErr {
			log.WithField("currentEpoch", blockEpoch).WithField("err", currentErr).
				Error("could not retrieve epoch in subscription")
		}

		consensusInfoChannel <- consensusInfo
	}
}
