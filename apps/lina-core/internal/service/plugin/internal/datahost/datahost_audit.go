// This file builds plugin data audit metadata and bridges the datahost package
// to the reusable plugindb host-side governance layer.

package datahost

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/pkg/pluginbridge"
	plugindbhost "lina-core/pkg/plugindb/host"
)

func withPluginDataAudit(ctx context.Context, metadata *plugindbhost.AuditMetadata) context.Context {
	return plugindbhost.WithAudit(ctx, metadata)
}

func buildPluginDataAuditMetadata(
	execCtx *executionContext,
	resource *catalog.ResourceSpec,
	method string,
	inTransaction bool,
) *plugindbhost.AuditMetadata {
	metadata := &plugindbhost.AuditMetadata{
		Method:      strings.ToLower(strings.TrimSpace(method)),
		Transaction: inTransaction,
	}
	if execCtx != nil {
		metadata.PluginID = strings.TrimSpace(execCtx.pluginID)
		metadata.Table = strings.TrimSpace(execCtx.table)
		metadata.ExecutionSource = pluginbridge.NormalizeExecutionSource(execCtx.executionSource)
		if execCtx.identity != nil {
			metadata.UserID = execCtx.identity.UserID
		}
	}
	if resource != nil {
		metadata.ResourceTable = strings.TrimSpace(resource.Table)
	}
	return metadata
}

func getPluginDataDB() (gdb.DB, error) {
	return plugindbhost.DB()
}
