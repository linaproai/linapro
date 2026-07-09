// This file owns the plugin-scoped data service table contract cache and binds
// entries to the plugin migration ledger plus host-service authorization shape.

package datahost

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"
	"sync"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/pkg/logger"
)

// tableContractCacheKey identifies one authorized table contract cache entry.
type tableContractCacheKey struct {
	pluginID                 string
	table                    string
	migrationFingerprint     string
	authorizationFingerprint string
}

// tableContractCacheStore holds detached resource contracts for one process.
type tableContractCacheStore struct {
	mu      sync.RWMutex
	entries map[tableContractCacheKey]*catalog.ResourceSpec
}

// tableContractCache stores process-local table contracts with versioned keys.
var tableContractCache = &tableContractCacheStore{}

// InvalidateTableContractCache removes cached table contracts for one plugin.
func InvalidateTableContractCache(ctx context.Context, pluginID string, reason string) {
	normalizedPluginID := strings.TrimSpace(pluginID)
	if normalizedPluginID == "" {
		return
	}
	invalidated := tableContractCache.invalidatePlugin(normalizedPluginID)
	logger.Debugf(
		ctx,
		"plugin datahost table contract cache invalidated plugin=%s entries=%d reason=%s",
		normalizedPluginID,
		invalidated,
		strings.TrimSpace(reason),
	)
}

func (s *tableContractCacheStore) get(key tableContractCacheKey) *catalog.ResourceSpec {
	if s == nil || !key.valid() {
		return nil
	}
	s.mu.RLock()
	resource := catalog.CloneResourceSpec(s.entries[key])
	s.mu.RUnlock()
	return resource
}

func (s *tableContractCacheStore) set(key tableContractCacheKey, resource *catalog.ResourceSpec) {
	if s == nil || !key.valid() || resource == nil {
		return
	}
	s.mu.Lock()
	if s.entries == nil {
		s.entries = make(map[tableContractCacheKey]*catalog.ResourceSpec)
	}
	s.entries[key] = catalog.CloneResourceSpec(resource)
	s.mu.Unlock()
}

func (s *tableContractCacheStore) invalidatePlugin(pluginID string) int {
	if s == nil || pluginID == "" {
		return 0
	}
	s.mu.Lock()
	invalidated := 0
	for key := range s.entries {
		if key.pluginID == pluginID {
			delete(s.entries, key)
			invalidated++
		}
	}
	s.mu.Unlock()
	return invalidated
}

func (key tableContractCacheKey) valid() bool {
	return key.pluginID != "" &&
		key.table != "" &&
		key.migrationFingerprint != "" &&
		key.authorizationFingerprint != ""
}

func buildTableContractAuthorizationFingerprint(methods []string) string {
	normalizedMethods := normalizeAuthorizedTableMethods(methods)
	if len(normalizedMethods) == 0 {
		return "methods:none"
	}
	return "methods:" + strings.Join(normalizedMethods, ",")
}

func buildTableContractMigrationFingerprint(ctx context.Context, pluginID string) (string, error) {
	pluginID = strings.TrimSpace(pluginID)
	if pluginID == "" {
		return plugintypes.MigrationStateNone.String(), nil
	}

	cols := dao.SysPluginMigration.Columns()
	var rows []*entity.SysPluginMigration
	err := dao.SysPluginMigration.Ctx(ctx).
		Fields(
			cols.Id,
			cols.ReleaseId,
			cols.Phase,
			cols.MigrationKey,
			cols.Checksum,
			cols.ExecutionOrder,
			cols.Status,
		).
		Where(do.SysPluginMigration{PluginId: pluginID}).
		WhereNot(cols.Phase, plugintypes.MigrationDirectionMock.String()).
		OrderAsc(cols.Id).
		Scan(&rows)
	if err != nil {
		return "", err
	}
	if len(rows) == 0 {
		return plugintypes.MigrationStateNone.String(), nil
	}

	hash := sha256.New()
	for _, row := range rows {
		if row == nil {
			continue
		}
		fmt.Fprintf(
			hash,
			"%d|%d|%s|%s|%s|%d|%s\n",
			row.Id,
			row.ReleaseId,
			row.Phase,
			row.MigrationKey,
			row.Checksum,
			row.ExecutionOrder,
			row.Status,
		)
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
