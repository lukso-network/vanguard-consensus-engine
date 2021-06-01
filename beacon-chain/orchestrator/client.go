package orchestrator

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/rpc"
)

const confirmVanBlockHashesMethod = "orc_confirmVanBlockHashes"

type Client interface {
	Dial(string) (*Client, error)
	DialContext(context.Context, string) (*Client, error)
	NewClient(*rpc.Client) *Client
	Close()
	ConfirmVanBlockHashes(context.Context, []*BlockHash) ([]*BlockStatus, error)
}

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

func (orcRpc *RPCClient) Close() {
	orcRpc.client.Close()
}

func (orcRpc *RPCClient) ConfirmVanBlockHashes(ctx context.Context, blockHashes []*BlockHash) ([]*BlockStatus, error) {
	var blockStatuses []*BlockStatus

	err := orcRpc.client.CallContext(ctx, &blockStatuses, confirmVanBlockHashesMethod, blockHashes)
	if err != nil {
		log.WithField("context", "ConfirmVanBlockHashes").
			WithField("requestedBlockHashed", blockHashes)
		return nil, fmt.Errorf("rpcClient call context error, error is: %s", err.Error())
	}

	return blockStatuses, nil
}
