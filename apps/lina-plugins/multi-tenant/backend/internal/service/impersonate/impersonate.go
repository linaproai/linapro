// Package impersonate implements the plugin-side impersonation command shape.
package impersonate

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/guid"
	"github.com/golang-jwt/jwt/v5"
	"github.com/mssola/useragent"

	"lina-core/pkg/authtoken"
	"lina-core/pkg/bizerr"
	plugincontract "lina-core/pkg/pluginservice/contract"
	"lina-plugin-multi-tenant/backend/internal/service/shared"
	tenantsvc "lina-plugin-multi-tenant/backend/internal/service/tenant"
)

// Service defines platform impersonation operations.
type Service interface {
	// Start validates an impersonation request and returns token metadata.
	Start(ctx context.Context, in StartInput) (*StartOutput, error)
	// Stop revokes one current impersonation token.
	Stop(ctx context.Context, in StopInput) error
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	bizCtxSvc   plugincontract.BizCtxService
	configSvc   plugincontract.ConfigService
	tenantSvc   tenantsvc.Service
	tokenSigner tokenSigner
}

// New creates and returns an impersonation service.
func New(
	bizCtxSvc plugincontract.BizCtxService,
	configSvc plugincontract.ConfigService,
	tenantSvc tenantsvc.Service,
) Service {
	return &serviceImpl{
		bizCtxSvc:   bizCtxSvc,
		configSvc:   configSvc,
		tenantSvc:   tenantSvc,
		tokenSigner: jwtTokenSigner{},
	}
}

// StartInput defines impersonation start input.
type StartInput struct {
	TenantID int64
	Reason   string
}

// StartOutput defines impersonation start output.
type StartOutput struct {
	Token          string
	TenantID       int64
	ActingUserID   int64
	IsImpersonated bool
}

// StopInput defines impersonation stop input.
type StopInput struct {
	TenantID int64
	Token    string
}

// userRow is the sys_user projection needed for compatible token claims.
type userRow struct {
	Id       int64  `json:"id" orm:"id"`
	Username string `json:"username" orm:"username"`
	Status   int    `json:"status" orm:"status"`
}

// tokenClaims mirrors the host JWT claim shape consumed by middleware.
type tokenClaims struct {
	TokenId         string `json:"tokenId"`
	TokenType       string `json:"tokenType"`
	UserId          int    `json:"userId"`
	Username        string `json:"username"`
	Status          int    `json:"status"`
	TenantId        int    `json:"tenantId"`
	IsImpersonation bool   `json:"isImpersonation"`
	ActingUserId    int    `json:"actingUserId"`
	jwt.RegisteredClaims
}

// tokenSigner signs and parses compatible host JWT tokens.
type tokenSigner interface {
	Sign(secret string, ttl time.Duration, user *userRow, tenantID int64, tokenID string) (string, error)
	Parse(secret string, tokenString string) (*tokenClaims, error)
}

// jwtTokenSigner signs HS256 JWT tokens compatible with the host auth service.
type jwtTokenSigner struct{}

// onlineSessionData is a typed insert payload for sys_online_session.
type onlineSessionData struct {
	TokenID        string      `orm:"token_id"`
	UserID         int64       `orm:"user_id"`
	Username       string      `orm:"username"`
	DeptName       string      `orm:"dept_name"`
	IP             string      `orm:"ip"`
	Browser        string      `orm:"browser"`
	OS             string      `orm:"os"`
	LoginTime      *gtime.Time `orm:"login_time"`
	LastActiveTime *gtime.Time `orm:"last_active_time"`
	TenantID       int64       `orm:"tenant_id"`
}

// loginLogData is a typed insert payload for plugin_monitor_loginlog.
type loginLogData struct {
	TenantID           int64  `orm:"tenant_id"`
	ActingUserID       int64  `orm:"acting_user_id"`
	OnBehalfOfTenantID int64  `orm:"on_behalf_of_tenant_id"`
	IsImpersonation    bool   `orm:"is_impersonation"`
	UserName           string `orm:"user_name"`
	Status             int    `orm:"status"`
	IP                 string `orm:"ip"`
	Browser            string `orm:"browser"`
	OS                 string `orm:"os"`
	Msg                string `orm:"msg"`
}

// operLogData is a typed insert payload for plugin_monitor_operlog.
type operLogData struct {
	TenantID           int64  `orm:"tenant_id"`
	ActingUserID       int64  `orm:"acting_user_id"`
	OnBehalfOfTenantID int64  `orm:"on_behalf_of_tenant_id"`
	IsImpersonation    bool   `orm:"is_impersonation"`
	Title              string `orm:"title"`
	OperSummary        string `orm:"oper_summary"`
	RouteOwner         string `orm:"route_owner"`
	RouteMethod        string `orm:"route_method"`
	RoutePath          string `orm:"route_path"`
	RouteDocKey        string `orm:"route_doc_key"`
	OperType           string `orm:"oper_type"`
	Method             string `orm:"method"`
	RequestMethod      string `orm:"request_method"`
	OperName           string `orm:"oper_name"`
	OperURL            string `orm:"oper_url"`
	OperIP             string `orm:"oper_ip"`
	OperParam          string `orm:"oper_param"`
	JsonResult         string `orm:"json_result"`
	Status             int    `orm:"status"`
	ErrorMsg           string `orm:"error_msg"`
	CostTime           int    `orm:"cost_time"`
}

// Start validates an impersonation request and returns token metadata.
func (s *serviceImpl) Start(ctx context.Context, in StartInput) (*StartOutput, error) {
	bizCtx := s.bizCtxSvc.Current(ctx)
	actingUserID := int64(bizCtx.UserID)
	if actingUserID <= 0 {
		return nil, bizerr.NewCode(CodeImpersonationPermissionDenied)
	}
	platformAdmin, err := s.isPlatformAdmin(ctx, actingUserID)
	if err != nil {
		return nil, err
	}
	if !platformAdmin {
		return nil, bizerr.NewCode(CodeImpersonationPermissionDenied)
	}
	tenant, err := s.tenantSvc.Get(ctx, in.TenantID)
	if err != nil {
		return nil, err
	}
	if tenant.Status != string(shared.TenantStatusActive) {
		return nil, bizerr.NewCode(CodeImpersonationTenantUnavailable)
	}
	user, err := s.currentUser(ctx, actingUserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, bizerr.NewCode(CodeImpersonationPermissionDenied)
	}
	secret, ttl, err := s.tokenConfig(ctx)
	if err != nil {
		return nil, err
	}
	tokenID := guid.S()
	token, err := s.tokenSigner.Sign(secret, ttl, user, in.TenantID, tokenID)
	if err != nil {
		return nil, err
	}
	client := clientInfoFromCtx(ctx)
	if err = s.createOnlineSession(ctx, onlineSessionData{
		TokenID:        tokenID,
		UserID:         actingUserID,
		Username:       user.Username,
		IP:             client.IP,
		Browser:        client.Browser,
		OS:             client.OS,
		LoginTime:      gtime.Now(),
		LastActiveTime: gtime.Now(),
		TenantID:       in.TenantID,
	}); err != nil {
		return nil, err
	}
	if err = s.writeAuditLogs(ctx, auditInput{
		TenantID:     in.TenantID,
		ActingUserID: actingUserID,
		Username:     user.Username,
		Reason:       in.Reason,
		Client:       client,
	}); err != nil {
		return nil, err
	}
	return &StartOutput{
		Token:          token,
		TenantID:       in.TenantID,
		ActingUserID:   actingUserID,
		IsImpersonated: true,
	}, nil
}

// Stop revokes one current impersonation token.
func (s *serviceImpl) Stop(ctx context.Context, in StopInput) error {
	tokenString := strings.TrimSpace(strings.TrimPrefix(in.Token, "Bearer "))
	if tokenString == "" {
		return bizerr.NewCode(CodeImpersonationTokenInvalid)
	}
	secret, _, err := s.tokenConfig(ctx)
	if err != nil {
		return err
	}
	claims, err := s.tokenSigner.Parse(secret, tokenString)
	if err != nil {
		return err
	}
	if claims == nil || !claims.IsImpersonation || claims.TokenId == "" {
		return bizerr.NewCode(CodeImpersonationTokenInvalid)
	}
	if in.TenantID > 0 && int64(claims.TenantId) != in.TenantID {
		return bizerr.NewCode(CodeImpersonationTokenInvalid)
	}
	_, err = shared.Model(ctx, "sys_online_session").
		Where("tenant_id", claims.TenantId).
		Where("token_id", claims.TokenId).
		Delete()
	return err
}

// Sign signs one compatible impersonation JWT.
func (jwtTokenSigner) Sign(secret string, ttl time.Duration, user *userRow, tenantID int64, tokenID string) (string, error) {
	if strings.TrimSpace(secret) == "" || user == nil {
		return "", bizerr.NewCode(CodeImpersonationTokenUnavailable)
	}
	now := time.Now()
	claims := tokenClaims{
		TokenId:         tokenID,
		TokenType:       authtoken.KindAccess,
		UserId:          int(user.Id),
		Username:        user.Username,
		Status:          user.Status,
		TenantId:        int(tenantID),
		IsImpersonation: true,
		ActingUserId:    int(user.Id),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		return "", err
	}
	return signed, nil
}

// Parse parses one compatible impersonation JWT.
func (jwtTokenSigner) Parse(secret string, tokenString string) (*tokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &tokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, bizerr.NewCode(CodeImpersonationTokenInvalid)
	}
	claims, ok := token.Claims.(*tokenClaims)
	if !ok || !token.Valid {
		return nil, bizerr.NewCode(CodeImpersonationTokenInvalid)
	}
	return claims, nil
}

// tokenConfig reads signing configuration from the host-published config service.
func (s *serviceImpl) tokenConfig(ctx context.Context) (string, time.Duration, error) {
	secret, err := s.configSvc.String(ctx, "jwt.secret", "")
	if err != nil {
		return "", 0, err
	}
	if strings.TrimSpace(secret) == "" {
		return "", 0, bizerr.NewCode(CodeImpersonationTokenUnavailable)
	}
	ttl, err := s.configSvc.Duration(ctx, "jwt.expire", 24*time.Hour)
	if err != nil {
		return "", 0, err
	}
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	return secret, ttl, nil
}

// currentUser returns the current platform user projection.
func (s *serviceImpl) currentUser(ctx context.Context, userID int64) (*userRow, error) {
	var user *userRow
	err := shared.Model(ctx, shared.TableSysUser).
		Fields("id", "username", "status").
		Where("id", userID).
		Scan(&user)
	return user, err
}

// isPlatformAdmin reports whether userID is bound to an all-data role in platform context.
func (s *serviceImpl) isPlatformAdmin(ctx context.Context, userID int64) (bool, error) {
	count, err := shared.Model(ctx, "sys_user_role").As("ur").
		InnerJoin("sys_role r", "r.id = ur.role_id").
		Where("ur.user_id", userID).
		Where("ur.tenant_id", shared.PlatformTenantID).
		Where("r.data_scope", 1).
		Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// createOnlineSession creates the session row required by host middleware.
func (s *serviceImpl) createOnlineSession(ctx context.Context, data onlineSessionData) error {
	_, err := shared.Model(ctx, "sys_online_session").Data(data).Insert()
	return err
}

// auditInput defines impersonation audit fields.
type auditInput struct {
	TenantID     int64
	ActingUserID int64
	Username     string
	Reason       string
	Client       clientInfo
}

// clientInfo contains normalized request client metadata.
type clientInfo struct {
	IP      string
	Browser string
	OS      string
	URL     string
}

// writeAuditLogs writes optional login and operation log rows when monitor tables exist.
func (s *serviceImpl) writeAuditLogs(ctx context.Context, in auditInput) error {
	tables, err := g.DB().Tables(ctx)
	if err != nil {
		return err
	}
	exists := make(map[string]struct{}, len(tables))
	for _, table := range tables {
		exists[table] = struct{}{}
	}
	if _, ok := exists["plugin_monitor_loginlog"]; ok {
		if _, err = shared.Model(ctx, "plugin_monitor_loginlog").Data(loginLogData{
			TenantID:           in.TenantID,
			ActingUserID:       in.ActingUserID,
			OnBehalfOfTenantID: in.TenantID,
			IsImpersonation:    true,
			UserName:           in.Username,
			Status:             0,
			IP:                 in.Client.IP,
			Browser:            in.Client.Browser,
			OS:                 in.Client.OS,
			Msg:                "Impersonation started",
		}).Insert(); err != nil {
			return err
		}
	}
	if _, ok := exists["plugin_monitor_operlog"]; ok {
		if _, err = shared.Model(ctx, "plugin_monitor_operlog").Data(operLogData{
			TenantID:           in.TenantID,
			ActingUserID:       in.ActingUserID,
			OnBehalfOfTenantID: in.TenantID,
			IsImpersonation:    true,
			Title:              "Tenant Impersonation",
			OperSummary:        in.Reason,
			RouteOwner:         "multi-tenant",
			RouteMethod:        "POST",
			RoutePath:          "/platform/tenants/{id}/impersonate",
			RouteDocKey:        "platform.tenant.impersonate",
			OperType:           "other",
			Method:             "impersonate.Start",
			RequestMethod:      "POST",
			OperName:           in.Username,
			OperURL:            in.Client.URL,
			OperIP:             in.Client.IP,
			OperParam:          "{}",
			JsonResult:         "{}",
			Status:             0,
			ErrorMsg:           "",
			CostTime:           0,
		}).Insert(); err != nil {
			return err
		}
	}
	return nil
}

// clientInfoFromCtx extracts browser, OS, IP, and URL metadata from the request.
func clientInfoFromCtx(ctx context.Context) clientInfo {
	request := g.RequestFromCtx(ctx)
	if request == nil {
		return clientInfo{}
	}
	browser, osName := parseUserAgent(request)
	return clientInfo{
		IP:      request.GetClientIp(),
		Browser: browser,
		OS:      osName,
		URL:     request.URL.String(),
	}
}

// parseUserAgent parses browser and OS names from a request.
func parseUserAgent(request *ghttp.Request) (string, string) {
	if request == nil {
		return "", ""
	}
	ua := useragent.New(request.GetHeader("User-Agent"))
	browserName, browserVersion := ua.Browser()
	return strings.TrimSpace(browserName + " " + browserVersion), ua.OS()
}
