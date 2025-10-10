package broadcaster

import "github.com/pkg/errors"

var (
	errWithdrawalInProcess     = errors.New("withdrawal in process")
	errWithdrawalAlreadyExists = errors.New("withdrawal already exists")
	errWithdraw                = errors.New("failed to process withdrawal")
)
