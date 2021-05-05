package events

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/rpc"
	eth2Types "github.com/prysmaticlabs/eth2-types"
	eth "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"github.com/prysmaticlabs/prysm/beacon-chain/blockchain"
	"github.com/prysmaticlabs/prysm/beacon-chain/core/feed"
	statefeed "github.com/prysmaticlabs/prysm/beacon-chain/core/feed/state"
	"github.com/prysmaticlabs/prysm/beacon-chain/rpc/beacon"
	"github.com/prysmaticlabs/prysm/beacon-chain/state"
	"github.com/prysmaticlabs/prysm/shared/params"
	"time"
)

type Backend interface {
	SubscribeNewEpochEvent(chan<- *MinimalEpochConsensusInfo) event.Subscription
	GetProposerListForEpoch(context.Context, eth2Types.Epoch) (*eth.ValidatorAssignments, error)
	GetMinimalConsensusInfo(context.Context, eth2Types.Epoch) (*MinimalEpochConsensusInfo, error)
	GetMinimalConsensusInfoRange(context.Context, eth2Types.Epoch) ([]*MinimalEpochConsensusInfo, error)
	GetBeaconChain() beacon.Server
}

// PublicFilterAPI offers support to create and manage filters. This will allow external clients to retrieve various
// information related to the Ethereum protocol such als blocks, transactions and logs.
type PublicFilterAPI struct {
	backend Backend
	events  *EventSystem
	timeout time.Duration
}

// NewPublicFilterAPI returns a new PublicFilterAPI instance.
func NewPublicFilterAPI(backend Backend, timeout time.Duration) *PublicFilterAPI {
	api := &PublicFilterAPI{
		backend: backend,
		events:  NewEventSystem(backend),
		timeout: timeout,
	}

	return api
}

// MinimalConsensusInfo is used to serve information about epochs from certain epoch until most recent
// This should be used as a pub/sub live subscription by Orchestrator client
// Due to the fact that a lot of notifications could happen you should use it wisely
func (api *PublicFilterAPI) MinimalConsensusInfo(ctx context.Context, epoch uint64) (*rpc.Subscription, error) {
	notifier, supported := rpc.NotifierFromContext(ctx)
	if !supported {
		return &rpc.Subscription{}, rpc.ErrNotificationsUnsupported
	}

	rpcSub := notifier.CreateSubscription()
	notifier, ok := rpc.NotifierFromContext(ctx)

	if !ok {
		err := fmt.Errorf("could not create notifier")
		log.WithField("context", "GetMinimalConsensusInfoRange").
			WithField("requestedEpoch", epoch).Error(err.Error())

		return nil, err
	}

	events := api.events
	backend := api.backend

	go sendMinimalConsensusRange(
		ctx,
		eth2Types.Epoch(epoch),
		backend,
		notifier,
		rpcSub,
	)

	consensusInfoSub := events.consensusInfoSub
	beaconChain := backend.GetBeaconChain()
	headFetcher := beaconChain.HeadFetcher
	headState, err := headFetcher.HeadState(ctx)

	if err != nil {
		log.Warn("Could not access head state for infostream")
		return nil, err
	}

	if headState == nil {
		err := fmt.Errorf("we are not ready to serve information")
		log.WithField("fromEpoch", epoch).Error(err.Error())
		return nil, err
	}

	go handleMinimalConsensusSubscription(
		ctx,
		headFetcher,
		headState,
		backend,
		notifier,
		rpcSub,
		consensusInfoSub,
	)

	return rpcSub, nil
}

func sendMinimalConsensusRange(
	ctx context.Context,
	epoch eth2Types.Epoch,
	backend Backend,
	notifier *rpc.Notifier,
	rpcSub *rpc.Subscription,
) {
	minimalInfos, err := backend.GetMinimalConsensusInfoRange(ctx, epoch)

	if nil != err {
		log.WithField("err", err.Error()).WithField("epoch", epoch).Error("could not get minimal info")

		return
	}

	log.WithField("range", len(minimalInfos)).Info("I will be sending epochs range")

	for _, consensusInfo := range minimalInfos {
		log.WithField("epoch", consensusInfo.Epoch).Info("sending consensus range to subscriber")
		err = notifier.Notify(rpcSub.ID, consensusInfo)

		if nil != err {
			log.WithField("err", err.Error()).WithField("epoch", epoch).Error("invalid notification")

			return
		}
	}
}

func handleMinimalConsensusSubscription(
	ctx context.Context,
	headFetcher blockchain.HeadFetcher,
	headState *state.BeaconState,
	backend Backend,
	notifier *rpc.Notifier,
	rpcSub *rpc.Subscription,
	consensusInfoSub event.Subscription,
) {
	subscriptionStartEpoch := eth2Types.Epoch(headState.Slot() / params.BeaconConfig().SlotsPerEpoch)
	stateChannel := make(chan *feed.Event, params.BeaconConfig().SlotsPerEpoch)
	beaconChain := backend.GetBeaconChain()
	stateNotifier := beaconChain.StateNotifier
	stateFeed := stateNotifier.StateFeed()
	stateSubscription := stateFeed.Subscribe(stateChannel)

	defer stateSubscription.Unsubscribe()

	if nil == headState {
		panic("head state cannot be nil")
	}

	log.WithField("fromEpoch", subscriptionStartEpoch).Info("registered new subscriber for consensus info")

	for {
		select {
		case stateEvent := <-stateChannel:
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

			log.WithField("epoch", blockEpoch).Info("sending consensus info to subscriber")
			currentErr = notifier.Notify(rpcSub.ID, consensusInfo)

			if nil != currentErr {
				log.WithField("err", currentErr).Error("subscriber notify error")

				continue
			}
		case <-rpcSub.Err():
			log.Info("unsubscribing registered subscriber")
			consensusInfoSub.Unsubscribe()
			return
		case <-notifier.Closed():
			log.Info("unsubscribing registered subscriber")
			consensusInfoSub.Unsubscribe()
			return
		}
	}
}
