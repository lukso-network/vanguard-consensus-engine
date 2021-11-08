package blockchain

import (
	"bytes"
	"context"
	"flag"
	"github.com/bazelbuild/rules_go/go/tools/bazel"
	beacondb "github.com/prysmaticlabs/prysm/beacon-chain/db"
	"github.com/prysmaticlabs/prysm/beacon-chain/db/kv"
	"github.com/prysmaticlabs/prysm/shared/cmd"
	"github.com/urfave/cli/v2"
	"io/ioutil"
	"path"
	"testing"
	"time"

	types "github.com/prysmaticlabs/eth2-types"
	mock "github.com/prysmaticlabs/prysm/beacon-chain/blockchain/testing"
	"github.com/prysmaticlabs/prysm/beacon-chain/core/helpers"
	testDB "github.com/prysmaticlabs/prysm/beacon-chain/db/testing"
	pb "github.com/prysmaticlabs/prysm/proto/beacon/p2p/v1"
	ethpbv1 "github.com/prysmaticlabs/prysm/proto/eth/v1"
	ethpb "github.com/prysmaticlabs/prysm/proto/eth/v1alpha1"
	"github.com/prysmaticlabs/prysm/proto/eth/v1alpha1/wrapper"
	"github.com/prysmaticlabs/prysm/shared/featureconfig"
	"github.com/prysmaticlabs/prysm/shared/params"
	"github.com/prysmaticlabs/prysm/shared/testutil"
	"github.com/prysmaticlabs/prysm/shared/testutil/assert"
	"github.com/prysmaticlabs/prysm/shared/testutil/require"
	logTest "github.com/sirupsen/logrus/hooks/test"
)

const (
	restoreSrcFilePath = "fixtures/vm4_backup_beaconchain.db"
	vm4HeadBlockSlot   = 27982
	vm4HeadForkSlot    = types.Slot(18962)
)

func TestSaveHead_Same(t *testing.T) {
	beaconDB := testDB.SetupDB(t)
	service := setupBeaconChain(t, beaconDB)

	r := [32]byte{'A'}
	service.head = &head{slot: 0, root: r}

	require.NoError(t, service.saveHead(context.Background(), r))
	assert.Equal(t, types.Slot(0), service.headSlot(), "Head did not stay the same")
	assert.Equal(t, r, service.headRoot(), "Head did not stay the same")
}

func TestSaveHead_Different(t *testing.T) {
	ctx := context.Background()
	beaconDB := testDB.SetupDB(t)
	service := setupBeaconChain(t, beaconDB)

	testutil.NewBeaconBlock()
	oldBlock := wrapper.WrappedPhase0SignedBeaconBlock(
		testutil.NewBeaconBlock(),
	)
	require.NoError(t, service.cfg.BeaconDB.SaveBlock(context.Background(), oldBlock))
	oldRoot, err := oldBlock.Block().HashTreeRoot()
	require.NoError(t, err)
	service.head = &head{
		slot:  0,
		root:  oldRoot,
		block: oldBlock,
	}

	newHeadSignedBlock := testutil.NewBeaconBlock()
	newHeadSignedBlock.Block.Slot = 1
	newHeadBlock := newHeadSignedBlock.Block

	require.NoError(t, service.cfg.BeaconDB.SaveBlock(context.Background(), wrapper.WrappedPhase0SignedBeaconBlock(newHeadSignedBlock)))
	newRoot, err := newHeadBlock.HashTreeRoot()
	require.NoError(t, err)
	headState, err := testutil.NewBeaconState()
	require.NoError(t, err)
	require.NoError(t, headState.SetSlot(1))
	require.NoError(t, service.cfg.BeaconDB.SaveStateSummary(context.Background(), &pb.StateSummary{Slot: 1, Root: newRoot[:]}))
	require.NoError(t, service.cfg.BeaconDB.SaveState(context.Background(), headState, newRoot))
	require.NoError(t, service.saveHead(context.Background(), newRoot))

	assert.Equal(t, types.Slot(1), service.HeadSlot(), "Head did not change")

	cachedRoot, err := service.HeadRoot(context.Background())
	require.NoError(t, err)
	assert.DeepEqual(t, cachedRoot, newRoot[:], "Head did not change")
	assert.DeepEqual(t, newHeadSignedBlock, service.headBlock().Proto(), "Head did not change")
	assert.DeepSSZEqual(t, headState.CloneInnerState(), service.headState(ctx).CloneInnerState(), "Head did not change")
}

func TestSaveHead_Different_Reorg(t *testing.T) {
	tests := []struct {
		name                   string
		headSlot               types.Slot
		newHeadSignedBlockSlot types.Slot
		expectedHeadSlot       types.Slot
		expectedLogOutput      []string
		stateSummarySlot       types.Slot
		vanguardNodeEnabled    bool
		loadBeaconChain        bool
	}{
		{
			name:                   "Checks for standard reorg feature",
			headSlot:               0,
			newHeadSignedBlockSlot: 1,
			expectedHeadSlot:       1,
			expectedLogOutput: []string{
				"Chain reorg occurred",
			},
			stateSummarySlot:    1,
			vanguardNodeEnabled: false,
			loadBeaconChain:     false,
		},
		{
			name:                   "Checks for Vanguard reorg feature",
			headSlot:               0,
			newHeadSignedBlockSlot: 1,
			expectedHeadSlot:       1,
			expectedLogOutput: []string{
				"Chain reorg occurred",
				"Setting latest sent epoch - vanguard node is enabled",
			},
			stateSummarySlot:    1,
			vanguardNodeEnabled: true,
			loadBeaconChain:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			hook := logTest.NewGlobal()
			beaconDB := testDB.SetupDB(t)
			service := setupBeaconChain(t, beaconDB)
			service.enableVanguardNode = tt.vanguardNodeEnabled
			oldBlock := wrapper.WrappedPhase0SignedBeaconBlock(
				testutil.NewBeaconBlock(),
			)
			require.NoError(t, service.cfg.BeaconDB.SaveBlock(context.Background(), oldBlock))

			oldRoot, err := oldBlock.Block().HashTreeRoot()
			require.NoError(t, err)

			service.head = &head{
				slot:  tt.headSlot,
				root:  oldRoot,
				block: oldBlock,
			}
			reorgChainParent := [32]byte{'B'}
			newHeadSignedBlock := testutil.NewBeaconBlock()
			newHeadSignedBlock.Block.Slot = tt.newHeadSignedBlockSlot
			newHeadSignedBlock.Block.ParentRoot = reorgChainParent[:]
			newHeadBlock := newHeadSignedBlock.Block
			require.NoError(t, service.cfg.BeaconDB.SaveBlock(context.Background(), wrapper.WrappedPhase0SignedBeaconBlock(newHeadSignedBlock)))

			newRoot, err := newHeadBlock.HashTreeRoot()
			require.NoError(t, err)

			headState, err := testutil.NewBeaconState()
			require.NoError(t, err)
			require.NoError(t, headState.SetSlot(1))
			require.NoError(t, service.cfg.BeaconDB.SaveStateSummary(context.Background(), &pb.StateSummary{Slot: tt.stateSummarySlot, Root: newRoot[:]}))
			require.NoError(t, service.cfg.BeaconDB.SaveState(context.Background(), headState, newRoot))
			require.NoError(t, service.saveHead(context.Background(), newRoot))
			assert.Equal(t, tt.expectedHeadSlot, service.HeadSlot(), "Head did not change")

			cachedRoot, err := service.HeadRoot(context.Background())
			require.NoError(t, err)
			if !bytes.Equal(cachedRoot, newRoot[:]) {
				t.Error("Head did not change")
			}
			assert.DeepEqual(t, newHeadSignedBlock, service.headBlock().Proto(), "Head did not change")
			assert.DeepSSZEqual(t, headState.CloneInnerState(), service.headState(ctx).CloneInnerState(), "Head did not change")

			for _, logOutput := range tt.expectedLogOutput {
				require.LogsContain(t, hook, logOutput)
			}
		})
	}
}

func TestSaveHead_Different_ReorgFix(t *testing.T) {
	ctx := context.Background()
	hook := logTest.NewGlobal()
	restoreDir := t.TempDir()

	app := cli.App{}
	set := flag.NewFlagSet("test", 0)
	set.String(cmd.RestoreSourceFileFlag.Name, "", "")
	set.String(cmd.RestoreTargetDirFlag.Name, "", "")
	bazelFilePath, err := bazel.Runfile(restoreSrcFilePath)
	assert.NoError(t, err)
	require.NoError(t, set.Set(cmd.RestoreSourceFileFlag.Name, bazelFilePath))
	require.NoError(t, set.Set(cmd.RestoreTargetDirFlag.Name, restoreDir))
	cliCtx := cli.NewContext(&app, set, nil)

	assert.NoError(t, beacondb.Restore(cliCtx))

	files, err := ioutil.ReadDir(path.Join(restoreDir, kv.BeaconNodeDbDirName))
	require.NoError(t, err)
	assert.Equal(t, 1, len(files))
	assert.Equal(t, kv.DatabaseFileName, files[0].Name())

	restoredDb := testDB.LoadDB(t, path.Join(restoreDir, kv.BeaconNodeDbDirName))
	assert.NotNil(t, restoredDb)
	headBlock, err := restoredDb.HeadBlock(ctx)
	require.NoError(t, err)
	assert.Equal(t, types.Slot(vm4HeadBlockSlot), headBlock.Block().Slot(), "Restored database has incorrect data")

	service := setupBeaconChain(t, restoredDb)
	service.enableVanguardNode = true

	oldBlock := wrapper.WrappedPhase0SignedBeaconBlock(
		testutil.NewBeaconBlock(),
	)
	require.NoError(t, service.cfg.BeaconDB.SaveBlock(context.Background(), oldBlock))

	oldRoot, err := oldBlock.Block().HashTreeRoot()
	require.NoError(t, err)

	service.head = &head{
		slot:  vm4HeadForkSlot - 1,
		root:  oldRoot,
		block: oldBlock,
	}

	reorgChainParent := [32]byte{'B'}
	newHeadSignedBlock := testutil.NewBeaconBlock()
	newHeadSignedBlock.Block.Slot = vm4HeadForkSlot
	newHeadSignedBlock.Block.ParentRoot = reorgChainParent[:]
	newHeadBlock := newHeadSignedBlock.Block

	require.NoError(t, service.cfg.BeaconDB.SaveBlock(context.Background(), wrapper.WrappedPhase0SignedBeaconBlock(newHeadSignedBlock)))
	newRoot, err := newHeadBlock.HashTreeRoot()
	require.NoError(t, err)
	headState, err := testutil.NewBeaconState()
	require.NoError(t, err)
	require.NoError(t, headState.SetSlot(vm4HeadForkSlot))
	require.NoError(t, service.cfg.BeaconDB.SaveStateSummary(context.Background(), &pb.StateSummary{Slot: vm4HeadForkSlot, Root: newRoot[:]}))
	require.NoError(t, service.cfg.BeaconDB.SaveState(context.Background(), headState, newRoot))
	require.NoError(t, service.saveHead(context.Background(), newRoot))

	assert.Equal(t, vm4HeadForkSlot, service.HeadSlot(), "Head did not change")

	cachedRoot, err := service.HeadRoot(context.Background())
	require.NoError(t, err)
	if !bytes.Equal(cachedRoot, newRoot[:]) {
		t.Error("Head did not change")
	}
	assert.DeepEqual(t, newHeadSignedBlock, service.headBlock().Proto(), "Head did not change")
	assert.DeepSSZEqual(t, headState.CloneInnerState(), service.headState(ctx).CloneInnerState(), "Head did not change")
	require.LogsContain(t, hook, "Chain reorg occurred")
	require.LogsContain(t, hook, "Setting latest sent epoch - vanguard node is enabled")

	assert.Equal(t, service.getLatestSentEpoch(), helpers.SlotToEpoch(newHeadSignedBlock.Block.Slot))
}

func TestCacheJustifiedStateBalances_CanCache(t *testing.T) {
	beaconDB := testDB.SetupDB(t)
	service := setupBeaconChain(t, beaconDB)

	state, _ := testutil.DeterministicGenesisState(t, 100)
	r := [32]byte{'a'}
	require.NoError(t, service.cfg.BeaconDB.SaveStateSummary(context.Background(), &pb.StateSummary{Root: r[:]}))
	require.NoError(t, service.cfg.BeaconDB.SaveState(context.Background(), state, r))
	require.NoError(t, service.cacheJustifiedStateBalances(context.Background(), r))
	require.DeepEqual(t, service.getJustifiedBalances(), state.Balances(), "Incorrect justified balances")
}

func TestUpdateHead_MissingJustifiedRoot(t *testing.T) {
	beaconDB := testDB.SetupDB(t)
	service := setupBeaconChain(t, beaconDB)

	b := testutil.NewBeaconBlock()
	require.NoError(t, service.cfg.BeaconDB.SaveBlock(context.Background(), wrapper.WrappedPhase0SignedBeaconBlock(b)))
	r, err := b.Block.HashTreeRoot()
	require.NoError(t, err)

	service.justifiedCheckpt = &ethpb.Checkpoint{Root: r[:]}
	service.finalizedCheckpt = &ethpb.Checkpoint{}
	service.bestJustifiedCheckpt = &ethpb.Checkpoint{}

	require.NoError(t, service.updateHead(context.Background(), []uint64{}))
}

func Test_notifyNewHeadEvent(t *testing.T) {
	t.Run("genesis_state_root", func(t *testing.T) {
		bState, _ := testutil.DeterministicGenesisState(t, 10)
		notifier := &mock.MockStateNotifier{RecordEvents: true}
		srv := &Service{
			cfg: &Config{
				StateNotifier: notifier,
			},
			genesisRoot: [32]byte{1},
		}
		newHeadStateRoot := [32]byte{2}
		newHeadRoot := [32]byte{3}
		err := srv.notifyNewHeadEvent(1, bState, newHeadStateRoot[:], newHeadRoot[:])
		require.NoError(t, err)
		events := notifier.ReceivedEvents()
		require.Equal(t, 1, len(events))

		eventHead, ok := events[0].Data.(*ethpbv1.EventHead)
		require.Equal(t, true, ok)
		wanted := &ethpbv1.EventHead{
			Slot:                      1,
			Block:                     newHeadRoot[:],
			State:                     newHeadStateRoot[:],
			EpochTransition:           false,
			PreviousDutyDependentRoot: srv.genesisRoot[:],
			CurrentDutyDependentRoot:  srv.genesisRoot[:],
		}
		require.DeepSSZEqual(t, wanted, eventHead)
	})
	t.Run("non_genesis_values", func(t *testing.T) {
		bState, _ := testutil.DeterministicGenesisState(t, 10)
		notifier := &mock.MockStateNotifier{RecordEvents: true}
		genesisRoot := [32]byte{1}
		srv := &Service{
			cfg: &Config{
				StateNotifier: notifier,
			},
			genesisRoot: genesisRoot,
		}
		epoch1Start, err := helpers.StartSlot(1)
		require.NoError(t, err)
		epoch2Start, err := helpers.StartSlot(1)
		require.NoError(t, err)
		require.NoError(t, bState.SetSlot(epoch1Start))

		newHeadStateRoot := [32]byte{2}
		newHeadRoot := [32]byte{3}
		err = srv.notifyNewHeadEvent(epoch2Start, bState, newHeadStateRoot[:], newHeadRoot[:])
		require.NoError(t, err)
		events := notifier.ReceivedEvents()
		require.Equal(t, 1, len(events))

		eventHead, ok := events[0].Data.(*ethpbv1.EventHead)
		require.Equal(t, true, ok)
		wanted := &ethpbv1.EventHead{
			Slot:                      epoch2Start,
			Block:                     newHeadRoot[:],
			State:                     newHeadStateRoot[:],
			EpochTransition:           false,
			PreviousDutyDependentRoot: genesisRoot[:],
			CurrentDutyDependentRoot:  make([]byte, 32),
		}
		require.DeepSSZEqual(t, wanted, eventHead)
	})
}

func TestSaveOrphanedAtts(t *testing.T) {
	resetCfg := featureconfig.InitWithReset(&featureconfig.Flags{
		CorrectlyInsertOrphanedAtts: true,
	})
	defer resetCfg()

	genesis, keys := testutil.DeterministicGenesisState(t, 64)
	b, err := testutil.GenerateFullBlock(genesis, keys, testutil.DefaultBlockGenConfig(), 1)
	assert.NoError(t, err)
	r, err := b.Block.HashTreeRoot()
	require.NoError(t, err)

	ctx := context.Background()
	beaconDB := testDB.SetupDB(t)
	service := setupBeaconChain(t, beaconDB)
	service.genesisTime = time.Now()

	require.NoError(t, service.cfg.BeaconDB.SaveBlock(ctx, wrapper.WrappedPhase0SignedBeaconBlock(b)))
	require.NoError(t, service.saveOrphanedAtts(ctx, r))

	require.Equal(t, len(b.Block.Body.Attestations), service.cfg.AttPool.AggregatedAttestationCount())
	savedAtts := service.cfg.AttPool.AggregatedAttestations()
	atts := b.Block.Body.Attestations
	require.DeepSSZEqual(t, atts, savedAtts)
}

func TestSaveOrphanedAtts_CanFilter(t *testing.T) {
	resetCfg := featureconfig.InitWithReset(&featureconfig.Flags{
		CorrectlyInsertOrphanedAtts: true,
	})
	defer resetCfg()

	genesis, keys := testutil.DeterministicGenesisState(t, 64)
	b, err := testutil.GenerateFullBlock(genesis, keys, testutil.DefaultBlockGenConfig(), 1)
	assert.NoError(t, err)
	r, err := b.Block.HashTreeRoot()
	require.NoError(t, err)

	ctx := context.Background()
	beaconDB := testDB.SetupDB(t)
	service := setupBeaconChain(t, beaconDB)
	service.genesisTime = time.Now().Add(time.Duration(-1*int64(params.BeaconConfig().SlotsPerEpoch+1)*int64(params.BeaconConfig().SecondsPerSlot)) * time.Second)

	require.NoError(t, service.cfg.BeaconDB.SaveBlock(ctx, wrapper.WrappedPhase0SignedBeaconBlock(b)))
	require.NoError(t, service.saveOrphanedAtts(ctx, r))

	require.Equal(t, 0, service.cfg.AttPool.AggregatedAttestationCount())
	savedAtts := service.cfg.AttPool.AggregatedAttestations()
	atts := b.Block.Body.Attestations
	require.DeepNotSSZEqual(t, atts, savedAtts)
}
