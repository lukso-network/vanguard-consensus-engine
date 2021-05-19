package beacon

import (
	"context"
	ptypes "github.com/gogo/protobuf/types"
	"github.com/golang/mock/gomock"
	types "github.com/prysmaticlabs/eth2-types"
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	chainMock "github.com/prysmaticlabs/prysm/beacon-chain/blockchain/testing"
	"github.com/prysmaticlabs/prysm/beacon-chain/core/helpers"
	"github.com/prysmaticlabs/prysm/beacon-chain/core/state"
	dbTest "github.com/prysmaticlabs/prysm/beacon-chain/db/testing"
	mockSync "github.com/prysmaticlabs/prysm/beacon-chain/sync/initial-sync/testing"
	"github.com/prysmaticlabs/prysm/shared/bytesutil"
	"github.com/prysmaticlabs/prysm/shared/params"
	"github.com/prysmaticlabs/prysm/shared/testutil"
	"github.com/prysmaticlabs/prysm/shared/testutil/assert"
	"github.com/prysmaticlabs/prysm/shared/testutil/require"
	vmock "github.com/prysmaticlabs/prysm/shared/van_mock"
	"testing"
	"time"
)

// TestServer_StreamMinimalConsensusInfo_ContextCanceled
func TestServer_StreamMinimalConsensusInfo_ContextCanceled(t *testing.T) {
	helpers.ClearCache()
	db := dbTest.SetupDB(t)

	genesis := testutil.NewBeaconBlock()
	depChainStart := params.BeaconConfig().MinGenesisActiveValidatorCount
	deposits, _, err := testutil.DeterministicDepositsAndKeys(depChainStart)
	require.NoError(t, err)
	eth1Data, err := testutil.DeterministicEth1Data(len(deposits))
	require.NoError(t, err)
	bs, err := state.GenesisBeaconState(deposits, 0, eth1Data)
	require.NoError(t, err, "Could not setup genesis bs")
	genesisRoot, err := genesis.Block.HashTreeRoot()
	require.NoError(t, err, "Could not get signing root")

	pubKeys := make([][]byte, len(deposits))
	indices := make([]uint64, len(deposits))
	for i := 0; i < len(deposits); i++ {
		pubKeys[i] = deposits[i].Data.PublicKey
		indices[i] = uint64(i)
	}

	pubkeysAs48ByteType := make([][48]byte, len(pubKeys))
	for i, pk := range pubKeys {
		pubkeysAs48ByteType[i] = bytesutil.ToBytes48(pk)
	}

	ctx, cancel := context.WithCancel(context.Background())
	c := &chainMock.ChainService{
		Genesis: time.Now(),
	}

	server := &Server{
		Ctx:                ctx,
		BeaconDB:           db,
		HeadFetcher:        &chainMock.ChainService{State: bs, Root: genesisRoot[:]},
		SyncChecker:        &mockSync.Sync{IsSyncing: false},
		GenesisTimeFetcher: c,
		StateNotifier:      &chainMock.MockStateNotifier{},
	}

	wantedRes := &ethpb.MinimalConsensusInfo{
		Epoch:            0,
		ValidatorList:    nil,
		EpochTimeStart:   0,
		SlotTimeDuration: nil,
	}

	// to check minimal output I must get it from blockchain
	// 1. move from blockchain to helpers or to rpc
	// 2. use like bs.ConsensusFetcher.MinimalConsino

	require.NoError(t, err)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	exitRoutine := make(chan bool)
	mockStream := vmock.NewMockBeaconChain_StreamMinimalConsensusInfoServer(ctrl)
	mockStream.EXPECT().Send(wantedRes).Do(func(arg0 interface{}) {
		exitRoutine <- true
	})
	mockStream.EXPECT().Context().Return(ctx).AnyTimes()
	epoch := types.Epoch(0)
	go func(tt *testing.T) {
		assert.ErrorContains(t, "context canceled", server.StreamMinimalConsensusInfo(&ptypes.Empty{}, &epoch, mockStream))
	}(t)
	<-exitRoutine
	cancel()
}
