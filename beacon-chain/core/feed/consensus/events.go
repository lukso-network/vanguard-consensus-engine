// Package consensus contains types for block operation-specific events fired
// during the runtime of a beacon node such as attestations, voluntary
// exits, and slashings.
package consensus

import (
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
)

// MinimalConsensusData is the data sent with ReceivedConsensusData events.
type MinimalConsensusData struct {
	MinimalConsensusInfo *ethpb.MinimalConsensusInfo
}
