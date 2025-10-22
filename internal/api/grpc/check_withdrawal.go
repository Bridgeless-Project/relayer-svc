package grpc

import (
	"context"
	"fmt"

	"github.com/Bridgeless-Project/relayer-svc/internal/api/common"
	apiCtx "github.com/Bridgeless-Project/relayer-svc/internal/api/ctx"
	"github.com/Bridgeless-Project/relayer-svc/internal/api/types"
	internalTypes "github.com/Bridgeless-Project/relayer-svc/internal/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (i Implementation) CheckWithdrawal(ctx context.Context, identifier *internalTypes.DepositIdentifier) (*types.CheckWithdrawalResponse, error) {
	var (
		clients = apiCtx.Clients(ctx)
		logger  = apiCtx.Logger(ctx)
		db      = apiCtx.DB(ctx)
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

	withdrawalData, err := db.Get(common.ToDbIdentifier(identifier))
	if err != nil {
		logger.WithError(err).Error("failed to fetch withdrawal data")
		return nil, status.Error(codes.Internal, "failed to get withdrawal details")
	}

	if withdrawalData == nil {
		return nil, status.Error(codes.NotFound, "withdrawal data not found")
	}

	return common.ToStatusResponse(withdrawalData), nil
}
