package blockchain

import (
	"context"
	types "github.com/prysmaticlabs/eth2-types"
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	blockchainTesting "github.com/prysmaticlabs/prysm/beacon-chain/blockchain/testing"
	"github.com/prysmaticlabs/prysm/shared/testutil"
	"github.com/prysmaticlabs/prysm/shared/testutil/assert"
	"github.com/prysmaticlabs/prysm/shared/testutil/require"
	"math/rand"
	"sort"
	"sync"
	"testing"
)

// TestService_PublishAndStorePendingBlock checks PublishAndStorePendingBlock method
func TestService_PublishAndStorePendingBlock(t *testing.T) {
	ctx := context.Background()
	cfg := &Config{
		BlockNotifier: &blockchainTesting.MockBlockNotifier{},
	}
	s, err := NewService(ctx, cfg)
	require.NoError(t, err)

	b := testutil.NewBeaconBlock()
	require.NoError(t, s.publishAndStorePendingBlock(ctx, b.Block))
	cachedBlock, err := s.pendingBlockCache.PendingBlock(b.Block.GetSlot())
	require.NoError(t, err)
	assert.DeepEqual(t, b.Block, cachedBlock)
}

// TestService_PublishAndStorePendingBlockBatch checks PublishAndStorePendingBlockBatch method
func TestService_PublishAndStorePendingBlockBatch(t *testing.T) {
	ctx := context.Background()
	cfg := &Config{
		BlockNotifier: &blockchainTesting.MockBlockNotifier{},
	}
	s, err := NewService(ctx, cfg)
	require.NoError(t, err)

	blks := make([]*ethpb.SignedBeaconBlock, 10)
	for i := 0; i < 10; i++ {
		b := testutil.NewBeaconBlock()
		blks[i] = b
	}

	require.NoError(t, s.publishAndStorePendingBlockBatch(ctx, blks))
	require.NoError(t, err)
	for _, blk := range blks {
		cachedBlock, err := s.pendingBlockCache.PendingBlock(blk.Block.GetSlot())
		require.NoError(t, err)
		assert.DeepEqual(t, blk.Block, cachedBlock)
	}
}

// TestService_SortedUnConfirmedBlocksFromCache checks SortedUnConfirmedBlocksFromCache method
func TestService_SortedUnConfirmedBlocksFromCache(t *testing.T) {
	ctx := context.Background()
	cfg := &Config{
		BlockNotifier: &blockchainTesting.MockBlockNotifier{},
	}
	s, err := NewService(ctx, cfg)
	require.NoError(t, err)

	blks := make([]*ethpb.BeaconBlock, 10)
	for i := 0; i < 10; i++ {
		b := testutil.NewBeaconBlock()
		b.Block.Slot = types.Slot(rand.Uint64() % 100)
		blks[i] = b.Block
		require.NoError(t, s.pendingBlockCache.AddPendingBlock(b.Block))
	}

	sort.Slice(blks, func(i, j int) bool {
		return blks[i].Slot < blks[j].Slot
	})

	sortedBlocks, err := s.SortedUnConfirmedBlocksFromCache()
	require.NoError(t, err)
	require.DeepEqual(t, blks, sortedBlocks)
}

// TestService_fetchOrcConfirmations checks fetchOrcConfirmations
func TestService_fetchOrcConfirmations(t *testing.T) {
	ctx := context.Background()
	cfg := &Config{
		BlockNotifier: &blockchainTesting.MockBlockNotifier{RecordEvents: true},
	}
	s, err := NewService(ctx, cfg)
	require.NoError(t, err)

	blks := make([]*ethpb.BeaconBlock, 10)
	for i := 0; i < 10; i++ {
		b := testutil.NewBeaconBlock()
		b.Block.Slot = types.Slot(rand.Uint64() % 100)
		blks[i] = b.Block
		require.NoError(t, s.pendingBlockCache.AddPendingBlock(b.Block))
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		require.NoError(t, s.fetchOrcConfirmations(ctx))
		wg.Done()
	}()
	wg.Wait()

	if recvd := len(s.blockNotifier.(*blockchainTesting.MockBlockNotifier).ReceivedEvents()); recvd < 1 {
		t.Errorf("Received %d pending block notifications, expected at least 1", recvd)
	}
}

// TestService_waitForConfirmationBlock checks waitForConfirmationBlock method
// When the confirmation result of the block is verified then waitForConfirmationBlock gives you nil return
// and delete the verified block from cache
func TestService_waitForConfirmationBlock_WithVerifiedBlock(t *testing.T) {
	ctx := context.Background()
	cfg := &Config{
		BlockNotifier: &blockchainTesting.MockBlockNotifier{},
	}
	s, err := NewService(ctx, cfg)
	require.NoError(t, err)

	blks := make([]*ethpb.SignedBeaconBlock, 10)
	for i := 0; i < 10; i++ {
		b := testutil.NewBeaconBlock()
		b.Block.Slot = types.Slot(rand.Uint64() % 100)
		blks[i] = b
		require.NoError(t, s.pendingBlockCache.AddPendingBlock(b.Block))
	}

	sort.Slice(blks, func(i, j int) bool {
		return blks[i].Block.Slot < blks[j].Block.Slot
	})

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		require.NoError(t, s.fetchOrcConfirmations(ctx))
		wg.Done()
	}()

	go func() {
		require.NoError(t, s.waitForConfirmationBlock(ctx, blks[6]))
		wg.Done()
	}()

	wg.Wait()
	b, err := s.pendingBlockCache.PendingBlock(blks[6].Block.GetSlot())
	require.NoError(t, err)
	var expected *ethpb.BeaconBlock
	assert.Equal(t, expected, b)
}

// TestService_waitForConfirmationBlock_WithInvalidBlock checks waitForConfirmationBlock method
// When the confirmation result of the block is verified then waitForConfirmationBlock gives you error return
// Not delete the invalid block because, when node gets an valid block, then it will be replaced and then it will be deleted
func TestService_waitForConfirmationBlock_WithInvalidBlock(t *testing.T) {
	ctx := context.Background()
	cfg := &Config{
		BlockNotifier: &blockchainTesting.MockBlockNotifier{},
	}
	s, err := NewService(ctx, cfg)
	require.NoError(t, err)

	blks := make([]*ethpb.SignedBeaconBlock, 10)
	for i := 0; i < 10; i++ {
		b := testutil.NewBeaconBlock()
		b.Block.Slot = types.Slot(rand.Uint64() % 100)
		blks[i] = b
		require.NoError(t, s.pendingBlockCache.AddPendingBlock(b.Block))
	}

	sort.Slice(blks, func(i, j int) bool {
		return blks[i].Block.Slot < blks[j].Block.Slot
	})

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		require.NoError(t, s.fetchOrcConfirmations(ctx))
		wg.Done()
	}()

	go func() {
		want := "invalid block found, discarded block batch"
		require.ErrorContains(t, want, s.waitForConfirmationBlock(ctx, blks[7]))
		wg.Done()
	}()

	wg.Wait()
	b, err := s.pendingBlockCache.PendingBlock(blks[7].Block.GetSlot())
	require.NoError(t, err)
	assert.Equal(t, blks[7].Block, b)
}

func TestService_waitForConfirmationBlockBatch_WithAllValidBlocks(t *testing.T) {

}