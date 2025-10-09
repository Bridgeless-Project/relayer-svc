package grpc

import (
	"context"

	"github.com/Bridgeless-Project/relayer-svc/internal/api/types"
	internalTypes "github.com/Bridgeless-Project/relayer-svc/internal/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	ErrInternal           = status.Error(codes.Internal, "internal error")
	ErrTxAlreadySubmitted = status.Error(codes.AlreadyExists, "transaction already submitted")
	ErrDepositPending     = status.Error(codes.FailedPrecondition, "deposit pending")
)

var _ types.APIServer = Implementation{}

type Implementation struct{}

func (i Implementation) SubmitWithdrawal(ctx context.Context, identifier *internalTypes.DepositIdentifier) (*emptypb.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (i Implementation) CheckHealth(ctx context.Context, empty *emptypb.Empty) (*emptypb.Empty, error) {
	//TODO implement me
	panic("implement me")
}
