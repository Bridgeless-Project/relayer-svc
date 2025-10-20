package broadcaster

import "github.com/pkg/errors"

const bufferChannelSize = 1000

var (
	errAlreadyExists = errors.New("withdrawal already exists")
	errWithdraw      = errors.New("failed to process withdrawal")
)
