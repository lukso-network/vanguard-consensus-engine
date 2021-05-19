package beacon

import (
	ptypes "github.com/gogo/protobuf/types"
	types "github.com/prysmaticlabs/eth2-types"
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"github.com/prysmaticlabs/prysm/beacon-chain/core/helpers"
	"github.com/prysmaticlabs/prysm/shared/params"
	"github.com/prysmaticlabs/prysm/shared/slotutil"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// StreamMinimalConsensusInfo to orchestrator client every single time an unconfirmed block is received by the beacon node.
func (bs *Server) StreamMinimalConsensusInfo(
	empty *ptypes.Empty,
	fromEpoch *types.Epoch,
	stream ethpb.BeaconChain_StreamMinimalConsensusInfoServer,
) error {
	if bs.SyncChecker.Syncing() {
		return status.Error(codes.Unavailable, "Syncing to latest head, not ready to respond")
	}

	// If we are post-genesis time, then set the current epoch to
	// the number epochs since the genesis time, otherwise 0 by default.
	genesisTime := bs.GenesisTimeFetcher.GenesisTime()
	if genesisTime.IsZero() {
		return status.Error(codes.Unavailable, "genesis time is not set")
	}

	minimalConsensusInfoRange, err := bs.MinimalConsensusInfoFetcher.MinimalConsensusInfoRange(*fromEpoch)
	if err != nil {
		return status.Errorf(codes.Unavailable, "Could not send minimalConsensusInfo over stream: %v", err)
	}

	for _, minimalConsensusInfo := range minimalConsensusInfoRange {
		if err := stream.Send(minimalConsensusInfo); err != nil {
			return status.Errorf(codes.Unavailable, "Could not send minimalConsensusInfo over stream: %v", err)
		}
	}

	secondsPerEpoch := params.BeaconConfig().SecondsPerSlot * uint64(params.BeaconConfig().SlotsPerEpoch)
	epochTicker := slotutil.NewSlotTicker(bs.GenesisTimeFetcher.GenesisTime(), secondsPerEpoch)

	for {
		select {
		case slot := <-epochTicker.C():
			endSlot, err := helpers.EndSlot(types.Epoch(slot))
			if err != nil {
				return status.Errorf(codes.Unavailable, "Could not retrieve end slot of epoch: %v, err: %v", err, types.Epoch(slot))
			}
			if slot == endSlot {
				res, err := bs.MinimalConsensusInfoFetcher.MinimalConsensusInfo(types.Epoch(slot))
				if err != nil {
					return status.Errorf(codes.Unavailable, "Could not send minimalConsensusInfo over stream: %v", err)
				}
				if err := stream.Send(res); err != nil {
					return status.Errorf(codes.Internal, "Could not send minimalConsensusInfo response over stream: %v", err)
				}
			}
		case <-stream.Context().Done():
			return status.Error(codes.Canceled, "[minimalConsensusInfo] Stream context canceled")
		case <-bs.Ctx.Done():
			return status.Error(codes.Canceled, "[minimalConsensusInfo] RPC context canceled")
		}
	}
}
