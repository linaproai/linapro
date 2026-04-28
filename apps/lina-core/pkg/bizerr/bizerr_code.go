// This file defines reusable business error codes and metadata accessors.

package bizerr

import (
	"strings"

	"github.com/gogf/gf/v2/errors/gcode"
)

// Meta carries runtime i18n metadata for one business error definition.
type Meta struct {
	ErrorCode  string     // ErrorCode is the stable machine-readable business code.
	MessageKey string     // MessageKey is the runtime i18n key used to render the error.
	Fallback   string     // Fallback is the English source text used when translation is missing.
	TypeCode   gcode.Code // TypeCode is the GoFrame semantic category returned in response code.
}

// Code defines one reusable business error without allocating a custom integer
// response code. The concrete response code always comes from TypeCode.
type Code struct {
	meta Meta
}

// MustDefine creates one reusable business error definition. It panics when
// required metadata is invalid so broken definitions fail during startup or
// tests.
func MustDefine(errorCode string, fallback string, typeCode gcode.Code) *Code {
	return MustDefineWithKey(errorCode, MessageKey(errorCode), fallback, typeCode)
}

// MustDefineWithKey creates one reusable business error definition with an
// explicit runtime i18n key. Prefer MustDefine unless a legacy key must be
// preserved.
func MustDefineWithKey(errorCode string, messageKey string, fallback string, typeCode gcode.Code) *Code {
	meta := Meta{
		ErrorCode:  strings.TrimSpace(errorCode),
		MessageKey: strings.TrimSpace(messageKey),
		Fallback:   strings.TrimSpace(fallback),
		TypeCode:   typeCode,
	}
	if meta.ErrorCode == "" {
		panic("bizerr error code is required")
	}
	if meta.MessageKey == "" {
		panic("bizerr message key is required")
	}
	if meta.Fallback == "" {
		panic("bizerr fallback is required")
	}
	if meta.TypeCode == nil || meta.TypeCode == gcode.CodeNil {
		meta.TypeCode = gcode.CodeUnknown
	}
	return &Code{meta: meta}
}

// Metadata extracts business error metadata from one reusable definition.
func Metadata(code *Code) (Meta, bool) {
	if code == nil {
		return Meta{}, false
	}
	return code.meta, true
}

// TypeCode returns the GoFrame semantic category for this business error
// definition.
func (c *Code) TypeCode() gcode.Code {
	if c == nil || c.meta.TypeCode == nil || c.meta.TypeCode == gcode.CodeNil {
		return gcode.CodeUnknown
	}
	return c.meta.TypeCode
}

// RuntimeCode returns the stable machine-readable business error code.
func (c *Code) RuntimeCode() string {
	if c == nil {
		return ""
	}
	return c.meta.ErrorCode
}

// MessageKey returns the runtime i18n key used to localize this error
// definition.
func (c *Code) MessageKey() string {
	if c == nil {
		return ""
	}
	return c.meta.MessageKey
}

// Fallback returns the English source fallback for this error definition.
func (c *Code) Fallback() string {
	if c == nil {
		return ""
	}
	return c.meta.Fallback
}
