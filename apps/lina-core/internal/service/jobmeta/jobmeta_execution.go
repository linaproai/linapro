// This file defines execution-identity helpers used by scheduled-job handlers.
// The persistent scheduler delivers at-least-once semantics: a process crash may
// leave unknown outcomes, and handlers that perform side effects must be
// idempotent. Missed cron ticks are not catch-up replayed after restart; startup
// only reclaims orphan running logs for the current node.

package jobmeta

import "context"

// executionLogIDKey is the private context key for the active job-log id.
type executionLogIDKey struct{}

// WithExecutionLogID attaches the current sys_job_log.id to ctx.
// The scheduler injects this value before invoking handler callbacks so
// handlers can use it as a stable execution id for de-duplication.
// logID <= 0 leaves ctx unchanged.
func WithExecutionLogID(ctx context.Context, logID int64) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if logID <= 0 {
		return ctx
	}
	return context.WithValue(ctx, executionLogIDKey{}, logID)
}

// ExecutionLogID returns the active execution log id when the current call is
// running under the persistent scheduler. Handlers with non-idempotent side
// effects should treat this id (or an equivalent business key) as an
// idempotency key. Missing or invalid values return ok=false.
func ExecutionLogID(ctx context.Context) (int64, bool) {
	if ctx == nil {
		return 0, false
	}
	logID, ok := ctx.Value(executionLogIDKey{}).(int64)
	if !ok || logID <= 0 {
		return 0, false
	}
	return logID, true
}
