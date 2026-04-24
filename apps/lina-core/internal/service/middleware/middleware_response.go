// This file normalizes handler responses and localizes user-facing error
// messages before the host writes the unified JSON payload.

package middleware

import (
	"mime"
	"net/http"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
)

const (
	// responseContentTypeEventStream identifies SSE responses.
	responseContentTypeEventStream = "text/event-stream"
	// responseContentTypeOctetStream identifies binary stream downloads.
	responseContentTypeOctetStream = "application/octet-stream"
	// responseContentTypeMixedReplace identifies multipart stream responses.
	responseContentTypeMixedReplace = "multipart/x-mixed-replace"
)

// Response serializes one unified JSON payload and localizes user-facing error
// text using the request-scoped locale.
func (s *serviceImpl) Response(r *ghttp.Request) {
	if r == nil {
		return
	}

	r.Middleware.Next()

	// Downstream handlers that already wrote bytes own the response body.
	if r.Response.BufferLength() > 0 || r.Response.BytesWritten() > 0 {
		return
	}

	mediaType, _, _ := mime.ParseMediaType(r.Response.Header().Get("Content-Type"))
	switch mediaType {
	case responseContentTypeEventStream, responseContentTypeOctetStream, responseContentTypeMixedReplace:
		return
	}

	var (
		err  = r.GetError()
		res  = r.GetHandlerResponse()
		code = gerror.Code(err)
		msg  string
	)

	if err != nil {
		if code == gcode.CodeNil {
			code = gcode.CodeInternalError
		}
		msg = s.i18nSvc.LocalizeError(r.Context(), err)
	} else {
		if r.Response.Status > 0 && r.Response.Status != http.StatusOK {
			switch r.Response.Status {
			case http.StatusNotFound:
				code = gcode.CodeNotFound
			case http.StatusForbidden:
				code = gcode.CodeNotAuthorized
			default:
				code = gcode.CodeUnknown
			}
			err = gerror.NewCode(code, code.Message())
			r.SetError(err)
		} else {
			code = gcode.CodeOK
		}
		msg = s.i18nSvc.LocalizeError(r.Context(), gerror.New(code.Message()))
	}

	if msg == "" {
		msg = code.Message()
	}

	r.Response.WriteJson(ghttp.DefaultHandlerResponse{
		Code:    code.Code(),
		Message: msg,
		Data:    res,
	})
}
