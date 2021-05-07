package beacon

import (
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"github.com/prysmaticlabs/prysm/beacon-chain/core/feed"
	blockfeed "github.com/prysmaticlabs/prysm/beacon-chain/core/feed/block"
	"github.com/prysmaticlabs/prysm/shared/event"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// StreamNewPendingBlocks to orchestrator client every single time an unconfirmed block is received by the beacon node.
func (bs *Server) StreamNewPendingBlocks(req *ethpb.StreamPendingBlocksRequest, stream ethpb.BeaconChain_StreamNewPendingBlocksServer) error {
	unConfirmedblocksChannel := make(chan *feed.Event, 1)
	var unConfirmedblockSub event.Subscription

	unConfirmedblockSub = bs.BlockNotifier.BlockFeed().Subscribe(unConfirmedblocksChannel)
	defer unConfirmedblockSub.Unsubscribe()

	for {
		select {
		case blockEvent := <-unConfirmedblocksChannel:
			if blockEvent.Type == blockfeed.UnConfirmedBlock {
				data, ok := blockEvent.Data.(*blockfeed.UnConfirmedBlockData)
				if !ok || data == nil {
					continue
				}
				if err := stream.Send(data.Block); err != nil {
					return status.Errorf(codes.Unavailable, "Could not send un-confirmed block over stream: %v", err)
				}
				log.WithField("slot", data.Block.Slot).Debug(
					"New pending block has been published successfully")
			}
		case <-unConfirmedblockSub.Err():
			return status.Error(codes.Aborted, "Subscriber closed, exiting goroutine")
		case <-bs.Ctx.Done():
			return status.Error(codes.Canceled, "Context canceled")
		case <-stream.Context().Done():
			return status.Error(codes.Canceled, "Context canceled")
		}
	}
}
