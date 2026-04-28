// This file implements opaque lock ticket encoding and validation helpers.

package hostlock

import (
	"encoding/base64"
	"encoding/json"
	"strings"

	"lina-core/pkg/bizerr"
)

// lockTicketClaims stores the serialized metadata required to renew or release
// one acquired plugin lock safely.
type lockTicketClaims struct {
	LockID      int64  `json:"lockId"`
	PluginID    string `json:"pluginId"`
	ResourceRef string `json:"resourceRef"`
	Holder      string `json:"holder"`
	LeaseMillis int64  `json:"leaseMillis"`
}

// encodeLockTicket serializes one opaque lock ticket for plugin callers.
func encodeLockTicket(claims lockTicketClaims) (string, error) {
	content, err := json.Marshal(claims)
	if err != nil {
		return "", bizerr.WrapCode(err, CodeHostLockTicketMarshalFailed)
	}
	return base64.RawURLEncoding.EncodeToString(content), nil
}

// decodeAndValidateTicket decodes one opaque lock ticket and verifies it
// matches the expected plugin and logical resource.
func decodeAndValidateTicket(ticket string, pluginID string, resourceRef string) (*lockTicketClaims, error) {
	normalizedTicket := strings.TrimSpace(ticket)
	if normalizedTicket == "" {
		return nil, bizerr.NewCode(CodeHostLockTicketRequired)
	}

	content, err := base64.RawURLEncoding.DecodeString(normalizedTicket)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeHostLockTicketParseFailed)
	}

	var claims lockTicketClaims
	if err = json.Unmarshal(content, &claims); err != nil {
		return nil, bizerr.WrapCode(err, CodeHostLockTicketUnmarshalFailed)
	}
	if claims.LockID <= 0 || strings.TrimSpace(claims.Holder) == "" || claims.LeaseMillis <= 0 {
		return nil, bizerr.NewCode(CodeHostLockTicketInvalid)
	}
	if strings.TrimSpace(claims.PluginID) != strings.TrimSpace(pluginID) {
		return nil, bizerr.NewCode(CodeHostLockTicketPluginMismatch)
	}
	if strings.TrimSpace(claims.ResourceRef) != strings.TrimSpace(resourceRef) {
		return nil, bizerr.NewCode(CodeHostLockTicketResourceMismatch)
	}
	return &claims, nil
}
