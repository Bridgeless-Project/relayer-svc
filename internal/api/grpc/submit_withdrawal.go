package grpc

import (
	"context"
	"strings"

	"github.com/Bridgeless-Project/relayer-svc/internal/api/common"
	apiCtx "github.com/Bridgeless-Project/relayer-svc/internal/api/ctx"
	"github.com/Bridgeless-Project/relayer-svc/internal/api/types"
	internalTypes "github.com/Bridgeless-Project/relayer-svc/internal/types"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (i Implementation) SubmitWithdrawal(ctx context.Context, identifier *internalTypes.DepositIdentifier) (*types.SubmitResponse, error) {
	var (
		clients     = apiCtx.Clients(ctx)
		logger      = apiCtx.Logger(ctx)
		connector   = apiCtx.Connector(ctx)
		broadcaster = apiCtx.Broadcaster(ctx)
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

	deposit, err := connector.GetDeposit(ctx, common.ToDbIdentifier(identifier))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, status.Error(codes.NotFound, "deposit not found")
		}

		return nil, status.Error(codes.Internal, "unable to process withdrawal")
	}

	id, err := broadcaster.Broadcast(ctx, *deposit)
	if err != nil {
		if errors.Is(err, internalTypes.ErrFailedToBroadcast) {
			return nil, status.Error(codes.Internal, "failed to broadcast withdrawal")
		}

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &types.SubmitResponse{WithdrawalId: *id}, nil
}
