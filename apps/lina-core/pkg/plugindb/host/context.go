// This file defines host-side plugindb audit context propagation helpers.

package host

import (
	"context"

	"lina-core/pkg/pluginbridge"
)

type auditContextKey struct{}

// AuditMetadata carries the governed execution metadata used by host-side
// audit, authorization, and DoCommit interception.
type AuditMetadata struct {
	// PluginID identifies the calling plugin.
	PluginID string
	// Table is the requested table identifier from the guest.
	Table string
	// Method is the structured data method currently being executed.
	Method string
	// ResourceTable is the resolved authorized host table.
	ResourceTable string
	// ExecutionSource identifies the calling source such as route or hook.
	ExecutionSource pluginbridge.ExecutionSource
	// UserID is the current authenticated user when present.
	UserID int32
	// Transaction reports whether the current call is inside one governed transaction.
	Transaction bool
}

// WithAudit attaches one audit metadata snapshot to the context.
func WithAudit(ctx context.Context, metadata *AuditMetadata) context.Context {
	if metadata == nil {
		return ctx
	}
	return context.WithValue(ctx, auditContextKey{}, metadata)
}

// AuditFromContext extracts one audit metadata snapshot from the context.
func AuditFromContext(ctx context.Context) *AuditMetadata {
	if ctx == nil {
		return nil
	}
	metadata, _ := ctx.Value(auditContextKey{}).(*AuditMetadata)
	return metadata
}
