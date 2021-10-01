package beacon

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	ethpb "github.com/prysmaticlabs/prysm/proto/eth/v1alpha1"
)

func (bs *Server) IsValidBlock(ctx context.Context, request *ethpb.BlockStatusValidationRequest) (*ethpb.BlockStatusValidationResponse, error) {
	ok, err := bs.CanonicalFetcher.IsCanonical(ctx, common.BytesToHash(request.GetHash()))
	return &ethpb.BlockStatusValidationResponse{IsValid: ok}, err
}
