package blockchain

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	types "github.com/prysmaticlabs/eth2-types"
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	blockchainTesting "github.com/prysmaticlabs/prysm/beacon-chain/blockchain/testing"
	"github.com/prysmaticlabs/prysm/beacon-chain/core/helpers"
	dbTest "github.com/prysmaticlabs/prysm/beacon-chain/db/testing"
	"github.com/prysmaticlabs/prysm/beacon-chain/rpc/beacon"
	"github.com/prysmaticlabs/prysm/beacon-chain/rpc/subscriber/api/events"
	"github.com/prysmaticlabs/prysm/beacon-chain/state/stategen"
	"github.com/prysmaticlabs/prysm/shared/params"
	"github.com/prysmaticlabs/prysm/shared/testutil"
	"github.com/prysmaticlabs/prysm/shared/testutil/assert"
	"github.com/prysmaticlabs/prysm/shared/testutil/require"
	"testing"
	"time"
)

func TestServer_MinimalConsensusSuite(t *testing.T) {
	helpers.ClearCache()
	db := dbTest.SetupDB(t)
	ctx := context.Background()
	cfg := &Config{
		ConsensusNotifier: blockchainTesting.MockConsensusNotifier{},
	}
	service, err := NewService(ctx, cfg)
	require.NoError(t, err)
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

	testStartTime := time.Now()
	validTestEpochs := 5
	totalSec := int64(params.BeaconConfig().SlotsPerEpoch.Mul(uint64(validTestEpochs) * params.BeaconConfig().SecondsPerSlot))
	genTime := testStartTime.Unix() - totalSec
	blks := make([]*ethpb.SignedBeaconBlock, validTestEpochs*int(params.BeaconConfig().SlotsPerEpoch))
	lastValidTestSlot := params.BeaconConfig().SlotsPerEpoch * types.Slot(validTestEpochs)
	parentRoot := [32]byte{1, 2, 3}

	blk := testutil.NewBeaconBlock().Block
	blk.ParentRoot = parentRoot[:]
	blockRoot, err := blk.HashTreeRoot()
	require.NoError(t, err)
	s, err := testutil.NewBeaconState()
	require.NoError(t, err)
	require.NoError(t, s.SetValidators(validators))
	require.NoError(t, db.SaveState(ctx, s, blockRoot))
	require.NoError(t, db.SaveGenesisBlockRoot(ctx, blockRoot))

	parentRoot = blockRoot

	bs := &beacon.Server{
		Ctx:      ctx,
		BeaconDB: db,
		FinalizationFetcher: &blockchainTesting.ChainService{
			Genesis: time.Unix(genTime, 0),
			FinalizedCheckPoint: &ethpb.Checkpoint{
				Epoch: 0,
			},
		},
		GenesisTimeFetcher: &blockchainTesting.ChainService{
			Genesis: time.Unix(genTime, 0),
		},
		StateGen: stategen.New(db),
	}

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
		require.NoError(t, db.SaveState(ctx, st, currentRoot))
		assert.Equal(t, true, db.HasState(ctx, currentRoot))
		hasState, err := bs.StateGen.HasState(ctx, currentRoot)
		require.NoError(t, err)
		assert.Equal(t, true, hasState)
	}

	require.NoError(t, db.SaveBlocks(ctx, blks))

	t.Run("should GetMinimalConsensusInfo", func(t *testing.T) {
		for epoch := types.Epoch(0); epoch < types.Epoch(validTestEpochs); epoch++ {

			assignments, err := MinimalConsensusInfo(ctx, epoch)

			require.NoError(t, err)
			assert.Equal(t, epoch, types.Epoch(assignments.Epoch))

			if types.Epoch(0) == epoch {
				publicKeyBytes := make([]byte, params.BeaconConfig().BLSPubkeyLength)
				currentString := fmt.Sprintf("0x%s", hex.EncodeToString(publicKeyBytes))
				assert.Equal(t, currentString, assignments.ValidatorList[0])
			}
		}
	})

	t.Run("should not GetMinimalConsensusInfo for future epoch", func(t *testing.T) {
		epoch := types.Epoch(validTestEpochs + 1)
		minimalConsensusInfo, err := bs.GetMinimalConsensusInfo(ctx, epoch)
		assert.NotNil(t, err)
		assert.DeepEqual(t, (*events.MinimalEpochConsensusInfo)(nil), minimalConsensusInfo)
	})

	t.Run("should GetMinimalConsensusInfoRange", func(t *testing.T) {
		consensusInformation, err := bs.GetMinimalConsensusInfoRange(ctx, 0)
		require.NoError(t, err)
		require.DeepEqual(t, validTestEpochs, len(consensusInformation))
	})

	t.Run("should fail GetMinimalConsensusInfoRange", func(t *testing.T) {
		consensusInformation, err := bs.GetMinimalConsensusInfoRange(ctx, types.Epoch(validTestEpochs+1))
		require.NotNil(t, err)
		require.DeepEqual(t, 0, len(consensusInformation))
	})
}
