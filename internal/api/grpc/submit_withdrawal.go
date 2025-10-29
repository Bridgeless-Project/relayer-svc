package grpc

import (
	"context"
	"fmt"
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
		connector   = apiCtx.Connector(ctx)
		broadcaster = apiCtx.Broadcaster(ctx)
	)

	err := common.ValidateIdentifier(identifier)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid identifier %s: %v", identifier, err)
	}

	deposit, err := connector.GetDeposit(ctx, common.ToDbIdentifier(identifier))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, status.Error(codes.NotFound, "deposit not found")
		}

		return nil, status.Error(codes.Internal, "unable to process withdrawal")
	}

	if ok := clients.SupportsChain(deposit.WithdrawalChainId); !ok {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("client does not support the chain %s",
			deposit.WithdrawalChainId))
	}

	if deposit.WithdrawalTxHash != nil {
		return nil, status.Error(codes.InvalidArgument, "deposit is already withdrawn")
	}

	err = broadcaster.Broadcast(*deposit)
	if err != nil {
		if errors.Is(err, internalTypes.ErrFailedToBroadcast) {
			return nil, status.Error(codes.Internal, "failed to broadcast withdrawal")
		}

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return nil, nil
}
