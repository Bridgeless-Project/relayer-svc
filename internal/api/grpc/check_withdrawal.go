package grpc

import (
	"context"

	"github.com/Bridgeless-Project/relayer-svc/internal/api/common"
	apiCtx "github.com/Bridgeless-Project/relayer-svc/internal/api/ctx"
	"github.com/Bridgeless-Project/relayer-svc/internal/api/types"
	internalTypes "github.com/Bridgeless-Project/relayer-svc/internal/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (i Implementation) CheckWithdrawal(ctx context.Context, identifier *internalTypes.DepositIdentifier) (*types.CheckWithdrawalResponse, error) {
	var (
		logger = apiCtx.Logger(ctx)
		db     = apiCtx.DB(ctx)
		cfg    = apiCtx.Config(ctx)
	)

	err := common.ValidateIdentifier(identifier)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid identifier %s: %v", identifier, err)
	}

	withdrawalData, err := db.Get(common.ToDbIdentifier(identifier))
	if err != nil {
		logger.WithError(err).Error("failed to fetch withdrawal data from db")
		return nil, status.Error(codes.Internal, "failed to get withdrawal details")
	}

	if withdrawalData == nil {
		return nil, status.Error(codes.NotFound, "withdrawal data not found")
	}

	_, timeout := cfg.RecoveryParams()
	withdrawalData.RecoveryTimestamp.Add(timeout)

	return common.ToStatusResponse(withdrawalData), nil
}
