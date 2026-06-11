// Package hostservicedispatch provides explicit service/method registration for
// dynamic-plugin host service handlers. It owns only dispatch lookup and common
// response helpers; the parent wasm package keeps runtime dependency ownership.
package hostservicedispatch

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	bridgehostcall "lina-core/pkg/plugin/pluginbridge/protocol"
)

// Context carries one authorized host-service invocation into a registered
// handler without exposing the parent wasm package's private context type.
type Context struct {
	// HostContext carries the parent package request context. Parent-owned
	// adapters validate and cast it before calling existing handlers.
	HostContext any
	// Service is the host-service family name.
	Service string
	// Method is the service method name.
	Method string
	// ResourceRef is the authorized resource reference for resource-scoped calls.
	ResourceRef string
	// Table is the authorized table name for data service calls.
	Table string
	// Payload is the raw method payload.
	Payload []byte
}

// Handler dispatches one authorized host-service invocation.
type Handler func(context.Context, Context) *bridgehostcall.HostCallResponseEnvelope

// Registry stores explicitly registered host-service handlers.
type Registry struct {
	handlers map[string]Handler
	methods  []Method
}

// Method describes one registered service/method pair.
type Method struct {
	// Service is the host-service family name.
	Service string
	// Method is the host-service method name.
	Method string
}

// NewRegistry creates an empty host-service dispatch registry.
func NewRegistry() *Registry {
	return &Registry{handlers: make(map[string]Handler)}
}

// Register binds one service/method pair to a handler and rejects duplicates.
func (r *Registry) Register(service string, method string, handler Handler) error {
	if r == nil {
		return gerror.New("host service dispatch registry is nil")
	}
	service = strings.TrimSpace(service)
	method = strings.TrimSpace(method)
	if service == "" || method == "" {
		return gerror.New("host service dispatch registration requires service and method")
	}
	if handler == nil {
		return gerror.Newf("host service dispatch handler is nil: %s.%s", service, method)
	}
	key := registryKey(service, method)
	if _, ok := r.handlers[key]; ok {
		return gerror.Newf("host service dispatch handler already registered: %s.%s", service, method)
	}
	r.handlers[key] = handler
	r.methods = append(r.methods, Method{Service: service, Method: method})
	return nil
}

// Lookup resolves a registered host-service handler.
func (r *Registry) Lookup(service string, method string) (Handler, bool) {
	if r == nil {
		return nil, false
	}
	handler, ok := r.handlers[registryKey(service, method)]
	return handler, ok
}

// Methods returns a snapshot of all registered service/method pairs.
func (r *Registry) Methods() []Method {
	if r == nil || len(r.methods) == 0 {
		return nil
	}
	methods := make([]Method, len(r.methods))
	copy(methods, r.methods)
	return methods
}

// Dispatch invokes a registered handler or returns a not-found response.
func (r *Registry) Dispatch(ctx context.Context, input Context) *bridgehostcall.HostCallResponseEnvelope {
	handler, ok := r.Lookup(input.Service, input.Method)
	if !ok {
		return NotFound(input.Service, input.Method)
	}
	return handler(ctx, input)
}

// NotFound returns the common error response for missing dispatch handlers.
func NotFound(service string, method string) *bridgehostcall.HostCallResponseEnvelope {
	return bridgehostcall.NewHostCallErrorResponse(
		bridgehostcall.HostCallStatusNotFound,
		"host service method not registered: "+strings.TrimSpace(service)+"."+strings.TrimSpace(method),
	)
}

func registryKey(service string, method string) string {
	return strings.TrimSpace(service) + "\x00" + strings.TrimSpace(method)
}
