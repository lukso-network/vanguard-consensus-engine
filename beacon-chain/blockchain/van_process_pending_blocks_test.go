package blockchain

// TestService_PublishAndStorePendingBlock
//func TestService_PublishAndStorePendingBlock(t *testing.T) {
//	ctx := context.Background()
//	cfg := &Config{
//		BlockNotifier: &blockchainTesting.MockBlockNotifier{},
//	}
//	s, err := NewService(ctx, cfg)
//	require.NoError(t, err)
//
//	b := testutil.NewBeaconBlock()
//	require.NoError(t, s.publishAndStorePendingBlock(ctx, b.Block))
//	cachedBlock, err := s.pendingBlockCache.PendingBlock(b.Block.GetSlot())
//	require.NoError(t, err)
//	assert.DeepEqual(t, b, cachedBlock)
//}
