package depositcache

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/pkg/errors"
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"github.com/prysmaticlabs/prysm/shared/fileutil"
	"github.com/prysmaticlabs/prysm/shared/params"
	"github.com/prysmaticlabs/prysm/shared/testutil/assert"
	"github.com/prysmaticlabs/prysm/shared/testutil/require"
	"github.com/prysmaticlabs/prysm/shared/trieutil"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestFinalizeDeposit_TrieRoot(t *testing.T) {
	ctx := context.Background()
	expanded, err := fileutil.ExpandPath("./test-data/deposit_data_32_keys.json")
	require.NoError(t, err)
	inputJSON, err := os.Open(expanded)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, inputJSON.Close())
	}()
	enc, err := ioutil.ReadAll(inputJSON)
	require.NoError(t, err)
	var depositJSON []*DepositDataJSON
	require.NoError(t, json.Unmarshal(enc, &depositJSON))

	depositDataList := make([]*ethpb.Deposit_Data, len(depositJSON))
	depositDataRoots := make([][]byte, len(depositJSON))
	for i, val := range depositJSON {
		data, dataRootBytes, err := depositJSONToDepositData(val)
		require.NoError(t, err)
		depositDataList[i] = data
		depositDataRoots[i] = dataRootBytes
	}

	trie, err := trieutil.GenerateTrieFromItems(depositDataRoots, params.BeaconConfig().DepositContractTreeDepth)
	require.NoError(t, err)
	expectedDepositRoot := trie.Root()
	expectedDepositRootHex := hexutil.Encode(expectedDepositRoot[:])

	dc, err := New()
	require.NoError(t, err)

	depositTrie, err := trieutil.NewTrie(params.BeaconConfig().DepositContractTreeDepth)
	require.NoError(t, err)

	initialtrieRoot := dc.finalizedDeposits.Deposits.HashTreeRoot()
	initialtrieRootHex := hexutil.Encode(initialtrieRoot[:])
	assert.Equal(t, "0xd70a234731285c6804c2a4f56711ddb8c82c99740f207854891028af34e27e5e", initialtrieRootHex)

	deposits, err := generateDepositsFromData(depositDataList, depositTrie)
	require.NoError(t, err)
	topTrieRoot := depositTrie.Root()
	topTrieRootHex := hexutil.Encode(topTrieRoot[:])
	assert.Equal(t, expectedDepositRootHex, topTrieRootHex)

	trie1, err := trieutil.NewTrie(params.BeaconConfig().DepositContractTreeDepth)
	require.NoError(t, err)
	for i := 0; i < len(deposits); i++ {
		depositRoot, err := deposits[i].Data.HashTreeRoot()
		require.NoError(t, err)

		depositRootHex := hexutil.Encode(depositRoot[:])
		actualDepositRootHex := hexutil.Encode(depositDataRoots[i])
		assert.DeepEqual(t, actualDepositRootHex, depositRootHex)

		depositHash, err := deposits[i].Data.HashTreeRoot()
		require.NoError(t, err)
		trie1.Insert(depositHash[:], i)

		dc.InsertDeposit(ctx, deposits[i], uint64(i), int64(i), trie.Root())
		dc.InsertFinalizedDeposits(ctx, int64(i))
	}

	actualDepositRoot := dc.finalizedDeposits.Deposits.Root()
	actualDepositRootHex := hexutil.Encode(actualDepositRoot[:])
	assert.DeepEqual(t, expectedDepositRootHex, actualDepositRootHex)
}

type DepositDataJSON struct {
	PubKey                string `json:"pubkey"`
	Amount                uint64 `json:"amount"`
	WithdrawalCredentials string `json:"withdrawal_credentials"`
	DepositDataRoot       string `json:"deposit_data_root"`
	Signature             string `json:"signature"`
}

func depositJSONToDepositData(input *DepositDataJSON) (depositData *ethpb.Deposit_Data, dataRoot []byte, err error) {
	pubKeyBytes, err := hex.DecodeString(strings.TrimPrefix(input.PubKey, "0x"))
	if err != nil {
		return
	}
	withdrawalbytes, err := hex.DecodeString(strings.TrimPrefix(input.WithdrawalCredentials, "0x"))
	if err != nil {
		return
	}
	signatureBytes, err := hex.DecodeString(strings.TrimPrefix(input.Signature, "0x"))
	if err != nil {
		return
	}
	dataRootBytes, err := hex.DecodeString(strings.TrimPrefix(input.DepositDataRoot, "0x"))
	if err != nil {
		return
	}

	depositData = &ethpb.Deposit_Data{
		PublicKey:             pubKeyBytes,
		WithdrawalCredentials: withdrawalbytes,
		Amount:                input.Amount,
		Signature:             signatureBytes,
	}

	//copy(dataRoot[:], dataRootBytes)
	dataRoot = dataRootBytes
	return
}

// generateDepositsFromData a list of deposit items by creating proofs for each of them from a sparse Merkle trie.
func generateDepositsFromData(depositDataItems []*ethpb.Deposit_Data, trie *trieutil.SparseMerkleTrie) ([]*ethpb.Deposit, error) {
	deposits := make([]*ethpb.Deposit, len(depositDataItems))
	for i := 0; i < len(depositDataItems); i++ {

		depositHash, err := depositDataItems[i].HashTreeRoot()
		if err != nil {
			return nil, errors.Wrap(err, "Unable to determine hashed value of deposit")
		}
		trie.Insert(depositHash[:], i)
		proof, err := trie.MerkleProof(i)
		if err != nil {
			return nil, errors.Wrapf(err, "could not generate proof for deposit %d", i)
		}
		deposits[i] = &ethpb.Deposit{
			Proof: proof,
			Data:  depositDataItems[i],
		}
	}
	return deposits, nil
}
