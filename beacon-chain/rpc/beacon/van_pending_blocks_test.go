package beacon

//func TestServer_StreamNewPendingBlocks_ContextCanceled(t *testing.T) {
//	db := dbTest.SetupDB(t)
//	ctx := context.Background()
//
//	ctx, cancel := context.WithCancel(ctx)
//	chainService := &chainMock.ChainService{}
//	server := &Server{
//		Ctx:           ctx,
//		StateNotifier: chainService.StateNotifier(),
//		BeaconDB:      db,
//	}
//
//	exitRoutine := make(chan bool)
//	ctrl := gomock.NewController(t)
//	defer ctrl.Finish()
//	mockStream := mock.NewMockBeaconChain_StreamChainHeadServer(ctrl)
//	mockStream.EXPECT().Context().Return(ctx)
//	go func(tt *testing.T) {
//		assert.ErrorContains(tt, "Context canceled", server.StreamChainHead(&ptypes.Empty{}, mockStream))
//		<-exitRoutine
//	}(t)
//	cancel()
//	exitRoutine <- true
//}
