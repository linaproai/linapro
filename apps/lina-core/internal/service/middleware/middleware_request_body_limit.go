// This file applies request-body size limits early enough for multipart
// uploads to honor the runtime-effective upload ceiling instead of the
// framework's static default.

package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"

	"lina-core/pkg/bizerr"
)

// Request-body limit constants define the default non-multipart ceiling and the
// additional multipart envelope allowance.
const (
	defaultRequestBodyLimitBytes   int64 = 8 * 1024 * 1024
	bytesPerMegabyte               int64 = 1024 * 1024
	multipartEnvelopeOverheadBytes int64 = 1 * bytesPerMegabyte
)

// RequestBodyLimit applies one request-scoped body-size guard.
// Non-multipart requests keep the historical 8MB ceiling, while multipart
// uploads reserve a small envelope overhead on top of the runtime-effective
// sys.upload.maxSize file limit so valid files are not rejected by transport
// framing before the business upload validation runs.
func (s *serviceImpl) RequestBodyLimit(r *ghttp.Request) {
	if r == nil {
		return
	}

	var uploadMaxSizeMB int64
	if s != nil && s.configSvc != nil {
		var err error
		uploadMaxSizeMB, err = s.configSvc.GetUploadMaxSize(r.Context())
		if err != nil {
			r.SetError(err)
			return
		}
	}
	// GoFrame may surface request-body overflow either as a normal request error
	// or as a panic raised while multipart parsing is still unwinding. Capture
	// the panic path here so both code paths can be normalized into the same
	// business-facing validation error.
	defer func() {
		if recovered := recover(); recovered != nil {
			if friendlyErr := applyRequestBodyLimitFriendlyError(r, recovered, uploadMaxSizeMB); friendlyErr == nil {
				panic(recovered)
			} else {
				writeRequestBodyLimitError(r, s.i18nSvc, friendlyErr)
			}
		}
	}()

	limit := requestBodyLimitForContentType(r.GetHeader("Content-Type"), uploadMaxSizeMB)
	if limit > 0 && r.Request != nil && r.Request.Body != nil && r.Response != nil {
		// MaxBytesReader enforces the transport-level cap before handlers or
		// middleware parse multipart form bodies, which prevents the framework
		// default 8MB ceiling from rejecting otherwise valid uploads.
		r.Request.Body = http.MaxBytesReader(r.Response.RawWriter(), r.Request.Body, limit)
	}

	r.Middleware.Next()
	// Requests that do not panic still attach the overflow error to the request,
	// so apply the same friendly translation after downstream middleware returns.
	if friendlyErr := applyRequestBodyLimitFriendlyError(r, r.GetError(), uploadMaxSizeMB); friendlyErr != nil {
		writeRequestBodyLimitError(r, s.i18nSvc, friendlyErr)
	}
}

// requestBodyLimitForContentType chooses the effective transport ceiling for
// the current request content type.
func requestBodyLimitForContentType(contentType string, uploadMaxSizeMB int64) int64 {
	if isMultipartContentType(contentType) {
		return multipartRequestBodyLimitBytes(uploadMaxSizeMB)
	}
	return defaultRequestBodyLimitBytes
}

// multipartRequestBodyLimitBytes converts the runtime file-size limit into one
// request-body limit by reserving extra bytes for multipart boundaries and
// headers that wrap the actual uploaded file.
func multipartRequestBodyLimitBytes(uploadMaxSizeMB int64) int64 {
	if uploadMaxSizeMB <= 0 {
		return defaultRequestBodyLimitBytes
	}
	return uploadMaxSizeMB*bytesPerMegabyte + multipartEnvelopeOverheadBytes
}

// isMultipartContentType reports whether the request should use multipart-aware
// transport sizing and friendly overflow messaging.
func isMultipartContentType(contentType string) bool {
	normalized := strings.ToLower(strings.TrimSpace(contentType))
	return strings.Contains(normalized, "multipart/")
}

// requestBodyLimitFriendlyError converts transport-level overflow failures into
// one stable business validation error that matches the configured upload size.
func requestBodyLimitFriendlyError(
	contentType string,
	recovered any,
	uploadMaxSizeMB int64,
) error {
	if !isMultipartContentType(contentType) {
		return nil
	}

	err := recoveredToError(recovered)
	if !isRequestBodyTooLargeError(err) {
		return nil
	}
	if uploadMaxSizeMB > 0 {
		return bizerr.NewCode(
			CodeMiddlewareUploadFileTooLarge,
			bizerr.P("maxSizeMB", uploadMaxSizeMB),
		)
	}
	return bizerr.NewCode(CodeMiddlewareUploadRequestBodyTooLarge)
}

// applyRequestBodyLimitFriendlyError writes the normalized overflow error back
// onto the request and clears any partially written transport response so the
// middleware can emit one stable business error payload instead of a raw server
// failure or empty body.
func applyRequestBodyLimitFriendlyError(r *ghttp.Request, recovered any, uploadMaxSizeMB int64) error {
	if r == nil {
		return nil
	}

	friendlyErr := requestBodyLimitFriendlyError(
		r.GetHeader("Content-Type"),
		recovered,
		uploadMaxSizeMB,
	)
	if friendlyErr == nil {
		return nil
	}
	if r.Response != nil {
		// Reset the partially written response so later middleware can serialize
		// the normalized business error with the project's standard payload shape.
		r.Response.Status = http.StatusOK
		r.Response.ClearBuffer()
	}
	r.SetError(friendlyErr)
	return friendlyErr
}

// writeRequestBodyLimitError serializes one stable JSON error payload for
// multipart body-size overflows, ensuring the client sees a business error
// message even when transport parsing aborted early.
func writeRequestBodyLimitError(r *ghttp.Request, i18nSvc middlewareI18nService, err error) {
	if r == nil || err == nil || r.Response == nil {
		return
	}

	var code = 1
	if errorCode := gerror.Code(err); errorCode != gcode.CodeNil {
		code = errorCode.Code()
	}

	r.Response.Status = http.StatusOK
	message := err.Error()
	if i18nSvc != nil {
		if localized := i18nSvc.LocalizeError(r.Context(), err); localized != "" {
			message = localized
		}
	}
	response := runtimeHandlerResponse{
		Code:    code,
		Data:    nil,
		Message: message,
	}
	applyRuntimeErrorMetadata(&response, err)
	r.Response.WriteJson(response)
	r.ExitAll()
}

// recoveredToError normalizes panic payloads and request-attached error values
// into one error instance for downstream inspection.
func recoveredToError(recovered any) error {
	switch value := recovered.(type) {
	case nil:
		return nil
	case error:
		return value
	case string:
		return errors.New(value)
	default:
		return errors.New(fmt.Sprint(value))
	}
}

// isRequestBodyTooLargeError detects both typed MaxBytesReader failures and the
// string-based framework errors emitted during multipart parsing.
func isRequestBodyTooLargeError(err error) bool {
	if err == nil {
		return false
	}
	var maxBytesErr *http.MaxBytesError
	if errors.As(err, &maxBytesErr) {
		return true
	}
	return strings.Contains(strings.ToLower(err.Error()), "request body too large")
}
