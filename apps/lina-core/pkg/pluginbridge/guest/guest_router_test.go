// This file tests reflected guest controller dispatch and guest route fallback behavior.

package guest

import (
	"context"
	"errors"
	"testing"
)

// guestRouteTestController provides reflected handlers for guest route
// dispatcher tests.
type guestRouteTestController struct{}

// guestTypedRouteTestController provides typed `(ctx, *Req) (*Res, error)`
// handlers for dispatcher coverage.
type guestTypedRouteTestController struct{}

// UpdateDemoReq exercises typed request binding from path params, query
// values, and JSON body.
type UpdateDemoReq struct {
	Id          string `json:"id"`
	PageNum     int    `json:"pageNum"`
	SkipNetwork bool   `json:"skipNetwork"`
	Title       string `json:"title"`
}

// UpdateDemoRes captures the typed response payload emitted by the test
// controller.
type UpdateDemoRes struct {
	Id          string `json:"id"`
	PageNum     int    `json:"pageNum"`
	SkipNetwork bool   `json:"skipNetwork"`
	Title       string `json:"title"`
	PluginID    string `json:"pluginId"`
	RequestID   string `json:"requestId"`
}

// DownloadReq exercises manual binary response writing from typed guest
// controllers.
type DownloadReq struct {
	Id string `json:"id"`
}

// DownloadRes is the placeholder typed response for manual binary writes.
type DownloadRes struct{}

// ClassifiedReq exercises typed response errors.
type ClassifiedReq struct{}

// ClassifiedRes is the placeholder typed response for classified error tests.
type ClassifiedRes struct{}

// BackendSummary returns a deterministic JSON payload for successful dispatch
// assertions.
func (c *guestRouteTestController) BackendSummary(
	request *BridgeRequestEnvelopeV1,
) (*BridgeResponseEnvelopeV1, error) {
	return NewJSONResponse(200, []byte(`{"requestId":"`+request.RequestID+`"}`)), nil
}

// Failure returns a controller error so dispatcher error propagation can be
// asserted.
func (c *guestRouteTestController) Failure(
	_ *BridgeRequestEnvelopeV1,
) (*BridgeResponseEnvelopeV1, error) {
	return nil, errors.New("boom")
}

// IgnoredMethod does not match the dispatcher signature and should be ignored
// during controller reflection.
func (c *guestRouteTestController) IgnoredMethod() {}

// UpdateDemo verifies the typed dispatcher can bind request DTOs and still
// expose bridge metadata through context helpers.
func (c *guestTypedRouteTestController) UpdateDemo(
	ctx context.Context,
	req *UpdateDemoReq,
) (*UpdateDemoRes, error) {
	if err := SetResponseHeader(ctx, "X-Guest-Typed", "ok"); err != nil {
		return nil, err
	}
	envelope := RequestEnvelopeFromContext(ctx)
	if envelope == nil {
		return nil, errors.New("missing envelope from typed guest context")
	}
	return &UpdateDemoRes{
		Id:          req.Id,
		PageNum:     req.PageNum,
		SkipNetwork: req.SkipNetwork,
		Title:       req.Title,
		PluginID:    envelope.PluginID,
		RequestID:   envelope.RequestID,
	}, nil
}

// Download verifies typed guest controllers can emit manual non-JSON
// responses.
func (c *guestTypedRouteTestController) Download(
	ctx context.Context,
	req *DownloadReq,
) (*DownloadRes, error) {
	if err := SetResponseHeader(ctx, "Content-Disposition", `attachment; filename="demo.txt"`); err != nil {
		return nil, err
	}
	if err := WriteResponse(ctx, 200, "text/plain; charset=utf-8", []byte(req.Id)); err != nil {
		return nil, err
	}
	return nil, nil
}

// Classified verifies typed guest controllers can return prebuilt bridge
// responses through the error channel.
func (c *guestTypedRouteTestController) Classified(
	_ context.Context,
	_ *ClassifiedReq,
) (*ClassifiedRes, error) {
	return nil, NewResponseError(NewForbiddenResponse("nope"))
}

// TestGuestControllerRouteDispatcherDispatchesByRequestType verifies reflected
// request-type dispatch finds the matching controller method.
func TestGuestControllerRouteDispatcherDispatchesByRequestType(t *testing.T) {
	dispatcher, err := NewGuestControllerRouteDispatcher(&guestRouteTestController{})
	if err != nil {
		t.Fatalf("expected dispatcher creation to succeed, got error: %v", err)
	}

	response, err := dispatcher.HandleRequest(&BridgeRequestEnvelopeV1{
		RequestID: "req-1",
		Route: &RouteMatchSnapshotV1{
			RequestType: "BackendSummaryReq",
		},
	})
	if err != nil {
		t.Fatalf("expected request dispatch to succeed, got error: %v", err)
	}
	if response == nil || response.StatusCode != 200 {
		t.Fatalf("expected dispatcher to return 200 response, got %#v", response)
	}
	if string(response.Body) != `{"requestId":"req-1"}` {
		t.Fatalf("expected dispatcher to preserve controller response body, got %q", string(response.Body))
	}
}

// TestGuestControllerRouteDispatcherReturnsControllerError verifies controller
// errors are returned without a synthesized bridge response.
func TestGuestControllerRouteDispatcherReturnsControllerError(t *testing.T) {
	dispatcher, err := NewGuestControllerRouteDispatcher(&guestRouteTestController{})
	if err != nil {
		t.Fatalf("expected dispatcher creation to succeed, got error: %v", err)
	}

	response, dispatchErr := dispatcher.HandleRequest(&BridgeRequestEnvelopeV1{
		Route: &RouteMatchSnapshotV1{
			RequestType: "FailureReq",
		},
	})
	if dispatchErr == nil || dispatchErr.Error() != "boom" {
		t.Fatalf("expected controller error boom, got %v", dispatchErr)
	}
	if response != nil {
		t.Fatalf("expected dispatcher to return nil response on error, got %#v", response)
	}
}

// TestGuestControllerRouteDispatcherRejectsMissingRequestType verifies the
// dispatcher responds with a bad request when requestType is absent.
func TestGuestControllerRouteDispatcherRejectsMissingRequestType(t *testing.T) {
	dispatcher, err := NewGuestControllerRouteDispatcher(&guestRouteTestController{})
	if err != nil {
		t.Fatalf("expected dispatcher creation to succeed, got error: %v", err)
	}

	response, err := dispatcher.HandleRequest(&BridgeRequestEnvelopeV1{
		Route: &RouteMatchSnapshotV1{},
	})
	if err != nil {
		t.Fatalf("expected missing request type to return bridge response, got error: %v", err)
	}
	if response == nil || response.StatusCode != 400 {
		t.Fatalf("expected 400 response for missing request type, got %#v", response)
	}
}

// TestGuestControllerRouteDispatcherFallsBackToInternalPath verifies internal
// path dispatch works when requestType is not supplied.
func TestGuestControllerRouteDispatcherFallsBackToInternalPath(t *testing.T) {
	dispatcher, err := NewGuestControllerRouteDispatcher(&guestRouteTestController{})
	if err != nil {
		t.Fatalf("expected dispatcher creation to succeed, got error: %v", err)
	}

	response, err := dispatcher.HandleRequest(&BridgeRequestEnvelopeV1{
		RequestID: "req-2",
		Route: &RouteMatchSnapshotV1{
			InternalPath: "/backend-summary",
		},
	})
	if err != nil {
		t.Fatalf("expected internal path fallback to succeed, got error: %v", err)
	}
	if response == nil || response.StatusCode != 200 {
		t.Fatalf("expected 200 response for internal path fallback, got %#v", response)
	}
}

// TestGuestControllerRouteDispatcherReturnsNotFound verifies unknown reflected
// handlers return a not-found bridge response.
func TestGuestControllerRouteDispatcherReturnsNotFound(t *testing.T) {
	dispatcher, err := NewGuestControllerRouteDispatcher(&guestRouteTestController{})
	if err != nil {
		t.Fatalf("expected dispatcher creation to succeed, got error: %v", err)
	}

	response, err := dispatcher.HandleRequest(&BridgeRequestEnvelopeV1{
		Route: &RouteMatchSnapshotV1{
			RequestType: "UnknownReq",
		},
	})
	if err != nil {
		t.Fatalf("expected unknown request type to return bridge response, got error: %v", err)
	}
	if response == nil || response.StatusCode != 404 {
		t.Fatalf("expected 404 response for unknown request type, got %#v", response)
	}
}

// TestGuestControllerRouteDispatcherBindsTypedRequest verifies typed guest
// handlers receive request DTOs hydrated from body, path params, and query
// values while still being able to emit custom headers.
func TestGuestControllerRouteDispatcherBindsTypedRequest(t *testing.T) {
	dispatcher, err := NewGuestControllerRouteDispatcher(&guestTypedRouteTestController{})
	if err != nil {
		t.Fatalf("expected typed dispatcher creation to succeed, got error: %v", err)
	}

	response, err := dispatcher.HandleRequest(&BridgeRequestEnvelopeV1{
		PluginID:  "plugin-demo-dynamic",
		RequestID: "req-typed",
		Route: &RouteMatchSnapshotV1{
			Method:      "PUT",
			RequestType: "UpdateDemoReq",
			PathParams:  map[string]string{"id": "demo-1"},
			QueryValues: map[string][]string{
				"pageNum":     {"7"},
				"skipNetwork": {"yes"},
			},
		},
		Request: &HTTPRequestSnapshotV1{
			Method: "PUT",
			Body:   []byte(`{"title":"updated title"}`),
		},
	})
	if err != nil {
		t.Fatalf("expected typed request dispatch to succeed, got error: %v", err)
	}
	if response == nil || response.StatusCode != 200 {
		t.Fatalf("expected 200 response for typed request, got %#v", response)
	}
	if got := response.Headers["X-Guest-Typed"]; len(got) != 1 || got[0] != "ok" {
		t.Fatalf("expected custom typed response header, got %#v", response.Headers)
	}
	if string(response.Body) != `{"id":"demo-1","pageNum":7,"skipNetwork":true,"title":"updated title","pluginId":"plugin-demo-dynamic","requestId":"req-typed"}` {
		t.Fatalf("unexpected typed response payload: %q", string(response.Body))
	}
}

// TestGuestControllerRouteDispatcherRequiresTypedJSONBody verifies typed POST
// or PUT handlers keep the BindJSON-style empty-body 400 response.
func TestGuestControllerRouteDispatcherRequiresTypedJSONBody(t *testing.T) {
	dispatcher, err := NewGuestControllerRouteDispatcher(&guestTypedRouteTestController{})
	if err != nil {
		t.Fatalf("expected typed dispatcher creation to succeed, got error: %v", err)
	}

	response, err := dispatcher.HandleRequest(&BridgeRequestEnvelopeV1{
		Route: &RouteMatchSnapshotV1{
			Method:      "PUT",
			RequestType: "UpdateDemoReq",
			PathParams:  map[string]string{"id": "demo-1"},
		},
		Request: &HTTPRequestSnapshotV1{Method: "PUT"},
	})
	if err != nil {
		t.Fatalf("expected missing typed JSON body to return bridge response, got error: %v", err)
	}
	if response == nil || response.StatusCode != 400 {
		t.Fatalf("expected 400 response for missing typed JSON body, got %#v", response)
	}
}

// TestGuestControllerRouteDispatcherSupportsTypedManualResponse verifies typed
// guest handlers can write raw responses through context helpers.
func TestGuestControllerRouteDispatcherSupportsTypedManualResponse(t *testing.T) {
	dispatcher, err := NewGuestControllerRouteDispatcher(&guestTypedRouteTestController{})
	if err != nil {
		t.Fatalf("expected typed dispatcher creation to succeed, got error: %v", err)
	}

	response, err := dispatcher.HandleRequest(&BridgeRequestEnvelopeV1{
		Route: &RouteMatchSnapshotV1{
			Method:       "GET",
			RequestType:  "DownloadReq",
			InternalPath: "/download",
			PathParams:   map[string]string{"id": "demo.txt"},
		},
		Request: &HTTPRequestSnapshotV1{Method: "GET"},
	})
	if err != nil {
		t.Fatalf("expected typed manual response dispatch to succeed, got error: %v", err)
	}
	if response == nil || response.StatusCode != 200 || response.ContentType != "text/plain; charset=utf-8" {
		t.Fatalf("expected text response for typed manual response, got %#v", response)
	}
	if string(response.Body) != "demo.txt" {
		t.Fatalf("expected manual response body demo.txt, got %q", string(response.Body))
	}
	if got := response.Headers["Content-Disposition"]; len(got) != 1 || got[0] != `attachment; filename="demo.txt"` {
		t.Fatalf("expected attachment header, got %#v", response.Headers)
	}
}

// TestGuestControllerRouteDispatcherSupportsTypedResponseErrors verifies typed
// handlers can surface prebuilt bridge responses through ResponseError.
func TestGuestControllerRouteDispatcherSupportsTypedResponseErrors(t *testing.T) {
	dispatcher, err := NewGuestControllerRouteDispatcher(&guestTypedRouteTestController{})
	if err != nil {
		t.Fatalf("expected typed dispatcher creation to succeed, got error: %v", err)
	}

	response, err := dispatcher.HandleRequest(&BridgeRequestEnvelopeV1{
		Route: &RouteMatchSnapshotV1{
			Method:      "GET",
			RequestType: "ClassifiedReq",
		},
		Request: &HTTPRequestSnapshotV1{Method: "GET"},
	})
	if err != nil {
		t.Fatalf("expected classified typed error to return bridge response, got error: %v", err)
	}
	if response == nil || response.StatusCode != 403 {
		t.Fatalf("expected 403 response for typed response error, got %#v", response)
	}
}
