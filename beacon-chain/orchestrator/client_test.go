package orchestrator_test

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/prysmaticlabs/prysm/beacon-chain/orchestrator"
	"github.com/prysmaticlabs/prysm/shared/testutil/assert"
	"testing"
)

func TestRPCClient_ConfirmVanBlockHashes(t *testing.T) {
	ctx := context.Background()
	httpPath := "http://127.0.0.1:8545"
	rpcClient, err := rpc.Dial(httpPath)
	assert.NoError(t, err)

	orcRpcClient := orchestrator.NewClient(rpcClient)
	assert.NotNil(t, orcRpcClient)

	defer orcRpcClient.Close()

	var blockHashes []*orchestrator.BlockHash
	blockHash := &orchestrator.BlockHash{
		Slot: 0,
		Hash: common.Hash{},
	}
	blockHashes = append(blockHashes, blockHash)

	t.Run("connection refused", func(t *testing.T) {
		blockStatuses, err := orcRpcClient.ConfirmVanBlockHashes(ctx, blockHashes)
		assert.ErrorContains(t, "dial tcp 127.0.0.1:8545: connect: connection refused", err)
		assert.Equal(t, 0, len(blockStatuses))
	})
}
