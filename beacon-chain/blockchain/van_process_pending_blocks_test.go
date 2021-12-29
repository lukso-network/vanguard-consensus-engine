package blockchain

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	gethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/golang/mock/gomock"
	types "github.com/prysmaticlabs/eth2-types"
	mock "github.com/prysmaticlabs/prysm/beacon-chain/blockchain/testing"
	"github.com/prysmaticlabs/prysm/beacon-chain/core/blocks"
	testDB "github.com/prysmaticlabs/prysm/beacon-chain/db/testing"
	"github.com/prysmaticlabs/prysm/beacon-chain/forkchoice/protoarray"
	"github.com/prysmaticlabs/prysm/beacon-chain/state/stategen"
	eth "github.com/prysmaticlabs/prysm/proto/eth/v1alpha1"
	ethpb "github.com/prysmaticlabs/prysm/proto/eth/v1alpha1"
	"github.com/prysmaticlabs/prysm/proto/eth/v1alpha1/wrapper"
	"github.com/prysmaticlabs/prysm/proto/interfaces"
	"github.com/prysmaticlabs/prysm/shared/bls"
	vanTypes "github.com/prysmaticlabs/prysm/shared/params"
	"github.com/prysmaticlabs/prysm/shared/testutil"
	"github.com/prysmaticlabs/prysm/shared/testutil/assert"
	"github.com/prysmaticlabs/prysm/shared/testutil/require"
	"github.com/prysmaticlabs/prysm/shared/van_mock"
	"math/big"
	"sort"
	"testing"
	"time"
)

// TestService_PublishAndStorePendingBlock checks PublishAndStorePendingBlock method
func TestService_PublishBlock(t *testing.T) {
	ctx := context.Background()
	beaconDB := testDB.SetupDB(t)
	cfg := &Config{
		BeaconDB:      beaconDB,
		StateGen:      stategen.New(beaconDB),
		BlockNotifier: &mock.MockBlockNotifier{RecordEvents: true},
		StateNotifier: &mock.MockStateNotifier{RecordEvents: true},
	}
	s, err := NewService(ctx, cfg)
	require.NoError(t, err)
	genesisStateRoot := [32]byte{}
	genesis := blocks.NewGenesisBlock(genesisStateRoot[:])
	wrappedGenesisBlk := wrapper.WrappedPhase0SignedBeaconBlock(genesis)
	assert.NoError(t, beaconDB.SaveBlock(ctx, wrappedGenesisBlk))
	require.NoError(t, err)
	b := testutil.NewBeaconBlock()
	wrappedBlk := wrapper.WrappedPhase0SignedBeaconBlock(b)
	s.publishBlock(wrappedBlk)
	time.Sleep(3 * time.Second)
	if recvd := len(s.blockNotifier.(*mock.MockBlockNotifier).ReceivedEvents()); recvd < 1 {
		t.Errorf("Received %d pending block notifications, expected at least 1", recvd)
	}
}

// TestService_SortedUnConfirmedBlocksFromCache checks SortedUnConfirmedBlocksFromCache method
func TestService_SortedUnConfirmedBlocksFromCache(t *testing.T) {
	ctx := context.Background()
	s, err := NewService(ctx, &Config{})
	require.NoError(t, err)
	blks := make([]interfaces.BeaconBlock, 10)
	for i := 0; i < 10; i++ {
		b := testutil.NewBeaconBlock()
		b.Block.Slot = types.Slot(10 - i)
		wrappedBlk := wrapper.WrappedPhase0BeaconBlock(b.Block)
		blks[i] = wrappedBlk
		require.NoError(t, s.pendingBlockCache.AddPendingBlock(wrappedBlk))
	}
	sort.Slice(blks, func(i, j int) bool {
		return blks[i].Slot() < blks[j].Slot()
	})
	sortedBlocks, err := s.SortedUnConfirmedBlocksFromCache()
	require.NoError(t, err)
	require.DeepEqual(t, blks, sortedBlocks)
}

// TestService_fetchOrcConfirmations checks fetchOrcConfirmations
func TestService_fetchOrcConfirmations(t *testing.T) {
	ctx := context.Background()
	var mockedOrcClient *van_mock.MockClient
	ctrl := gomock.NewController(t)
	mockedOrcClient = van_mock.NewMockClient(ctrl)
	cfg := &Config{
		BlockNotifier:      &mock.MockBlockNotifier{RecordEvents: true},
		OrcRPCClient:       mockedOrcClient,
		EnableVanguardNode: true,
	}
	confirmationStatus := make([]*vanTypes.ConfirmationResData, 10)
	for i := 0; i < 10; i++ {
		confirmationStatus[i] = &vanTypes.ConfirmationResData{Slot: types.Slot(i), Status: vanTypes.Verified}
	}
	mockedOrcClient.EXPECT().ConfirmVanBlockHashes(
		gomock.Any(),
		gomock.Any(),
	).AnyTimes().Return(confirmationStatus, nil)
	s, err := NewService(ctx, cfg)
	go s.processOrcConfirmationRoutine()
	require.NoError(t, err)
	blks := make([]interfaces.BeaconBlock, 10)
	for i := 0; i < 10; i++ {
		b := testutil.NewBeaconBlock()
		b.Block.Slot = types.Slot(i)
		wrappedBlk := wrapper.WrappedPhase0BeaconBlock(b.Block)
		blks[i] = wrappedBlk
		confirmationStatus[i] = &vanTypes.ConfirmationResData{Slot: types.Slot(i), Status: vanTypes.Verified}
		require.NoError(t, s.pendingBlockCache.AddPendingBlock(wrappedBlk))
	}
}

func TestService_VerifyPandoraShardConsecutiveness(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	mockedOrcClient := van_mock.NewMockClient(ctrl)
	beaconDB := testDB.SetupDB(t)
	stateGen := stategen.New(beaconDB)
	cfg := &Config{
		BlockNotifier:      &mock.MockBlockNotifier{RecordEvents: true},
		OrcRPCClient:       mockedOrcClient,
		EnableVanguardNode: true,
		ForkChoiceStore:    protoarray.New(0, 0, [32]byte{}),
		BeaconDB:           beaconDB,
		StateGen:           stateGen,
	}
	s, err := NewService(ctx, cfg)

	require.NoError(t, err)
	require.NotNil(t, s)

	t.Run("should throw an error when signed block is empty", func(t *testing.T) {
		currentErr := s.VerifyPandoraShardConsecutiveness(nil, nil)
		require.NotNil(t, currentErr)
		require.ErrorContains(t, errInvalidPandoraShardInfo.Error(), currentErr)
		require.ErrorContains(t, "signed block is nil", currentErr)
	})

	t.Run("should throw an error when parent block is empty", func(t *testing.T) {
		signedBlock := &eth.SignedBeaconBlock{Block: &eth.BeaconBlock{}}
		currentErr := s.VerifyPandoraShardConsecutiveness(nil, signedBlock)
		require.DeepEqual(t, errInvalidBeaconBlock, currentErr)
	})

	t.Run("should throw an error when signed block is without sharding part", func(t *testing.T) {
		signedBlock := &eth.SignedBeaconBlock{Block: &eth.BeaconBlock{}}
		parentBlock := &eth.SignedBeaconBlock{Block: &eth.BeaconBlock{}}
		currentErr := s.VerifyPandoraShardConsecutiveness(parentBlock, signedBlock)
		require.NotNil(t, currentErr)
		require.ErrorContains(t, errInvalidPandoraShardInfo.Error(), currentErr)
		require.ErrorContains(t, "empty signed Pandora shards", currentErr)
	})

	t.Run("should return an error when signed shard is nil", func(t *testing.T) {
		wrappedBlock := wrapper.WrappedPhase0SignedBeaconBlock(testutil.NewBeaconBlock())
		parentBlock, currentErr := wrappedBlock.PbPhase0Block()
		require.NoError(t, currentErr)
		pandoraShards := make([]*eth.PandoraShard, 1)
		hashTreeRoot, currentErr := parentBlock.HashTreeRoot()
		signedBlock := &eth.SignedBeaconBlock{Block: &eth.BeaconBlock{
			ParentRoot: hashTreeRoot[:],
			Body:       &eth.BeaconBlockBody{PandoraShard: pandoraShards},
		}}
		require.NoError(t, currentErr)
		currentErr = s.VerifyPandoraShardConsecutiveness(parentBlock, signedBlock)
		require.ErrorContains(t, "signed shard cannot be nil", currentErr)
		require.ErrorContains(t, errInvalidPandoraShardInfo.Error(), currentErr)
	})

	st, keys := testutil.DeterministicGenesisState(t, 64)
	genesisStateRoot := [32]byte{}
	genesis := blocks.NewGenesisBlock(genesisStateRoot[:])
	assert.NoError(t, beaconDB.SaveBlock(ctx, wrapper.WrappedPhase0SignedBeaconBlock(genesis)))
	gRoot, err := genesis.Block.HashTreeRoot()
	require.NoError(t, err)
	s.finalizedCheckpt = &ethpb.Checkpoint{
		Root: gRoot[:],
	}
	s.genesisRoot = gRoot
	s.cfg.ForkChoiceStore = protoarray.New(0, 0, [32]byte{})
	s.saveInitSyncBlock(gRoot, wrapper.WrappedPhase0SignedBeaconBlock(genesis))
	require.NoError(t, s.cfg.BeaconDB.SaveBlock(ctx, wrapper.WrappedPhase0SignedBeaconBlock(genesis)))
	require.NoError(t, s.cfg.BeaconDB.SaveState(ctx, st.Copy(), genesisStateRoot))
	require.NoError(t, s.cfg.BeaconDB.SaveGenesisData(ctx, st.Copy()))

	t.Run("should return an error when there are no pandora shards", func(t *testing.T) {
		wrappedBlock := wrapper.WrappedPhase0SignedBeaconBlock(testutil.NewBeaconBlock())
		parentBlock, currentErr := wrappedBlock.PbPhase0Block()
		require.NoError(t, currentErr)
		signedBlock := &eth.SignedBeaconBlock{Block: &eth.BeaconBlock{
			ParentRoot: gRoot[:],
			Body:       &eth.BeaconBlockBody{},
		}}
		currentErr = s.VerifyPandoraShardConsecutiveness(parentBlock, signedBlock)

		require.ErrorContains(t, "empty signed Pandora shards", currentErr)
	})

	pandoraShards0Block0Hash := common.HexToHash("0xabcd").Bytes()
	pandoraShards0Block1Hash := common.HexToHash("0xabc").Bytes()

	// TODO: this test suite can be refactored a bit to not rely on state of the service, but to create new service for each case
	// Order of the tests matter from that point from logical point of view (state).
	t.Run("should return nil if there is a genesis root match at parent and signature is present", func(t *testing.T) {
		pandoraShards := make([]*eth.PandoraShard, 1)
		pandoraShards[0] = &ethpb.PandoraShard{
			BlockNumber: 1,
			Hash:        pandoraShards0Block1Hash,
			ParentHash:  pandoraShards0Block0Hash,
			StateRoot:   common.HexToHash("0xabcde").Bytes(),
			TxHash:      common.HexToHash("0xabcdf").Bytes(),
			ReceiptHash: common.HexToHash("0xabcda").Bytes(),
			SealHash:    common.HexToHash("0xabcdw").Bytes(),
			Signature:   nil,
		}

		signature := keys[0].Sign(pandoraShards[0].SealHash)
		pandoraShards[0].Signature = signature.Marshal()

		signedBlock := &eth.SignedBeaconBlock{Block: &eth.BeaconBlock{
			ParentRoot: gRoot[:],
			Body:       &eth.BeaconBlockBody{PandoraShard: pandoraShards},
		}}
		parentBlock, currentErr := s.cfg.BeaconDB.GenesisBlock(ctx)
		require.NoError(t, currentErr)
		signedParentBlock0, currentErr := parentBlock.PbPhase0Block()
		require.NoError(t, currentErr)

		currentErr = s.VerifyPandoraShardConsecutiveness(
			signedParentBlock0,
			signedBlock,
		)

		require.NoError(t, currentErr)

		pandoraState := st.Copy()

		fullBlock0, currentErr := testutil.GenerateFullBlock(
			pandoraState,
			keys,
			testutil.DefaultBlockGenConfig(),
			0,
			pandoraShards...,
		)

		require.NoError(t, currentErr)

		signedBlock1HashTreeRoot, currentErr := fullBlock0.HashTreeRoot()
		require.NoError(t, currentErr)

		s.saveInitSyncBlock(gRoot, wrapper.WrappedPhase0SignedBeaconBlock(fullBlock0))
		require.NoError(t, s.cfg.BeaconDB.SaveBlock(ctx, wrapper.WrappedPhase0SignedBeaconBlock(fullBlock0)))
		require.NoError(t, s.cfg.BeaconDB.SaveState(ctx, pandoraState.Copy(), signedBlock1HashTreeRoot))

		t.Run("should throw an error when pandora parent does not match and is nonconsecutive", func(t *testing.T) {
			signedBlock2 := &eth.SignedBeaconBlock{Block: &eth.BeaconBlock{
				ParentRoot: signedBlock1HashTreeRoot[:],
				Body:       &eth.BeaconBlockBody{PandoraShard: pandoraShards},
			}}

			pandoraShardErr := s.VerifyPandoraShardConsecutiveness(
				signedBlock,
				signedBlock2,
			)

			require.ErrorContains(t, errNonConsecutivePandoraShardInfo.Error(), pandoraShardErr)
			require.ErrorContains(t, common.Bytes2Hex(pandoraShards[0].GetHash()), pandoraShardErr)
		})

		// TODO: WORTH TO CONSIDER TO CHECK IF PANDORA HASH WAS PREVIOUSLY IN WHOLE BRANCH TREE (heavy)
		t.Run("should pass for next block when consecutiveness is present", func(t *testing.T) {
			signedBlock2PandoraShards := make([]*ethpb.PandoraShard, 1)

			signedBlock2PandoraShards[0] = &ethpb.PandoraShard{
				BlockNumber: 2,
				Hash:        common.HexToHash("0xfacd1234").Bytes(),
				ParentHash:  pandoraShards0Block1Hash,
				StateRoot:   common.HexToHash("0xabcde").Bytes(),
				TxHash:      common.HexToHash("0xabcdf").Bytes(),
				ReceiptHash: common.HexToHash("0xabcda").Bytes(),
				SealHash:    common.HexToHash("0xabcdw").Bytes(),
				Signature:   nil,
			}

			pandoraBlock2Signature := keys[0].Sign(signedBlock2PandoraShards[0].SealHash)
			signedBlock2PandoraShards[0].Signature = pandoraBlock2Signature.Marshal()

			signedBlock2 := &eth.SignedBeaconBlock{Block: &eth.BeaconBlock{
				ParentRoot: signedBlock1HashTreeRoot[:],
				Body:       &eth.BeaconBlockBody{PandoraShard: signedBlock2PandoraShards},
			}}

			pandoraShardErr := s.VerifyPandoraShardConsecutiveness(
				signedBlock,
				signedBlock2,
			)

			require.NoError(t, pandoraShardErr)
		})

		// This case is edgy, but Hash is not a function but constant. Somebody might mistakenly pass this invalid
		// state during implementation in consumers of this code. It will prevent that lack of precision
		t.Run("should not pass for next block when hash is the same as parent hash", func(t *testing.T) {
			signedBlock2PandoraShards := make([]*ethpb.PandoraShard, 1)
			signedBlock2PandoraShards[0] = &ethpb.PandoraShard{
				BlockNumber: 2,
				Hash:        common.HexToHash("0xabc").Bytes(),
				ParentHash:  common.HexToHash("0xabc").Bytes(),
				StateRoot:   common.HexToHash("0xabcde").Bytes(),
				TxHash:      common.HexToHash("0xabcdf").Bytes(),
				ReceiptHash: common.HexToHash("0xabcda").Bytes(),
				SealHash:    common.HexToHash("0xabcdw").Bytes(),
				Signature:   nil,
			}

			pandoraBlock2Signature := keys[0].Sign(signedBlock2PandoraShards[0].SealHash)
			signedBlock2PandoraShards[0].Signature = pandoraBlock2Signature.Marshal()

			signedBlock2 := &eth.SignedBeaconBlock{Block: &eth.BeaconBlock{
				ParentRoot: signedBlock1HashTreeRoot[:],
				Body:       &eth.BeaconBlockBody{PandoraShard: signedBlock2PandoraShards},
			}}

			pandoraShardErr := s.VerifyPandoraShardConsecutiveness(
				signedBlock,
				signedBlock2,
			)

			require.ErrorContains(t, errInvalidPandoraShardInfo.Error(), pandoraShardErr)
		})
	})
}

func TestGuardPandoraShardHeader(t *testing.T) {
	pandoraBlock := &ethpb.PandoraShard{}

	t.Run("should throw an error when hash is empty", func(t *testing.T) {
		require.Equal(t, errInvalidPandoraShardInfo, GuardPandoraShard(pandoraBlock))
	})

	pandoraBlock.Hash = []byte("0xde6f0b6c17077334abd585da38b251871251cb26fa3456be135825ea45c06f12")

	t.Run("should throw an error when parent hash is empty", func(t *testing.T) {
		require.Equal(t, errInvalidPandoraShardInfo, GuardPandoraShard(pandoraBlock))
	})

	pandoraBlock.ParentHash = pandoraBlock.Hash

	t.Run("should throw an error when parent hash is equal hash", func(t *testing.T) {
		require.Equal(t, errInvalidPandoraShardInfo, GuardPandoraShard(pandoraBlock))
	})

	pandoraBlock.ParentHash = []byte("0x67b96c7bbdbf2186c868ac7565a24d250c8ecbf4f43cb50bd78f11b73681c025")

	t.Run("should pass when parent hash and hash is not empty", func(t *testing.T) {
		require.NoError(t, GuardPandoraShard(pandoraBlock))
	})
}

func TestGuardPandoraConsecutiveness(t *testing.T) {
	t.Run("should return err when parent hash does not match", func(t *testing.T) {
		require.Equal(
			t,
			errNonConsecutivePandoraShardInfo,
			GuardPandoraConsecutiveness(
				common.HexToHash("0xa"),
				common.HexToHash("0xb"),
				2,
				8,
			),
		)
	})

	t.Run("should return err when block number is nonconsecutive", func(t *testing.T) {
		require.Equal(
			t,
			errNonConsecutivePandoraShardInfo,
			GuardPandoraConsecutiveness(
				common.HexToHash("0xa"),
				common.HexToHash("0xa"),
				2,
				8,
			),
		)
	})

	t.Run("should pass when consecutiveness is present", func(t *testing.T) {
		require.NoError(
			t,
			GuardPandoraConsecutiveness(
				common.HexToHash("0xa"),
				common.HexToHash("0xa"),
				3,
				2,
			),
		)
	})
}

func Test_GuardPandoraShardSignature(t *testing.T) {
	randKey, err := bls.RandKey()
	require.NoError(t, err)

	t.Run("should return error when shard is nil", func(t *testing.T) {
		currentErr := GuardPandoraShardSignature(nil, nil)
		require.ErrorContains(t, "pandora shard is missing", currentErr)
		require.ErrorContains(t, errInvalidPandoraShardInfo.Error(), currentErr)
	})

	t.Run("should return error because of lack of bls public key", func(t *testing.T) {
		currentErr := GuardPandoraShardSignature(&ethpb.PandoraShard{}, nil)
		require.ErrorContains(t, "bls signature is missing", currentErr)
		require.ErrorContains(t, errInvalidBlsSignature.Error(), currentErr)
	})

	t.Run("should return error because of lack of seal hash", func(t *testing.T) {
		currentErr := GuardPandoraShardSignature(&ethpb.PandoraShard{}, randKey.PublicKey())
		require.ErrorContains(t, "seal hash is invalid", currentErr)
		require.ErrorContains(t, errInvalidPandoraShardInfo.Error(), currentErr)
	})

	t.Run("should return error because of lack of signature", func(t *testing.T) {
		currentErr := GuardPandoraShardSignature(
			&ethpb.PandoraShard{SealHash: common.BytesToHash([]byte{}).Bytes()},
			randKey.PublicKey(),
		)
		require.ErrorContains(t, "pandora shard signature is invalid", currentErr)
		require.ErrorContains(t, errInvalidBlsSignature.Error(), currentErr)

		currentErr = GuardPandoraShardSignature(
			&ethpb.PandoraShard{
				SealHash:  common.BytesToHash([]byte{}).Bytes(),
				Signature: common.BytesToHash([]byte{}).Bytes(),
			},
			randKey.PublicKey(),
		)
		require.ErrorContains(t, "pandora shard signature is invalid", currentErr)
		require.ErrorContains(t, errInvalidBlsSignature.Error(), currentErr)
	})

	t.Run("should return error because of invalid signature", func(t *testing.T) {
		signatureBytes := make([]byte, 96)
		currentErr := GuardPandoraShardSignature(
			&ethpb.PandoraShard{
				SealHash:  common.BytesToHash([]byte{}).Bytes(),
				Signature: signatureBytes,
			},
			randKey.PublicKey(),
		)
		require.ErrorContains(t, "could not create bls signature", currentErr)
		require.ErrorContains(t, "could not unmarshal bytes into signature", currentErr)
	})

	pandoraHeader := &gethTypes.Header{
		ParentHash:  common.Hash{},
		UncleHash:   common.Hash{},
		Coinbase:    common.Address{},
		Root:        common.Hash{},
		TxHash:      common.Hash{},
		ReceiptHash: common.Hash{},
		Bloom:       gethTypes.Bloom{},
		Difficulty:  big.NewInt(0),
		Number:      big.NewInt(1),
		GasLimit:    0,
		GasUsed:     0,
		Time:        0,
		Extra:       nil,
		MixDigest:   common.Hash{},
		Nonce:       gethTypes.BlockNonce{},
		BaseFee:     nil,
	}

	pandoraExtraData := testutil.ExtraData{Slot: 56, Epoch: 2}
	pandoraHeader.Extra, err = rlp.EncodeToBytes(pandoraExtraData)
	require.NoError(t, err)
	pandoraSealHash := testutil.SealHash(pandoraHeader)
	validPandoraSignature := randKey.Sign(pandoraSealHash.Bytes())

	t.Run("should not pass if signature is not verified", func(t *testing.T) {
		invalidKey, currentErr := bls.RandKey()
		require.NoError(t, currentErr)
		currentErr = GuardPandoraShardSignature(
			&ethpb.PandoraShard{
				SealHash:  pandoraSealHash.Bytes(),
				Signature: validPandoraSignature.Marshal(),
			},
			invalidKey.PublicKey(),
		)

		require.ErrorContains(t, "signature of shard did not verify", currentErr)
		require.ErrorContains(t, "invalid bls signature", currentErr)
	})

	t.Run("should pass if signature is verified", func(t *testing.T) {
		require.NoError(t, GuardPandoraShardSignature(
			&ethpb.PandoraShard{
				SealHash:  pandoraSealHash.Bytes(),
				Signature: validPandoraSignature.Marshal(),
			},
			randKey.PublicKey(),
		))
	})
}

// TestService_waitForConfirmationBlock checks waitForConfirmationBlock method
// When the confirmation result of the block is verified then waitForConfirmationBlock gives you error return
// Not delete the invalid block because, when node gets an valid block, then it will be replaced and then it will be deleted
func TestService_waitForConfirmationBlock(t *testing.T) {
	tests := []struct {
		name                 string
		pendingBlocksInQueue []interfaces.SignedBeaconBlock
		incomingBlock        interfaces.SignedBeaconBlock
		confirmationStatus   []*vanTypes.ConfirmationResData
		expectedOutput       string
	}{
		{
			name:                 "Returns nil when orchestrator sends verified status for all blocks",
			pendingBlocksInQueue: getBeaconBlocks(0, 3),
			incomingBlock:        getBeaconBlock(2),
			confirmationStatus: []*vanTypes.ConfirmationResData{
				{
					Slot:   0,
					Status: vanTypes.Verified,
				},
				{
					Slot:   1,
					Status: vanTypes.Verified,
				},
				{
					Slot:   2,
					Status: vanTypes.Verified,
				},
			},
			expectedOutput: "",
		},
		{
			name:                 "Returns error when orchestrator sends invalid status",
			pendingBlocksInQueue: getBeaconBlocks(0, 3),
			incomingBlock:        getBeaconBlock(1),
			confirmationStatus: []*vanTypes.ConfirmationResData{
				{
					Slot:   0,
					Status: vanTypes.Verified,
				},
				{
					Slot:   1,
					Status: vanTypes.Invalid,
				},
				{
					Slot:   2,
					Status: vanTypes.Verified,
				},
			},
			expectedOutput: "invalid block found in orchestrator",
		},
		{
			name:                 "Retry for the block with pending status",
			pendingBlocksInQueue: getBeaconBlocks(0, 3),
			incomingBlock:        getBeaconBlock(1),
			confirmationStatus: []*vanTypes.ConfirmationResData{
				{
					Slot:   0,
					Status: vanTypes.Verified,
				},
				{
					Slot:   1,
					Status: vanTypes.Pending,
				},
				{
					Slot:   2,
					Status: vanTypes.Verified,
				},
			},
			expectedOutput: "maximum wait is exceeded and orchestrator can not verify the block",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			var mockedOrcClient *van_mock.MockClient
			ctrl := gomock.NewController(t)
			mockedOrcClient = van_mock.NewMockClient(ctrl)
			cfg := &Config{
				BlockNotifier:      &mock.MockBlockNotifier{},
				OrcRPCClient:       mockedOrcClient,
				EnableVanguardNode: true,
			}
			s, err := NewService(ctx, cfg)
			require.NoError(t, err)
			go s.processOrcConfirmationRoutine()
			mockedOrcClient.EXPECT().ConfirmVanBlockHashes(
				gomock.Any(),
				gomock.Any(),
			).AnyTimes().Return(tt.confirmationStatus, nil)
			for i := 0; i < len(tt.pendingBlocksInQueue); i++ {
				require.NoError(t, s.pendingBlockCache.AddPendingBlock(tt.pendingBlocksInQueue[i].Block()))
			}
			if tt.expectedOutput == "" {
				require.NoError(t, s.waitForConfirmationBlock(ctx, tt.incomingBlock))
			} else {
				require.ErrorContains(t, tt.expectedOutput, s.waitForConfirmationBlock(ctx, tt.incomingBlock))
			}
		})
	}
}

// Helper method to generate pending queue with random blocks
func getBeaconBlocks(from, to int) []interfaces.SignedBeaconBlock {
	pendingBlks := make([]interfaces.SignedBeaconBlock, to-from)
	for i := 0; i < to-from; i++ {
		b := testutil.NewBeaconBlock()
		b.Block.Slot = types.Slot(from + i)
		wrappedBlk := wrapper.WrappedPhase0SignedBeaconBlock(b)
		pendingBlks[i] = wrappedBlk
	}
	return pendingBlks
}

// Helper method to generate pending queue with random block
func getBeaconBlock(slot types.Slot) interfaces.SignedBeaconBlock {
	b := testutil.NewBeaconBlock()
	b.Block.Slot = types.Slot(slot)
	return wrapper.WrappedPhase0SignedBeaconBlock(b)
}
