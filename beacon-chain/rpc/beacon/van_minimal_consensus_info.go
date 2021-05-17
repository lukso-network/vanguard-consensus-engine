package beacon

import (
	ptypes "github.com/gogo/protobuf/types"
	types "github.com/prysmaticlabs/eth2-types"
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"github.com/prysmaticlabs/prysm/beacon-chain/core/feed"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// StreamMinimalConsensusInfo to orchestrator client every single time an unconfirmed block is received by the beacon node.
func (bs *Server) StreamMinimalConsensusInfo(
	empty *ptypes.Empty,
	stream ethpb.BeaconChain_StreamMinimalConsensusInfoServer,
) error {
	minimalConsensusChannel := make(chan *feed.Event, 1)
	minimalConsensusSub := bs.ConsensusNotifier.ConsensusFeed().Subscribe(minimalConsensusChannel)
	defer minimalConsensusSub.Unsubscribe()

	// Sends all past minimalConsensusInfo to the orchestrator client
	minimalConsensusInfoRange, err := bs.MinimalConsensusInfoFetcher.MinimalConsensusInfoRange(types.Epoch(0))
	if err != nil {
		return status.Errorf(codes.Unavailable, "Could not send minimalConsensusInfo over stream: %v", err)
	}
	for _, blk := range minimalConsensusInfoRange {
		if err := stream.Send(blk); err != nil {
			return status.Errorf(codes.Unavailable, "Could not send minimalConsensusInfo over stream: %v", err)
		}
	}

	for {
		select {
		case minConsensusEvent := <-minimalConsensusChannel:
			epoch, ok := minConsensusEvent.Data.(*types.Epoch)
			if !ok || epoch == nil {
				continue
			}
			minimalConsensusInfoRange, err = bs.MinimalConsensusInfoFetcher.MinimalConsensusInfoRange(*epoch)
			if err != nil {
				return status.Errorf(codes.Unavailable, "Could not send minimalConsensusInfo over stream: %v", err)
			}
			if err := stream.Send(minimalConsensusInfoRange); err != nil {
				return status.Errorf(codes.Unavailable, "Could not send minimalConsensusInfo over stream: %v", err)
			}
			log.WithField("epoch", epoch).Debug(
				"New minimalConsensusInfo has been published successfully")
		case <-minimalConsensusSub.Err():
			return status.Error(codes.Aborted, "Subscriber closed, exiting goroutine")
		case <-bs.Ctx.Done():
			return status.Error(codes.Canceled, "Context canceled")
		case <-stream.Context().Done():
			return status.Error(codes.Canceled, "Context canceled")
		}
	}
}
