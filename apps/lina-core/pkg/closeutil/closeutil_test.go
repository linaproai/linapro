// This file verifies resource close helpers fold errors without panicking.

package closeutil

import (
	"context"
	"errors"
	"testing"
)

// failingCloser is a test closer that always returns one configured error.
type failingCloser struct {
	err error
}

// Close returns the configured test error.
func (c failingCloser) Close() error {
	return c.err
}

// TestCloseFoldsError verifies close failures are folded into the caller's
// named error return path.
func TestCloseFoldsError(t *testing.T) {
	var err error

	Close(context.Background(), failingCloser{err: errors.New("close failed")}, &err, "close resource")
	if err == nil {
		t.Fatal("expected close error to be folded into err pointer")
	}
}

// TestCloseWithoutErrorPointerDoesNotPanic verifies helper misuse is logged
// rather than converted into a process-level panic.
func TestCloseWithoutErrorPointerDoesNotPanic(t *testing.T) {
	Close(context.Background(), failingCloser{err: errors.New("close failed")}, nil, "close resource")
}
