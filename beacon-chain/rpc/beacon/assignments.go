package beacon

import (
	"context"
	"encoding/hex"
	"fmt"
	ptypes "github.com/gogo/protobuf/types"
	"github.com/prysmaticlabs/prysm/beacon-chain/rpc/subscriber/api/events"
	"github.com/prysmaticlabs/prysm/beacon-chain/state"
	"github.com/prysmaticlabs/prysm/shared/params"
	"strconv"
	"time"

	types "github.com/prysmaticlabs/eth2-types"
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"github.com/prysmaticlabs/prysm/beacon-chain/core/helpers"
	"github.com/prysmaticlabs/prysm/shared/bytesutil"
	"github.com/prysmaticlabs/prysm/shared/cmd"
	"github.com/prysmaticlabs/prysm/shared/pagination"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const errEpoch = "Cannot retrieve information about an epoch in the future, current epoch %d, requesting %d"

// ListValidatorAssignments retrieves the validator assignments for a given epoch,
// optional validator indices or public keys may be included to filter validator assignments.
func (bs *Server) ListValidatorAssignments(
	ctx context.Context, req *ethpb.ListValidatorAssignmentsRequest,
) (*ethpb.ValidatorAssignments, error) {
	if int(req.PageSize) > cmd.Get().MaxRPCPageSize {
		return nil, status.Errorf(
			codes.InvalidArgument,
			"Requested page size %d can not be greater than max size %d",
			req.PageSize,
			cmd.Get().MaxRPCPageSize,
		)
	}

	var res []*ethpb.ValidatorAssignments_CommitteeAssignment
	filtered := map[types.ValidatorIndex]bool{} // track filtered validators to prevent duplication in the response.
	filteredIndices := make([]types.ValidatorIndex, 0)
	var requestedEpoch types.Epoch
	switch q := req.QueryFilter.(type) {
	case *ethpb.ListValidatorAssignmentsRequest_Genesis:
		if q.Genesis {
			requestedEpoch = 0
		}
	case *ethpb.ListValidatorAssignmentsRequest_Epoch:
		requestedEpoch = q.Epoch
	}

	currentEpoch := helpers.SlotToEpoch(bs.GenesisTimeFetcher.CurrentSlot())
	if requestedEpoch > currentEpoch {
		return nil, status.Errorf(
			codes.InvalidArgument,
			errEpoch,
			currentEpoch,
			requestedEpoch,
		)
	}

	startSlot, err := helpers.StartSlot(requestedEpoch)
	if err != nil {
		return nil, err
	}
	requestedState, err := bs.StateGen.StateBySlot(ctx, startSlot)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not retrieve archived state for epoch %d: %v", requestedEpoch, err)
	}

	// Filter out assignments by public keys.
	for _, pubKey := range req.PublicKeys {
		index, ok := requestedState.ValidatorIndexByPubkey(bytesutil.ToBytes48(pubKey))
		if !ok {
			return nil, status.Errorf(codes.NotFound, "Could not find validator index for public key %#x", pubKey)
		}
		filtered[index] = true
		filteredIndices = append(filteredIndices, index)
	}

	// Filter out assignments by validator indices.
	for _, index := range req.Indices {
		if !filtered[index] {
			filteredIndices = append(filteredIndices, index)
		}
	}

	activeIndices, err := helpers.ActiveValidatorIndices(requestedState, requestedEpoch)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not retrieve active validator indices: %v", err)
	}
	if len(filteredIndices) == 0 {
		if len(activeIndices) == 0 {
			return &ethpb.ValidatorAssignments{
				Assignments:   make([]*ethpb.ValidatorAssignments_CommitteeAssignment, 0),
				TotalSize:     int32(0),
				NextPageToken: strconv.Itoa(0),
			}, nil
		}
		// If no filter was specified, return assignments from active validator indices with pagination.
		filteredIndices = activeIndices
	}

	start, end, nextPageToken, err := pagination.StartAndEndPage(req.PageToken, int(req.PageSize), len(filteredIndices))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not paginate results: %v", err)
	}

	// Initialize all committee related data.
	committeeAssignments, proposerIndexToSlots, err := helpers.CommitteeAssignments(requestedState, requestedEpoch)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not compute committee assignments: %v", err)
	}

	for _, index := range filteredIndices[start:end] {
		if uint64(index) >= uint64(requestedState.NumValidators()) {
			return nil, status.Errorf(codes.OutOfRange, "Validator index %d >= validator count %d",
				index, requestedState.NumValidators())
		}
		comAssignment := committeeAssignments[index]
		pubkey := requestedState.PubkeyAtIndex(index)
		assign := &ethpb.ValidatorAssignments_CommitteeAssignment{
			BeaconCommittees: comAssignment.Committee,
			CommitteeIndex:   comAssignment.CommitteeIndex,
			AttesterSlot:     comAssignment.AttesterSlot,
			ProposerSlots:    proposerIndexToSlots[index],
			PublicKey:        pubkey[:],
			ValidatorIndex:   index,
		}
		res = append(res, assign)
	}

	return &ethpb.ValidatorAssignments{
		Epoch:         requestedEpoch,
		Assignments:   res,
		NextPageToken: nextPageToken,
		TotalSize:     int32(len(filteredIndices)),
	}, nil
}

// Deprecated: use GetMinimalConsensusInfoRange or GetMinimalConsensusInfo
// NextEpochProposerList retrieves the validator assignments for a given epoch, [This api is specially used for Orchestrator client]
// optional validator indices or public keys may be included to filter validator assignments.
func (bs *Server) NextEpochProposerList(
	ctx context.Context,
	empty *ptypes.Empty,
) (assignments *ethpb.ValidatorAssignments, err error) {
	curEpoch := helpers.SlotToEpoch(bs.GenesisTimeFetcher.CurrentSlot())
	assignments, err = bs.getProposerListForEpoch(ctx, curEpoch)

	return
}

// GetMinimalConsensusInfoRange allows to subscribe into feed about minimalConsensusInformation from particular epoch
// TODO: Serve it in chunks, if recent epoch will be a very high number flood of responses could kill the connection
func (bs *Server) GetMinimalConsensusInfoRange(
	ctx context.Context,
	fromEpoch types.Epoch,
) (consensusInfos []*events.MinimalEpochConsensusInfo, err error) {
	consensusInfo, err := bs.GetMinimalConsensusInfo(ctx, fromEpoch)

	if nil != err {
		log.WithField("currentEpoch", "unknown").
			WithField("requestedEpoch", fromEpoch).Error(err.Error())

		return nil, err
	}

	consensusInfos = make([]*events.MinimalEpochConsensusInfo, 0)
	consensusInfos = append(consensusInfos, consensusInfo)
	tempEpochIndex := consensusInfo.Epoch

	for {
		tempEpochIndex++
		minimalConsensusInfo, currentErr := bs.GetMinimalConsensusInfo(ctx, types.Epoch(tempEpochIndex))

		if nil != currentErr {
			log.WithField("currentEpoch", tempEpochIndex).
				WithField("context", "epochNotFound").
				WithField("requestedEpoch", fromEpoch).Error(currentErr.Error())

			break
		}

		consensusInfos = append(consensusInfos, minimalConsensusInfo)
	}

	log.WithField("currentEpoch", tempEpochIndex).
		WithField("gathered", len(consensusInfos)).
		WithField("requestedEpoch", fromEpoch).Info("I should send epoch list")

	return
}

// GetMinimalConsensusInfo will give simple information about particular epoch
// If epoch is not present it will return an error
func (bs *Server) GetMinimalConsensusInfo(
	ctx context.Context,
	curEpoch types.Epoch,
) (minConsensusInfo *events.MinimalEpochConsensusInfo, err error) {
	log.WithField("prefix", "GetMinimalConsensusInfo")

	assignments, err := bs.getProposerListForEpoch(ctx, curEpoch)
	if nil != err {
		log.Errorf("[VAN_SUB] getProposerListForEpoch err = %s", err.Error())
		return nil, err
	}

	assignmentsSlice := make([]string, 32)
	for _, assigment := range assignments.Assignments {
		assignmentsSlice = append(assignmentsSlice, hex.EncodeToString(assigment.PublicKey))
	}

	genesisTime := bs.GenesisTimeFetcher.GenesisTime()
	startSlot, err := helpers.StartSlot(curEpoch)
	if nil != err {
		log.Errorf("[VAN_SUB] StartSlot err = %s", err.Error())
		return nil, err
	}
	epochStartTime, err := helpers.SlotToTime(uint64(genesisTime.Unix()), startSlot)
	if nil != err {
		log.Errorf("[VAN_SUB] SlotToTime err = %s", err.Error())
		return nil, err
	}

	minConsensusInfo = &events.MinimalEpochConsensusInfo{
		Epoch:            uint64(curEpoch),
		ValidatorList:    assignmentsSlice,
		EpochStartTime:   uint64(epochStartTime.Unix()),
		SlotTimeDuration: time.Duration(params.BeaconConfig().SecondsPerSlot),
	}

	log.Infof("[VAN_SUB] currEpoch = %#v", uint64(curEpoch))

	return minConsensusInfo, nil
}

func (bs *Server) getProposerListForEpoch(
	ctx context.Context,
	curEpoch types.Epoch,
) (*ethpb.ValidatorAssignments, error) {
	var (
		res         []*ethpb.ValidatorAssignments_CommitteeAssignment
		latestState *state.BeaconState
	)
	startSlot, err := helpers.StartSlot(curEpoch)

	if err != nil {
		return nil, status.Errorf(
			codes.Internal, "Could not retrieve startSlot for epoch %d: %v", curEpoch, err)
	}

	endSlot, err := helpers.EndSlot(curEpoch)

	if nil != err {
		return nil, status.Errorf(
			codes.Internal, "Could not retrieve endSlot for epoch %d: %v", curEpoch, err)
	}

	states, err := bs.BeaconDB.HighestSlotStatesBelow(ctx, endSlot)

	if err != nil {
		return nil, status.Errorf(
			codes.Internal, "Could not retrieve archived state for epoch %d: %v", curEpoch, err)
	}

	// Any state should return same proposer assignments so I pick first in slice
	for _, currentState := range states {
		if currentState.Slot() >= startSlot && currentState.Slot() <= endSlot {
			latestState = currentState

			break
		}
	}

	if nil == latestState {
		return nil, status.Errorf(
			codes.Internal, "Could not retrieve any state for epoch %d: %v", curEpoch, err)
	}

	// Initialize all committee related data.
	proposerIndexToSlots, err := helpers.ProposerAssignments(latestState, curEpoch)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not compute committee assignments: %v", err)
	}

	for index, proposerSlots := range proposerIndexToSlots {
		pubkey := latestState.PubkeyAtIndex(index)
		assign := &ethpb.ValidatorAssignments_CommitteeAssignment{
			ProposerSlots:  proposerSlots,
			PublicKey:      pubkey[:],
			ValidatorIndex: index,
		}
		res = append(res, assign)
	}

	maxValidators := params.BeaconConfig().SlotsPerEpoch

	// We omit the genesis slot
	if curEpoch == 0 {
		maxValidators = maxValidators - 1
	}

	if len(res) != int(maxValidators) {
		return nil, fmt.Errorf("invalid validators len, expected: %d, got: %d, epoch: %#v", maxValidators, len(res), curEpoch)
	}

	return &ethpb.ValidatorAssignments{
		Epoch:       curEpoch,
		Assignments: res,
	}, nil
}
