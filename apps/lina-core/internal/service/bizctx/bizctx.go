// Package bizctx stores and mutates request-scoped host business context values
// such as authenticated user identity and resolved locale.
package bizctx

import (
	"context"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gctx"

	"lina-core/internal/model"
	"lina-core/pkg/plugin/capability/bizctxcap"
)

// ContextKey is the key for business context in request context.
const ContextKey gctx.StrKey = "BizCtx"

// Service defines request-scoped business context mutation helpers used by
// middleware and downstream services. Values are in-memory only for the active
// request and must not be treated as a durable cache.
type Service interface {
	// Init initializes and injects a model.Context into the GoFrame request
	// context before authentication and tenant middleware run.
	Init(r *ghttp.Request, ctx *model.Context)
	// Get retrieves the current business context from ctx and returns nil when
	// middleware has not initialized it.
	Get(ctx context.Context) *model.Context
	// Current returns the plugin-visible read-only projection of the current
	// business context for public pluginservice contracts.
	Current(ctx context.Context) bizctxcap.CurrentContext
	// SetLocale records the resolved request locale for downstream i18n lookups.
	SetLocale(ctx context.Context, locale string)
	// SetUser records authenticated token, user identity, and session client
	// type after token validation.
	SetUser(ctx context.Context, tokenId string, userId int, username string, status int, clientType string)
	// SetTenant records the resolved tenant boundary for tenant-aware services.
	SetTenant(ctx context.Context, tenantId int)
	// SetImpersonation records platform impersonation metadata used by tenancy
	// and permission checks.
	SetImpersonation(ctx context.Context, actingUserId int, tenantId int, actingAsTenant bool, isImpersonation bool)
	// SetUserAccess records cached role access-snapshot fields for data-scope
	// and tenant-bypass decisions during the current request.
	SetUserAccess(ctx context.Context, dataScope int, dataScopeUnsupported bool, unsupportedDataScope int)
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct{}

// New creates and returns a new Service instance.
func New() Service {
	return &serviceImpl{}
}
