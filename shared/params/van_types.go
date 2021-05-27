package params

import (
	types "github.com/prysmaticlabs/eth2-types"
)

type Status int

const (
	Pending Status = iota
	Verified
	Invalid
)

// ConfirmationReqData is used as a request param for getting confirmation from orchestrator
type ConfirmationReqData struct {
	Slot types.Slot
	Hash [32]byte
}

// ConfirmationResData is used as a response param for getting confirmation from orchestrator
type ConfirmationResData struct {
	Slot types.Slot
	Hash [32]byte
	Status Status
}
