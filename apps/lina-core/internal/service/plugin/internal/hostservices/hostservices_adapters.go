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
	"lina-core/internal/service/plugin/internal/runtime"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/apidoccap"
	"lina-core/pkg/plugin/capability/authcap/token"
	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/i18ncap"
	"lina-core/pkg/plugin/capability/routecap"
)

// apiDocAdapter bridges the internal apidoc service into the published plugin contract.
type apiDocAdapter struct {
	service APIDocResolver
}

// newAPIDocAdapter creates the source-plugin apidoc service adapter.
func newAPIDocAdapter(service APIDocResolver) apidoccap.Service {
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
	tokenIssuer TenantTokenIssuer
}

// newAuthAdapter creates the source-plugin auth service adapter.
func newAuthAdapter(tokenIssuer TenantTokenIssuer) token.Service {
	return &authAdapter{tokenIssuer: tokenIssuer}
}

// SelectTenant consumes a pre-login token and issues a tenant-bound token.
func (s *authAdapter) SelectTenant(ctx context.Context, in token.SelectTenantInput) (*token.TenantTokenOutput, error) {
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
func (s *authAdapter) SwitchTenant(ctx context.Context, in token.SwitchTenantInput) (*token.TenantTokenOutput, error) {
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
	in token.ImpersonationTokenIssueInput,
) (*token.ImpersonationTokenOutput, error) {
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
	if s == nil || s.tokenIssuer == nil {
		return bizerr.NewCode(CodePluginHostAuthTokenStateUnavailable)
	}
	if strings.TrimSpace(in.BearerToken) == "" {
		return bizerr.NewCode(CodePluginHostAuthTokenInvalid)
	}
	return s.tokenIssuer.RevokeImpersonationToken(ctx, in.BearerToken, in.TenantID)
}

// tenantTokenOutput maps host auth token output into the published plugin contract.
func tenantTokenOutput(out *auth.TenantTokenOutput) *token.TenantTokenOutput {
	if out == nil {
		return nil
	}
	return &token.TenantTokenOutput{AccessToken: out.AccessToken, RefreshToken: out.RefreshToken}
}

// bizCtxAdapter bridges the internal bizctx service into the published plugin contract.
type bizCtxAdapter struct {
	service BizContextProvider
}

// newBizCtxAdapter creates the source-plugin business-context service adapter.
func newBizCtxAdapter(service BizContextProvider) bizctxcap.Service {
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
	service RuntimeI18nService
}

// newI18nAdapter creates the source-plugin i18n service adapter.
func newI18nAdapter(service RuntimeI18nService) i18ncap.Service {
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

// routeAdapter bridges internal dynamic-route helpers into the published contract.
type routeAdapter struct{}

// newRouteAdapter creates the source-plugin dynamic-route service adapter.
func newRouteAdapter() routecap.Service {
	return &routeAdapter{}
}

// DynamicRouteMetadata returns metadata attached to the current dynamic-route request.
func (s *routeAdapter) DynamicRouteMetadata(request *ghttp.Request) *routecap.DynamicRouteMetadata {
	metadata := runtime.GetDynamicRouteMetadata(request)
	if metadata == nil {
		return nil
	}
	return &routecap.DynamicRouteMetadata{
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
