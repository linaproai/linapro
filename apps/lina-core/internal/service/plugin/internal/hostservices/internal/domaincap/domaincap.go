// Package domaincap contains shared primitives used by host-owned domain
// capability adapters. It does not define concrete capability services.
package domaincap

import (
	"context"
	"strconv"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/capmodel"
)

const (
	// DefaultPageNum is the fallback page number for domain list adapters.
	DefaultPageNum = 1
	// DefaultPageSize is the fallback page size for domain list adapters.
	DefaultPageSize = 20
	// MaxPageSize is the hard page-size limit for domain list adapters.
	MaxPageSize = 200

	// PlatformTenantID identifies platform-global rows.
	PlatformTenantID = 0

	// PluginRuntimeCacheDomain is the shared plugin-runtime cache revision domain.
	PluginRuntimeCacheDomain = "plugin-runtime"
	// PluginRuntimeCacheScopeGlobal is the global plugin-runtime revision scope.
	PluginRuntimeCacheScopeGlobal = "global"
	// TenantPluginRuntimeChangeReason is the revision reason for tenant plugin state changes.
	TenantPluginRuntimeChangeReason = "tenant_plugin_enablement_changed"
	// AuthorizationCacheDomain is the shared authorization cache revision domain.
	AuthorizationCacheDomain = "permission-access"
	// AuthorizationCacheScopeGlobal is the global authorization revision scope.
	AuthorizationCacheScopeGlobal = "global"
	// AuthorizationChangeReason is the revision reason for authorization changes.
	AuthorizationChangeReason = "authorization_changed"
	// DictionaryCacheDomain is the shared dictionary cache revision domain.
	DictionaryCacheDomain = "dictionary"
	// DictionaryRefreshReason is the revision reason for dictionary refreshes.
	DictionaryRefreshReason = "dictionary_refreshed"
	// RuntimeConfigDomain is the shared runtime-config cache revision domain.
	RuntimeConfigDomain = "runtime-config"
	// RuntimeConfigChangeReason is the revision reason for runtime-config changes.
	RuntimeConfigChangeReason = "runtime_config_changed"
	// RuntimeConfigGlobalScope is the shared runtime-config revision scope.
	RuntimeConfigGlobalScope = "global"
	// AuthorizationPlatformAllDataScope is the role data-scope value for platform admins.
	AuthorizationPlatformAllDataScope = 1
)

// NormalizePage applies conservative defaults and hard limits to one page request.
func NormalizePage(page capmodel.PageRequest) (int, int) {
	pageNum := page.PageNum
	if pageNum <= 0 {
		pageNum = DefaultPageNum
	}
	pageSize := page.PageSize
	if pageSize <= 0 {
		pageSize = page.Limit
	}
	if pageSize <= 0 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}
	return pageNum, pageSize
}

// TenantID decodes a tenant domain ID into the host integer tenant key.
func TenantID(value capmodel.DomainID) (int, error) {
	tenantID, err := strconv.Atoi(strings.TrimSpace(string(value)))
	if err != nil {
		return 0, err
	}
	return tenantID, nil
}

// BumpSharedRevision advances one cache-revision row in the caller's transaction.
func BumpSharedRevision(ctx context.Context, tx gdb.TX, domain string, scope string, reason string) error {
	_, err := tx.Model(dao.SysCacheRevision.Table()).Safe().Ctx(ctx).Data(do.SysCacheRevision{
		TenantId: PlatformTenantID,
		Domain:   domain,
		Scope:    scope,
		Revision: 0,
		Reason:   reason,
	}).InsertIgnore()
	if err != nil {
		return err
	}
	var row *entity.SysCacheRevision
	err = tx.Model(dao.SysCacheRevision.Table()).Safe().Ctx(ctx).
		Fields(dao.SysCacheRevision.Columns().Id, dao.SysCacheRevision.Columns().Revision).
		Where(do.SysCacheRevision{
			TenantId: PlatformTenantID,
			Domain:   domain,
			Scope:    scope,
		}).
		LockUpdate().
		Scan(&row)
	if err != nil {
		return err
	}
	if row == nil {
		return bizerr.NewCode(capmodel.CodeCapabilityUnavailable, bizerr.P("capability", domain+"-revision"))
	}
	_, err = tx.Model(dao.SysCacheRevision.Table()).Safe().Ctx(ctx).
		Where(do.SysCacheRevision{Id: row.Id}).
		Data(do.SysCacheRevision{
			Revision: row.Revision + 1,
			Reason:   reason,
		}).
		Update()
	return err
}

// ParseInt64IDs decodes string domain IDs into host-owned int64 keys.
func ParseInt64IDs[ID ~string](ids []ID, invalid func(ID)) ([]int64, map[int64]ID) {
	parsedIDs := make([]int64, 0, len(ids))
	requested := make(map[int64]ID, len(ids))
	for _, id := range ids {
		parsedID, err := strconv.ParseInt(strings.TrimSpace(string(id)), 10, 64)
		if err != nil || parsedID <= 0 {
			if invalid != nil {
				invalid(id)
			}
			continue
		}
		if _, exists := requested[parsedID]; exists {
			continue
		}
		requested[parsedID] = id
		parsedIDs = append(parsedIDs, parsedID)
	}
	return parsedIDs, requested
}

// ParsePositiveInt64Strings decodes positive host-owned integer IDs.
func ParsePositiveInt64Strings(values []string) ([]int64, error) {
	ids := make([]int64, 0, len(values))
	seen := make(map[int64]struct{}, len(values))
	for _, value := range values {
		parsedID, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
		if err != nil || parsedID <= 0 {
			if err != nil {
				return nil, err
			}
			return nil, bizerr.NewCode(capmodel.CodeCapabilityDenied)
		}
		if _, exists := seen[parsedID]; exists {
			continue
		}
		seen[parsedID] = struct{}{}
		ids = append(ids, parsedID)
	}
	return ids, nil
}

// FirstNonEmpty returns the first non-empty value.
func FirstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

// MimeTypeFromSuffix returns a stable coarse media type for plugin projections.
func MimeTypeFromSuffix(suffix string) string {
	normalizedSuffix := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(suffix)), ".")
	switch normalizedSuffix {
	case "jpg", "jpeg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "gif":
		return "image/gif"
	case "pdf":
		return "application/pdf"
	case "txt", "log":
		return "text/plain"
	case "json":
		return "application/json"
	default:
		return ""
	}
}

// Contains reports whether values already contains value.
func Contains[Value comparable](values []Value, value Value) bool {
	for _, existing := range values {
		if existing == value {
			return true
		}
	}
	return false
}
