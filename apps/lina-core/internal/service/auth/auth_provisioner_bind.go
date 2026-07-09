// auth_provisioner_bind.go implements the post-startup binding of the
// user-owner external provisioning seam. Auth is constructed before the user
// service during runtime assembly, so the provisioner cannot be a constructor
// parameter; runtime wiring calls BindExternalProvisioner once both services
// exist. A nil provisioner keeps auto-provisioning disabled fail-closed.

package auth

// BindExternalProvisioner attaches the user-owner provisioning seam. It is
// called once from runtime assembly after the user service is constructed.
func (s *serviceImpl) BindExternalProvisioner(provisioner ExternalProvisioner) {
	if s == nil {
		return
	}
	s.provisioner = provisioner
}
