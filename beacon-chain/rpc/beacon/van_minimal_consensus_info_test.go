package beacon

import (
	"context"
	ptypes "github.com/gogo/protobuf/types"
	"github.com/golang/mock/gomock"
	types "github.com/prysmaticlabs/eth2-types"
	chainMock "github.com/prysmaticlabs/prysm/beacon-chain/blockchain/testing"
	"github.com/prysmaticlabs/prysm/beacon-chain/core/feed"
	consensusfeed "github.com/prysmaticlabs/prysm/beacon-chain/core/feed/consensus"
	"github.com/prysmaticlabs/prysm/beacon-chain/core/helpers"
	dbTest "github.com/prysmaticlabs/prysm/beacon-chain/db/testing"
	"github.com/prysmaticlabs/prysm/shared/testutil"
	"github.com/prysmaticlabs/prysm/shared/testutil/assert"
	"github.com/prysmaticlabs/prysm/shared/van_mock"
	"testing"
)

// TestServer_StreamMinimalConsensusInfo_ContextCanceled
func TestServer_StreamMinimalConsensusInfo_ContextCanceled(t *testing.T) {
	helpers.ClearCache()
	db := dbTest.SetupDB(t)
	ctx := context.Background()

	ctx, cancel := context.WithCancel(ctx)
	chainService := &chainMock.ChainService{}
	server := &Server{
		Ctx:                     ctx,
		ConsensusNotifier:       chainMock.MockConsensusNotifier{},
		BeaconDB:                db,
		UnconfirmedBlockFetcher: chainService,
	}

	exitRoutine := make(chan bool)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockStream := mock.NewMockBeaconChain_StreamMinimalConsensusInfoServer(ctrl)
	mockStream.EXPECT().Context().Return(ctx)
	go func(tt *testing.T) {
		assert.ErrorContains(tt, "Context canceled", server.StreamMinimalConsensusInfo(&ptypes.Empty{}, mockStream))
		<-exitRoutine
	}(t)
	cancel()
	exitRoutine <- true
}

// TestServer_StreamMinimalConsensusInfo_OnNewConsensusInfo checks the pipeline of the new epoch
func TestServer_StreamMinimalConsensusInfo_OnNewConsensusInfo(t *testing.T) {
	helpers.ClearCache()
	ctx := context.Background()
	beaconState, _ := testutil.DeterministicGenesisState(t, 32)
	chainService := &chainMock.ChainService{State: beaconState}
	server := &Server{
		Ctx:                     ctx,
		ConsensusNotifier:       chainMock.MockConsensusNotifier{},
		HeadFetcher:             chainService,
		UnconfirmedBlockFetcher: chainService,
	}
	minConsensusInfo, err := server.MinimalConsensusInfoFetcher.MinimalConsensusInfo(types.Epoch(0))
	if err != nil {
		t.Fail()
	}
	exitRoutine := make(chan bool)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockStream := mock.NewMockBeaconChain_StreamMinimalConsensusInfoServer(ctrl)
	mockStream.EXPECT().Send(minConsensusInfo).Do(func(arg0 interface{}) {
		exitRoutine <- true
	})
	mockStream.EXPECT().Context().Return(ctx).AnyTimes()

	go func(tt *testing.T) {
		assert.NoError(tt, server.StreamMinimalConsensusInfo(&ptypes.Empty{}, mockStream), "Could not call RPC method")
	}(t)

	// Send in a loop to ensure it is delivered (busy wait for the service to subscribe to the state feed).
	for sent := 0; sent == 0; {
		sent = server.ConsensusNotifier.ConsensusFeed().Send(&feed.Event{
			Data: &consensusfeed.MinimalConsensusData{
				MinimalConsensusInfo: minConsensusInfo,
			},
		})
	}
	<-exitRoutine
}
