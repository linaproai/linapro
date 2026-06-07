// This file builds plugin data audit metadata and bridges the datahost package
// to the reusable record store capability host-side governance layer.

package datahost

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/internal/service/plugin/internal/catalog"
	datahost "lina-core/internal/service/plugin/internal/datahost/internal/host"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// withPluginDataAudit attaches audit metadata for downstream record store capability host logging.
func withPluginDataAudit(ctx context.Context, metadata *datahost.AuditMetadata) context.Context {
	return datahost.WithAudit(ctx, metadata)
}

// buildPluginDataAuditMetadata builds the audit metadata snapshot for one governed request.
func buildPluginDataAuditMetadata(
	execCtx *executionContext,
	resource *catalog.ResourceSpec,
	method string,
	inTransaction bool,
) *datahost.AuditMetadata {
	metadata := &datahost.AuditMetadata{
		Method:      strings.ToLower(strings.TrimSpace(method)),
		Transaction: inTransaction,
	}
	if execCtx != nil {
		metadata.PluginID = strings.TrimSpace(execCtx.pluginID)
		metadata.Table = strings.TrimSpace(execCtx.table)
		metadata.ExecutionSource = protocol.NormalizeExecutionSource(execCtx.executionSource)
		if execCtx.identity != nil {
			metadata.UserID = execCtx.identity.UserID
		}
	}
	if resource != nil {
		metadata.ResourceTable = strings.TrimSpace(resource.Table)
	}
	return metadata
}

// getPluginDataDB returns the governed record store capability host database handle.
func getPluginDataDB() (gdb.DB, error) {
	return datahost.DB()
}
