package orchestrator_test

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/prysmaticlabs/prysm/beacon-chain/orchestrator"
	vanTypes "github.com/prysmaticlabs/prysm/shared/params"
	"github.com/prysmaticlabs/prysm/shared/testutil/assert"
	"net"
	"testing"
)

func listenTCP() (net.Listener, string) {
	l, e := net.Listen("tcp", "127.0.0.1:8545") // any available address
	if e != nil {
		print(e.Error())
	}
	if l == nil {
		panic("error: Listener is nil!")
	}

	return l, l.Addr().String()
}

type orchestratorTestService struct {
	unsubscribed            chan string
	gotHangSubscriptionReq  chan struct{}
	unblockHangSubscription chan struct{}
}

func (apiMock *orchestratorTestService) ConfirmVanBlockHashes(args []*vanTypes.ConfirmationReqData, reply []*vanTypes.ConfirmationResData) error {
	for _, confirmationReq := range args {
		exampleResp := &vanTypes.ConfirmationResData{
			Slot:   confirmationReq.Slot,
			// only for test purpose, for now
			Hash:   confirmationReq.Hash,
			Status: "Verified",
		}
		reply = append(reply, exampleResp)
	}

	return nil
}

func newTestServer() *rpc.Server {
	newServer := rpc.NewServer()
	if err := newServer.RegisterName("orc", new(orchestratorTestService)); err != nil {
		panic(err)
	}

	return newServer
}

func TestRPCClient_ConfirmVanBlockHashes(t *testing.T) {
	ctx := context.Background()
	rpcServer := newTestServer()

	defer rpcServer.Stop()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("can't listen: %s", err.Error())
	}

	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			fmt.Println(err.Error())
		}
	}(listener)

	go func() {
		err := rpcServer.ServeListener(listener)
		if err != nil {
			fmt.Println(err.Error())
		}
	}()

	listenerAddr := "http://" + listener.Addr().String()

	orcRpcClient, err := orchestrator.Dial(ctx, listenerAddr)
	assert.NoError(t, err)

	defer orcRpcClient.Close()

	blockHashes := make([]*vanTypes.ConfirmationReqData, 1)
	blockHash := &vanTypes.ConfirmationReqData{
		Slot: 0,
		Hash: common.Hash{},
	}
	blockHashes[0] = blockHash

	t.Run("connection success", func(t *testing.T) {
		blockStatuses, err := orcRpcClient.ConfirmVanBlockHashes(ctx, blockHashes)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(blockStatuses))
	})

	//t.Run("test another rpc client flow", func(t *testing.T) {
	//	client, err := orchestrator.Dial(ctx, listenerAddr)
	//	if err != nil {
	//		t.Fatalf("dialing error: %s", err.Error())
	//	}
	//
	//	defer client.Close()
	//
	//	// Synchronous calls
	//	args := &vanTypes.ConfirmationReqData{
	//		Slot: 0,
	//		Hash: [32]byte{},
	//	}
	//	var argsSlice []*vanTypes.ConfirmationReqData
	//	argsSlice = append(argsSlice, args)
	//
	//	hashes, err := client.ConfirmVanBlockHashes(ctx, argsSlice)
	//	if err != nil {
	//		t.Fatalf("ConfirmVanBlockHashes error: %s", err.Error())
	//	}
	//
	//	wantedResp := &vanTypes.ConfirmationResData{
	//		Slot:   0,
	//		Hash:   [32]byte{},
	//		Status: "",
	//	}
	//
	//	assert.DeepEqual(t, wantedResp, hashes)
	//})

	//t.Run("test request and response flow", func(t *testing.T) {
	//	var (
	//		request  = `{"jsonrpc":"1.0","id":1,"method":"orc_confirmVanBlockHashes","params":{"subtrahend": 23, "minuend": 42}}` + "\n"
	//		wantResp = `{"jsonrpc":"2.0","id":1,"error":{"code":-32602,"message":"non-array args"}}` + "\n"
	//		deadline = time.Now().Add(10 * time.Second)
	//	)
	//
	//	conn, err := net.Dial("tcp", listener.Addr().String())
	//	if err != nil {
	//		t.Fatal("can't dial:", err)
	//	}
	//
	//	defer func(conn net.Conn) {
	//		err := conn.Close()
	//		if err != nil {
	//			t.Error(err)
	//		}
	//	}(conn)
	//
	//	err = conn.SetDeadline(deadline)
	//	if err != nil {
	//		t.Error(err)
	//	}
	//	// Write the request, then half-close the connection so the server stops reading.
	//	_, err = conn.Write([]byte(request))
	//	if err != nil {
	//		t.Error(err)
	//	}
	//	err = conn.(*net.TCPConn).CloseWrite()
	//	if err != nil {
	//		t.Error(err)
	//	}
	//	// Now try to get the response.
	//	buf := make([]byte, 2000)
	//	n, err := conn.Read(buf)
	//	if err != nil {
	//		t.Fatal("read error:", err)
	//	}
	//	assert.DeepEqual(t, buf[:n], []byte(wantResp))
	//	//if !bytes.Equal(buf[:n], []byte(wantResp)) {
	//	//	t.Fatalf("wrong response: %s", buf[:n])
	//	//}
	//})
}
