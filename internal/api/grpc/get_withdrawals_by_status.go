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

func (i Implementation) GetWithdrawalsByStatus(ctx context.Context, request *types.GetWithdrawalsByStatusRequest) (*types.GetWithdrawalsByStatusResponse, error) {
	var (
		logger = apiCtx.Logger(ctx)
		db     = apiCtx.DB(ctx)
	)
	statusInt := int32(request.GetWithdrawalStatus())

	if statusInt == 0 {
		return nil, status.Error(codes.InvalidArgument, "withdrawal_status is required")
	}

	if _, isValid := internalTypes.WithdrawalStatus_name[statusInt]; !isValid {
		return nil, status.Errorf(codes.InvalidArgument, "unsupported withdrawal status integer: %d", statusInt)
	}
	withdrawals, err := db.GetWithStatus(request.GetWithdrawalStatus())
	if err != nil {
		logger.WithError(err).Error("failed to fetch withdrawal list from db")
		return nil, status.Error(codes.Internal, "failed to get withdrawals")
	}
	result := common.ToWithdrawalByStatusResponse(withdrawals)
	return result, nil
}
