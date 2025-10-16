package broadcaster

import "github.com/pkg/errors"

var (
	errWithdrawalInProcess = errors.New("withdrawal in process")
	errWithdraw            = errors.New("failed to process withdrawal")
)
