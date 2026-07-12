// user_create_from_external.go implements system-level user creation for
// verified external identities (external-login auto-create). Unlike the
// operator-facing Create path, CreateFromExternalUser has no acting operator:
// the username is derived from the verified email local part with numeric
// de-duplication, the password is random and unusable (external-login only),
// and no roles or tenants are assigned so the new account starts with least
// privilege. Authorization for WHEN external create may run stays with the
// external-identity provider plugin; this file only owns HOW a created user is
// shaped.

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

// CreateFromExternalUser creates one platform user for a verified external
// identity. See the Service interface for the full contract.
func (s *serviceImpl) CreateFromExternalUser(ctx context.Context, in CreateFromExternalInput) (int, error) {
	email := strings.TrimSpace(in.Email)
	anchor := strings.TrimSpace(in.UsernameAnchor)
	// Email-backed providers derive the username from the email local part.
	// Email-less providers (for example WeChat) MUST supply a deterministic
	// anchor; only reject when neither a valid email nor an anchor is present.
	if email == "" || !strings.Contains(email, "@") {
		if anchor == "" {
			return 0, bizerr.NewCode(CodeUserProvisionEmailInvalid)
		}
		// Email-less provisioning: do not persist an invalid email.
		email = ""
	}
	var (
		username string
		err      error
	)
	if email != "" {
		// Minting-safety invariant: never mint a second account for an email
		// that an existing local account already uses, otherwise an IdP email
		// assertion could take over or shadow that account. The lookup is an
		// unfiltered host-side query because system provisioning runs without
		// an actor context; callers (the external-identity provider plugin) map
		// this into their caller-visible conflict policy.
		count, countErr := dao.SysUser.Ctx(ctx).
			Where(do.SysUser{Email: email}).
			Count()
		if countErr != nil {
			return 0, bizerr.WrapCode(countErr, CodeUserProvisionFailed)
		}
		if count > 0 {
			return 0, bizerr.NewCode(CodeUserProvisionEmailConflict)
		}
		username, err = s.resolveProvisionUsername(ctx, email)
	} else {
		username, err = s.resolveProvisionUsernameFromAnchor(ctx, anchor)
	}
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

// resolveProvisionUsernameFromAnchor derives a deterministic username from a
// collision-resistant anchor supplied by an email-less provider. Unlike the
// email path, it MUST NOT append numeric de-duplication suffixes: the anchor is
// the provider's stable per-identity key, so a repeated provisioning attempt
// for the same external identity resolves to the same username and reuses the
// same account (idempotent). When the derived username already exists, the
// existing account ID is returned by the caller's link-conflict path; here we
// only sanitize and bound the anchor into a legal username.
func (s *serviceImpl) resolveProvisionUsernameFromAnchor(_ context.Context, anchor string) (string, error) {
	base := strings.ToLower(strings.TrimSpace(anchor))
	var builder strings.Builder
	for _, r := range base {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			builder.WriteRune(r)
		case r == '.' || r == '-' || r == '_':
			builder.WriteRune(r)
		}
	}
	candidate := strings.Trim(builder.String(), "._-")
	if candidate == "" {
		return "", bizerr.NewCode(CodeUserProvisionEmailInvalid)
	}
	if len(candidate) > provisionUsernameMaxLength {
		candidate = candidate[:provisionUsernameMaxLength]
	}
	return candidate, nil
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
