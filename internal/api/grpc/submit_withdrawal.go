package grpc

import (
	"context"
	"strings"

	"github.com/Bridgeless-Project/relayer-svc/internal/api/common"
	apiCtx "github.com/Bridgeless-Project/relayer-svc/internal/api/ctx"
	internalTypes "github.com/Bridgeless-Project/relayer-svc/internal/types"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (i Implementation) SubmitWithdrawal(ctx context.Context, identifier *internalTypes.DepositIdentifier) (*emptypb.Empty, error) {
	var (
		clients     = apiCtx.Clients(ctx)
		logger      = apiCtx.Logger(ctx)
		connector   = apiCtx.Connector(ctx)
		broadcaster = apiCtx.Broadcaster(ctx)
		db          = apiCtx.DB(ctx)
	)

	err := common.ValidateIdentifier(identifier)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid identifier %s: %v", identifier, err)
	}

	if !clients.SupportsChain(identifier.ChainId) {
		return nil, status.Error(codes.InvalidArgument, "chain is not supported")
	}

	client, err := clients.Client(identifier.ChainId)
	if err != nil {
		logger.Errorf("failed to get chain client %s: %v", identifier.ChainId, err)
		return nil, status.Errorf(codes.Internal, "unable to process withdrawal")
	}

	if !client.TransactionHashValid(identifier.TxHash) {
		return nil, status.Error(codes.InvalidArgument, "invalid transaction hash")
	}

	deposit, err := db.Get(common.ToDbIdentifier(identifier))
	if err != nil {
		logger.Errorf("failed to get deposit from database: %v", err)
		return nil, status.Error(codes.Internal, "unable to process withdrawal")
	}

	if deposit != nil {
		return nil, status.Error(codes.AlreadyExists, "deposit already exists")
	}

	deposit, err = connector.GetDeposit(ctx, common.ToDbIdentifier(identifier))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, status.Error(codes.NotFound, "deposit not found")
		}

		return nil, status.Error(codes.Internal, "unable to process withdrawal")
	}

	if err := broadcaster.Broadcast(ctx, *deposit); err != nil {
		if errors.Is(err, internalTypes.ErrFailedToBroadcast) {
			return nil, status.Error(codes.Internal, "failed to broadcast withdrawal")
		}

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &emptypb.Empty{}, nil
}
