package core

import "github.com/pkg/errors"

const BufferChannelSize = 1000

var (
	ErrTransactionAlreadySubmitted = errors.New("transaction already submitted")
)
