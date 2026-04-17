// This file defines built-in runtime parameters backed by sys_config and their
// validation rules.

package config

import (
	"context"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
)

const (
	// RuntimeParamKeyJWTExpire stores the runtime JWT token lifetime.
	RuntimeParamKeyJWTExpire = "sys.jwt.expire"
	// RuntimeParamKeySessionTimeout stores the runtime online-session inactivity timeout.
	RuntimeParamKeySessionTimeout = "sys.session.timeout"
	// RuntimeParamKeyUploadMaxSize stores the runtime upload size ceiling in MB.
	RuntimeParamKeyUploadMaxSize = "sys.upload.maxSize"
	// RuntimeParamKeyLoginBlackIPList stores the runtime login IP blacklist.
	RuntimeParamKeyLoginBlackIPList = "sys.login.blackIPList"
	// RuntimeParamKeyUserInitPassword stores the default reset-password value shown by the user module.
	RuntimeParamKeyUserInitPassword = "sys.user.initPassword"
)

// RuntimeParamSpec describes one built-in runtime parameter managed through
// sys_config.
type RuntimeParamSpec struct {
	Key          string // Key is the sys_config key.
	Name         string // Name is the display name shown in the config page.
	DefaultValue string // DefaultValue is the host fallback value.
	Remark       string // Remark explains the supported format and behavior.
}

var runtimeParamSpecs = []RuntimeParamSpec{
	{
		Key:          RuntimeParamKeyJWTExpire,
		Name:         "认证管理-JWT Token 有效期",
		DefaultValue: "24h",
		Remark:       "控制新签发 JWT Token 的有效期，支持 Go duration 格式，如 12h、24h。",
	},
	{
		Key:          RuntimeParamKeySessionTimeout,
		Name:         "在线用户-会话超时时间",
		DefaultValue: "24h",
		Remark:       "控制在线会话无活动超时时长，支持 Go duration 格式，如 30m、24h。",
	},
	{
		Key:          RuntimeParamKeyUploadMaxSize,
		Name:         "文件管理-上传大小上限",
		DefaultValue: "10",
		Remark:       "控制单个上传文件大小上限，单位为 MB，必须为正整数。",
	},
	{
		Key:          RuntimeParamKeyLoginBlackIPList,
		Name:         "用户登录-IP 黑名单列表",
		DefaultValue: "",
		Remark:       "禁止登录的 IP 或 CIDR 网段，多个值以英文分号分隔，例如 127.0.0.1;10.0.0.0/8。",
	},
	{
		Key:          RuntimeParamKeyUserInitPassword,
		Name:         "用户管理-账号初始密码",
		DefaultValue: "123456",
		Remark:       "用户重置密码弹窗默认回填值，长度必须为 5-20 个字符。",
	},
}

var runtimeParamSpecByKey = map[string]RuntimeParamSpec{
	RuntimeParamKeyJWTExpire:        runtimeParamSpecs[0],
	RuntimeParamKeySessionTimeout:   runtimeParamSpecs[1],
	RuntimeParamKeyUploadMaxSize:    runtimeParamSpecs[2],
	RuntimeParamKeyLoginBlackIPList: runtimeParamSpecs[3],
	RuntimeParamKeyUserInitPassword: runtimeParamSpecs[4],
}

var runtimeParamKeys = []string{
	RuntimeParamKeyJWTExpire,
	RuntimeParamKeySessionTimeout,
	RuntimeParamKeyUploadMaxSize,
	RuntimeParamKeyLoginBlackIPList,
	RuntimeParamKeyUserInitPassword,
}

// RuntimeParamSpecs returns all built-in runtime parameter specs.
func RuntimeParamSpecs() []RuntimeParamSpec {
	specs := make([]RuntimeParamSpec, len(runtimeParamSpecs))
	copy(specs, runtimeParamSpecs)
	return specs
}

// LookupRuntimeParamSpec returns one built-in runtime parameter spec by key.
func LookupRuntimeParamSpec(key string) (RuntimeParamSpec, bool) {
	spec, ok := runtimeParamSpecByKey[strings.TrimSpace(key)]
	return spec, ok
}

// IsProtectedRuntimeParam reports whether the key belongs to one built-in
// runtime parameter that must not be renamed or deleted.
func IsProtectedRuntimeParam(key string) bool {
	_, ok := LookupRuntimeParamSpec(key)
	return ok
}

// ValidateRuntimeParamValue validates one built-in runtime parameter value.
func ValidateRuntimeParamValue(key string, value string) error {
	switch strings.TrimSpace(key) {
	case RuntimeParamKeyJWTExpire:
		_, err := validatePositiveDurationValue(key, value)
		return err

	case RuntimeParamKeySessionTimeout:
		_, err := validatePositiveDurationValue(key, value)
		return err

	case RuntimeParamKeyUploadMaxSize:
		_, err := validatePositiveInt64Value(key, value)
		return err

	case RuntimeParamKeyLoginBlackIPList:
		return validateIPBlacklistValue(key, value)

	case RuntimeParamKeyUserInitPassword:
		return validatePasswordValue(key, value)
	}
	return nil
}

func (s *serviceImpl) lookupRuntimeParamValue(ctx context.Context, key string) (value string, ok bool) {
	snapshot := s.getRuntimeParamSnapshot(ctx)
	if snapshot == nil {
		return "", false
	}
	return snapshot.lookupValue(key)
}

func (s *serviceImpl) applyRuntimeDurationOverride(
	ctx context.Context,
	key string,
	current time.Duration,
) time.Duration {
	snapshot := s.getRuntimeParamSnapshot(ctx)
	if snapshot == nil {
		return current
	}
	duration, ok, err := snapshot.lookupDuration(key)
	if err != nil {
		panic(err)
	}
	if !ok {
		return current
	}
	return duration
}

func (s *serviceImpl) applyRuntimeInt64Override(
	ctx context.Context,
	key string,
	current int64,
) int64 {
	snapshot := s.getRuntimeParamSnapshot(ctx)
	if snapshot == nil {
		return current
	}
	parsed, ok, err := snapshot.lookupInt64(key)
	if err != nil {
		panic(err)
	}
	if !ok {
		return current
	}
	return parsed
}

func splitSemicolonValues(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ";")
	values := make([]string, 0, len(parts))
	for _, item := range parts {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		values = append(values, trimmed)
	}
	return values
}

func validatePositiveDurationValue(key string, value string) (time.Duration, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 0, gerror.Newf("参数 %s 不能为空", key)
	}
	duration, err := time.ParseDuration(trimmed)
	if err != nil {
		return 0, gerror.Wrapf(err, "参数 %s 必须为合法时长", key)
	}
	if duration <= 0 {
		return 0, gerror.Newf("参数 %s 必须大于 0", key)
	}
	return duration, nil
}

func validatePositiveInt64Value(key string, value string) (int64, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 0, gerror.Newf("参数 %s 不能为空", key)
	}
	parsed, err := strconv.ParseInt(trimmed, 10, 64)
	if err != nil {
		return 0, gerror.Wrapf(err, "参数 %s 必须为整数", key)
	}
	if parsed <= 0 {
		return 0, gerror.Newf("参数 %s 必须大于 0", key)
	}
	return parsed, nil
}

func validateIPBlacklistValue(key string, value string) error {
	for _, item := range splitSemicolonValues(value) {
		if net.ParseIP(item) != nil {
			continue
		}
		if _, _, err := net.ParseCIDR(item); err == nil {
			continue
		}
		return gerror.Newf("参数 %s 包含非法 IP 或 CIDR：%s", key, item)
	}
	return nil
}

func validatePasswordValue(key string, value string) error {
	length := len(strings.TrimSpace(value))
	if length < 5 || length > 20 {
		return gerror.Newf("参数 %s 的长度必须为 5-20 个字符", key)
	}
	return nil
}
