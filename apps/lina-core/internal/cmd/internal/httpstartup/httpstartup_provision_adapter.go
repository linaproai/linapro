// httpstartup_provision_adapter.go narrows the user owner service to the auth
// ExternalProvisioner boundary so auth can auto-provision platform users for
// verified external identities without importing the user package (user
// already depends on auth for password hashing).

package httpstartup

import (
	"context"

	"lina-core/internal/service/auth"
	"lina-core/internal/service/user"
)

// externalProvisionAdapter adapts user.Service to auth.ExternalProvisioner.
type externalProvisionAdapter struct {
	// userSvc is the user domain owner that shapes provisioned accounts.
	userSvc user.Service
}

// ProvisionExternalUser delegates to the user owner's system provisioning path.
func (a externalProvisionAdapter) ProvisionExternalUser(ctx context.Context, in auth.ExternalProvisionInput) (int, error) {
	return a.userSvc.ProvisionExternalUser(ctx, user.ProvisionExternalInput{
		Email:       in.Email,
		DisplayName: in.DisplayName,
		Remark:      in.Remark,
	})
}
