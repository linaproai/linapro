// This file implements structured business error construction, extraction, and matching.

package bizerr

import (
	"errors"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

// Error carries business i18n metadata while delegating stack, cause, and code
// behavior to GoFrame gerror.
type Error struct {
	err    error
	meta   Meta
	params map[string]any
}

// NewCode constructs a structured business error for one predefined definition.
func NewCode(code *Code, params ...Param) error {
	return newError(nil, code, params...)
}

// New is an alias for NewCode for call sites that prefer concise construction.
func New(code *Code, params ...Param) error {
	return NewCode(code, params...)
}

// WrapCode constructs a structured business error that wraps a lower-level cause.
func WrapCode(cause error, code *Code, params ...Param) error {
	if cause == nil {
		return nil
	}
	return newError(cause, code, params...)
}

// Wrap is an alias for WrapCode for call sites that prefer concise wrapping.
func Wrap(cause error, code *Code, params ...Param) error {
	return WrapCode(cause, code, params...)
}

// As extracts one structured business error from an error chain.
func As(err error) (*Error, bool) {
	var target *Error
	if errors.As(err, &target) {
		return target, true
	}
	return nil, false
}

// Is reports whether err carries the provided structured business error code.
func Is(err error, code *Code) bool {
	messageErr, ok := As(err)
	if !ok {
		return false
	}
	return messageErr.Matches(code)
}

// Error returns the English fallback when available, otherwise the message key
// or machine-readable code. It intentionally avoids localized text.
func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.err != nil {
		return e.err.Error()
	}
	return Format(e.Fallback(), e.params)
}

// Unwrap returns the GoFrame error that carries stack and optional cause.
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.err
}

// Stack returns the stack captured by the delegated GoFrame error.
func (e *Error) Stack() string {
	if e == nil || e.err == nil {
		return ""
	}
	return gerror.Stack(e.err)
}

// Code returns the GoFrame type code carried by this error.
func (e *Error) Code() gcode.Code {
	return e.TypeCode()
}

// TypeCode returns the GoFrame semantic category for this business error.
func (e *Error) TypeCode() gcode.Code {
	if e == nil || e.meta.TypeCode == nil || e.meta.TypeCode == gcode.CodeNil {
		return gcode.CodeUnknown
	}
	return e.meta.TypeCode
}

// RuntimeCode returns the stable machine-readable business error code.
func (e *Error) RuntimeCode() string {
	if e == nil {
		return ""
	}
	return e.meta.ErrorCode
}

// MessageKey returns the runtime i18n key used to localize the error.
func (e *Error) MessageKey() string {
	if e == nil {
		return ""
	}
	return e.meta.MessageKey
}

// Fallback returns the English source fallback for this error.
func (e *Error) Fallback() string {
	if e == nil {
		return ""
	}
	return e.meta.Fallback
}

// Matches reports whether this structured error was created from code.
func (e *Error) Matches(code *Code) bool {
	if e == nil || code == nil {
		return false
	}
	return e.RuntimeCode() != "" && e.RuntimeCode() == code.RuntimeCode()
}

// Params returns a copy of runtime message parameters.
func (e *Error) Params() map[string]any {
	if e == nil || len(e.params) == 0 {
		return nil
	}
	params := make(map[string]any, len(e.params))
	for key, value := range e.params {
		params[key] = value
	}
	return params
}

// newError constructs one structured business error after normalizing definition
// metadata and named parameters. The actual error is created by GoFrame gerror
// so stack traces and wrapped causes follow the framework's native behavior.
func newError(cause error, code *Code, params ...Param) *Error {
	meta, ok := Metadata(code)
	if !ok {
		meta = Meta{
			Fallback: "Unknown error",
			TypeCode: gcode.CodeUnknown,
		}
	}
	normalizedParams := make(map[string]any, len(params))
	for _, item := range params {
		if item.Name == "" {
			continue
		}
		normalizedParams[item.Name] = item.Value
	}
	message := Format(meta.Fallback, normalizedParams)
	var err error
	if cause == nil {
		err = gerror.NewCodeSkip(meta.TypeCode, 2, message)
	} else {
		err = gerror.WrapCodeSkip(meta.TypeCode, 2, cause, message)
	}
	return &Error{
		err:    err,
		meta:   meta,
		params: normalizedParams,
	}
}
