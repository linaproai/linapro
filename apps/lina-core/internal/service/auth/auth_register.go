// This file implements public self-registration for platform accounts.

package auth

import (
	"context"
	"strings"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/logger"
	"lina-core/pkg/statusflag"

	"github.com/gogf/gf/v2/database/gdb"
)

const defaultRegisterRoleKey = "user"

// Register creates one public platform account when self-registration is enabled.
func (s *serviceImpl) Register(ctx context.Context, in RegisterInput) (*RegisterOutput, error) {
	publicCfg, err := s.configSvc.GetPublicFrontend(ctx)
	if err != nil {
		return nil, err
	}
	if publicCfg == nil || !publicCfg.Auth.RegisterEnabled {
		return nil, bizerr.NewCode(CodeAuthRegisterDisabled)
	}

	var (
		username = strings.TrimSpace(in.Username)
		password = strings.TrimSpace(in.Password)
		email    = strings.ToLower(strings.TrimSpace(in.Email))
		nickname = strings.TrimSpace(in.Nickname)
	)
	if username == "" || password == "" || email == "" {
		return nil, bizerr.NewCode(CodeAuthClientTypeInvalid)
	}
	if nickname == "" {
		nickname = username
	}

	if s.rateLimit != nil {
		ok, limitErr := s.rateLimit.Allow(ctx, "register:username:"+strings.ToLower(username), registerRateLimitMax)
		if limitErr != nil {
			return nil, bizerr.WrapCode(limitErr, CodeAuthTokenStateUnavailable)
		}
		if !ok {
			return nil, bizerr.NewCode(CodeAuthRegisterRateLimited)
		}
	}

	usernameCount, err := dao.SysUser.Ctx(ctx).Where(do.SysUser{Username: username}).Count()
	if err != nil {
		return nil, err
	}
	if usernameCount > 0 {
		return nil, bizerr.NewCode(CodeAuthRegisterUsernameExists)
	}
	emailCount, err := dao.SysUser.Ctx(ctx).Where(do.SysUser{Email: email}).Count()
	if err != nil {
		return nil, err
	}
	if emailCount > 0 {
		return nil, bizerr.NewCode(CodeAuthRegisterEmailExists)
	}

	roleID, err := s.lookupBuiltinUserRoleID(ctx)
	if err != nil {
		return nil, err
	}

	hash, err := s.HashPassword(password)
	if err != nil {
		return nil, err
	}

	var userID int
	err = dao.SysUser.Ctx(ctx).Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		id, insertErr := dao.SysUser.Ctx(ctx).Data(do.SysUser{
			Username: username,
			Password: hash,
			Nickname: nickname,
			Email:    email,
			Sex:      0,
			Status:   statusflag.EnabledValue.Int(),
			Remark:   "self-registered",
			TenantId: 0,
		}).InsertAndGetId()
		if insertErr != nil {
			return insertErr
		}
		userID = int(id)
		_, relErr := dao.SysUserRole.Ctx(ctx).Data(do.SysUserRole{
			UserId:   userID,
			RoleId:   roleID,
			TenantId: 0,
		}).Insert()
		return relErr
	})
	if err != nil {
		return nil, err
	}

	logger.Infof(ctx, "public registration created user id=%d username=%s", userID, username)
	return &RegisterOutput{UserID: userID}, nil
}

// lookupBuiltinUserRoleID resolves the platform built-in standard user role.
func (s *serviceImpl) lookupBuiltinUserRoleID(ctx context.Context) (int, error) {
	var role *entity.SysRole
	err := dao.SysRole.Ctx(ctx).
		Where(do.SysRole{Key: defaultRegisterRoleKey, TenantId: 0}).
		Scan(&role)
	if err != nil {
		return 0, err
	}
	if role == nil || role.Id <= 0 || role.Status == 0 {
		return 0, bizerr.NewCode(CodeAuthDefaultRoleMissing)
	}
	return role.Id, nil
}
