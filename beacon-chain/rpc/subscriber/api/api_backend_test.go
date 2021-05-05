package api

import (
	"context"
	"encoding/binary"
	"github.com/ethereum/go-ethereum/rpc"
	types "github.com/prysmaticlabs/eth2-types"
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	mock "github.com/prysmaticlabs/prysm/beacon-chain/blockchain/testing"
	"github.com/prysmaticlabs/prysm/beacon-chain/core/helpers"
	dbTest "github.com/prysmaticlabs/prysm/beacon-chain/db/testing"
	"github.com/prysmaticlabs/prysm/beacon-chain/rpc/beacon"
	"github.com/prysmaticlabs/prysm/beacon-chain/state/stategen"
	"github.com/prysmaticlabs/prysm/shared/params"
	"github.com/prysmaticlabs/prysm/shared/testutil"
	"github.com/prysmaticlabs/prysm/shared/testutil/require"
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
	"time"
)

type OrchestratorMock struct {
	consensusInfoChannel chan interface{}
}

func (api *OrchestratorMock) SyncNode() error {
	return nil
}


func TestAPIBackend_GetMinimalConsensusInfo(t *testing.T) {
	consensusChannel := make(chan interface{})
	listener, server, location := makeOrchestratorServer(t, consensusChannel)
	helpers.ClearCache()
	db := dbTest.SetupDB(t)
	ctx := context.Background()
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
		if recovery := recover(); recovery != nil {
			t.Log("Recovered in server stop", recovery)
		}
		server.Stop()
	}()

	require.Equal(t, location, listener.Addr().String())

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
		BeaconDB: db,
		FinalizationFetcher: &mock.ChainService{
			Genesis: time.Unix(genTime, 0),
			FinalizedCheckPoint: &ethpb.Checkpoint{
				Epoch: 0,
			},
		},
		GenesisTimeFetcher: &mock.ChainService{
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

	t.Run("Should notify about MinimalConsensusInfo", func(t *testing.T) {

	})

}

func makeOrchestratorServer(
	t *testing.T,
	consensusInfoChannel chan interface{},
) (listener net.Listener, server *rpc.Server, location string) {
	location = "./test.ipc"
	apis := make([]rpc.API, 0)
	api := &OrchestratorMock{consensusInfoChannel: consensusInfoChannel}

	apis = append(apis, rpc.API{
		Namespace: "orc",
		Version:   "1.0",
		Service:   api,
		Public:    true,
	})

	listener, server, err := rpc.StartIPCEndpoint(location, apis)
	require.NoError(t, err)

	return
}

