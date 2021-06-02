package orchestrator_test

import (
	"bytes"
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/prysmaticlabs/prysm/beacon-chain/orchestrator"
	vanTypes "github.com/prysmaticlabs/prysm/shared/params"
	"github.com/prysmaticlabs/prysm/shared/testutil/assert"
	"net"
	"testing"
	"time"
)

type orchestratorTestService struct {
	unsubscribed            chan string
	gotHangSubscriptionReq  chan struct{}
	unblockHangSubscription chan struct{}
}

func newTestServer() *rpc.Server {
	server := rpc.NewServer()
	if err := server.RegisterName("orc", new(orchestratorTestService)); err != nil {
		panic(err)
	}
	return server
}

func TestRPCClient_ConfirmVanBlockHashes(t *testing.T) {
	ctx := context.Background()
	httpPath := "http://127.0.0.1:8545"
	rpcClient, err := rpc.Dial(httpPath)
	assert.NoError(t, err)

	orcRpcClient := orchestrator.NewClient(rpcClient)
	assert.NotNil(t, orcRpcClient)
	defer orcRpcClient.Close()

	var blockHashes []*vanTypes.ConfirmationReqData
	blockHash := &vanTypes.ConfirmationReqData{
		Slot: 0,
		Hash: common.Hash{},
	}
	blockHashes = append(blockHashes, blockHash)

	t.Run("connection refused", func(t *testing.T) {
		blockStatuses, err := orcRpcClient.ConfirmVanBlockHashes(ctx, blockHashes)
		assert.ErrorContains(t, "dial tcp 127.0.0.1:8545: connect: connection refused", err)
		assert.Equal(t, 0, len(blockStatuses))
	})

	t.Run("", func(t *testing.T) {
		// TODO: mock comes here
		// - add blocks to mocked server
		// - retrieve confirmations and compare with expected ones
	})
}

func TestName(t *testing.T) {
	server := newTestServer()
	defer server.Stop()

	listener, err := net.Listen("tcp", "127.0.0.1:8545")
	if err != nil {
		t.Fatal("can't listen:", err)
	}
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {

		}
	}(listener)

	go func() {
		err := server.ServeListener(listener)
		if err != nil {
			t.Error("listener error:", err)
			return
		}
	}()

	var (
		request  = `{"jsonrpc":"1.0","id":1,"method":"orc_confirmVanBlockHashes"}` + "\n"
		wantResp = `{"jsonrpc":"1.0","id":1,"result":{"nftest":"1.0","rpc":"1.0","test":"1.0"}}` + "\n"
		deadline = time.Now().Add(10 * time.Second)
	)
	for i := 0; i < 20; i++ {
		conn, err := net.Dial("tcp", listener.Addr().String())
		if err != nil {
			t.Fatal("can't dial:", err)
		}
		defer func(conn net.Conn) {
			err := conn.Close()
			if err != nil {
				t.Error(err)
			}
		}(conn)
		err = conn.SetDeadline(deadline)
		if err != nil {
			t.Error(err)
		}
		// Write the request, then half-close the connection so the server stops reading.
		_, err = conn.Write([]byte(request))
		if err != nil {
			t.Error(err)
		}
		err = conn.(*net.TCPConn).CloseWrite()
		if err != nil {
			t.Error(err)
		}
		// Now try to get the response.
		buf := make([]byte, 2000)
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatal("read error:", err)
		}
		if !bytes.Equal(buf[:n], []byte(wantResp)) {
			t.Fatalf("wrong response: %s", buf[:n])
		}
	}
}
