// Package consensus contains types for block operation-specific events fired
// during the runtime of a beacon node such as attestations, voluntary
// exits, and slashings.
package consensus

import types "github.com/prysmaticlabs/eth2-types"

// ReceivedMinimalConsensusData is the data sent with ReceivedConsensusData events.
type ReceivedMinimalConsensusData struct {
	epoch types.Epoch
}
