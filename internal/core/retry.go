package core

import (
	"context"
	"time"

	"github.com/avast/retry-go"
	"github.com/pkg/errors"
	"gitlab.com/distributed_lab/logan/v3"
)

var (
	Retries      uint
	RetryTimeout time.Duration
	Logger       *logan.Entry
)

func DoWithRetry(ctx context.Context, function func() error) error {
	err := retry.Do(
		function,
		retry.Attempts(Retries),
		retry.Delay(RetryTimeout),
		retry.DelayType(retry.BackOffDelay),
		retry.Context(ctx),
		retry.OnRetry(func(n uint, err error) {
			Logger.WithError(err).WithField("attempt", n+1).Warn("retrying step")
		}),
	)

	return errors.Wrap(err, "failed to execute function")
}
