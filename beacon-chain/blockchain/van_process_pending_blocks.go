package blockchain

import (
	"context"
	"github.com/pkg/errors"
	types "github.com/prysmaticlabs/eth2-types"
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"github.com/prysmaticlabs/prysm/beacon-chain/core/feed"
	blockfeed "github.com/prysmaticlabs/prysm/beacon-chain/core/feed/block"
	"github.com/prysmaticlabs/prysm/shared/event"
	vanTypes "github.com/prysmaticlabs/prysm/shared/params"
	"github.com/prysmaticlabs/prysm/shared/slotutil"
	"go.opencensus.io/trace"
	"sort"
	"time"
)

type OrcClient interface {
	ConfirmVanBlockHashes(ctx context.Context, request []*vanTypes.ConfirmationReqData) (response []*vanTypes.ConfirmationResData, err error)
}

var (
	processPendingBlocksPeriod = slotutil.DivideSlotBy(3 /* times per slot */)
	errInvalidBlock = errors.New("invalid block found, discarded block batch")
	errPendingBlockCtxIsDone = errors.New("pending block confirmation context is done, reinitialize")
)

type blockRoot [32]byte

// PendingBlocksFetcher retrieves the cached un-confirmed beacon blocks from cache
type PendingBlocksFetcher interface {
	SortedUnConfirmedBlocksFromCache() ([]*ethpb.BeaconBlock, error)
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
func (s *Service) SortedUnConfirmedBlocksFromCache() ([]*ethpb.BeaconBlock, error) {
	blks, err := s.pendingBlockCache.PendingBlocks()
	if err != nil {
		return nil, errors.Wrap(err, "Could not retrieve cached unconfirmed blocks from cache")
	}

	sort.Slice(blks, func(i, j int) bool {
		return blks[i].Slot < blks[j].Slot
	})

	return blks, nil
}

// processOrcConfirmation
func (s *Service) processOrcConfirmationLoop(ctx context.Context) {
	ticker := time.NewTicker(processPendingBlocksPeriod)
	go func() {
		for {
			select {
			case <-ticker.C:
				log.WithField("function", "processOrcConfirmation").Trace("running")
				err := s.fetchOrcConfirmations(ctx)
				log.WithError(err).Error("got error when calling fetchOrcConfirmations method. exiting!")
				return
			case <-ctx.Done():
				log.WithField("function", "processOrcConfirmation").Debug("context is closed, exiting")
				ticker.Stop()
				return
			}
		}
	}()
}

// fetchOrcConfirmations process confirmation for pending blocks
// -> After getting confirmation for a list of pending slots, it iterates through the list
// -> If any slot gets invalid status then stop the iteration and start again from that slot
// -> If any slot gets verified status then, publish the slots and block hashes to the blockchain service
//    who actually waiting for confirmed blocks
// -> If any slot gets un
func (s *Service) fetchOrcConfirmations(ctx context.Context) error {
	reqData, err := s.sortedPendingSlots()
	if err != nil {
		log.WithError(err).Error("got error when preparing sorted confirmation request data")
		return err
	}

	resData, err := s.ConfirmVanBlockHashes(ctx, reqData)
	if err != nil {
		log.WithError(err).Error("got error when fetching confirmations from orchestrator client")
		return err
	}

	for i := 0; i < len(resData); i++ {
		log.WithField("slot", resData[i].Slot).WithField(
			"status", resData[i].Status).Debug("got confirmation status from orchestrator")

		s.blockNotifier.BlockFeed().Send(&feed.Event{
			Type: blockfeed.ConfirmedBlock,
			Data: &blockfeed.ConfirmedData{
				Slot: resData[i].Slot,
				BlockRootHash: resData[i].Hash,
				Status: resData[i].Status,
			},
		})
	}

	return nil
}

// waitForConfirmationBlock
func(s *Service) waitForConfirmationBlock(ctx context.Context, b *ethpb.SignedBeaconBlock) error {
	confirmedblocksCh := make(chan *feed.Event, 1)
	var confirmedblockSub event.Subscription

	confirmedblockSub = s.blockNotifier.BlockFeed().Subscribe(confirmedblocksCh)
	defer confirmedblockSub.Unsubscribe()

	for {
		select {
		case statusData := <-confirmedblocksCh:
			if statusData.Type == blockfeed.ConfirmedBlock {
				data, ok := statusData.Data.(*blockfeed.ConfirmedData)
				if !ok || data == nil {
					continue
				}
				// Checks slot number with incoming confirmation data slot
				if data.Slot == b.Block.Slot {
					switch status := data.Status; status {
					case vanTypes.Verified:
						log.WithField("slot", data.Slot).WithField(
							"blockHash", data.BlockRootHash).Debug("verified by orchestrator")
						if err := s.pendingBlockCache.DeleteConfirmedBlock(b.Block.Slot); err != nil {
							log.WithError(err).Error("couldn't delete the verified blocks from cache")
							return err
						}
						return nil
					case vanTypes.Invalid:
						log.WithField("slot", data.Slot).WithField(
							"blockHash", data.BlockRootHash).Debug("invalid by orchestrator, exiting goroutine")
						return errInvalidBlock
					default:
						continue		// for pending status, we can not make instant decision so we are waiting for the confirmation
					}
				}
			}
		case err := <-confirmedblockSub.Err():
			log.WithError(err).Error("Confirmation fetcher closed, exiting goroutine")
			return err
		case <-s.ctx.Done():
			log.WithField("function", "waitForConfirmationBlock").Debug("context is closed, exiting")
			return errPendingBlockCtxIsDone
		}
	}
}

// waitForConfirmationsBatch
// responsibilities:
// -> subscribe on ConfirmedBlock
// -> when getting confirmation data from fetchOrcConfirmations
// -> There are 3 status: Pending, Unknown/Invalid, Verified
// -> When Verified: it increments the pointer
// -> When Pending: it sets the pointer on the previous slot and keep waiting for new requested confirmation data
// -> When Unknown/Invalid: it return error
// -> At the end, if it runs successfully, then it deletes all the batch from cache and return nil
func (s *Service) waitForConfirmationsBlockBatch(ctx context.Context, blocks []*ethpb.SignedBeaconBlock) error {
	confirmedblocksCh := make(chan *feed.Event, 1)
	var confirmedblockSub event.Subscription

	confirmedblockSub = s.blockNotifier.BlockFeed().Subscribe(confirmedblocksCh)
	defer confirmedblockSub.Unsubscribe()

	lastVerifiedSlot := types.Slot(0)
	lastSlot := blocks[len(blocks) - 1].Block.Slot

	for {
		select {
		case statusData := <-confirmedblocksCh:
			if statusData.Type == blockfeed.ConfirmedBlock {
				data, ok := statusData.Data.(*blockfeed.ConfirmedData)
				if !ok || data == nil {
					continue
				}

				switch status := data.Status; status {
				case vanTypes.Verified:
					log.WithField("slot", data.Slot).WithField(
						"blockHash", data.BlockRootHash).Debug("verified by orchestrator")
					lastVerifiedSlot = data.Slot
					if lastVerifiedSlot == lastSlot {
						for _, b := range blocks {
							if err := s.pendingBlockCache.DeleteConfirmedBlock(b.Block.Slot); err != nil {
								log.WithError(err).Error("couldn't delete the verified blocks from cache")
								return err
							}
						}
						return nil
					}
					continue
				case vanTypes.Invalid:
					log.WithField("slot", data.Slot).WithField(
						"blockHash", data.BlockRootHash).Debug("invalid by orchestrator, exiting goroutine")
					return errInvalidBlock
				default:
					continue		// for pending status, we can not make instant decision so we are waiting for the confirmation
				}
			}
		case err := <-confirmedblockSub.Err():
			log.WithError(err).Error("Confirmation fetcher closed, exiting goroutine")
			return err
		case <-s.ctx.Done():
			log.WithField("function", "waitForConfirmationsBlockBatch").Debug("context is closed, exiting")
			return errPendingBlockCtxIsDone
		}
	}
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
		if i == 7 {
			resData = append(resData, &vanTypes.ConfirmationResData{
				Slot: request[i].Slot,
				Hash: request[i].Hash,
				Status: vanTypes.Invalid,
			})
			continue
		}
		resData = append(resData, &vanTypes.ConfirmationResData{
			Slot: request[i].Slot,
			Hash: request[i].Hash,
			Status: vanTypes.Verified,
		})
	}
	return resData, nil
}