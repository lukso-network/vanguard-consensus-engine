package client

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	eth1Types "github.com/ethereum/go-ethereum/core/types"
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	types "github.com/prysmaticlabs/eth2-types"
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"github.com/prysmaticlabs/prysm/shared/bls"
	"github.com/prysmaticlabs/prysm/shared/bytesutil"
	"github.com/prysmaticlabs/prysm/shared/testutil"
	"github.com/prysmaticlabs/prysm/shared/testutil/require"
	logTest "github.com/sirupsen/logrus/hooks/test"
	"testing"
)

// TestVerifyPandoraHeader_Ok method checks pandora header validation method
func TestVerifyPandoraShardHeader(t *testing.T) {
	validator, _, _, finish := setup(t)
	defer finish()

	blk := testutil.NewBeaconBlock()
	blk.Block.Slot = 98
	blk.Block.ProposerIndex = 23
	epoch := types.Epoch(uint64(blk.Block.Slot) / 32)

	header, extraData := testutil.NewPandoraBlock(blk.Block.Slot, uint64(blk.Block.ProposerIndex))
	headerHash := sealHash(header)

	// Checks all the validations
	err := validator.verifyPandoraShardHeader(blk.Block.Slot, epoch, header, headerHash, extraData)
	require.NoError(t, err, "Should pass without any error")

	// Should get an `errInvalidHeaderHash` error
	header.Time = uint64(14265167)
	want := "invalid header hash"
	err = validator.verifyPandoraShardHeader(blk.Block.Slot, epoch, header, headerHash, extraData)
	require.ErrorContains(t, want, err, "Should get an errInvalidHeaderHash error")

	// Should get an `errInvalidSlot` error
	header.Time = uint64(1426516743)
	blk.Block.Slot = 90
	want = "invalid slot"
	err = validator.verifyPandoraShardHeader(blk.Block.Slot, epoch, header, headerHash, extraData)
	require.ErrorContains(t, want, err, "Should get an errInvalidSlot error")

	// Should get an `errInvalidEpoch` error
	blk.Block.Slot = 98
	epoch = 2
	want = "invalid epoch"
	err = validator.verifyPandoraShardHeader(blk.Block.Slot, epoch, header, headerHash, extraData)
	require.ErrorContains(t, want, err, "Should get an errInvalidEpoch error")
}

// TestProcessPandoraShardHeader method checks the `processPandoraShardHeader`
func TestProcessPandoraShardHeader(t *testing.T) {
	validator, m, _, finish := setup(t)
	defer finish()

	secretKey, err := bls.SecretKeyFromBytes(bytesutil.PadTo([]byte{1}, 32))
	require.NoError(t, err, "Failed to generate key from bytes")
	publicKey := secretKey.PublicKey()
	var pubKey [48]byte
	copy(pubKey[:], publicKey.Marshal())
	km := &mockKeymanager{
		keysMap: map[[48]byte]bls.SecretKey{
			pubKey: secretKey,
		},
	}
	validator.keyManager = km
	blk := testutil.NewBeaconBlock()
	blk.Block.Slot = 98
	blk.Block.ProposerIndex = 23
	epoch := types.Epoch(uint64(blk.Block.Slot) / 32)

	// Check with happy path
	header, extraData := testutil.NewPandoraBlock(blk.Block.Slot, uint64(blk.Block.ProposerIndex))
	headerHash := sealHash(header)

	m.pandoraService.EXPECT().GetShardBlockHeader(
		gomock.Any(), // ctx
	).Return(header, headerHash, extraData, nil) // nil - error

	m.pandoraService.EXPECT().SubmitShardBlockHeader(
		gomock.Any(), // ctx
		gomock.Any(), // blockNonce
		gomock.Any(), // headerHash
		gomock.Any(), // sig
	).Return(true, nil)

	status, err := validator.processPandoraShardHeader(context.Background(), blk.Block, blk.Block.Slot, epoch, pubKey)
	require.NoError(t, err, "Should sucess")
	require.Equal(t, true, status, "Pandora shard processing should be successful")

	// Return rlp decoding error when calls `GetWork` api
	ErrRlpDecoding := errors.New("rlp: input contains more than one value")
	m.pandoraService.EXPECT().GetShardBlockHeader(
		gomock.Any(), // ctx
	).Return(nil, common.Hash{}, nil, ErrRlpDecoding)
	_, err = validator.processPandoraShardHeader(context.Background(), blk.Block, blk.Block.Slot, epoch, pubKey)
	require.ErrorContains(t, "rlp: input contains more than one value", ErrRlpDecoding)
}

// TestValidator_ProposeBlock_Failed_WhenSubmitShardInfoFails methods checks when `SubmitShardInfo` fails
func TestValidator_ProposeBlock_Failed_WhenSubmitShardInfoFails(t *testing.T) {
	hook := logTest.NewGlobal()
	validator, m, validatorKey, finish := setup(t)
	defer finish()
	pubKey := [48]byte{}
	copy(pubKey[:], validatorKey.PublicKey().Marshal())

	m.validatorClient.EXPECT().DomainData(
		gomock.Any(), // ctx
		gomock.Any(), // epoch
	).Times(2).Return(&ethpb.DomainResponse{SignatureDomain: make([]byte, 32)}, nil /*err*/)

	m.validatorClient.EXPECT().GetBlock(
		gomock.Any(), // ctx
		gomock.Any(),
	).Times(2).Return(testutil.NewBeaconBlock().Block, nil /*err*/)

	header, extraData := testutil.NewPandoraBlock(types.Slot(1), 0)
	headerHash := sealHash(header)

	m.pandoraService.EXPECT().GetShardBlockHeader(
		gomock.Any(), // ctx
	).Return(header, headerHash, extraData, nil) // nil - error

	header, extraData = testutil.NewPandoraBlock(types.Slot(1), 0)
	headerHash = sealHash(header)

	m.pandoraService.EXPECT().GetShardBlockHeader(
		gomock.Any(), // ctx
	).Return(header, headerHash, extraData, nil) // nil - error

	// When `SubmitShardInfo` api returns false status
	m.pandoraService.EXPECT().SubmitShardBlockHeader(
		gomock.Any(), // ctx
		gomock.Any(), // blockNonce
		gomock.Any(), // headerHash
		gomock.Any(), // sig
	).Return(false, nil)

	validator.ProposeBlock(context.Background(), 1, pubKey)
	require.LogsContain(t, hook, "Failed to process pandora chain shard header")

	// When `SubmitShardInfo` api returns error
	m.pandoraService.EXPECT().SubmitShardBlockHeader(
		gomock.Any(), // ctx
		gomock.Any(), // blockNonce
		gomock.Any(), // headerHash
		gomock.Any(), // sig
	).Return(true, errors.New("Failed to submit shard header"))

	validator.ProposeBlock(context.Background(), 1, pubKey)
	require.LogsContain(t, hook, "Failed to process pandora chain shard header")
}

// TestValidator_panShardingCanonicalInfo
func TestValidator_panShardingCanonicalInfo(t *testing.T) {
	ctx := context.Background()
	validator, m, validatorKey, finish := setup(t)
	defer finish()
	pubKey := [48]byte{}
	copy(pubKey[:], validatorKey.PublicKey().Marshal())

	panBlockNum := uint64(100)
	slot := types.Slot(35)
	beaconBlock := testutil.NewBeaconBlockWithPandoraSharding(panBlockNum, slot)
	m.beaconChainClient.EXPECT().GetCanonicalBlock(
		gomock.Any(),
		gomock.Any(),
	).Return(beaconBlock, nil)
	err, hash, blkNum := validator.panShardingCanonicalInfo(ctx, slot, pubKey)
	require.NoError(t, err, "Failed to get pandora sharding info")
	require.DeepEqual(t, eth1Types.EmptyRootHash, hash)
	require.DeepEqual(t, panBlockNum, blkNum)
}
