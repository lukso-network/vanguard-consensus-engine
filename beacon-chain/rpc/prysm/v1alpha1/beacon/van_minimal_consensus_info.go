package beacon

import (
	"encoding/hex"
	"fmt"
	duration "github.com/golang/protobuf/ptypes/duration"
	types "github.com/prysmaticlabs/eth2-types"
	"github.com/prysmaticlabs/prysm/beacon-chain/core/feed"
	statefeed "github.com/prysmaticlabs/prysm/beacon-chain/core/feed/state"
	"github.com/prysmaticlabs/prysm/beacon-chain/core/helpers"
	ethpbv1 "github.com/prysmaticlabs/prysm/proto/eth/v1"
	ethpb "github.com/prysmaticlabs/prysm/proto/eth/v1alpha1"
	"github.com/prysmaticlabs/prysm/shared/params"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// StreamMinimalConsensusInfo to orchestrator client every single time an unconfirmed block is received by the beacon node.
func (bs *Server) StreamMinimalConsensusInfo(
	req *ethpb.MinimalConsensusInfoRequest,
	stream ethpb.BeaconChain_StreamMinimalConsensusInfoServer,
) error {

	sender := func(epoch types.Epoch, epochInfo *ethpb.MinimalConsensusInfo) error {
		if err := stream.Send(epochInfo); err != nil {
			return status.Errorf(codes.Unavailable,
				"Could not send over stream: %v  err: %v", epoch, err)
		}
		log.WithField("epoch", epoch).Info("Published epoch info")
		log.WithField("epoch", epoch).WithField("epochInfo", fmt.Sprintf("%+v", epochInfo)).Debug("Sent epoch info")
		return nil
	}

	batchSender := func(startEpoch, endEpoch types.Epoch) error {
		for epoch := startEpoch; epoch <= endEpoch; epoch++ {
			startSlot, err := helpers.StartSlot(epoch)
			if err != nil {
				return status.Errorf(codes.Internal, "Could not send over stream: %v", err)
			}
			// retrieve state from cache or db
			s, err := bs.StateGen.StateBySlot(bs.Ctx, startSlot)
			if err != nil {
				return status.Errorf(codes.Internal, "Could not send over stream: %v", err)
			}
			// retrieve proposer
			proposerIndices, pubKeys, err := helpers.ProposerIndicesInCache(s)
			if err != nil {
				return status.Errorf(codes.Internal, "Could not send over stream: %v", err)
			}
			epochInfo, err := bs.prepareEpochInfo(epoch, proposerIndices, pubKeys)
			if err != nil {
				return status.Errorf(codes.Internal, "Could not send over stream: %v", err)
			}
			if !s.IsNil() {
				if err := sender(epoch, epochInfo); err != nil {
					return err
				}
			}
		}
		return nil
	}

	cp, err := bs.BeaconDB.FinalizedCheckpoint(bs.Ctx)
	if err != nil {
		return status.Errorf(codes.Internal,
			"Could not send over stream: %v", err)
	}
	startEpoch := req.FromEpoch
	endEpoch := cp.Epoch
	log.WithField("startEpoch", startEpoch).WithField("endEpoch", endEpoch).Info("Publishing previous epoch infos")
	if startEpoch == 0 || startEpoch < endEpoch {
		if err := batchSender(startEpoch, endEpoch); err != nil {
			return err
		}
		log.WithField("startEpoch", startEpoch).WithField("endEpoch", endEpoch).Debug("Successfully published previous epoch infos")
	}

	stateChannel := make(chan *feed.Event, 1)
	stateSub := bs.StateNotifier.StateFeed().Subscribe(stateChannel)
	firstTime := true
	defer stateSub.Unsubscribe()

	for {
		select {
		case stateEvent := <-stateChannel:
			if stateEvent.Type == statefeed.EpochInfo {
				epochInfoData, ok := stateEvent.Data.(*statefeed.EpochInfoData)
				if !ok {
					log.Warn("Failed to send epoch info to orchestrator")
					continue
				}
				curEpoch := helpers.SlotToEpoch(epochInfoData.Slot)
				// Executes for a single time
				if firstTime {
					startEpoch = endEpoch + 1
					endEpoch = curEpoch - 1
					if startEpoch <= endEpoch {
						log.WithField("startEpoch", startEpoch).WithField("endEpoch", endEpoch).WithField("liveSyncStart", curEpoch).
							Info("Publishing left over epoch infos")
						if err := batchSender(startEpoch, endEpoch); err != nil {
							return err
						}
						log.WithField("startEpoch", startEpoch).WithField("endEpoch", endEpoch).WithField("liveSyncStart", curEpoch).
							Debug("Successfully published left over epoch infos")
					}
					firstTime = false
					log.WithField("curEpoch", curEpoch).Debug("Start sending live epoch info")
				}

				epochInfo, err := bs.prepareEpochInfo(curEpoch, epochInfoData.ProposerIndices, epochInfoData.PublicKeys)
				if err != nil {
					return status.Errorf(codes.Internal, "Could not send over stream: %v", err)
				}
				if err := sender(curEpoch, epochInfo); err != nil {
					return err
				}
			}
			// If a reorg occurred, we recompute duties for the connected validator clients
			// and send another response over the server stream right away.
			if stateEvent.Type == statefeed.Reorg {
				data, ok := stateEvent.Data.(*ethpbv1.EventChainReorg)
				if !ok {
					return status.Errorf(codes.Internal, "Received incorrect data type over reorg feed: %v", data)
				}
				log.WithField("newSlot", data.Slot).WithField("newEpoch", data.Epoch).Debug("Encountered a reorg. Re-sending updated epoch info")
				if err := batchSender(data.Epoch, data.Epoch); err != nil {
					return err
				}
				log.WithField("startEpoch", startEpoch).WithField("endEpoch", endEpoch).Info("Published reorg epoch infos")
			}

		case <-stateSub.Err():
			return status.Error(codes.Aborted, "Subscriber closed, exiting go routine")
		case <-stream.Context().Done():
			return status.Error(codes.Canceled, "Stream context canceled")
		case <-bs.Ctx.Done():
			return status.Error(codes.Canceled, "RPC context canceled")
		}
	}
}

// prepareEpochInfo
func (bs *Server) prepareEpochInfo(
	epoch types.Epoch,
	proposerIndices []types.ValidatorIndex,
	pubKeys map[types.ValidatorIndex][48]byte,
) (*ethpb.MinimalConsensusInfo, error) {
	startSlot, err := helpers.StartSlot(epoch)
	if err != nil {
		return nil, err
	}

	epochStartTime, err := helpers.SlotToTime(uint64(bs.GenesisTimeFetcher.GenesisTime().Unix()), startSlot)
	if nil != err {
		return nil, err
	}

	publicKeyList := make([]string, 0)
	for i := 0; i < len(proposerIndices); i++ {
		pi := proposerIndices[i]
		pubKey := pubKeys[pi]
		var pubKeyStr string
		if startSlot == 0 && i == 0 {
			publicKeyBytes := make([]byte, 48)
			pubKeyStr = fmt.Sprintf("0x%s", hex.EncodeToString(publicKeyBytes))
		} else {
			pubKeyStr = fmt.Sprintf("0x%s", hex.EncodeToString(pubKey[:]))
		}
		publicKeyList = append(publicKeyList, pubKeyStr)
	}

	return &ethpb.MinimalConsensusInfo{
		Epoch:            epoch,
		ValidatorList:    publicKeyList,
		EpochTimeStart:   uint64(epochStartTime.Unix()),
		SlotTimeDuration: &duration.Duration{Seconds: int64(params.BeaconConfig().SecondsPerSlot)},
	}, nil
}