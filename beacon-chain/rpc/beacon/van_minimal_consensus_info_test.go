package beacon

import (
	"context"
	"github.com/golang/mock/gomock"
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	chainMock "github.com/prysmaticlabs/prysm/beacon-chain/blockchain/testing"
	dbTest "github.com/prysmaticlabs/prysm/beacon-chain/db/testing"
	"github.com/prysmaticlabs/prysm/beacon-chain/state/stategen"
	"github.com/prysmaticlabs/prysm/shared/mock"
	"github.com/prysmaticlabs/prysm/shared/testutil"
	"github.com/prysmaticlabs/prysm/shared/testutil/assert"
	"github.com/prysmaticlabs/prysm/shared/testutil/require"
	"testing"
)

func TestServer_StreamMinimalConsensusInfo_ContextCanceled(t *testing.T) {
	db := dbTest.SetupDB(t)
	ctx := context.Background()

	beaconState, privs := testutil.DeterministicGenesisState(t, 32)
	b, err := testutil.GenerateFullBlock(beaconState, privs, testutil.DefaultBlockGenConfig(), 1)
	require.NoError(t, err)
	require.NoError(t, db.SaveBlock(ctx, b))
	ctx, cancel := context.WithCancel(ctx)
	chainService := &chainMock.ChainService{}

	server := &Server{
		Ctx:           ctx,
		StateNotifier: chainService.StateNotifier(),
		BeaconDB:      db,
		StateGen:      stategen.New(db),
	}

	exitRoutine := make(chan bool)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockStream := mock.NewMockBeaconChain_StreamMinimalConsensusInfoServer(ctrl)
	mockStream.EXPECT().Context().Return(ctx)

	go func(tt *testing.T) {
		assert.ErrorContains(tt, "Context canceled", server.StreamMinimalConsensusInfo(&ethpb.MinimalConsensusInfoRequest{
			FromEpoch: 0,
		}, mockStream))
		<-exitRoutine
	}(t)

	cancel()
	exitRoutine <- true
}
