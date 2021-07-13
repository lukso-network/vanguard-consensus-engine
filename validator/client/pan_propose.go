package client

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	eth1Types "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/pkg/errors"
	types "github.com/prysmaticlabs/eth2-types"
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"github.com/prysmaticlabs/prysm/beacon-chain/core/helpers"
	validatorpb "github.com/prysmaticlabs/prysm/proto/validator/accounts/v2"
	"github.com/prysmaticlabs/prysm/shared/params"
	"github.com/prysmaticlabs/prysm/shared/timeutils"
	"github.com/prysmaticlabs/prysm/validator/pandora"
	"golang.org/x/crypto/sha3"
	"time"
)

var (
	// errInvalidHeaderHash is returned if the header hash does not match with incoming header hash
	errInvalidHeaderHash = errors.New("invalid header hash")
	// errInvalidSlot is returned if the current slot does not match with incoming slot
	errInvalidSlot = errors.New("invalid slot")
	// errInvalidEpoch is returned if the epoch does not match with incoming epoch
	errInvalidEpoch = errors.New("invalid epoch")
	// errInvalidProposerIndex is returned if the proposer index does not match with incoming proposer index
	errInvalidProposerIndex = errors.New("invalid proposer index")
	// errInvalidTimestamp is returned if the timestamp of a block is higher than the current time
	errInvalidTimestamp = errors.New("invalid timestamp")
	// errNilHeader
	errNilHeader = errors.New("pandora header is nil")
	// errPanShardingInfoNotFound
	errPanShardingInfoNotFound = errors.New("pandora sharding info not found in canonical head")
	// errSubmitShardingSignatureFailed
	errSubmitShardingSignatureFailed = errors.New("pandora sharding signature submission failed")
)

// processPandoraShardHeader method does the following tasks:
// - Get pandora block header, header hash, extraData from remote pandora node
// - Validate block header hash and extraData fields
// - Signs header hash using a validator key
// - Submit signature and header to pandora node
func (v *validator) processPandoraShardHeader(
	ctx context.Context,
	beaconBlk *ethpb.BeaconBlock,
	slot types.Slot,
	epoch types.Epoch,
	pubKey [48]byte,
) error {

	fmtKey := fmt.Sprintf("%#x", pubKey[:])
	latestPandoraHash := eth1Types.EmptyRootHash
	latestPandoraBlkNum := uint64(0)

	// if pandoraShard is nil means there is no pandora blocks in pandora chain except block-0
	if beaconBlk.Body != nil && beaconBlk.Body.PandoraShard != nil {
		latestPandoraHash = common.BytesToHash(beaconBlk.Body.PandoraShard[0].Hash)
		latestPandoraBlkNum = beaconBlk.Body.PandoraShard[0].BlockNumber
	}

	// Request for pandora chain header
	header, headerHash, extraData, err := v.pandoraService.GetShardBlockHeader(ctx, latestPandoraHash, latestPandoraBlkNum+1)
	if err != nil {
		log.WithField("blockSlot", slot).
			WithField("fmtKey", fmtKey).
			WithError(err).Error("Failed to request block header from pandora node")
		if v.emitAccountMetrics {
			ValidatorProposeFailVec.WithLabelValues(fmtKey).Inc()
		}
		return err
	}

	// Validate pandora chain header hash, extraData fields
	if err := v.verifyPandoraShardHeader(slot, epoch, header, headerHash, extraData); err != nil {
		log.WithField("blockSlot", slot).
			WithField("fmtKey", fmtKey).
			WithError(err).Error("Failed to validate pandora block header")
		if v.emitAccountMetrics {
			ValidatorProposeFailVec.WithLabelValues(fmtKey).Inc()
		}
		return err
	}
	headerHashSig, err := v.keyManager.Sign(ctx, &validatorpb.SignRequest{
		PublicKey:       pubKey[:],
		SigningRoot:     headerHash[:],
		SignatureDomain: nil,
		Object:          nil,
	})

	if err != nil {
		log.WithField("blockSlot", slot).WithError(err).Error("Failed to sign pandora header hash")
		if v.emitAccountMetrics {
			ValidatorProposeFailVec.WithLabelValues(fmtKey).Inc()
		}
		return err
	}

	header.MixDigest = common.BytesToHash(headerHashSig.Marshal())
	var headerHashSig96Bytes [96]byte
	copy(headerHashSig96Bytes[:], headerHashSig.Marshal())

	// Submit bls signature to pandora
	if status, err := v.pandoraService.SubmitShardBlockHeader(
		ctx, header.Nonce.Uint64(), headerHash, headerHashSig96Bytes); !status || err != nil {
		// err nil means got success in api request but pandora does not write the header
		if err == nil {
			log.WithField("slot", slot).Debug("pandora refused to accept the sharding signature")
			err = errSubmitShardingSignatureFailed
		}
		log.WithError(err).
			WithField("pubKey", fmt.Sprintf("%#x", pubKey)).
			WithField("slot", slot).
			Error("Failed to process pandora chain shard header")
		if v.emitAccountMetrics {
			ValidatorProposeFailVec.WithLabelValues(fmtKey).Inc()
		}
		return err
	}

	//var headerHashWithSig common.Hash
	//if headerHashWithSig, err = calculateHeaderHashWithSig(header, *extraData, headerHashSig96Bytes); err != nil {
	//	log.WithError(err).
	//		WithField("pubKey", fmt.Sprintf("%#x", pubKey)).
	//		WithField("slot", slot).
	//		Error("Failed to process pandora chain shard header hash with signature")
	//	if v.emitAccountMetrics {
	//		ValidatorProposeFailVec.WithLabelValues(fmtKey).Inc()
	//	}
	//	return err
	//}
	// fill pandora shard info with pandora header
	pandoraShard := v.preparePandoraShardingInfo(header, common.Hash{}, headerHash, headerHashSig.Marshal())
	pandoraShards := make([]*ethpb.PandoraShard, 1)
	pandoraShards[0] = pandoraShard
	beaconBlk.Body.PandoraShard = pandoraShards
	log.WithField("slot", beaconBlk.Slot).Debug("successfully created pandora sharding block")
	printHeader(header)

	// calling UpdateStateRoot api of beacon-chain so that state root will be updated after adding pandora shard
	updateBeaconBlk, err := v.updateStateRoot(ctx, beaconBlk)
	if err != nil {
		log.WithError(err).
			WithField("pubKey", fmt.Sprintf("%#x", pubKey)).
			WithField("slot", slot).
			Error("Failed to process pandora chain shard header")
		if v.emitAccountMetrics {
			ValidatorProposeFailVec.WithLabelValues(fmtKey).Inc()
		}
		return err
	}

	beaconBlk.StateRoot = updateBeaconBlk.StateRoot
	return nil
}

// verifyPandoraShardHeader verifies pandora sharding chain header hash and extraData field
func (v *validator) verifyPandoraShardHeader(
	slot types.Slot,
	epoch types.Epoch,
	header *eth1Types.Header,
	headerHash common.Hash,
	extraData *pandora.ExtraData,
) error {

	// verify header hash
	if sealHash(header) != headerHash {
		log.WithError(errInvalidHeaderHash).Error("invalid header hash from pandora chain")
		return errInvalidHeaderHash
	}
	// verify timestamp. Timestamp should not be future time
	if header.Time > uint64(timeutils.Now().Unix()) {
		log.WithError(errInvalidTimestamp).Error("invalid timestamp from pandora chain")
		return errInvalidTimestamp
	}
	// verify epoch number
	if extraData.Epoch != uint64(epoch) {
		log.WithError(errInvalidEpoch).Error("invalid epoch from pandora chain")
		return errInvalidEpoch
	}

	expectedTimeStart, err := helpers.SlotToTime(v.genesisTime, slot)

	if nil != err {
		return err
	}

	// verify slot number
	if extraData.Slot != uint64(slot) {
		log.WithError(errInvalidSlot).
			WithField("slot", slot).
			WithField("extraDataSlot", extraData.Slot).
			WithField("header", header.Extra).
			WithField("headerTime", header.Time).
			WithField("expectedTimeStart", expectedTimeStart.Unix()).
			Error("invalid slot from pandora chain")
		return errInvalidSlot
	}

	err = helpers.VerifySlotTime(
		v.genesisTime,
		types.Slot(extraData.Slot),
		params.BeaconNetworkConfig().MaximumGossipClockDisparity,
	)

	if nil != err {
		log.WithError(errInvalidSlot).
			WithField("slot", slot).
			WithField("extraDataSlot", extraData.Slot).
			WithField("header", header.Extra).
			WithField("headerTime", header.Time).
			WithField("expectedTimeStart", expectedTimeStart.Unix()).
			WithField("currentSlot", helpers.CurrentSlot(v.genesisTime)).
			WithField("unixTimeNow", time.Now().Unix()).
			Error(err)

		return err
	}

	return nil
}

// panShardingCanonicalInfo method gets header hash and block number from sharding head
func (v *validator) updateStateRoot(
	ctx context.Context,
	beaconBlk *ethpb.BeaconBlock,
) (*ethpb.BeaconBlock, error) {

	log.WithField("slot", beaconBlk.Slot).WithField(
		"stateRoot", fmt.Sprintf("%X", beaconBlk.StateRoot)).Debug(
		"trying to update state root of the beacon block after adding pandora shard")
	// Request block from beacon node
	updatedBeaconBlk, err := v.validatorClient.UpdateStateRoot(ctx, beaconBlk)
	if err != nil {
		log.WithField("slot", updatedBeaconBlk.Slot).WithError(err).Error("Failed to update state root in beacon node")
		return beaconBlk, err
	}
	log.WithField("slot", updatedBeaconBlk.Slot).WithField(
		"stateRoot", fmt.Sprintf("%X", updatedBeaconBlk.StateRoot)).Debug(
		"successfully compute and update state root in beacon node")
	return updatedBeaconBlk, nil
}

// preparePandoraShardingInfo
func (v *validator) preparePandoraShardingInfo(
	header *eth1Types.Header,
	headerHashWithSig common.Hash,
	sealedHeaderHash common.Hash,
	sig []byte,
) *ethpb.PandoraShard {
	return &ethpb.PandoraShard{
		BlockNumber: header.Number.Uint64(),
		Hash:        header.Hash().Bytes(),
		ParentHash:  header.ParentHash.Bytes(),
		StateRoot:   header.Root.Bytes(),
		TxHash:      header.TxHash.Bytes(),
		ReceiptHash: header.ReceiptHash.Bytes(),
		SealHash:    sealedHeaderHash.Bytes(),
		Signature:   sig,
	}
}

// SealHash returns the hash of a block prior to it being sealed.
func sealHash(header *eth1Types.Header) (hash common.Hash) {
	hasher := sha3.NewLegacyKeccak256()

	if err := rlp.Encode(hasher, []interface{}{
		header.ParentHash,
		header.UncleHash,
		header.Coinbase,
		header.Root,
		header.TxHash,
		header.ReceiptHash,
		header.Bloom,
		header.Difficulty,
		header.Number,
		header.GasLimit,
		header.GasUsed,
		header.Time,
		header.Extra,
	}); err != nil {
		return eth1Types.EmptyRootHash
	}
	hasher.Sum(hash[:0])
	return hash
}

func calculateHeaderHashWithSig(
	header *eth1Types.Header,
	pandoraExtraData pandora.ExtraData,
	signatureBytes [96]byte,
) (headerHash common.Hash, err error) {

	var blsSignatureBytes pandora.BlsSignatureBytes
	copy(blsSignatureBytes[:], signatureBytes[:])

	extraDataWithSig := new(pandora.PandoraExtraDataSig)
	extraDataWithSig.ExtraData = pandoraExtraData
	extraDataWithSig.BlsSignatureBytes = &blsSignatureBytes

	log.WithField("headerHash", header.Hash().Hex()).Debug("before calculation header hash with signature")
	header.Extra, err = rlp.EncodeToBytes(extraDataWithSig)
	headerHash = header.Hash()
	log.WithField("headerHashWithSig", headerHash.Hex()).Debug("calculated header hash with signature")
	return
}

/**
{
  difficulty: "0x1",
  extraData: "0xf866c30a800ab860a899054e1dd5ada5f5174edc532ffa39662cbfc90470233028096d7e41a3263114572cb7d0493ba213becec37f43145d041e0bfbaaf4bf8c2a7aeaebdd0d7fd6c326831b986a9802bf5e9ad1f180553ae0af77334cd4eb606ed71b0dc7db424e",
  gasLimit: "0x47ff2c",
  gasUsed: "0x0",
  hash: "0xaa1193c7d0d3cb6fbd33f5ddb748cd1e70e92b7bfc9667d4ecb4f61be63deb6c",
  logsBloom: "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
  miner: "0xb46d14ef42ac9bb01303ba1842ea784e2460c7e7",
  mixHash: "0xa899054e1dd5ada5f5174edc532ffa39662cbfc90470233028096d7e41a32631",
  nonce: "0x0000000000000000",
  number: "0x4",
  parentHash: "0x3244474eb97faefc26df91a8c3d0f2a8f859855ba87b76b1cc6044cca29add40",
  receiptsRoot: "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
  sha3Uncles: "0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347",
  size: "0x288",
  stateRoot: "0x03906b0760f3bec421d8a71c44273a5994c5f0e35b8b8d9e2112dc95a182aae6",
  timestamp: "0x60ed66d1",
  totalDifficulty: "0x80004",
  transactionsRoot: "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421"
}
*/
func printHeader(header *eth1Types.Header) {
	log.WithField("difficulty", header.Difficulty).WithField(
		"extraData", common.Bytes2Hex(header.Extra)).WithField(
		"gasLimit", header.GasLimit).WithField("gasUsed", header.GasUsed).WithField(
		"hash", header.Hash()).WithField("logsBloom", header.Bloom.Bytes()).WithField(
		"miner", header.Coinbase.Hex()).WithField("mixHash", header.MixDigest.Hex()).WithField(
		"nonce", header.Nonce).WithField("number", header.Number).WithField(
		"parentHash", header.ParentHash.Hex()).WithField(
		"receiptsRoot", header.ReceiptHash.Hex()).WithField(
		"sha3Uncles", header.UncleHash.Hex()).WithField("stateRoot", header.Root.Hex()).WithField(
		"timestamp", header.Time).WithField("transactionsRoot", header.TxHash.Hex()).Debug("header info")

}
