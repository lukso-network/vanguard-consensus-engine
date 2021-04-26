package events

import (
	"github.com/ethereum/go-ethereum/event"
	types "github.com/prysmaticlabs/eth2-types"
	"time"
)

var (
	deadline = 5 * time.Minute
)

type MockBackend struct {
	ConsensusInfoFeed    event.Feed
	ConsensusInfoMapping map[types.Epoch]*MinimalEpochConsensusInfo
	CurEpoch             types.Epoch
}

func (backend *MockBackend) CurrentEpoch() types.Epoch {
	return backend.CurEpoch
}

func (backend *MockBackend) ConsensusInfoByEpochRange(fromEpoch, toEpoch types.Epoch,
) map[types.Epoch]*MinimalEpochConsensusInfo {

	consensusInfoMapping := make(map[types.Epoch]*MinimalEpochConsensusInfo)
	for epoch := fromEpoch; epoch <= toEpoch; epoch++ {
		item, exists := backend.ConsensusInfoMapping[epoch]
		if exists && item != nil {
			consensusInfoMapping[epoch] = item
		}
	}
	return consensusInfoMapping
}

func (b *MockBackend) SubscribeNewEpochEvent(ch chan<- *MinimalEpochConsensusInfo) event.Subscription {
	return b.ConsensusInfoFeed.Subscribe(ch)
}
