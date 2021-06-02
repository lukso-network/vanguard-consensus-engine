package orchestrator

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
	types "github.com/prysmaticlabs/eth2-types"
	vanTypes "github.com/prysmaticlabs/prysm/shared/params"
)

const confirmVanBlockHashesMethod = "orc_confirmVanBlockHashes"

type Client interface {
	ConfirmVanBlockHashes(context.Context, []*vanTypes.ConfirmationReqData) ([]*vanTypes.ConfirmationResData, error)
}

// Assure that RPCClient struct will implement Client interface
var _ Client = &RPCClient{}

// RPCClient defines typed wrappers for the Ethereum RPC API.
type RPCClient struct {
	client *rpc.Client
}

// Dial connects a client to the given URL.
func Dial(rawurl string) (*RPCClient, error) {
	return DialContext(context.Background(), rawurl)
}

func DialContext(ctx context.Context, rawurl string) (*RPCClient, error) {
	c, err := rpc.DialContext(ctx, rawurl)
	if err != nil {
		return nil, err
	}
	return NewClient(c), nil
}

// NewClient creates a client that uses the given RPC client.
func NewClient(c *rpc.Client) *RPCClient {
	return &RPCClient{c}
}

func (orc *RPCClient) Close() {
	orc.client.Close()
}

func (orc *RPCClient) ConfirmVanBlockHashes(ctx context.Context, blockHashes []*vanTypes.ConfirmationReqData) ([]*vanTypes.ConfirmationResData, error) {
	var (
		orcBlockStatuses []*BlockStatus
		orcBlockHashes   []*BlockHash
		blockStatuses    []*vanTypes.ConfirmationResData
	)

	for _, blockHash := range blockHashes {
		orcBlockHash := &BlockHash{
			Slot: uint64(blockHash.Slot),
			Hash: common.BytesToHash(blockHash.Hash[:]),
		}
		orcBlockHashes = append(orcBlockHashes, orcBlockHash)
	}

	err := orc.client.CallContext(ctx, &orcBlockStatuses, confirmVanBlockHashesMethod, orcBlockHashes)
	if err != nil {
		log.WithField("context", "ConfirmVanBlockHashes").
			WithField("requestedBlockHashed", blockHashes)
		return nil, fmt.Errorf("rpcClient call context error, error is: %s", err.Error())
	}

	for _, orcBlockStatus := range orcBlockStatuses {
		blockStatus := &vanTypes.ConfirmationResData{
			Slot:   types.Slot(orcBlockStatus.Slot),
			Hash:   orcBlockStatus.Hash,
			Status: vanTypes.Status(orcBlockStatus.Status),
		}
		blockStatuses = append(blockStatuses, blockStatus)
	}

	return blockStatuses, nil
}
