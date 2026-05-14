// This file implements the Tieta HTTP client used by token-based media authorization.

package media

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/gclient"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/logger"
)

// Tieta configuration keys and response constants.
const (
	tietaConfigBaseURLKey             = "tieta.baseUrl"
	tietaConfigMockKey                = "tieta.mock"
	tietaConfigTimeoutKey             = "tieta.timeout"
	tietaUserInfoEndpoint             = "open-apis/user/info"
	tietaTenantDeviceEndpointTemplate = "open-apis/video/device/queryDevicePerm/%s/%s"
	tietaAuthorizationPrefix          = "Bearer "
	tietaDefaultTimeout               = 3 * time.Second
	tietaMockTenantID                 = "1"
	tietaHTTPMethodGet                = "GET"
	tietaHTTPMethodPost               = "POST"
	tietaUserInfoSuccessCode          = 200
	tietaTenantDeviceSuccessCode      = 0
)

var (
	// mediaTietaClient is replaceable by unit tests.
	mediaTietaClient tietaClient = &httpTietaClient{}
)

// tietaClient defines the Tieta operations used by media services.
type tietaClient interface {
	UserInfoByToken(ctx context.Context, token string) (*TietaUser, error)
	CheckTenantHasDevice(ctx context.Context, token string, tenantID string, deviceID string) (bool, error)
}

// httpTietaClient implements tietaClient with GoFrame HTTP client.
type httpTietaClient struct{}

// UserInfoByToken validates one Tieta token and returns the corresponding user identity.
func (c *httpTietaClient) UserInfoByToken(ctx context.Context, token string) (*TietaUser, error) {
	normalizedToken := normalizeTietaToken(token)
	if normalizedToken == "" {
		return nil, bizerr.NewCode(CodeMediaTietaTokenRequired)
	}
	if isTietaMock(ctx) {
		return mockTietaUser(normalizedToken), nil
	}

	result, err := c.call(ctx, normalizedToken, tietaHTTPMethodPost, tietaUserInfoEndpoint, nil)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaTietaUserInfoFailed)
	}

	var response tietaUserResponse
	responseBody := result.String()
	if err = gjson.Unmarshal([]byte(responseBody), &response); err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaTietaUserInfoInvalid)
	}
	if response.Code != tietaUserInfoSuccessCode {
		return nil, bizerr.NewCode(CodeMediaTietaTokenInvalid, bizerr.P("message", response.Msg))
	}
	if response.Data == nil || response.Data.ID <= 0 {
		return nil, bizerr.NewCode(CodeMediaTietaTokenInvalid, bizerr.P("message", "用户信息为空"))
	}

	user := buildTietaUser(response.Data)
	return user, nil
}

// CheckTenantHasDevice checks whether one Tieta tenant can access a device.
func (c *httpTietaClient) CheckTenantHasDevice(
	ctx context.Context,
	token string,
	tenantID string,
	deviceID string,
) (bool, error) {
	normalizedTenantID, err := normalizeTenantID(tenantID)
	if err != nil {
		return false, err
	}
	normalizedDeviceID, err := normalizeDeviceID(deviceID)
	if err != nil {
		return false, err
	}
	normalizedToken := normalizeTietaToken(token)
	if normalizedToken == "" {
		return false, bizerr.NewCode(CodeMediaTietaTokenRequired)
	}
	if isTietaMock(ctx) {
		logger.Infof(ctx, "铁塔 mock 模式: 租户设备权限通过 tenant=%s device=%s", normalizedTenantID, normalizedDeviceID)
		return true, nil
	}

	deviceType, upstreamDeviceID := splitTietaDeviceID(normalizedDeviceID)
	endpoint := fmt.Sprintf(tietaTenantDeviceEndpointTemplate, deviceType, upstreamDeviceID)
	result, err := c.call(ctx, normalizedToken, tietaHTTPMethodPost, endpoint, nil)
	if err != nil {
		return false, bizerr.WrapCode(err, CodeMediaTietaDevicePermissionFailed)
	}

	var response tietaTenantDeviceResponse
	responseBody := result.String()
	if err = gjson.Unmarshal([]byte(responseBody), &response); err != nil {
		return false, bizerr.WrapCode(err, CodeMediaTietaDevicePermissionInvalid)
	}
	if response.Code != tietaTenantDeviceSuccessCode {
		return false, bizerr.NewCode(CodeMediaTietaDevicePermissionDenied, bizerr.P("message", response.Msg))
	}
	return response.Data, nil
}

// call invokes one Tieta HTTP endpoint and parses the JSON response.
func (c *httpTietaClient) call(
	ctx context.Context,
	token string,
	method string,
	endpoint string,
	requestBody any,
) (*gjson.Json, error) {
	baseURL, err := tietaBaseURL(ctx)
	if err != nil {
		return nil, err
	}
	requestURL := strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(endpoint, "/")

	client := g.Client()
	client.SetTimeout(tietaTimeout(ctx))
	client.SetHeader("Authorization", tietaAuthorizationPrefix+token)

	var response *gclient.Response
	switch method {
	case tietaHTTPMethodPost:
		response, err = client.Post(ctx, requestURL, requestBody)
	case tietaHTTPMethodGet:
		response, err = client.Get(ctx, requestURL)
	default:
		return nil, gerror.Newf("unsupported Tieta request method: %s", method)
	}
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := response.Close(); closeErr != nil {
			logger.Warningf(ctx, "关闭铁塔响应体失败: %v", closeErr)
		}
	}()

	responseBody := response.ReadAll()
	result, err := gjson.LoadJson(responseBody)
	if err != nil {
		return nil, gerror.Wrapf(err, "parse Tieta response failed: %s", responseBody)
	}
	return result, nil
}

// buildTietaUser maps one raw Tieta user payload into the media identity structure.
func buildTietaUser(info *tietaUserInfo) *TietaUser {
	if info == nil {
		return nil
	}
	return &TietaUser{
		Id:           info.ID,
		DeptId:       info.DeptId,
		Username:     strings.TrimSpace(info.UserName),
		RealName:     strings.TrimSpace(info.NickName),
		Mobile:       strings.TrimSpace(info.Phone),
		UserType:     strings.TrimSpace(info.UserType),
		CustomerCode: strings.TrimSpace(info.CustomerCode),
		TenantId:     strings.TrimSpace(info.CustomerId),
		DeptName:     strings.TrimSpace(info.DeptName),
		RegionCode:   info.RegionCode,
		OrgId:        info.OrgID,
		Enable:       info.Enable,
	}
}

// tietaBaseURL returns the configured upstream Tieta base URL.
func tietaBaseURL(ctx context.Context) (string, error) {
	baseURL := strings.TrimSpace(g.Cfg().MustGet(ctx, tietaConfigBaseURLKey).String())
	if baseURL == "" {
		return "", bizerr.NewCode(CodeMediaTietaBaseURLMissing)
	}
	return baseURL, nil
}

// tietaTimeout returns the configured Tieta HTTP timeout.
func tietaTimeout(ctx context.Context) time.Duration {
	value, err := g.Cfg().Get(ctx, tietaConfigTimeoutKey)
	if err != nil {
		logger.Warningf(ctx, "读取铁塔 timeout 配置失败，使用默认值: %v", err)
		return tietaDefaultTimeout
	}
	if configValueMissing(value) {
		return tietaDefaultTimeout
	}
	raw := strings.TrimSpace(value.String())
	if raw == "" {
		return tietaDefaultTimeout
	}
	timeout, err := time.ParseDuration(raw)
	if err != nil || timeout <= 0 {
		logger.Warningf(ctx, "铁塔 timeout 配置无效，使用默认值: value=%s err=%v", raw, err)
		return tietaDefaultTimeout
	}
	return timeout
}

// isTietaMock reports whether Tieta mock mode is enabled.
func isTietaMock(ctx context.Context) bool {
	return g.Cfg().MustGet(ctx, tietaConfigMockKey).Bool()
}

// normalizeTietaToken removes the optional Bearer prefix from a Tieta token.
func normalizeTietaToken(token string) string {
	trimmed := strings.TrimSpace(token)
	if trimmed == "" {
		return ""
	}
	if strings.HasPrefix(strings.ToLower(trimmed), strings.ToLower(tietaAuthorizationPrefix)) {
		trimmed = strings.TrimSpace(trimmed[len(tietaAuthorizationPrefix):])
	}
	return trimmed
}

// splitTietaDeviceID maps device IDs like "uav_123" to Tieta's device type and device ID pair.
func splitTietaDeviceID(deviceID string) (deviceType string, upstreamDeviceID string) {
	parts := strings.SplitN(strings.TrimSpace(deviceID), "_", 2)
	if len(parts) == 1 {
		return "", parts[0]
	}
	return parts[0], parts[1]
}

// configValueMissing reports whether a GoFrame config lookup returned no concrete value.
func configValueMissing(value *gvar.Var) bool {
	return value == nil || value.IsNil()
}

// mockTietaUser returns a deterministic Tieta user for local development and tests.
func mockTietaUser(token string) *TietaUser {
	tenantID := strings.TrimSpace(token)
	if tenantID == "" || strings.EqualFold(tenantID, "mock") {
		tenantID = tietaMockTenantID
	}
	return &TietaUser{
		Id:       13,
		DeptId:   100,
		Username: "wj530",
		RealName: "王杰",
		Mobile:   "18213268117",
		UserType: "00",
		TenantId: tenantID,
		DeptName: "湖南铁塔",
		Enable:   true,
	}
}
