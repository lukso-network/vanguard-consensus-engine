package blockchain

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/prysmaticlabs/prysm/beacon-chain/core/feed"
	statefeed "github.com/prysmaticlabs/prysm/beacon-chain/core/feed/state"
	"github.com/prysmaticlabs/prysm/beacon-chain/core/helpers"
	"github.com/prysmaticlabs/prysm/beacon-chain/core/state"
	iface "github.com/prysmaticlabs/prysm/beacon-chain/state/interface"
	pb "github.com/prysmaticlabs/prysm/proto/beacon/p2p/v1"
	ethpbv1 "github.com/prysmaticlabs/prysm/proto/eth/v1"
	ethpb "github.com/prysmaticlabs/prysm/proto/eth/v1alpha1"
	"github.com/prysmaticlabs/prysm/proto/interfaces"
	"github.com/prysmaticlabs/prysm/shared/attestationutil"
	"github.com/prysmaticlabs/prysm/shared/bls"
	"github.com/prysmaticlabs/prysm/shared/bytesutil"
	"github.com/prysmaticlabs/prysm/shared/featureconfig"
	"github.com/prysmaticlabs/prysm/shared/params"
	"go.opencensus.io/trace"
)

// A custom slot deadline for processing state slots in our cache.
const slotDeadline = 5 * time.Second

// A custom deadline for deposit trie insertion.
const depositDeadline = 20 * time.Second

// This defines size of the upper bound for initial sync block cache.
var initialSyncBlockCacheSize = uint64(2 * params.BeaconConfig().SlotsPerEpoch)

// onBlock is called when a gossip block is received. It runs regular state transition on the block.
// The block's signing root should be computed before calling this method to avoid redundant
// computation in this method and methods it calls into.
//
// Spec pseudocode definition:
//   def on_block(store: Store, signed_block: SignedBeaconBlock) -> None:
//    block = signed_block.message
//    # Parent block must be known
//    assert block.parent_root in store.block_states
//    # Make a copy of the state to avoid mutability issues
//    pre_state = copy(store.block_states[block.parent_root])
//    # Blocks cannot be in the future. If they are, their consideration must be delayed until the are in the past.
//    assert get_current_slot(store) >= block.slot
//
//    # Check that block is later than the finalized epoch slot (optimization to reduce calls to get_ancestor)
//    finalized_slot = compute_start_slot_at_epoch(store.finalized_checkpoint.epoch)
//    assert block.slot > finalized_slot
//    # Check block is a descendant of the finalized block at the checkpoint finalized slot
//    assert get_ancestor(store, block.parent_root, finalized_slot) == store.finalized_checkpoint.root
//
//    # Check the block is valid and compute the post-state
//    state = pre_state.copy()
//    state_transition(state, signed_block, True)
//    # Add new block to the store
//    store.blocks[hash_tree_root(block)] = block
//    # Add new state for this block to the store
//    store.block_states[hash_tree_root(block)] = state
//
//    # Update justified checkpoint
//    if state.current_justified_checkpoint.epoch > store.justified_checkpoint.epoch:
//        if state.current_justified_checkpoint.epoch > store.best_justified_checkpoint.epoch:
//            store.best_justified_checkpoint = state.current_justified_checkpoint
//        if should_update_justified_checkpoint(store, state.current_justified_checkpoint):
//            store.justified_checkpoint = state.current_justified_checkpoint
//
//    # Update finalized checkpoint
//    if state.finalized_checkpoint.epoch > store.finalized_checkpoint.epoch:
//        store.finalized_checkpoint = state.finalized_checkpoint
//
//        # Potentially update justified if different from store
//        if store.justified_checkpoint != state.current_justified_checkpoint:
//            # Update justified if new justified is later than store justified
//            if state.current_justified_checkpoint.epoch > store.justified_checkpoint.epoch:
//                store.justified_checkpoint = state.current_justified_checkpoint
//                return
//
//            # Update justified if store justified is not in chain with finalized checkpoint
//            finalized_slot = compute_start_slot_at_epoch(store.finalized_checkpoint.epoch)
//            ancestor_at_finalized_slot = get_ancestor(store, store.justified_checkpoint.root, finalized_slot)
//            if ancestor_at_finalized_slot != store.finalized_checkpoint.root:
//                store.justified_checkpoint = state.current_justified_checkpoint
func (s *Service) onBlock(ctx context.Context, signed interfaces.SignedBeaconBlock, blockRoot [32]byte) error {
	ctx, span := trace.StartSpan(ctx, "blockChain.onBlock")
	defer span.End()

	if signed == nil || signed.IsNil() || signed.Block().IsNil() {
		return errors.New("nil block")
	}
	b := signed.Block()

	preState, err := s.getBlockPreState(ctx, b)
	if err != nil {
		return err
	}

	postState, err := state.ExecuteStateTransition(ctx, preState, signed)
	if err != nil {
		return err
	}
	// Vanguard: Validated by vanguard node. Now intercepting the execution and publishing the block
	// and waiting for confirmation from orchestrator. If Lukso vanguard flag is enabled then these segment of code will be executed
	if s.enableVanguardNode {
		curEpoch := helpers.CurrentEpoch(postState)
		nextEpoch := curEpoch + 1
		// TODO: this logic should be tested, In my opinion its not the best place to execute if below.
		// Its crucial for our system to publish epoch info to consumers
		if s.latestSentEpoch < nextEpoch {
			proposerIndices, pubKeys, err := helpers.ProposerIndicesInCache(postState.Copy(), nextEpoch)
			if err != nil {
				return errors.Wrap(err, "could not get proposer indices for publishing")
			}
			log.WithField("nextEpoch", nextEpoch).WithField("latestSentEpoch", s.latestSentEpoch).Debug("publishing latest epoch info")
			s.publishEpochInfo(signed.Block().Slot(), proposerIndices, pubKeys)
			s.latestSentEpoch = nextEpoch
		}

		parentRoot := bytesutil.ToBytes32(b.ParentRoot())
		parentBlk, err := s.cfg.BeaconDB.Block(s.ctx, parentRoot)
		if err != nil {
			return errors.Wrapf(errParentDoesNotExist, "vanguard node doesn't have a parent in db with slot: "+
				"%d and parentRoot: %#x", b.Slot(), b.ParentRoot())
		}

		isParentBlockNil := nil == parentBlk || parentBlk.IsNil()

		if isParentBlockNil && !s.hasInitSyncBlock(parentRoot) {
			return errors.Wrapf(errParentDoesNotExist, "vanguard node does not have block neither in db nor"+
				"in its init cache with slot: %d and parentRoot: %#x", b.Slot(), b.ParentRoot(),
			)
		}

		if isParentBlockNil {
			parentBlk = s.getInitSyncBlock(parentRoot)
		}

		parentBlkPhase0, err := parentBlk.PbPhase0Block()

		if nil != err {
			return errors.Wrap(err, "could not cast parent block to phase 0 block")
		}

		signedBlockPhase0, err := signed.PbPhase0Block()

		if nil != err {
			return errors.Wrap(err, "could not cast signed block to phase 0 block")
		}

		// this fallback is for nil-block received as a parent in genesis conditions
		// TODO: debug why for some cases for genesis block it is passing nil block
		// Possible cause is bad design of the genesis vanguard block (lacking of pandora shards)
		if s.genesisRoot == parentRoot && nil == parentBlkPhase0 {
			parentBlk, err = s.cfg.BeaconDB.GenesisBlock(s.ctx)
		}

		if nil != err {
			return errors.Wrap(err, "could not find genesis state in DB")
		}

		if err := s.VerifyPandoraShardConsecutiveness(parentBlkPhase0, signedBlockPhase0); err != nil {
			return errors.Wrap(err, "could not verify pandora shard info onBlock")
		}

		validator, err := preState.ValidatorAtIndex(b.ProposerIndex())

		if nil != err {
			return errors.Wrap(err, "could not get validator from state")
		}

		blsPubKey, err := bls.PublicKeyFromBytes(validator.GetPublicKey())

		if nil != err {
			return errors.Wrap(err, "could not cast validator from blsPubKey")
		}

		blockData := signed.Block()
		blockBody := blockData.Body()
		pandoraShards := blockBody.PandoraShards()

		if nil == pandoraShards || len(pandoraShards) < 1 {
			return errors.Wrap(errInvalidPandoraShardInfo, "empty pandora shards")
		}

		for shardIndex, shard := range pandoraShards {
			currentErr := GuardPandoraShardSignature(shard, blsPubKey)

			if nil != currentErr {
				return errors.Wrapf(
					currentErr,
					"signature of shard did not verify. Slot: %d ShardIndex: %d",
					blockData.Slot(),
					shardIndex,
				)
			}
		}

		// publish block to orchestrator and rpc service for sending minimal consensus info
		s.publishBlock(signed)

		if s.orcVerification {
			// waiting for orchestrator confirmation in live-sync mode
			if err := s.waitForConfirmation(ctx, signed); err != nil {
				return errors.Wrap(err, "could not publish and verified by orchestrator client onBlock")
			}
		}
	}

	if err := s.savePostStateInfo(ctx, blockRoot, signed, postState, false /* reg sync */); err != nil {
		return err
	}

	// Updating next slot state cache can happen in the background. It shouldn't block rest of the process.
	if featureconfig.Get().EnableNextSlotStateCache {
		go func() {
			// Use a custom deadline here, since this method runs asynchronously.
			// We ignore the parent method's context and instead create a new one
			// with a custom deadline, therefore using the background context instead.
			slotCtx, cancel := context.WithTimeout(context.Background(), slotDeadline)
			defer cancel()
			if err := state.UpdateNextSlotCache(slotCtx, blockRoot[:], postState); err != nil {
				log.WithError(err).Debug("could not update next slot state cache")
			}
		}()
	}

	// Update justified check point.
	if postState.CurrentJustifiedCheckpoint().Epoch > s.justifiedCheckpt.Epoch {
		if err := s.updateJustified(ctx, postState); err != nil {
			return err
		}
	}

	newFinalized := postState.FinalizedCheckpointEpoch() > s.finalizedCheckpt.Epoch
	if featureconfig.Get().UpdateHeadTimely {
		if newFinalized {
			if err := s.finalizedImpliesNewJustified(ctx, postState); err != nil {
				return errors.Wrap(err, "could not save new justified")
			}
			s.prevFinalizedCheckpt = s.finalizedCheckpt
			s.finalizedCheckpt = postState.FinalizedCheckpoint()
		}

		if err := s.updateHead(ctx, s.getJustifiedBalances()); err != nil {
			log.WithError(err).Warn("Could not update head")
		}

		if err := s.pruneCanonicalAttsFromPool(ctx, blockRoot, signed); err != nil {
			return err
		}

		// Send notification of the processed block to the state feed.
		s.cfg.StateNotifier.StateFeed().Send(&feed.Event{
			Type: statefeed.BlockProcessed,
			Data: &statefeed.BlockProcessedData{
				Slot:        signed.Block().Slot(),
				BlockRoot:   blockRoot,
				SignedBlock: signed,
				Verified:    true,
			},
		})
	}

	// Update finalized check point.
	if newFinalized {
		if err := s.updateFinalized(ctx, postState.FinalizedCheckpoint()); err != nil {
			return err
		}
		fRoot := bytesutil.ToBytes32(postState.FinalizedCheckpoint().Root)
		if err := s.cfg.ForkChoiceStore.Prune(ctx, fRoot); err != nil {
			return errors.Wrap(err, "could not prune proto array fork choice nodes")
		}
		if !featureconfig.Get().UpdateHeadTimely {
			if err := s.finalizedImpliesNewJustified(ctx, postState); err != nil {
				return errors.Wrap(err, "could not save new justified")
			}
		}
		go func() {
			// Send an event regarding the new finalized checkpoint over a common event feed.
			s.cfg.StateNotifier.StateFeed().Send(&feed.Event{
				Type: statefeed.FinalizedCheckpoint,
				Data: &ethpbv1.EventFinalizedCheckpoint{
					Epoch: postState.FinalizedCheckpoint().Epoch,
					Block: postState.FinalizedCheckpoint().Root,
					State: signed.Block().StateRoot(),
				},
			})

			// Use a custom deadline here, since this method runs asynchronously.
			// We ignore the parent method's context and instead create a new one
			// with a custom deadline, therefore using the background context instead.
			depCtx, cancel := context.WithTimeout(context.Background(), depositDeadline)
			defer cancel()
			if err := s.insertFinalizedDeposits(depCtx, fRoot); err != nil {
				log.WithError(err).Error("Could not insert finalized deposits.")
			}
		}()

	}

	defer reportAttestationInclusion(b)

	return s.handleEpochBoundary(ctx, postState)
}

func (s *Service) onBlockBatch(ctx context.Context, blks []interfaces.SignedBeaconBlock,
	blockRoots [][32]byte) ([]*ethpb.Checkpoint, []*ethpb.Checkpoint, error) {
	ctx, span := trace.StartSpan(ctx, "blockChain.onBlockBatch")
	defer span.End()

	if len(blks) == 0 || len(blockRoots) == 0 {
		return nil, nil, errors.New("no blocks provided")
	}
	if blks[0] == nil || blks[0].IsNil() || blks[0].Block().IsNil() {
		return nil, nil, errors.New("nil block")
	}
	b := blks[0].Block()

	// Retrieve incoming block's pre state.
	if err := s.verifyBlkPreState(ctx, b); err != nil {
		return nil, nil, err
	}
	preState, err := s.cfg.StateGen.StateByRootInitialSync(ctx, bytesutil.ToBytes32(b.ParentRoot()))
	if err != nil {
		return nil, nil, err
	}
	if preState == nil || preState.IsNil() {
		return nil, nil, fmt.Errorf("nil pre state for slot %d", b.Slot())
	}

	jCheckpoints := make([]*ethpb.Checkpoint, len(blks))
	fCheckpoints := make([]*ethpb.Checkpoint, len(blks))
	sigSet := &bls.SignatureSet{
		Signatures: [][]byte{},
		PublicKeys: []bls.PublicKey{},
		Messages:   [][32]byte{},
	}
	var set *bls.SignatureSet
	boundaries := make(map[[32]byte]iface.BeaconState)

	// TODO: verification signature set for pandora shards
	// TODO: verification of consecutiveness in shards
	for i, b := range blks {
		set, preState, err = state.ExecuteStateTransitionNoVerifyAnySig(ctx, preState, b)
		if err != nil {
			return nil, nil, err
		}
		// Save potential boundary states.
		if helpers.IsEpochStart(preState.Slot()) {
			boundaries[blockRoots[i]] = preState.Copy()
			if err := s.handleEpochBoundary(ctx, preState); err != nil {
				return nil, nil, errors.Wrap(err, "could not handle epoch boundary state")
			}
		}
		// TODO: this logic should be tested, In my opinion its not the best place to execute if below.
		// Its crucial for our system to publish epoch info to consumers
		if s.enableVanguardNode {
			curEpoch := helpers.CurrentEpoch(preState)
			nextEpoch := curEpoch + 1
			if s.latestSentEpoch < nextEpoch {
				proposerIndices, pubKeys, err := helpers.ProposerIndicesInCache(preState.Copy(), nextEpoch)
				if err != nil {
					return nil, nil, errors.Wrap(err, "could not get proposer indices for publishing")
				}
				log.WithField("nextEpoch", nextEpoch).WithField("latestSentEpoch", s.latestSentEpoch).Debug("publishing latest epoch info")
				s.publishEpochInfo(b.Block().Slot(), proposerIndices, pubKeys)
				s.latestSentEpoch = nextEpoch
			}

			blockData := b.Block()
			blockBody := blockData.Body()

			if nil == blockBody || blockBody.IsNil() {
				return nil, nil, errors.Wrap(err, "block body is nil, could not process batch")
			}

			proposerIndex := blockData.ProposerIndex()
			publicKey, currentErr := preState.ValidatorAtIndex(proposerIndex)

			if nil != currentErr {
				return nil, nil, errors.Wrap(
					currentErr,
					fmt.Sprintf("could not get proposer at index: %d", blockData.ProposerIndex()),
				)
			}

			blsPubKey, currentErr := bls.PublicKeyFromBytes(publicKey.GetPublicKey())

			if nil != currentErr {
				return nil, nil, errors.Wrap(currentErr, "could not cast bytes to pubkey")
			}

			pandoraShards := blockBody.PandoraShards()

			if nil == pandoraShards || len(pandoraShards) < 1 {
				return nil, nil, errors.Wrap(errInvalidPandoraShardInfo, "empty pandora shards")
			}

			for shardIndex, shard := range pandoraShards {
				currentErr = GuardPandoraShardSignature(shard, blsPubKey)

				if nil != currentErr {
					return nil, nil, errors.Wrapf(
						currentErr,
						"signature of shard did not verify. Slot: %d ShardIndex: %d",
						blockData.Slot(),
						shardIndex,
					)
				}
			}
		}
		jCheckpoints[i] = preState.CurrentJustifiedCheckpoint()
		fCheckpoints[i] = preState.FinalizedCheckpoint()
		sigSet.Join(set)
	}
	verify, err := sigSet.Verify()
	if err != nil {
		return nil, nil, err
	}
	if !verify {
		return nil, nil, errors.New("batch block signature verification failed")
	}
	// Verify only block batch that we received starting from block 1 and block 0 as a parent.
	// Do not dive into database because it might be missing due to round-robin sync
	// Whole verification should be done after transition from initial(optimistic) sync to regular sync.
	// TODO: check for the parent from initCache and db as in `onBlock`
	if s.enableVanguardNode {
		parentRoot := bytesutil.ToBytes32(b.ParentRoot())
		parentBlk, vanErr := s.cfg.BeaconDB.Block(s.ctx, parentRoot)

		if vanErr != nil {
			return nil, nil, errors.Wrapf(errParentDoesNotExist, "vanguard node doesn't have a parent in "+
				"db with slot: %d and parentRoot: %#x", b.Slot(), b.ParentRoot())
		}

		isParentBlockNil := nil == parentBlk || parentBlk.IsNil()

		if isParentBlockNil && !s.hasInitSyncBlock(parentRoot) {
			return nil, nil, errors.Wrapf(errParentDoesNotExist, "vanguard node does not have block"+
				"neither in db nor in its init cache with slot: %d and parentRoot: %#x", b.Slot(), b.ParentRoot(),
			)
		}

		if isParentBlockNil {
			parentBlk = s.getInitSyncBlock(parentRoot)
		}

		for i := 0; i < len(blks); i++ {
			if i > 0 {
				parentBlk = blks[i-1]
			}

			parentBlkPhase0, currentErr := parentBlk.PbPhase0Block()

			if nil != currentErr {
				return nil, nil, errors.Wrap(currentErr, "could not cast parent block to phase 0 block")
			}

			signedBlockPhase0, currentErr := blks[i].PbPhase0Block()

			if nil != currentErr {
				return nil, nil, errors.Wrap(currentErr, "could not cast signed block to phase 0 block")
			}

			currentErr = s.VerifyPandoraShardConsecutiveness(parentBlkPhase0, signedBlockPhase0)

			if nil != currentErr {
				return nil, nil, errors.Wrap(currentErr, "could not verify pandora shard info onBlockBatch")
			}

			s.publishBlock(blks[i])
		}
	}
	for r, st := range boundaries {
		if err := s.cfg.StateGen.SaveState(ctx, r, st); err != nil {
			return nil, nil, err
		}
	}
	// Also saves the last post state which to be used as pre state for the next batch.
	lastB := blks[len(blks)-1]
	lastBR := blockRoots[len(blockRoots)-1]
	if err := s.cfg.StateGen.SaveState(ctx, lastBR, preState); err != nil {
		return nil, nil, err
	}
	if err := s.saveHeadNoDB(ctx, lastB, lastBR, preState); err != nil {
		return nil, nil, err
	}
	return fCheckpoints, jCheckpoints, nil
}

// handles a block after the block's batch has been verified, where we can save blocks
// their state summaries and split them off to relative hot/cold storage.
func (s *Service) handleBlockAfterBatchVerify(ctx context.Context, signed interfaces.SignedBeaconBlock,
	blockRoot [32]byte, fCheckpoint, jCheckpoint *ethpb.Checkpoint) error {
	b := signed.Block()

	s.saveInitSyncBlock(blockRoot, signed)
	if err := s.insertBlockToForkChoiceStore(ctx, b, blockRoot, fCheckpoint, jCheckpoint); err != nil {
		return err
	}
	if err := s.cfg.BeaconDB.SaveStateSummary(ctx, &pb.StateSummary{
		Slot: signed.Block().Slot(),
		Root: blockRoot[:],
	}); err != nil {
		return err
	}

	// Rate limit how many blocks (2 epochs worth of blocks) a node keeps in the memory.
	if uint64(len(s.getInitSyncBlocks())) > initialSyncBlockCacheSize {
		if err := s.cfg.BeaconDB.SaveBlocks(ctx, s.getInitSyncBlocks()); err != nil {
			return err
		}
		s.clearInitSyncBlocks()
	}

	if jCheckpoint.Epoch > s.justifiedCheckpt.Epoch {
		if err := s.updateJustifiedInitSync(ctx, jCheckpoint); err != nil {
			return err
		}
	}

	// Update finalized check point. Prune the block cache and helper caches on every new finalized epoch.
	if fCheckpoint.Epoch > s.finalizedCheckpt.Epoch {
		if err := s.updateFinalized(ctx, fCheckpoint); err != nil {
			return err
		}
		if featureconfig.Get().UpdateHeadTimely {
			s.prevFinalizedCheckpt = s.finalizedCheckpt
			s.finalizedCheckpt = fCheckpoint
		}
	}
	return nil
}

// Epoch boundary bookkeeping such as logging epoch summaries.
func (s *Service) handleEpochBoundary(ctx context.Context, postState iface.BeaconState) error {
	if postState.Slot()+1 == s.nextEpochBoundarySlot {
		// Update caches for the next epoch at epoch boundary slot - 1.
		if err := helpers.UpdateCommitteeCache(postState, helpers.NextEpoch(postState)); err != nil {
			return err
		}
		copied := postState.Copy()
		copied, err := state.ProcessSlots(ctx, copied, copied.Slot()+1)
		if err != nil {
			return err
		}
		if err := helpers.UpdateProposerIndicesInCache(copied); err != nil {
			return err
		}
	} else if postState.Slot() >= s.nextEpochBoundarySlot {
		if err := reportEpochMetrics(ctx, postState, s.head.state); err != nil {
			return err
		}
		var err error
		s.nextEpochBoundarySlot, err = helpers.StartSlot(helpers.NextEpoch(postState))
		if err != nil {
			return err
		}

		// Update caches at epoch boundary slot.
		// The following updates have short cut to return nil cheaply if fulfilled during boundary slot - 1.
		if err := helpers.UpdateCommitteeCache(postState, helpers.CurrentEpoch(postState)); err != nil {
			return err
		}
		if err := helpers.UpdateProposerIndicesInCache(postState); err != nil {
			return err
		}
	}

	return nil
}

// This feeds in the block and block's attestations to fork choice store. It's allows fork choice store
// to gain information on the most current chain.
func (s *Service) insertBlockAndAttestationsToForkChoiceStore(ctx context.Context, blk interfaces.BeaconBlock, root [32]byte,
	st iface.BeaconState) error {
	fCheckpoint := st.FinalizedCheckpoint()
	jCheckpoint := st.CurrentJustifiedCheckpoint()
	if err := s.insertBlockToForkChoiceStore(ctx, blk, root, fCheckpoint, jCheckpoint); err != nil {
		return err
	}
	// Feed in block's attestations to fork choice store.
	for _, a := range blk.Body().Attestations() {
		committee, err := helpers.BeaconCommitteeFromState(st, a.Data.Slot, a.Data.CommitteeIndex)
		if err != nil {
			return err
		}
		indices, err := attestationutil.AttestingIndices(a.AggregationBits, committee)
		if err != nil {
			return err
		}
		s.cfg.ForkChoiceStore.ProcessAttestation(ctx, indices, bytesutil.ToBytes32(a.Data.BeaconBlockRoot), a.Data.Target.Epoch)
	}
	return nil
}

func (s *Service) insertBlockToForkChoiceStore(ctx context.Context, blk interfaces.BeaconBlock,
	root [32]byte, fCheckpoint, jCheckpoint *ethpb.Checkpoint) error {
	if err := s.fillInForkChoiceMissingBlocks(ctx, blk, fCheckpoint, jCheckpoint); err != nil {
		return err
	}
	// Feed in block to fork choice store.
	if err := s.cfg.ForkChoiceStore.ProcessBlock(ctx,
		blk.Slot(), root, bytesutil.ToBytes32(blk.ParentRoot()), bytesutil.ToBytes32(blk.Body().Graffiti()),
		jCheckpoint.Epoch,
		fCheckpoint.Epoch); err != nil {
		return errors.Wrap(err, "could not process block for proto array fork choice")
	}
	return nil
}

// This saves post state info to DB or cache. This also saves post state info to fork choice store.
// Post state info consists of processed block and state. Do not call this method unless the block and state are verified.
func (s *Service) savePostStateInfo(ctx context.Context, r [32]byte, b interfaces.SignedBeaconBlock, st iface.BeaconState, initSync bool) error {
	ctx, span := trace.StartSpan(ctx, "blockChain.savePostStateInfo")
	defer span.End()
	if initSync {
		s.saveInitSyncBlock(r, b)
	} else if err := s.cfg.BeaconDB.SaveBlock(ctx, b); err != nil {
		return errors.Wrapf(err, "could not save block from slot %d", b.Block().Slot())
	}
	if err := s.cfg.StateGen.SaveState(ctx, r, st); err != nil {
		return errors.Wrap(err, "could not save state")
	}
	if err := s.insertBlockAndAttestationsToForkChoiceStore(ctx, b.Block(), r, st); err != nil {
		return errors.Wrapf(err, "could not insert block %d to fork choice store", b.Block().Slot())
	}
	return nil
}

// This removes the attestations from the mem pool. It will only remove the attestations if input root `r` is canonical,
// meaning the block `b` is part of the canonical chain.
func (s *Service) pruneCanonicalAttsFromPool(ctx context.Context, r [32]byte, b interfaces.SignedBeaconBlock) error {
	if !featureconfig.Get().CorrectlyPruneCanonicalAtts {
		return nil
	}

	canonical, err := s.IsCanonical(ctx, r)
	if err != nil {
		return err
	}
	if !canonical {
		return nil
	}

	atts := b.Block().Body().Attestations()
	for _, att := range atts {
		if helpers.IsAggregated(att) {
			if err := s.cfg.AttPool.DeleteAggregatedAttestation(att); err != nil {
				return err
			}
		} else {
			if err := s.cfg.AttPool.DeleteUnaggregatedAttestation(att); err != nil {
				return err
			}
		}
	}
	return nil
}
