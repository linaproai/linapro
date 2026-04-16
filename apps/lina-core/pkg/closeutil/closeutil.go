// Package closeutil provides shared helpers for closing resources without
// dropping the returned error.
package closeutil

import "github.com/gogf/gf/v2/errors/gerror"

// Closer describes one resource that can be closed.
type Closer interface {
	Close() error
}

// Close folds one close error into errPtr when the caller already returns an error.
func Close(closer Closer, errPtr *error, action string) {
	if closer == nil {
		return
	}
	closeErr := closer.Close()
	if closeErr == nil {
		return
	}
	wrapped := gerror.Wrap(closeErr, action)
	if errPtr == nil {
		// Callers that do not expose an error return must still fail loudly so
		// close errors are never swallowed silently.
		panic(wrapped)
	}
	if *errPtr == nil {
		// Preserve the original business error when one already exists and only
		// surface the close failure when nothing else has failed yet.
		*errPtr = wrapped
	}
}
