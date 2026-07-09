// user_provision_external.go implements system-level user provisioning for
// verified external identities (external-login auto-provision). Unlike the
// operator-facing Create path, provisioning has no acting operator: the
// username is derived from the verified email local part with numeric
// de-duplication, the password is random and unusable (external-login only),
// and no roles or tenants are assigned so the new account starts with least
// privilege. Authorization for WHEN provisioning may run stays with the host
// auth owner; this file only owns HOW a provisioned user is shaped.

package user

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/logger"
	"lina-core/pkg/statusflag"
)

// provisionUsernameMaxLength bounds the generated username so derived names
// plus de-duplication suffixes stay inside the sys_user column length.
const provisionUsernameMaxLength = 30

// provisionUsernameMaxAttempts bounds the numeric de-duplication loop so a
// pathological collision storm fails fast instead of scanning forever.
const provisionUsernameMaxAttempts = 20

// provisionPasswordByteLength sizes the random unusable password. The value
// is never disclosed; the account can only sign in through external login
// until an administrator resets the password.
const provisionPasswordByteLength = 32

// ProvisionExternalUser creates one platform user for a verified external
// identity. See the Service interface for the full contract.
func (s *serviceImpl) ProvisionExternalUser(ctx context.Context, in ProvisionExternalInput) (int, error) {
	email := strings.TrimSpace(in.Email)
	if email == "" || !strings.Contains(email, "@") {
		return 0, bizerr.NewCode(CodeUserProvisionEmailInvalid)
	}
	username, err := s.resolveProvisionUsername(ctx, email)
	if err != nil {
		return 0, err
	}
	randomPassword, err := generateUnusablePassword()
	if err != nil {
		return 0, bizerr.WrapCode(err, CodeUserProvisionFailed)
	}
	hash, err := s.authSvc.HashPassword(randomPassword)
	if err != nil {
		return 0, bizerr.WrapCode(err, CodeUserProvisionFailed)
	}
	nickname := strings.TrimSpace(in.DisplayName)
	if nickname == "" {
		nickname = username
	}
	remark := strings.TrimSpace(in.Remark)
	if remark == "" {
		remark = "auto-provisioned by external login"
	}
	id, err := dao.SysUser.Ctx(ctx).Data(do.SysUser{
		Username: username,
		Password: hash,
		Nickname: nickname,
		Email:    email,
		Sex:      0,
		Status:   statusflag.EnabledValue.Int(),
		Remark:   remark,
		TenantId: 0,
	}).InsertAndGetId()
	if err != nil {
		return 0, bizerr.WrapCode(err, CodeUserProvisionFailed)
	}
	logger.Infof(ctx, "external login provisioned user id=%d username=%s email=%s", id, username, email)
	return int(id), nil
}

// resolveProvisionUsername derives a unique username from the email local
// part, appending numeric suffixes on collision.
func (s *serviceImpl) resolveProvisionUsername(ctx context.Context, email string) (string, error) {
	base := strings.ToLower(strings.TrimSpace(strings.SplitN(email, "@", 2)[0]))
	// Keep only characters that are safe for login names; collapse everything
	// else so IdP-specific punctuation cannot produce awkward usernames.
	var builder strings.Builder
	for _, r := range base {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			builder.WriteRune(r)
		case r == '.' || r == '-' || r == '_':
			builder.WriteRune(r)
		}
	}
	candidateBase := strings.Trim(builder.String(), "._-")
	if candidateBase == "" {
		candidateBase = "external-user"
	}
	if len(candidateBase) > provisionUsernameMaxLength-3 {
		candidateBase = candidateBase[:provisionUsernameMaxLength-3]
	}
	candidate := candidateBase
	for attempt := 0; attempt < provisionUsernameMaxAttempts; attempt++ {
		if attempt > 0 {
			candidate = fmt.Sprintf("%s%d", candidateBase, attempt+1)
		}
		count, err := dao.SysUser.Ctx(ctx).
			Where(do.SysUser{Username: candidate}).
			Count()
		if err != nil {
			return "", bizerr.WrapCode(err, CodeUserProvisionFailed)
		}
		if count == 0 {
			return candidate, nil
		}
	}
	return "", bizerr.WrapCode(
		gerror.Newf("username derivation exhausted %d attempts for base %q", provisionUsernameMaxAttempts, candidateBase),
		CodeUserProvisionFailed,
	)
}

// generateUnusablePassword returns a high-entropy random password that is
// never disclosed to anyone, making password login impossible until an
// administrator explicitly resets it.
func generateUnusablePassword() (string, error) {
	buffer := make([]byte, provisionPasswordByteLength)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return hex.EncodeToString(buffer), nil
}
