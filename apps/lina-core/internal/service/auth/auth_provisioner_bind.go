// auth_provisioner_bind.go implements the post-startup binding of the
// external-identity provider seam. Auth is constructed before the plugin
// provider managers finish wiring, so the provider cannot be a constructor
// parameter; runtime wiring calls BindExternalIdentityProvider once with the
// host manager-backed service. A nil provider keeps external login fail-closed.

package auth

import "lina-core/pkg/plugin/capability/authcap/externallogin/externalidentityspi"

// BindExternalIdentityProvider attaches the source-plugin external-identity
// provider seam. It is called once from runtime assembly with the host
// manager-backed service, which lazily resolves the enabled provider plugin
// (linapro-oidc-core) per call. A nil provider keeps external login
// fail-closed.
func (s *serviceImpl) BindExternalIdentityProvider(provider externalidentityspi.Provider) {
	if s == nil {
		return
	}
	s.identityProvider = provider
}
