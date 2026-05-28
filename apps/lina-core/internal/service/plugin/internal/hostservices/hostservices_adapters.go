// This file adapts runtime-owned host services to source-plugin service
// contracts without making public capability packages depend on internals.

package hostservices

import (
	"context"
	"sort"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/internal/service/apidoc"
	"lina-core/internal/service/auth"
	"lina-core/internal/service/datascope"
	"lina-core/internal/service/notify"
	"lina-core/internal/service/plugin/internal/runtime"
	"lina-core/internal/service/session"
	"lina-core/pkg/bizerr"
	plugincontract "lina-core/pkg/plugin/capability/contract"
	tenantcapsvc "lina-core/pkg/plugin/capability/tenantcap"
)

// apiDocAdapter bridges the internal apidoc service into the published plugin contract.
type apiDocAdapter struct {
	service APIDocResolver
}

// newAPIDocAdapter creates the source-plugin apidoc service adapter.
func newAPIDocAdapter(service APIDocResolver) plugincontract.APIDocService {
	return &apiDocAdapter{service: service}
}

// ResolveRouteText resolves one route's localized module tag and operation summary.
func (s *apiDocAdapter) ResolveRouteText(ctx context.Context, input plugincontract.RouteTextInput) plugincontract.RouteTextOutput {
	if s == nil || s.service == nil {
		return plugincontract.RouteTextOutput{Title: input.FallbackTitle, Summary: input.FallbackSummary}
	}
	output := s.service.ResolveRouteText(ctx, apidoc.RouteTextInput{
		OperationKey:    input.OperationKey,
		Method:          input.Method,
		Path:            input.Path,
		FallbackTitle:   input.FallbackTitle,
		FallbackSummary: input.FallbackSummary,
	})
	return plugincontract.RouteTextOutput{Title: output.Title, Summary: output.Summary}
}

// ResolveRouteTexts resolves multiple route texts with one apidoc catalog load.
func (s *apiDocAdapter) ResolveRouteTexts(ctx context.Context, inputs []plugincontract.RouteTextInput) []plugincontract.RouteTextOutput {
	outputs := make([]plugincontract.RouteTextOutput, 0, len(inputs))
	if s == nil || s.service == nil {
		for _, input := range inputs {
			outputs = append(outputs, plugincontract.RouteTextOutput{Title: input.FallbackTitle, Summary: input.FallbackSummary})
		}
		return outputs
	}
	internalInputs := make([]apidoc.RouteTextInput, 0, len(inputs))
	for _, input := range inputs {
		internalInputs = append(internalInputs, apidoc.RouteTextInput{
			OperationKey:    input.OperationKey,
			Method:          input.Method,
			Path:            input.Path,
			FallbackTitle:   input.FallbackTitle,
			FallbackSummary: input.FallbackSummary,
		})
	}
	for _, output := range s.service.ResolveRouteTexts(ctx, internalInputs) {
		outputs = append(outputs, plugincontract.RouteTextOutput{Title: output.Title, Summary: output.Summary})
	}
	return outputs
}

// FindRouteTitleOperationKeys finds route-title operation keys by keyword.
func (s *apiDocAdapter) FindRouteTitleOperationKeys(ctx context.Context, keyword string) []string {
	if s == nil || s.service == nil {
		return []string{}
	}
	return s.service.FindRouteTitleOperationKeys(ctx, keyword)
}

// authAdapter bridges the internal auth service into the published plugin contract.
type authAdapter struct {
	tokenIssuer TenantTokenIssuer
}

// newAuthAdapter creates the source-plugin auth service adapter.
func newAuthAdapter(tokenIssuer TenantTokenIssuer) plugincontract.AuthService {
	return &authAdapter{tokenIssuer: tokenIssuer}
}

// SelectTenant consumes a pre-login token and issues a tenant-bound token.
func (s *authAdapter) SelectTenant(ctx context.Context, in plugincontract.SelectTenantInput) (*plugincontract.TenantTokenOutput, error) {
	if s == nil || s.tokenIssuer == nil {
		return nil, bizerr.NewCode(CodePluginHostAuthTokenStateUnavailable)
	}
	out, err := s.tokenIssuer.IssueTenantToken(ctx, auth.TenantTokenIssueInput{
		PreToken: in.PreToken,
		TenantID: in.TenantID,
	})
	if err != nil {
		return nil, err
	}
	if out == nil {
		return nil, bizerr.NewCode(CodePluginHostAuthTokenStateUnavailable)
	}
	return tenantTokenOutput(out), nil
}

// SwitchTenant validates membership, revokes the current token, and issues a new token.
func (s *authAdapter) SwitchTenant(ctx context.Context, in plugincontract.SwitchTenantInput) (*plugincontract.TenantTokenOutput, error) {
	if s == nil || s.tokenIssuer == nil {
		return nil, bizerr.NewCode(CodePluginHostAuthTokenStateUnavailable)
	}
	if strings.TrimSpace(in.BearerToken) == "" {
		return nil, bizerr.NewCode(CodePluginHostAuthTokenInvalid)
	}
	out, err := s.tokenIssuer.ReissueTenantTokenFromBearer(ctx, in.BearerToken, in.TenantID)
	if err != nil {
		return nil, err
	}
	if out == nil {
		return nil, bizerr.NewCode(CodePluginHostAuthTokenStateUnavailable)
	}
	return tenantTokenOutput(out), nil
}

// IssueImpersonationToken asks the host auth service to sign and register one
// impersonation token without exposing JWT signing configuration to plugins.
func (s *authAdapter) IssueImpersonationToken(
	ctx context.Context,
	in plugincontract.ImpersonationTokenIssueInput,
) (*plugincontract.ImpersonationTokenOutput, error) {
	if s == nil || s.tokenIssuer == nil {
		return nil, bizerr.NewCode(CodePluginHostAuthTokenStateUnavailable)
	}
	out, err := s.tokenIssuer.IssueImpersonationToken(ctx, auth.ImpersonationTokenIssueInput{
		ActingUserID: in.ActingUserID,
		TenantID:     in.TenantID,
	})
	if err != nil {
		return nil, err
	}
	if out == nil {
		return nil, bizerr.NewCode(CodePluginHostAuthTokenStateUnavailable)
	}
	return &plugincontract.ImpersonationTokenOutput{
		AccessToken:  out.AccessToken,
		TokenID:      out.TokenID,
		TenantID:     out.TenantID,
		ActingUserID: out.ActingUserID,
	}, nil
}

// RevokeImpersonationToken delegates impersonation-token validation and
// session revocation to the host auth service.
func (s *authAdapter) RevokeImpersonationToken(ctx context.Context, in plugincontract.ImpersonationTokenRevokeInput) error {
	if s == nil || s.tokenIssuer == nil {
		return bizerr.NewCode(CodePluginHostAuthTokenStateUnavailable)
	}
	if strings.TrimSpace(in.BearerToken) == "" {
		return bizerr.NewCode(CodePluginHostAuthTokenInvalid)
	}
	return s.tokenIssuer.RevokeImpersonationToken(ctx, in.BearerToken, in.TenantID)
}

// tenantTokenOutput maps host auth token output into the published plugin contract.
func tenantTokenOutput(out *auth.TenantTokenOutput) *plugincontract.TenantTokenOutput {
	if out == nil {
		return nil
	}
	return &plugincontract.TenantTokenOutput{AccessToken: out.AccessToken, RefreshToken: out.RefreshToken}
}

// bizCtxAdapter bridges the internal bizctx service into the published plugin contract.
type bizCtxAdapter struct {
	service BizContextProvider
}

// newBizCtxAdapter creates the source-plugin business-context service adapter.
func newBizCtxAdapter(service BizContextProvider) plugincontract.BizCtxService {
	return &bizCtxAdapter{service: service}
}

// Current returns a read-only snapshot of the request context fields.
func (s *bizCtxAdapter) Current(ctx context.Context) plugincontract.CurrentContext {
	if s != nil && s.service != nil && ctx != nil {
		return s.service.Current(ctx)
	}
	return plugincontract.CurrentFromContext(ctx)
}

// i18nAdapter bridges the internal i18n service into the published plugin contract.
type i18nAdapter struct {
	service RuntimeI18nService
}

// newI18nAdapter creates the source-plugin i18n service adapter.
func newI18nAdapter(service RuntimeI18nService) plugincontract.I18nService {
	return &i18nAdapter{service: service}
}

// GetLocale returns the effective request locale stored in host business context.
func (s *i18nAdapter) GetLocale(ctx context.Context) string {
	if s == nil || s.service == nil {
		return ""
	}
	return s.service.GetLocale(ctx)
}

// Translate returns the localized value for one runtime i18n key and fallback text.
func (s *i18nAdapter) Translate(ctx context.Context, key string, fallback string) string {
	if s == nil || s.service == nil {
		return fallback
	}
	return s.service.Translate(ctx, key, fallback)
}

// FindMessageKeys returns runtime i18n keys under prefix whose localized value matches keyword.
func (s *i18nAdapter) FindMessageKeys(ctx context.Context, prefix string, keyword string) []string {
	if s == nil || s.service == nil {
		return []string{}
	}

	trimmedKeyword := strings.TrimSpace(keyword)
	if trimmedKeyword == "" {
		return []string{}
	}
	normalizedKeyword := strings.ToLower(trimmedKeyword)
	trimmedPrefix := strings.TrimSpace(prefix)

	messages := s.service.ExportMessages(ctx, s.service.GetLocale(ctx)).Messages
	keys := make([]string, 0)
	for key, value := range messages {
		if trimmedPrefix != "" && !strings.HasPrefix(key, trimmedPrefix) {
			continue
		}
		if strings.Contains(strings.ToLower(value), normalizedKeyword) {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return keys
}

// notifyAdapter bridges the internal notify service into the published plugin contract.
type notifyAdapter struct {
	service NotifyPublisher
}

// newNotifyAdapter creates the source-plugin notify service adapter.
func newNotifyAdapter(service NotifyPublisher) plugincontract.NotifyService {
	return &notifyAdapter{service: service}
}

// SendNoticePublication fans one published notice into the host inbox pipeline.
func (s *notifyAdapter) SendNoticePublication(ctx context.Context, in plugincontract.NoticePublishInput) (*plugincontract.SendOutput, error) {
	if s == nil || s.service == nil {
		return nil, nil
	}
	output, err := s.service.SendNoticePublication(ctx, notify.NoticePublishInput{
		NoticeID:     in.NoticeID,
		Title:        in.Title,
		Content:      in.Content,
		CategoryCode: notify.CategoryCode(in.CategoryCode),
		SenderUserID: in.SenderUserID,
	})
	if output == nil || err != nil {
		return nil, err
	}
	return &plugincontract.SendOutput{MessageID: output.MessageID, DeliveryCount: output.DeliveryCount}, nil
}

// DeleteBySource removes host notify records for the given business source identifiers.
func (s *notifyAdapter) DeleteBySource(ctx context.Context, sourceType plugincontract.SourceType, sourceIDs []string) error {
	if s == nil || s.service == nil {
		return nil
	}
	return s.service.DeleteBySource(ctx, notify.SourceType(sourceType), sourceIDs)
}

// routeAdapter bridges internal dynamic-route helpers into the published contract.
type routeAdapter struct{}

// newRouteAdapter creates the source-plugin dynamic-route service adapter.
func newRouteAdapter() plugincontract.RouteService {
	return &routeAdapter{}
}

// DynamicRouteMetadata returns metadata attached to the current dynamic-route request.
func (s *routeAdapter) DynamicRouteMetadata(request *ghttp.Request) *plugincontract.DynamicRouteMetadata {
	metadata := runtime.GetDynamicRouteMetadata(request)
	if metadata == nil {
		return nil
	}
	return &plugincontract.DynamicRouteMetadata{
		PluginID:            metadata.PluginID,
		Method:              metadata.Method,
		PublicPath:          metadata.PublicPath,
		Tags:                append([]string(nil), metadata.Tags...),
		Summary:             metadata.Summary,
		Meta:                cloneStringMap(metadata.Meta),
		ResponseBody:        metadata.ResponseBody,
		ResponseContentType: metadata.ResponseContentType,
	}
}

// sessionAdapter bridges host auth/session services into the published plugin contract.
type sessionAdapter struct {
	authSvc      AuthSessionRevoker
	scopeSvc     datascope.Service
	sessionStore session.Store
	tenantSvc    tenantcapsvc.RuntimeService
}

// newSessionAdapter creates the source-plugin session service adapter.
func newSessionAdapter(
	authSvc AuthSessionRevoker,
	scopeSvc datascope.Service,
	sessionStore session.Store,
	tenantSvc tenantcapsvc.RuntimeService,
) plugincontract.SessionService {
	return &sessionAdapter{
		authSvc:      authSvc,
		scopeSvc:     scopeSvc,
		sessionStore: sessionStore,
		tenantSvc:    tenantSvc,
	}
}

// ListPage returns one paginated online-session list for the optional filter.
func (s *sessionAdapter) ListPage(ctx context.Context, filter *plugincontract.ListFilter, pageNum, pageSize int) (*plugincontract.ListResult, error) {
	if s == nil || s.sessionStore == nil {
		return &plugincontract.ListResult{Items: []*plugincontract.Session{}, Total: 0}, nil
	}
	result, err := s.sessionStore.ListPageScoped(
		ctx,
		toInternalSessionFilter(filter),
		pageNum,
		pageSize,
		s.currentScopeSvc(),
		s.currentTenantSvc(),
	)
	if err != nil {
		return nil, err
	}
	return fromInternalSessionListResult(result), nil
}

// Revoke invalidates one online session by token ID.
func (s *sessionAdapter) Revoke(ctx context.Context, tokenID string) error {
	if s == nil {
		return nil
	}
	if s.sessionStore != nil {
		sessionItem, err := s.sessionStore.Get(ctx, tokenID)
		if err != nil {
			return err
		}
		if sessionItem != nil {
			if tenantSvc := s.currentTenantSvc(); tenantSvc != nil {
				if err = tenantSvc.EnsureTenantVisible(ctx, tenantcapsvc.TenantID(sessionItem.TenantId)); err != nil {
					return err
				}
			}
			if scopeSvc := s.currentScopeSvc(); scopeSvc != nil {
				if err = scopeSvc.EnsureUsersVisible(ctx, []int{sessionItem.UserId}); err != nil {
					return err
				}
			}
		}
	}
	if s.authSvc == nil {
		return nil
	}
	return s.authSvc.RevokeSession(ctx, tokenID)
}

// currentScopeSvc returns the shared data-scope service for plugin-facing session operations.
func (s *sessionAdapter) currentScopeSvc() datascope.Service {
	if s.scopeSvc != nil {
		return s.scopeSvc
	}
	return nil
}

// currentTenantSvc returns the shared tenant capability service for plugin-facing session operations.
func (s *sessionAdapter) currentTenantSvc() tenantcapsvc.RuntimeService {
	if s.tenantSvc != nil {
		return s.tenantSvc
	}
	return nil
}

// toInternalSessionFilter converts the published filter contract into the host-internal filter.
func toInternalSessionFilter(filter *plugincontract.ListFilter) *session.ListFilter {
	if filter == nil {
		return nil
	}
	return &session.ListFilter{Username: filter.Username, Ip: filter.Ip}
}

// fromInternalSessionListResult projects the host-internal paged session result into the plugin contract.
func fromInternalSessionListResult(result *session.ListResult) *plugincontract.ListResult {
	if result == nil {
		return &plugincontract.ListResult{Items: []*plugincontract.Session{}, Total: 0}
	}
	items := make([]*plugincontract.Session, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, fromInternalSession(item))
	}
	return &plugincontract.ListResult{Items: items, Total: result.Total}
}

// fromInternalSession copies one host-internal session projection into the plugin-facing DTO.
func fromInternalSession(sessionItem *session.Session) *plugincontract.Session {
	if sessionItem == nil {
		return nil
	}
	return &plugincontract.Session{
		TokenId:        sessionItem.TokenId,
		TenantId:       sessionItem.TenantId,
		UserId:         sessionItem.UserId,
		Username:       sessionItem.Username,
		ClientType:     sessionItem.ClientType,
		DeptName:       sessionItem.DeptName,
		Ip:             sessionItem.Ip,
		Browser:        sessionItem.Browser,
		Os:             sessionItem.Os,
		LoginTime:      sessionItem.LoginTime,
		LastActiveTime: sessionItem.LastActiveTime,
	}
}

// cloneStringMap returns a shallow copy of one string map.
func cloneStringMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return map[string]string{}
	}
	cloned := make(map[string]string, len(values))
	for key, value := range values {
		cloned[key] = value
	}
	return cloned
}
