package blockchain

import (
	"context"
	"github.com/pkg/errors"
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"github.com/prysmaticlabs/prysm/beacon-chain/core/feed"
	blockfeed "github.com/prysmaticlabs/prysm/beacon-chain/core/feed/block"
	vanTypes "github.com/prysmaticlabs/prysm/shared/params"
	"github.com/prysmaticlabs/prysm/shared/slotutil"
	"go.opencensus.io/trace"
	"sort"
	"time"
)

type OrcClient interface {
	ConfirmVanBlockHashes(ctx context.Context, request []*vanTypes.ConfirmationReqData) (response []*vanTypes.ConfirmationResData, err error)
}

var processPendingBlocksPeriod = slotutil.DivideSlotBy(3 /* times per slot */)

// PendingBlocksFetcher retrieves the cached un-confirmed beacon blocks from cache
type PendingBlocksFetcher interface {
	UnConfirmedBlocksFromCache() ([]*ethpb.BeaconBlock, error)
}

// publishAndStorePendingBlock method publishes and stores the pending block for final confirmation check
func (s *Service) publishAndStorePendingBlock(ctx context.Context, pendingBlk *ethpb.BeaconBlock) error {
	ctx, span := trace.StartSpan(ctx, "blockChain.publishAndStorePendingBlock")
	defer span.End()

	// Sending pending block feed to streaming api
	log.WithField("slot", pendingBlk.Slot).Debug("Unconfirmed block sends for publishing")
	s.blockNotifier.BlockFeed().Send(&feed.Event{
		Type: blockfeed.UnConfirmedBlock,
		Data: &blockfeed.UnConfirmedBlockData{Block: pendingBlk},
	})

	// Storing pending block into pendingBlockCache
	if err := s.pendingBlockCache.AddPendingBlock(pendingBlk); err != nil {
		return errors.Wrapf(err, "could not cache block of slot %d", pendingBlk.Slot)
	}

	return nil
}

// publishAndStorePendingBlockBatch method publishes and stores the batch of pending block for final confirmation check
func (s *Service) publishAndStorePendingBlockBatch(ctx context.Context, pendingBlkBatch []*ethpb.SignedBeaconBlock) error {
	ctx, span := trace.StartSpan(ctx, "blockChain.publishAndStorePendingBlockBatch")
	defer span.End()

	for _, b := range pendingBlkBatch {

		// Sending pending block feed to streaming api
		log.WithField("slot", b.Block.Slot).Debug("Unconfirmed block batch sends for publishing")
		s.blockNotifier.BlockFeed().Send(&feed.Event{
			Type: blockfeed.UnConfirmedBlock,
			Data: &blockfeed.UnConfirmedBlockData{Block: b.Block},
		})

		// Storing pending block into pendingBlockCache
		if err := s.pendingBlockCache.AddPendingBlock(b.Block); err != nil {
			return errors.Wrapf(err, "could not cache block of slot %d", b.Block.Slot)
		}
	}

	return nil
}

// UnConfirmedBlocksFromCache retrieves all the cached blocks from cache and send it back to event api
func (s *Service) UnConfirmedBlocksFromCache() ([]*ethpb.BeaconBlock, error) {
	blks, err := s.pendingBlockCache.PendingBlocks()
	if err != nil {
		return nil, errors.Wrap(err, "Could not retrieve cached unconfirmed blocks from cache")
	}
	return blks, nil
}

// processOrcConfirmation
func (s *Service) processOrcConfirmation(ctx context.Context) {
	ticker := time.NewTicker(processPendingBlocksPeriod)
	go func() {
		for {
			select {
			case <-ticker.C:
				log.WithField("function", "processOrcConfirmation").Trace("running")

			case <-ctx.Done():
				log.WithField("function", "processOrcConfirmation").Debug("context is closed, exiting")
				ticker.Stop()
				return
			}
		}
	}()
}

// processPendingBlocks process confirmation for pending blocks
// -> After getting confirmation for a list of pending slots, it iterates through the list
// -> If any slot gets invalid status then stop the iteration and start again from that slot
// -> If any slot gets verified status then, publish the slots and block hashes to the blockchain service
//    who actually waiting for confirmed blocks
// -> If any slot gets un
func (s *Service) processPendingBlocks(ctx context.Context) error {
	reqData, err := s.sortedPendingSlots()
	if err != nil {
		log.WithError(err).Error("got error when preparing sorted confirmation request data")
		return err
	}

	resData, err := s.ConfirmVanBlockHashes(ctx, reqData)
	for i := 0; i < len(resData); i++ {
		log.WithField("slot", resData[i].Slot).WithField("status", resData[i].Status).Debug("got confirmation status from orchestrator")

		//if resData[i].Status == vanTypes.Invalid {
		//	log.WithField("Slot", resData[i]).WithField("status", resData[i].Status).Warn("Invalid confirmation!")
		//	return nil
		//}

		// checking the block's confirmation status. If the block is being confirmed then add it to canonical blockchain
		//if resData[i].Status == vanTypes.Verified {
		s.blockNotifier.BlockFeed().Send(&feed.Event{
			Type: blockfeed.ConfirmedBlock,
			Data: &blockfeed.ConfirmedData{
				Slot: resData[i].Slot,
				BlockRootHash: resData[i].Hash,
				Status: resData[i].Status,
			},
		})
		// deleting block from cache because it is already verified
		//err := s.pendingBlockCache.DeleteConfirmedBlock(resData[i].Slot)
		//if err != nil {
		//	return err
		//}
		//}
	}

	return nil
}

// sortedPendingSlots retrieves pending blocks from pending block cache and prepare sorted request data
func (s *Service) sortedPendingSlots() ([]*vanTypes.ConfirmationReqData, error) {
	items, err := s.pendingBlockCache.PendingBlocks()
	if err != nil {
		return nil, err
	}

	reqData := make([]*vanTypes.ConfirmationReqData, 0, len(items))
	for _, blk := range items {
		blockRoot, err := blk.HashTreeRoot()
		if err != nil {
			return nil, err
		}
		reqData = append(reqData, &vanTypes.ConfirmationReqData{
			Slot: blk.Slot,
			Hash: blockRoot,
		})
	}

	sort.Slice(reqData, func(i, j int) bool {
		return reqData[i].Slot < reqData[j].Slot
	})

	return reqData, nil
}

// ConfirmVanBlockHashes is a dummy function
// TODO- Will remove this function when orchestrator client will be ready
func (s *Service) ConfirmVanBlockHashes(
	ctx context.Context, request []*vanTypes.ConfirmationReqData) (response []*vanTypes.ConfirmationResData, err error) {

	resData := make([]*vanTypes.ConfirmationResData, 0, len(request))
	for i := 0; i < len(request); i++ {
		resData = append(resData, &vanTypes.ConfirmationResData{
				request[i].Slot,
				request[i].Hash,
				vanTypes.Verified,
			})
	}
	return resData, nil
}