package blockchain

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	types "github.com/prysmaticlabs/eth2-types"
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"github.com/prysmaticlabs/prysm/beacon-chain/core/helpers"
	testDB "github.com/prysmaticlabs/prysm/beacon-chain/db/testing"
	"github.com/prysmaticlabs/prysm/beacon-chain/forkchoice/protoarray"
	"github.com/prysmaticlabs/prysm/beacon-chain/state/stategen"
	"github.com/prysmaticlabs/prysm/shared/params"
	"github.com/prysmaticlabs/prysm/shared/testutil"
	"github.com/prysmaticlabs/prysm/shared/testutil/assert"
	"github.com/prysmaticlabs/prysm/shared/testutil/require"
	"testing"
)

func TestServer_MinimalConsensusSuite(t *testing.T) {
	helpers.ClearCache()
	ctx := context.Background()
	beaconDB := testDB.SetupDB(t)

	count := 10000
	validators := make([]*ethpb.Validator, 0, count)
	withdrawCred := make([]byte, 32)
	for i := 0; i < count; i++ {
		pubKey := make([]byte, params.BeaconConfig().BLSPubkeyLength)
		binary.LittleEndian.PutUint64(pubKey, uint64(i))
		val := &ethpb.Validator{
			PublicKey:             pubKey,
			WithdrawalCredentials: withdrawCred,
			ExitEpoch:             params.BeaconConfig().FarFutureEpoch,
		}
		validators = append(validators, val)
	}

	config := params.BeaconConfig().Copy()
	oldConfig := config.Copy()
	config.SlotsPerEpoch = 32
	params.OverrideBeaconConfig(config)

	defer func() {
		params.OverrideBeaconConfig(oldConfig)
	}()

	parentRoot := [32]byte{1, 2, 3}
	blk := testutil.NewBeaconBlock().Block
	blk.ParentRoot = parentRoot[:]
	blockRoot, err := blk.HashTreeRoot()
	require.NoError(t, err)
	s, err := testutil.NewBeaconState()
	require.NoError(t, err)
	require.NoError(t, s.SetValidators(validators))
	require.NoError(t, beaconDB.SaveState(ctx, s, blockRoot))
	require.NoError(t, beaconDB.SaveGenesisBlockRoot(ctx, blockRoot))

	cfg := &Config{
		BeaconDB:        beaconDB,
		ForkChoiceStore: protoarray.New(0, 0, [32]byte{}),
		StateGen:        stategen.New(beaconDB),
	}

	service, err := NewService(ctx, cfg)
	require.NoError(t, err)

	validTestEpochs := 5
	blk.ParentRoot = parentRoot[:]
	require.NoError(t, err)
	blks := make([]*ethpb.SignedBeaconBlock, validTestEpochs*int(params.BeaconConfig().SlotsPerEpoch))
	lastValidTestSlot := params.BeaconConfig().SlotsPerEpoch * types.Slot(validTestEpochs)

	for i := types.Slot(0); i < lastValidTestSlot; i++ {
		b := testutil.NewBeaconBlock()
		b.Block.Slot = i
		b.Block.ParentRoot = parentRoot[:]
		blks[i] = b
		currentRoot, err := b.Block.HashTreeRoot()
		require.NoError(t, err)
		parentRoot = currentRoot
		st, err := testutil.NewBeaconState()
		require.NoError(t, err)
		require.NoError(t, st.SetSlot(i))
		require.NoError(t, st.SetValidators(validators))
		require.NoError(t, beaconDB.SaveState(ctx, st, currentRoot))
		assert.Equal(t, true, beaconDB.HasState(ctx, currentRoot))
		hasState, err := service.stateGen.HasState(ctx, currentRoot)
		require.NoError(t, err)
		assert.Equal(t, true, hasState)
	}

	require.NoError(t, beaconDB.SaveBlocks(ctx, blks))

	t.Run("should GetMinimalConsensusInfo", func(t *testing.T) {
		for epoch := types.Epoch(0); epoch < types.Epoch(validTestEpochs); epoch++ {
			minConsensusInfo, err := service.MinimalConsensusInfo(epoch)
			require.NoError(t, err)
			assert.Equal(t, epoch, minConsensusInfo.Epoch)
			if types.Epoch(0) == epoch {
				publicKeyBytes := make([]byte, params.BeaconConfig().BLSPubkeyLength)
				currentString := fmt.Sprintf("0x%s", hex.EncodeToString(publicKeyBytes))
				assert.Equal(t, currentString, minConsensusInfo.Epoch)
			}
		}
	})

	t.Run("should not GetMinimalConsensusInfo for future epoch", func(t *testing.T) {
		epoch := types.Epoch(validTestEpochs + 1)
		minimalConsensusInfo, err := service.MinimalConsensusInfo(epoch)

		assert.NotNil(t, err)
		assert.DeepEqual(t, (*ethpb.MinimalConsensusInfo)(nil), minimalConsensusInfo)
	})

	t.Run("should GetMinimalConsensusInfoRange", func(t *testing.T) {
		consensusInformation, err := service.MinimalConsensusInfoRange(0)
		require.NoError(t, err)
		require.DeepEqual(t, validTestEpochs, len(consensusInformation))
	})

	t.Run("should fail GetMinimalConsensusInfoRange", func(t *testing.T) {
		consensusInformation, err := service.MinimalConsensusInfoRange(types.Epoch(validTestEpochs + 1))
		require.NotNil(t, err)
		require.DeepEqual(t, 0, len(consensusInformation))
	})
}
