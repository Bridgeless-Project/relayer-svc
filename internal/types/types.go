package types

import "github.com/pkg/errors"

var (
	ErrFailedToBroadcast   = errors.New("failed to broadcast transaction")
	ErrAlreadyExists       = errors.New("transaction already exists")
	ErrWithdrawalProcessed = errors.New("withdrawal already processed")
)
