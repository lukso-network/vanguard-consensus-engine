package beacon

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	chainMock "github.com/prysmaticlabs/prysm/beacon-chain/blockchain/testing"
	"github.com/prysmaticlabs/prysm/beacon-chain/core/helpers"
	dbTest "github.com/prysmaticlabs/prysm/beacon-chain/db/testing"
	"github.com/prysmaticlabs/prysm/beacon-chain/state/stategen"
	mockSync "github.com/prysmaticlabs/prysm/beacon-chain/sync/initial-sync/testing"
	pb "github.com/prysmaticlabs/prysm/proto/beacon/p2p/v1"
	"github.com/prysmaticlabs/prysm/shared/params"
	"github.com/prysmaticlabs/prysm/shared/testutil"
	"github.com/prysmaticlabs/prysm/shared/testutil/assert"
	"github.com/prysmaticlabs/prysm/shared/testutil/require"
	"testing"
	"time"
)

func TestServer_StreamMinimalConsensusInfo_FromL15(t *testing.T) {
	genesisSszLocator := "l15-genesis.ssz"
	helpers.ClearCache()
	preBeaconStateFile, err := testutil.BazelFileBytes(genesisSszLocator)
	assert.NoError(t, err)

	config := params.BeaconConfig().Copy()
	oldConfig := config.Copy()
	config = params.L15Config()

	params.OverrideBeaconConfig(config)
	defer func() {
		params.OverrideBeaconConfig(oldConfig)
	}()

	beaconStateBase := &pb.BeaconState{}
	require.NoError(t, beaconStateBase.UnmarshalSSZ(preBeaconStateFile), "Failed to unmarshal")

	db := dbTest.SetupDB(t)
	ctx := context.Background()
	genesisTime := time.Unix(int64(config.MinGenesisTime), 0)
	stateNotifier := new(chainMock.ChainService).StateNotifier()

	beaconState, err := testutil.NewBeaconState(func(state *pb.BeaconState) (err error) {
		*state = *beaconStateBase

		return
	})
	assert.NoError(t, err)
	assert.Equal(t, "0x83a55317", hexutil.Encode(beaconState.Fork().CurrentVersion))

	server := &Server{
		BeaconDB: db,
		Ctx:      ctx,
		FinalizationFetcher: &chainMock.ChainService{
			Genesis: genesisTime,
		},
		GenesisTimeFetcher: &chainMock.ChainService{
			Genesis: genesisTime,
		},
		StateGen:      stategen.New(db),
		StateNotifier: stateNotifier,
		HeadFetcher: &chainMock.ChainService{
			State: beaconState,
		},
		SyncChecker: &mockSync.Sync{IsSyncing: false},
	}

	blockRoots := beaconState.BlockRoots()
	assert.Equal(t, true, len(blockRoots) > 0)
	blockRoot := [32]byte{}
	copy(blockRoot[:], blockRoots[0])

	require.NoError(t, db.SaveState(server.Ctx, beaconState, blockRoot))
	require.NoError(t, db.SaveGenesisBlockRoot(server.Ctx, blockRoot))

	resp, err := server.MinimalConsensusInfoRange(ctx, 0)
	require.NoError(t, err)
	assert.Equal(t, 2, len(resp))

	for index, info := range resp {
		if 0 == info.Epoch {
			publicKeyBytes := make([]byte, params.BeaconConfig().BLSPubkeyLength)
			currentString := fmt.Sprintf("0x%s", hexutil.Encode(publicKeyBytes))
			assert.Equal(
				t,
				currentString[:4],
				info.ValidatorList[0][:4],
				fmt.Sprintf(
					"Failed on inded: %d, wanted: %s, got: %s",
					index,
					publicKeyBytes,
					info.ValidatorList[0],
				))
		}
	}
}
