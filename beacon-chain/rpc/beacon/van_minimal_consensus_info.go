package beacon

import (
	"encoding/hex"
	"fmt"
	types2 "github.com/gogo/protobuf/types"
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
	req *ethpb.MinimalConsensusInfoRequest,
	stream ethpb.BeaconChain_StreamMinimalConsensusInfoServer,
) error {

	curEpoch := helpers.SlotToEpoch(bs.GenesisTimeFetcher.CurrentSlot())
	requestedEpoch := req.FromEpoch
	if requestedEpoch > curEpoch {
		log.WithField("curEpoch", curEpoch).
			WithField("requestedEpoch", requestedEpoch).
			Warn("requested epoch is future from current epoch")
		return status.Errorf(codes.InvalidArgument, errEpoch, curEpoch, req.FromEpoch)
	}

	// sending past proposer assignments info to orchestrator
	for epoch := requestedEpoch; epoch <= curEpoch; epoch++ {
		epochInfo, err := bs.prepareEpochInfo(epoch)
		if err != nil {
			log.WithField("epoch", epoch).
				WithError(err).
				Warn("Failed to prepare epoch info")
			return err
		}
		if err := stream.Send(epochInfo); err != nil {
			return status.Errorf(codes.Unavailable, "Could not send minimalConsensusInfo over stream: %v", err)
		}
	}

	secondsPerEpoch := params.BeaconConfig().SecondsPerSlot * uint64(params.BeaconConfig().SlotsPerEpoch)
	epochTicker := slotutil.NewSlotTicker(bs.GenesisTimeFetcher.GenesisTime(), secondsPerEpoch)

	for {
		select {
		case slot := <-epochTicker.C():
			epoch := types.Epoch(slot)
			log.WithField("epoch", epoch).Debug("Sending current epoch info to orchestrator")

			res, err := bs.prepareEpochInfo(epoch + 1)
			if err != nil {
				log.WithField("epoch", epoch).
					WithField("slot", slot).
					WithError(err).Warn("Failed to prepare epoch info")
				return status.Errorf(codes.Unavailable, "Could not send epoch info over stream: %v", err)
			}

			if err := stream.Send(res); err != nil {
				return status.Errorf(codes.Internal, "Could not send minimalConsensusInfo response over stream: %v", err)
			}
		case <-stream.Context().Done():
			return status.Error(codes.Canceled, "[minimalConsensusInfo] Stream context canceled")
		case <-bs.Ctx.Done():
			return status.Error(codes.Canceled, "[minimalConsensusInfo] RPC context canceled")
		}
	}
}

// prepareEpochInfo
func (bs *Server) prepareEpochInfo(epoch types.Epoch) (*ethpb.MinimalConsensusInfo, error) {
	startSlot, err := helpers.StartSlot(epoch)
	if err != nil {
		return nil, err
	}

	requestedState, err := bs.StateGen.StateBySlot(bs.Ctx, startSlot)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not retrieve archived state for epoch %d: %v", epoch, err)
	}

	proposerAssignmentInfo, err := helpers.ProposerAssignments(requestedState, epoch)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not generate proposer assignments %d: %v", epoch, err)
	}

	epochStartTime, err := helpers.SlotToTime(uint64(bs.GenesisTimeFetcher.GenesisTime().Unix()), startSlot)
	if nil != err {
		return nil, err
	}
	validatorList, err := prepareSortedValidatorList(epoch, proposerAssignmentInfo)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not prepare validator list %d: %v", epoch, err)
	}

	log.WithField("epoch", epoch).
		WithField("proposerList", fmt.Sprintf("%v", validatorList)).
		Debug("proposer public key list")

	return &ethpb.MinimalConsensusInfo{
		Epoch:            epoch,
		ValidatorList:    validatorList,
		EpochTimeStart:   uint64(epochStartTime.Unix()),
		SlotTimeDuration: &types2.Duration{Seconds: int64(params.BeaconConfig().SecondsPerSlot)},
	}, nil
}

// prepareEpochInfo
func prepareSortedValidatorList(
	epoch types.Epoch,
	proposerAssignmentInfo []*ethpb.ValidatorAssignments_CommitteeAssignment,
) ([]string, error) {

	publicKeyList := make([]string, 0)
	slotToPubKeyMapping := make(map[types.Slot]string)

	for _, assignment := range proposerAssignmentInfo {
		for _, slot := range assignment.ProposerSlots {
			slotToPubKeyMapping[slot] = fmt.Sprintf("0x%s", hex.EncodeToString(assignment.PublicKey))
		}
	}

	if epoch == 0 {
		publicKeyBytes := make([]byte, params.BeaconConfig().BLSPubkeyLength)
		emptyPubKey := fmt.Sprintf("0x%s", hex.EncodeToString(publicKeyBytes))
		slotToPubKeyMapping[0] = emptyPubKey
	}

	startSlot, err := helpers.StartSlot(epoch)
	if err != nil {
		return []string{}, err
	}

	endSlot, err := helpers.EndSlot(epoch)
	if err != nil {
		return []string{}, err
	}

	for slot := startSlot; slot <= endSlot; slot++ {
		publicKeyList = append(publicKeyList, slotToPubKeyMapping[slot])
	}
	return publicKeyList, nil
}
