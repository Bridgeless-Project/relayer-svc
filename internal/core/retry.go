package core

import (
	"context"
	"time"

	"github.com/avast/retry-go"
	"github.com/pkg/errors"
	"gitlab.com/distributed_lab/logan/v3"
)

func DoWithRetry(ctx context.Context, function func() error,
	retries uint, retryTimeout time.Duration, logger *logan.Entry) error {
	err := retry.Do(
		function,
		retry.Attempts(retries),
		retry.Delay(retryTimeout),
		retry.DelayType(retry.BackOffDelay),
		retry.Context(ctx),
		retry.OnRetry(func(n uint, err error) {
			logger.WithError(err).WithField("attempt", n+1).Info("retrying step")
		}),
	)

	return errors.Wrap(err, "failed to execute function")
}
