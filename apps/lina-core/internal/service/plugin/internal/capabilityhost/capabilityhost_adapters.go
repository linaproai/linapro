// This file adapts runtime-owned host services to source-plugin service
// contracts without making public capability packages depend on internals.

package capabilityhost

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/frame/g"

	"lina-core/internal/service/apidoc"
	"lina-core/internal/service/auth"
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/internal/service/plugin/internal/runtime"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/apidoccap"
	"lina-core/pkg/plugin/capability/authcap/externallogin"
	"lina-core/pkg/plugin/capability/authcap/token"
	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/i18ncap"
	"lina-core/pkg/plugin/capability/routecap"
	"lina-core/pkg/plugin/pluginhost"
)

// apiDocAdapter bridges the internal apidoc service into the published plugin contract.
type apiDocAdapter struct {
	service apidoc.Service
}

// newAPIDocAdapter creates the source-plugin apidoc service adapter.
func newAPIDocAdapter(service apidoc.Service) apidoccap.Service {
	return &apiDocAdapter{service: service}
}

// ResolveRouteText resolves one route's localized module tag and operation summary.
func (s *apiDocAdapter) ResolveRouteText(ctx context.Context, input apidoccap.RouteTextInput) apidoccap.RouteTextOutput {
	if s == nil || s.service == nil {
		return apidoccap.RouteTextOutput{Title: input.FallbackTitle, Summary: input.FallbackSummary}
	}
	output := s.service.ResolveRouteText(ctx, apidoc.RouteTextInput{
		OperationKey:    input.OperationKey,
		Method:          input.Method,
		Path:            input.Path,
		FallbackTitle:   input.FallbackTitle,
		FallbackSummary: input.FallbackSummary,
	})
	return apidoccap.RouteTextOutput{Title: output.Title, Summary: output.Summary}
}

// ResolveRouteTexts resolves multiple route texts with one apidoc catalog load.
func (s *apiDocAdapter) ResolveRouteTexts(ctx context.Context, inputs []apidoccap.RouteTextInput) []apidoccap.RouteTextOutput {
	outputs := make([]apidoccap.RouteTextOutput, 0, len(inputs))
	if s == nil || s.service == nil {
		for _, input := range inputs {
			outputs = append(outputs, apidoccap.RouteTextOutput{Title: input.FallbackTitle, Summary: input.FallbackSummary})
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
		outputs = append(outputs, apidoccap.RouteTextOutput{Title: output.Title, Summary: output.Summary})
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
	authSvc auth.Service
}

// newAuthAdapter creates the source-plugin auth service adapter.
func newAuthAdapter(authSvc auth.Service) token.Service {
	return &authAdapter{authSvc: authSvc}
}

// SelectTenant consumes a pre-login token and issues a tenant-bound token.
func (s *authAdapter) SelectTenant(ctx context.Context, in token.SelectTenantInput) (*token.TenantTokenOutput, error) {
	if s == nil || s.authSvc == nil {
		return nil, bizerr.NewCode(CodePluginHostAuthTokenStateUnavailable)
	}
	out, err := s.authSvc.IssueTenantToken(ctx, auth.TenantTokenIssueInput{
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
func (s *authAdapter) SwitchTenant(ctx context.Context, in token.SwitchTenantInput) (*token.TenantTokenOutput, error) {
	if s == nil || s.authSvc == nil {
		return nil, bizerr.NewCode(CodePluginHostAuthTokenStateUnavailable)
	}
	if strings.TrimSpace(in.BearerToken) == "" {
		return nil, bizerr.NewCode(CodePluginHostAuthTokenInvalid)
	}
	out, err := s.authSvc.ReissueTenantTokenFromBearer(ctx, in.BearerToken, in.TenantID)
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
	in token.ImpersonationTokenIssueInput,
) (*token.ImpersonationTokenOutput, error) {
	if s == nil || s.authSvc == nil {
		return nil, bizerr.NewCode(CodePluginHostAuthTokenStateUnavailable)
	}
	out, err := s.authSvc.IssueImpersonationToken(ctx, auth.ImpersonationTokenIssueInput{
		ActingUserID: in.ActingUserID,
		TenantID:     in.TenantID,
	})
	if err != nil {
		return nil, err
	}
	if out == nil {
		return nil, bizerr.NewCode(CodePluginHostAuthTokenStateUnavailable)
	}
	return &token.ImpersonationTokenOutput{
		AccessToken:  out.AccessToken,
		TokenID:      out.TokenID,
		TenantID:     out.TenantID,
		ActingUserID: out.ActingUserID,
	}, nil
}

// RevokeImpersonationToken delegates impersonation-token validation and
// session revocation to the host auth service.
func (s *authAdapter) RevokeImpersonationToken(ctx context.Context, in token.ImpersonationTokenRevokeInput) error {
	if s == nil || s.authSvc == nil {
		return bizerr.NewCode(CodePluginHostAuthTokenStateUnavailable)
	}
	if strings.TrimSpace(in.BearerToken) == "" {
		return bizerr.NewCode(CodePluginHostAuthTokenInvalid)
	}
	return s.authSvc.RevokeImpersonationToken(ctx, in.BearerToken, in.TenantID)
}

// tenantTokenOutput maps host auth token output into the published plugin contract.
func tenantTokenOutput(out *auth.TenantTokenOutput) *token.TenantTokenOutput {
	if out == nil {
		return nil
	}
	return &token.TenantTokenOutput{AccessToken: out.AccessToken, RefreshToken: out.RefreshToken}
}

// externalLoginAdapter bridges the host auth service into the plugin-facing
// external-login contract. It is plugin-scoped: the base adapter is unbound and
// fail-closed, and forPlugin returns a copy bound to one source-plugin identity.
// Provider ownership and plugin enablement are enforced here so the published
// contract never trusts a plugin-supplied plugin identity.
type externalLoginAdapter struct {
	authSvc     auth.Service
	pluginState pluginStateLookup
	pluginID    string
}

// newExternalLoginAdapter creates the unbound base external-login adapter. The
// base adapter fail-closes on every call because it is not bound to a plugin;
// callers reach a usable service only through the plugin-scoped directory.
func newExternalLoginAdapter(authSvc auth.Service, pluginState pluginStateLookup) *externalLoginAdapter {
	return &externalLoginAdapter{authSvc: authSvc, pluginState: pluginState}
}

// forPlugin returns an external-login service bound to one source plugin.
func (s *externalLoginAdapter) forPlugin(pluginID string) externallogin.Service {
	if s == nil {
		return nil
	}
	return &externalLoginAdapter{
		authSvc:     s.authSvc,
		pluginState: s.pluginState,
		pluginID:    strings.TrimSpace(pluginID),
	}
}

// LoginByVerifiedIdentity enforces plugin binding, provider ownership, and
// plugin enablement before delegating to the host auth service. The host-owned
// pluginID is stamped onto the host input so the plugin cannot spoof it.
func (s *externalLoginAdapter) LoginByVerifiedIdentity(
	ctx context.Context,
	in externallogin.LoginInput,
) (*externallogin.LoginOutput, error) {
	if s == nil || s.authSvc == nil {
		return nil, bizerr.NewCode(CodePluginHostAuthTokenStateUnavailable)
	}
	pluginID := strings.TrimSpace(s.pluginID)
	if pluginID == "" {
		return nil, bizerr.NewCode(CodePluginHostExternalLoginPluginRequired)
	}
	provider := strings.TrimSpace(in.Provider)
	if !s.ownsProvider(pluginID, provider) {
		return nil, bizerr.NewCode(CodePluginHostExternalLoginProviderForbidden)
	}
	if s.pluginState == nil || !s.pluginState.IsEnabled(ctx, pluginID) {
		return nil, bizerr.NewCode(CodePluginHostExternalLoginPluginDisabled)
	}
	out, err := s.authSvc.LoginByExternalIdentity(ctx, auth.ExternalLoginInput{
		PluginID:           pluginID,
		Provider:           provider,
		Subject:            in.Subject,
		Email:              in.Email,
		DisplayName:        in.DisplayName,
		ClientType:         token.ClientTypeWeb,
		AllowAutoProvision: in.AllowAutoProvision,
	})
	if err != nil {
		return nil, err
	}
	if out == nil {
		return nil, bizerr.NewCode(CodePluginHostAuthTokenStateUnavailable)
	}
	return externalLoginOutput(out), nil
}

// ownsProvider reports whether pluginID declared providerID through
// ProvideExternalIdentity at declaration time.
func (s *externalLoginAdapter) ownsProvider(pluginID string, providerID string) bool {
	if providerID == "" {
		return false
	}
	definition, ok := pluginhost.GetSourcePlugin(pluginID)
	if !ok || definition == nil {
		return false
	}
	for _, owned := range definition.GetExternalIdentityProviderIDs() {
		if owned == providerID {
			return true
		}
	}
	return false
}

// externalLoginOutput maps host external-login output into the published contract.
func externalLoginOutput(out *auth.ExternalLoginOutput) *externallogin.LoginOutput {
	if out == nil {
		return nil
	}
	candidates := make([]externallogin.TenantCandidate, 0, len(out.Tenants))
	for _, tenant := range out.Tenants {
		candidates = append(candidates, externallogin.TenantCandidate{
			ID:     tenant.Id,
			Code:   tenant.Code,
			Name:   tenant.Name,
			Status: tenant.Status,
		})
	}
	return &externallogin.LoginOutput{
		AccessToken:      out.AccessToken,
		RefreshToken:     out.RefreshToken,
		PreToken:         out.PreToken,
		TenantCandidates: candidates,
	}
}

// bizCtxAdapter bridges the internal bizctx service into the published plugin contract.
type bizCtxAdapter struct {
	service bizctxcap.Service
}

// newBizCtxAdapter creates the source-plugin business-context service adapter.
func newBizCtxAdapter(service bizctxcap.Service) bizctxcap.Service {
	return &bizCtxAdapter{service: service}
}

// Current returns a read-only snapshot of the request context fields.
func (s *bizCtxAdapter) Current(ctx context.Context) bizctxcap.CurrentContext {
	if s != nil && s.service != nil && ctx != nil {
		return s.service.Current(ctx)
	}
	return bizctxcap.CurrentFromContext(ctx)
}

// i18nAdapter bridges the internal i18n service into the published plugin contract.
type i18nAdapter struct {
	service i18nsvc.Service
}

// newI18nAdapter creates the source-plugin i18n service adapter.
func newI18nAdapter(service i18nsvc.Service) i18ncap.Service {
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

// routeAdapter bridges internal dynamic-route helpers into the published contract.
type routeAdapter struct{}

// newRouteAdapter creates the source-plugin dynamic-route service adapter.
func newRouteAdapter() routecap.Service {
	return &routeAdapter{}
}

// GetMetadata returns metadata attached to the current dynamic-route request.
func (s *routeAdapter) GetMetadata(ctx context.Context) *routecap.Metadata {
	request := g.RequestFromCtx(ctx)
	metadata := runtime.GetMetadata(request)
	if metadata == nil {
		return nil
	}
	return &routecap.Metadata{
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
