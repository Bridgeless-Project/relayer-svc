package broadcaster

import "github.com/pkg/errors"

const bufferChannelSize = 1000

var (
	errWithdrawalInProcess = errors.New("withdrawal in process")
	errWithdraw            = errors.New("failed to process withdrawal")
)
