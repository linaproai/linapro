// Package closeutil provides shared helpers for closing resources without
// dropping the returned error.
package closeutil

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/logger"
)

// Closer describes one resource that can be closed.
type Closer interface {
	// Close releases the underlying resource and returns any cleanup failure so
	// callers can fold it into their existing error return path.
	Close() error
}

// Close folds one close error into errPtr when the caller already returns an error.
func Close(ctx context.Context, closer Closer, errPtr *error, action string) {
	if closer == nil {
		return
	}
	closeErr := closer.Close()
	if closeErr == nil {
		return
	}
	wrapped := gerror.Wrap(closeErr, action)
	if errPtr == nil {
		// A nil error pointer means the caller misused this helper by omitting
		// the named return error path, so log the close failure instead of
		// panicking or silently dropping it.
		logger.Warningf(ctx, "resource close failed without error return path action=%s err=%v", action, wrapped)
		return
	}
	if *errPtr == nil {
		// Preserve the original business error when one already exists and only
		// surface the close failure when nothing else has failed yet.
		*errPtr = wrapped
	}
}
