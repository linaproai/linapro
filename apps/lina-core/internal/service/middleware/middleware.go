package middleware

import (
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/internal/model"
	"lina-core/internal/service/auth"
	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/operlog"
	pluginsvc "lina-core/internal/service/plugin"
	"lina-core/internal/service/role"
	"lina-core/internal/service/session"
	"lina-core/pkg/pluginhost"
)

// Service defines the middleware service contract.
type Service interface {
	// SessionStore returns the session store for external use (e.g., cleanup tasks).
	SessionStore() session.Store
	// PublishedRouteMiddlewares returns the published host middleware directory for plugin route composition.
	PublishedRouteMiddlewares() pluginhost.RouteMiddlewares
	// Ctx injects business context into request.
	Ctx(r *ghttp.Request)
	// CORS handles cross-origin requests.
	CORS(r *ghttp.Request)
	// Auth validates JWT token and injects user info into context.
	Auth(r *ghttp.Request)
	// OperLog records operation logs for write operations and specially tagged GET operations.
	OperLog(r *ghttp.Request)
	// Permission enforces declarative permission requirements declared on static host API handlers.
	Permission(r *ghttp.Request)
}

var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	authSvc    auth.Service      // Authentication service
	bizCtxSvc  bizctx.Service    // Business context service
	operLogSvc operlog.Service   // Operation log service
	pluginSvc  pluginsvc.Service // Plugin service
	roleSvc    role.Service      // Role and permission service
}

// New creates and returns a new Service instance.
func New() Service {
	return &serviceImpl{
		authSvc:    auth.New(),
		bizCtxSvc:  bizctx.New(),
		operLogSvc: operlog.New(),
		pluginSvc:  pluginsvc.New(),
		roleSvc:    role.New(),
	}
}

// SessionStore returns the session store for external use (e.g., cleanup tasks).
func (s *serviceImpl) SessionStore() session.Store {
	return s.authSvc.SessionStore()
}

// PublishedRouteMiddlewares returns the published host middleware directory for plugin route composition.
func (s *serviceImpl) PublishedRouteMiddlewares() pluginhost.RouteMiddlewares {
	if s == nil {
		return nil
	}

	return pluginhost.NewRouteMiddlewares(
		ghttp.MiddlewareNeverDoneCtx,
		ghttp.MiddlewareHandlerResponse,
		s.CORS,
		s.Ctx,
		s.Auth,
		s.OperLog,
		s.Permission,
	)
}

// Ctx injects business context into request.
func (s *serviceImpl) Ctx(r *ghttp.Request) {
	customCtx := &model.Context{}
	s.bizCtxSvc.Init(r, customCtx)
	r.Middleware.Next()
}

// CORS handles cross-origin requests.
func (s *serviceImpl) CORS(r *ghttp.Request) {
	r.Response.CORSDefault()
	r.Middleware.Next()
}

// Auth validates JWT token and injects user info into context.
func (s *serviceImpl) Auth(r *ghttp.Request) {
	tokenHeader := r.GetHeader("Authorization")
	if tokenHeader == "" {
		r.Response.WriteStatus(http.StatusUnauthorized)
		return
	}

	tokenString := strings.TrimPrefix(tokenHeader, "Bearer ")
	if tokenString == tokenHeader {
		r.Response.WriteStatus(http.StatusUnauthorized)
		return
	}

	claims, err := s.authSvc.ParseToken(r.Context(), tokenString)
	if err != nil {
		r.Response.WriteStatus(http.StatusUnauthorized)
		return
	}

	// Update last active time and validate session exists (supports forced logout and timeout cleanup)
	exists, err := s.authSvc.SessionStore().TouchOrValidate(r.Context(), claims.TokenId)
	if err != nil || !exists {
		s.roleSvc.InvalidateTokenAccessContext(r.Context(), claims.TokenId)
		r.Response.WriteStatus(http.StatusUnauthorized)
		return
	}

	// Inject user info into business context
	s.bizCtxSvc.SetUser(r.Context(), claims.TokenId, claims.UserId, claims.Username, claims.Status)
	s.pluginSvc.DispatchAfterAuthRequest(
		r.Context(),
		pluginhost.NewAfterAuthInput(
			r,
			claims.TokenId,
			claims.UserId,
			claims.Username,
			claims.Status,
		),
	)
	r.Middleware.Next()
}
