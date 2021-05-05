package events

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/rpc"
	eth2Types "github.com/prysmaticlabs/eth2-types"
	eth "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"time"
)

type Backend interface {
	SubscribeNewEpochEvent(ctx context.Context, epoch eth2Types.Epoch, consensusChannel chan interface{})
	GetProposerListForEpoch(context.Context, eth2Types.Epoch) (*eth.ValidatorAssignments, error)
	GetMinimalConsensusInfo(context.Context, eth2Types.Epoch) (*MinimalEpochConsensusInfo, error)
	GetMinimalConsensusInfoRange(context.Context, eth2Types.Epoch) ([]*MinimalEpochConsensusInfo, error)
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
		log.WithField("context", "MinimalConsensusInfo").
			WithField("requestedEpoch", epoch).Error(err.Error())

		return nil, err
	}

	backend := api.backend

	go sendMinimalConsensusRange(
		ctx,
		eth2Types.Epoch(epoch),
		backend,
		notifier,
		rpcSub,
	)

	subscriptionChannel := make(chan interface{})
	backend.SubscribeNewEpochEvent(ctx, eth2Types.Epoch(epoch), subscriptionChannel)

	go func() {
		for {
			information := <-subscriptionChannel
			consensusInfo, isConsensusInfo := information.(*MinimalEpochConsensusInfo)

			if !isConsensusInfo {
				log.WithField("context", "MinimalConsensusInfo").
					WithField("requestedEpoch", epoch).Error("received invalid type on channel")
				continue
			}

			currentErr := notifier.Notify(rpcSub.ID, consensusInfo)

			if nil != currentErr {
				log.WithField("context", "MinimalConsensusInfo").
					WithField("requestedEpoch", epoch).
					WithField("err", currentErr).
					Error("could not send notification")

				return
			}
		}
	}()

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
