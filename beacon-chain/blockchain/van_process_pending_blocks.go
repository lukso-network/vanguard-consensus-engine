package blockchain

import (
	"context"
	"github.com/pkg/errors"
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"github.com/prysmaticlabs/prysm/beacon-chain/core/feed"
	blockfeed "github.com/prysmaticlabs/prysm/beacon-chain/core/feed/block"
	"go.opencensus.io/trace"
)

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
func (s *Service) publishAndStorePendingBlockBatch(ctx context.Context, pendingBlkBatch []*ethpb.BeaconBlock) error {
	ctx, span := trace.StartSpan(ctx, "blockChain.publishAndStorePendingBlockBatch")
	defer span.End()

	for _, b := range pendingBlkBatch {

		// Sending pending block feed to streaming api
		log.WithField("slot", b.Slot).Debug("Unconfirmed block batch sends for publishing")
		s.blockNotifier.BlockFeed().Send(&feed.Event{
			Type: blockfeed.UnConfirmedBlock,
			Data: &blockfeed.UnConfirmedBlockData{Block: b},
		})

		// Storing pending block into pendingBlockCache
		if err := s.pendingBlockCache.AddPendingBlock(b); err != nil {
			return errors.Wrapf(err, "could not cache block of slot %d", b.Slot)
		}
	}

	return nil
}
