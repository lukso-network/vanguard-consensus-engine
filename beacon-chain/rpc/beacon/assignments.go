package beacon

import (
	"context"
	"encoding/hex"
	"fmt"
	ptypes "github.com/gogo/protobuf/types"
	"github.com/prysmaticlabs/prysm/beacon-chain/rpc/subscriber/api/events"
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

func (bs *Server) GetProposerListForEpoch(
	ctx context.Context,
	curEpoch types.Epoch,
) (*ethpb.ValidatorAssignments, error) {
	var res []*ethpb.ValidatorAssignments_CommitteeAssignment
	startSlot, err := helpers.StartSlot(curEpoch)
	if err != nil {
		return nil, err
	}

	// latestState is the state of last epoch.
	latestState, err := bs.StateGen.StateBySlot(ctx, startSlot)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal, "Could not retrieve archived state for epoch %d: %v", curEpoch, err)
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
		return nil, fmt.Errorf("invalid validators len, expected: %d, got: %d", maxValidators, len(res))
	}

	return &ethpb.ValidatorAssignments{
		Epoch:       curEpoch,
		Assignments: res,
	}, nil
}

// GetProposerList retrieves the validator assignments for a given epoch, [This api is specially used for Orchestrator client]
// optional validator indices or public keys may be included to filter validator assignments.
func (bs *Server) NextEpochProposerList(
	ctx context.Context,
	empty *ptypes.Empty,
) (assignments *ethpb.ValidatorAssignments, err error) {
	curEpoch := helpers.SlotToEpoch(bs.GenesisTimeFetcher.CurrentSlot())
	assignments, err = bs.GetProposerListForEpoch(ctx, curEpoch)

	return
}

func (bs *Server) GetMinimalConsensusInfoRange(
	ctx context.Context,
	fromEpoch types.Epoch,
) (consensusInfos []*events.MinimalEpochConsensusInfo, err error) {
	assignments, err := bs.NextEpochProposerList(ctx, &ptypes.Empty{})

	if nil != err {
		return
	}

	if fromEpoch > assignments.Epoch {
		err = fmt.Errorf("requested epoch too high")
		log.WithField("currentEpoch", assignments.Epoch).
			WithField("requestedEpoch", fromEpoch).Error(err.Error())

		return
	}

	epochRange := assignments.Epoch - fromEpoch
	consensusInfos = make([]*events.MinimalEpochConsensusInfo, 0)

	for index := 0; index < int(epochRange); index++ {
		currentEpoch := types.Epoch(index) + epochRange
		minimalConsensusInfo, err := bs.GetMinimalConsensusInfo(ctx, currentEpoch)

		if nil != err {
			currentErr := fmt.Errorf("requested epoch not present")
			log.WithField("currentEpoch", assignments.Epoch).
				WithField("internal", currentErr.Error()).
				WithField("requestedEpoch", fromEpoch).Error(err.Error())

			return nil, err
		}

		consensusInfos = append(consensusInfos, minimalConsensusInfo)
	}

	// get the latest
	currentInfo, err := bs.GetMinimalConsensusInfo(ctx, assignments.Epoch)

	if nil != err {
		// Do not send invalid slice back
		consensusInfos = nil
		currentErr := fmt.Errorf("error during fetching latest epoch")
		log.WithField("currentEpoch", assignments.Epoch).
			WithField("internal", currentErr.Error()).
			WithField("requestedEpoch", fromEpoch).Error(err.Error())

		return
	}

	consensusInfos = append(consensusInfos, currentInfo)

	return
}

func (bs *Server) GetMinimalConsensusInfo(
	ctx context.Context,
	curEpoch types.Epoch,
) (minConsensusInfo *events.MinimalEpochConsensusInfo, err error) {
	curSlot := bs.GenesisTimeFetcher.CurrentSlot()
	curEpoch = helpers.SlotToEpoch(curSlot)
	assignments, err := bs.GetProposerListForEpoch(ctx, curEpoch)
	if err != nil {
		return nil, err
	}

	assignmentsString := make([]string, 32)
	for _, assigment := range assignments.Assignments {
		assignmentsString = append(assignmentsString, hex.EncodeToString(assigment.PublicKey))
	}

	genesisTime := bs.GenesisTimeFetcher.GenesisTime()
	epochStartTime, err := helpers.SlotToTime(uint64(genesisTime.Unix()), curSlot)
	if nil != err {
		return nil, err
	}

	minConsensusInfo = &events.MinimalEpochConsensusInfo{
		Epoch:            uint64(curEpoch),
		ValidatorList:    assignmentsString,
		EpochStartTime:   uint64(epochStartTime.Unix()),
		SlotTimeDuration: time.Duration(params.BeaconConfig().SecondsPerSlot),
	}

	return minConsensusInfo, nil
}
