package beacon

import (
	ptypes "github.com/gogo/protobuf/types"
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
)

// StreamMinimalConsensusInfo to orchestrator client every single time an unconfirmed block is received by the beacon node.
func (bs *Server) StreamMinimalConsensusInfo(
	empty *ptypes.Empty,
	stream ethpb.BeaconChain_StreamMinimalConsensusInfoServer,
) error {
	return nil
}
