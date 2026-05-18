// This file provides shared validation, token, and value helpers for
// coordination backend implementations.

package core

import (
	"crypto/rand"
	"encoding/base64"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/bizerr"
)

// newOwnerToken creates an opaque owner token for compare-and-renew/delete
// lock operations.
func NewOwnerToken(owner string) (string, error) {
	trimmedOwner := strings.TrimSpace(owner)
	if trimmedOwner == "" {
		return "", bizerr.NewCode(CodeCoordinationKeyInvalid, bizerr.P("field", "owner"))
	}
	var randomBytes [18]byte
	if _, err := rand.Read(randomBytes[:]); err != nil {
		return "", err
	}
	return trimmedOwner + ":" + base64.RawURLEncoding.EncodeToString(randomBytes[:]), nil
}

// handleName validates a lock handle and returns its lock name.
func HandleName(handle *LockHandle) string {
	if handle == nil {
		return ""
	}
	return handle.Name
}

// RequireLogicalKey validates a backend key already produced by a key builder.
func RequireLogicalKey(key string) (string, error) {
	trimmed := strings.TrimSpace(key)
	if trimmed == "" {
		return "", bizerr.NewCode(CodeCoordinationKeyInvalid, bizerr.P("field", "key"))
	}
	return trimmed, nil
}

// normalizeKVWrite validates one KV write key and expiration.
func NormalizeKVWrite(key string, ttl time.Duration) (string, time.Time, error) {
	if ttl < 0 {
		return "", time.Time{}, BizerrTTLInvalid()
	}
	normalizedKey, err := RequireLogicalKey(key)
	if err != nil {
		return "", time.Time{}, err
	}
	var expireAt time.Time
	if ttl > 0 {
		expireAt = time.Now().Add(ttl)
	}
	return normalizedKey, expireAt, nil
}

// BizerrTTLInvalid returns the common TTL validation error.
func BizerrTTLInvalid() error {
	return bizerr.NewCode(CodeCoordinationTTLInvalid)
}

// recordExpired reports whether a stored record is no longer visible.
func RecordExpired(expireAt time.Time) bool {
	return !expireAt.IsZero() && !expireAt.After(time.Now())
}

// parseStoredInt parses one integer value stored in KV.
func ParseStoredInt(value string) (int64, error) {
	parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil {
		return 0, gerror.Wrap(err, "coordination kv value is not an integer")
	}
	return parsed, nil
}

// formatStoredInt formats one integer value for KV storage.
func FormatStoredInt(value int64) string {
	return strconv.FormatInt(value, 10)
}
