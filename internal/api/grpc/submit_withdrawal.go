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

	client, err := clients.Client(identifier.ChainId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("chain %s is not supported", identifier.ChainId))
	}

	err = common.ValidateChainIdentifier(identifier, client)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid chain identifier %s: %v", identifier, err)
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

	err = broadcaster.Broadcast(*deposit)
	if err != nil {
		if errors.Is(err, internalTypes.ErrFailedToBroadcast) {
			return nil, status.Error(codes.Internal, "failed to broadcast withdrawal")
		}

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return nil, nil
}
