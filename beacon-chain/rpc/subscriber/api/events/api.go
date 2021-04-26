package events

import (
	"context"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/rpc"
	eth2Types "github.com/prysmaticlabs/eth2-types"
	eth "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"time"
)

type Backend interface {
	SubscribeNewEpochEvent(chan<- *MinimalEpochConsensusInfo) event.Subscription
	GetProposerListForEpoch(context.Context, eth2Types.Epoch) (*eth.ValidatorAssignments, error)
	GetMinimalConsensusInfo(context.Context, eth2Types.Epoch) (*MinimalEpochConsensusInfo, error)
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

// MinimalConsensusInfo
func (api *PublicFilterAPI) MinimalConsensusInfo(ctx context.Context, epoch uint64) (*rpc.Subscription, error) {
	notifier, supported := rpc.NotifierFromContext(ctx)
	if !supported {
		return &rpc.Subscription{}, rpc.ErrNotificationsUnsupported
	}
	rpcSub := notifier.CreateSubscription()

	go func() {
		consensusInfo := make(chan *MinimalEpochConsensusInfo)
		consensusInfoSub := api.events.SubscribeConsensusInfo(consensusInfo, eth2Types.Epoch(epoch))
		log.WithField("fromEpoch", epoch).Debug("registered new subscriber for consensus info")

		for {
			select {
			case c := <-consensusInfo:
				log.WithField("epoch", c.Epoch).Info("sending consensus info to subscriber")
				err := notifier.Notify(rpcSub.ID, c)
				if nil != err {
					log.Info("subscriber notify error")
					return
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
	}()

	return rpcSub, nil
}
