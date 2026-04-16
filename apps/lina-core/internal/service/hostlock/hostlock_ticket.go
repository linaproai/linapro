// This file implements opaque lock ticket encoding and validation helpers.

package hostlock

import (
	"encoding/base64"
	"encoding/json"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
)

type lockTicketClaims struct {
	LockID      int64  `json:"lockId"`
	PluginID    string `json:"pluginId"`
	ResourceRef string `json:"resourceRef"`
	Holder      string `json:"holder"`
	LeaseMillis int64  `json:"leaseMillis"`
}

func encodeLockTicket(claims lockTicketClaims) (string, error) {
	content, err := json.Marshal(claims)
	if err != nil {
		return "", gerror.Wrap(err, "序列化锁票据失败")
	}
	return base64.RawURLEncoding.EncodeToString(content), nil
}

func decodeAndValidateTicket(ticket string, pluginID string, resourceRef string) (*lockTicketClaims, error) {
	normalizedTicket := strings.TrimSpace(ticket)
	if normalizedTicket == "" {
		return nil, gerror.New("锁票据不能为空")
	}

	content, err := base64.RawURLEncoding.DecodeString(normalizedTicket)
	if err != nil {
		return nil, gerror.Wrap(err, "解析锁票据失败")
	}

	var claims lockTicketClaims
	if err = json.Unmarshal(content, &claims); err != nil {
		return nil, gerror.Wrap(err, "反序列化锁票据失败")
	}
	if claims.LockID <= 0 || strings.TrimSpace(claims.Holder) == "" || claims.LeaseMillis <= 0 {
		return nil, gerror.New("锁票据内容无效")
	}
	if strings.TrimSpace(claims.PluginID) != strings.TrimSpace(pluginID) {
		return nil, gerror.New("锁票据插件身份不匹配")
	}
	if strings.TrimSpace(claims.ResourceRef) != strings.TrimSpace(resourceRef) {
		return nil, gerror.New("锁票据逻辑锁名不匹配")
	}
	return &claims, nil
}
