// This file implements request and response sanitization helpers used by the
// monitor-operlog audit middleware.

package middleware

import (
	"encoding/json"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"
)

// getRequestParam extracts request parameters as a JSON string.
func getRequestParam(request *ghttp.Request) string {
	body := request.GetBodyString()
	if body != "" {
		return body
	}
	params := request.GetQueryMap()
	if len(params) == 0 {
		return ""
	}
	buffer, err := json.Marshal(params)
	if err != nil {
		return ""
	}
	return string(buffer)
}

// sanitizeOperLogParam recursively masks sensitive request parameters before persistence.
func sanitizeOperLogParam(param string) string {
	if param == "" {
		return param
	}

	var data any
	if err := json.Unmarshal([]byte(param), &data); err != nil {
		return param
	}

	sanitized, changed := sanitizeOperLogValue(data)
	if !changed {
		return param
	}
	buffer, err := json.Marshal(sanitized)
	if err != nil {
		return param
	}
	return string(buffer)
}

// sanitizeOperLogValue traverses one decoded JSON value and masks password and environment payloads.
func sanitizeOperLogValue(value any) (any, bool) {
	switch typed := value.(type) {
	case map[string]any:
		sanitized := make(map[string]any, len(typed))
		changed := false
		for key, item := range typed {
			switch {
			case isOperLogPasswordField(key):
				sanitized[key] = operLogMaskedPassword
				changed = true
			case isOperLogEnvField(key):
				redacted, redactedChanged := redactOperLogEnvValue(item)
				sanitized[key] = redacted
				changed = changed || redactedChanged
			default:
				child, childChanged := sanitizeOperLogValue(item)
				sanitized[key] = child
				changed = changed || childChanged
			}
		}
		if !changed {
			return value, false
		}
		return sanitized, true

	case []any:
		sanitized := make([]any, len(typed))
		changed := false
		for index, item := range typed {
			child, childChanged := sanitizeOperLogValue(item)
			sanitized[index] = child
			changed = changed || childChanged
		}
		if !changed {
			return value, false
		}
		return sanitized, true
	}

	return value, false
}

// redactOperLogEnvValue masks one shell-environment payload while preserving visible keys when possible.
func redactOperLogEnvValue(value any) (any, bool) {
	switch typed := value.(type) {
	case map[string]any:
		if len(typed) == 0 {
			return value, false
		}
		sanitized := make(map[string]any, len(typed))
		for key := range typed {
			sanitized[key] = operLogRedactedValue
		}
		return sanitized, true

	case []any:
		if len(typed) == 0 {
			return value, false
		}
		sanitized := make([]any, len(typed))
		for index, item := range typed {
			sanitized[index] = redactOperLogEnvEntry(item)
		}
		return sanitized, true
	}

	return operLogRedactedValue, true
}

// redactOperLogEnvEntry masks one environment-variable entry inside an array payload.
func redactOperLogEnvEntry(value any) any {
	typed, ok := value.(map[string]any)
	if !ok {
		return operLogRedactedValue
	}

	sanitized := make(map[string]any, len(typed))
	for key, item := range typed {
		switch strings.ToLower(strings.TrimSpace(key)) {
		case "key", "name":
			sanitized[key] = item
		case "value":
			sanitized[key] = operLogRedactedValue
		default:
			sanitized[key] = operLogRedactedValue
		}
	}
	return sanitized
}

// isOperLogPasswordField reports whether the field name carries password semantics and must be masked.
func isOperLogPasswordField(field string) bool {
	switch strings.ToLower(strings.TrimSpace(field)) {
	case "password", "newpassword", "oldpassword":
		return true
	}
	return false
}

// isOperLogEnvField reports whether the field name carries shell environment variables and must be redacted.
func isOperLogEnvField(field string) bool {
	return strings.EqualFold(strings.TrimSpace(field), "env")
}

// isBinaryContentType reports whether the given content type represents binary data.
func isBinaryContentType(contentType string) bool {
	if contentType == "" {
		return false
	}
	lowerContentType := strings.ToLower(contentType)
	return strings.Contains(lowerContentType, "multipart/form-data") ||
		strings.Contains(lowerContentType, "application/octet-stream") ||
		strings.Contains(lowerContentType, "spreadsheetml") ||
		strings.Contains(lowerContentType, "image/") ||
		strings.Contains(lowerContentType, "audio/") ||
		strings.Contains(lowerContentType, "video/")
}

// truncate limits a string length and appends a suffix when truncation happens.
func truncate(value string, maxLen int) string {
	if len(value) <= maxLen {
		return value
	}
	return value[:maxLen] + "...(truncated)"
}
