// This file dispatches guest bridge requests to reflected controller methods
// and provides the guest-side route fallback mapping helpers.

package pluginbridge

import (
	"context"
	"reflect"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
)

// Reflected bridge envelope types reused during guest controller registration.
var (
	bridgeRequestEnvelopeType  = reflect.TypeOf(&BridgeRequestEnvelopeV1{})
	bridgeResponseEnvelopeType = reflect.TypeOf(&BridgeResponseEnvelopeV1{})
	contextInterfaceType       = reflect.TypeOf((*context.Context)(nil)).Elem()
	errorInterfaceType         = reflect.TypeOf((*error)(nil)).Elem()
)

// GuestControllerRouteDispatcher exposes reflected guest controller
// registration and request dispatch published to dynamic plugins.
type GuestControllerRouteDispatcher interface {
	// RegisterController registers all exported controller methods whose signature matches the guest bridge contract.
	RegisterController(controller any) error
	// HandleRequest dispatches the guest bridge request to the registered controller method.
	HandleRequest(request *BridgeRequestEnvelopeV1) (*BridgeResponseEnvelopeV1, error)
}

// guestControllerRouteDispatcher dispatches guest bridge requests to
// controller methods registered by reflection.
type guestControllerRouteDispatcher struct {
	handlersByRequestType map[string]GuestHandler
	handlersByPath        map[string]GuestHandler
}

// NewGuestControllerRouteDispatcher creates one reflection-based dispatcher for
// the provided controller object.
func NewGuestControllerRouteDispatcher(controller any) (GuestControllerRouteDispatcher, error) {
	dispatcher := &guestControllerRouteDispatcher{
		handlersByRequestType: make(map[string]GuestHandler),
		handlersByPath:        make(map[string]GuestHandler),
	}
	if err := dispatcher.RegisterController(controller); err != nil {
		return nil, err
	}
	return dispatcher, nil
}

// MustNewGuestControllerRouteDispatcher creates one reflection-based
// dispatcher and panics when registration fails.
func MustNewGuestControllerRouteDispatcher(controller any) GuestControllerRouteDispatcher {
	dispatcher, err := NewGuestControllerRouteDispatcher(controller)
	if err != nil {
		panic(err)
	}
	return dispatcher
}

// RegisterController registers all exported controller methods whose
// signature matches either `func(*BridgeRequestEnvelopeV1) (*BridgeResponseEnvelopeV1, error)`
// or `func(context.Context, *Req) (*Res, error)`. Typed handlers are exposed
// under the request DTO type name so runtime RequestType contracts can reuse
// the backend API DTO declaration directly.
func (d *guestControllerRouteDispatcher) RegisterController(controller any) error {
	if d == nil {
		return gerror.New("guest controller route dispatcher is nil")
	}

	controllerValue := reflect.ValueOf(controller)
	if !controllerValue.IsValid() {
		return gerror.New("guest route controller cannot be nil")
	}
	if controllerValue.Kind() == reflect.Pointer && controllerValue.IsNil() {
		return gerror.New("guest route controller cannot be nil")
	}
	if d.handlersByRequestType == nil {
		d.handlersByRequestType = make(map[string]GuestHandler)
	}
	if d.handlersByPath == nil {
		d.handlersByPath = make(map[string]GuestHandler)
	}

	controllerType := controllerValue.Type()
	registeredCount := 0
	for index := 0; index < controllerType.NumMethod(); index++ {
		method := controllerType.Method(index)
		handler, requestType, internalPath, ok, err := buildGuestControllerHandler(
			controllerValue,
			method,
		)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}
		if _, exists := d.handlersByRequestType[requestType]; exists {
			return gerror.Newf("guest route request type already registered: %s", requestType)
		}
		if _, exists := d.handlersByPath[internalPath]; exists {
			return gerror.Newf("guest route internal path already registered: %s", internalPath)
		}
		d.handlersByRequestType[requestType] = handler
		d.handlersByPath[internalPath] = handler
		registeredCount++
	}

	if registeredCount == 0 {
		return gerror.Newf(
			"guest route controller %T does not expose any bridge handler methods",
			controller,
		)
	}
	return nil
}

// HandleRequest dispatches the guest bridge request to the registered
// controller method resolved from request.Route.RequestType.
func (d *guestControllerRouteDispatcher) HandleRequest(
	request *BridgeRequestEnvelopeV1,
) (*BridgeResponseEnvelopeV1, error) {
	if request == nil || request.Route == nil {
		return NewBadRequestResponse("Dynamic bridge request is missing route metadata"), nil
	}

	if d == nil || len(d.handlersByRequestType) == 0 {
		return NewInternalErrorResponse("Dynamic guest route dispatcher is not initialized"), nil
	}

	requestType := strings.TrimSpace(request.Route.RequestType)
	if requestType != "" {
		handler, ok := d.handlersByRequestType[requestType]
		if ok {
			return handler(request)
		}
	}

	internalPath := strings.TrimSpace(request.Route.InternalPath)
	if internalPath != "" {
		handler, ok := d.handlersByPath[internalPath]
		if ok {
			return handler(request)
		}
	}

	if requestType == "" && internalPath == "" {
		return NewBadRequestResponse("Dynamic bridge request is missing route request type"), nil
	}

	handler, ok := d.handlersByRequestType[requestType]
	if !ok {
		return NewNotFoundResponse("Dynamic bridge route not found"), nil
	}
	return handler(request)
}

// isGuestEnvelopeControllerHandlerMethod reports whether one reflected method
// matches the legacy envelope-based bridge controller signature.
func isGuestEnvelopeControllerHandlerMethod(methodType reflect.Type) bool {
	return methodType.NumIn() == 2 &&
		methodType.In(1) == bridgeRequestEnvelopeType &&
		methodType.NumOut() == 2 &&
		methodType.Out(0) == bridgeResponseEnvelopeType &&
		methodType.Out(1) == errorInterfaceType
}

// isGuestTypedControllerHandlerMethod reports whether one reflected method
// matches the typed guest controller signature.
func isGuestTypedControllerHandlerMethod(methodType reflect.Type) bool {
	return methodType.NumIn() == 3 &&
		methodType.In(1).Implements(contextInterfaceType) &&
		methodType.In(2).Kind() == reflect.Pointer &&
		methodType.In(2).Elem().Kind() == reflect.Struct &&
		methodType.NumOut() == 2 &&
		methodType.Out(0).Kind() == reflect.Pointer &&
		methodType.Out(0).Elem().Kind() == reflect.Struct &&
		methodType.Out(1) == errorInterfaceType
}

// buildGuestControllerHandler creates one dispatch handler for either the
// legacy envelope signature or the typed guest controller signature.
func buildGuestControllerHandler(
	controllerValue reflect.Value,
	method reflect.Method,
) (GuestHandler, string, string, bool, error) {
	switch {
	case isGuestEnvelopeControllerHandlerMethod(method.Type):
		return buildGuestEnvelopeControllerHandler(controllerValue, method), method.Name + "Req", buildGuestControllerInternalPath(method.Name), true, nil
	case isGuestTypedControllerHandlerMethod(method.Type):
		requestType := strings.TrimSpace(method.Type.In(2).Elem().Name())
		if requestType == "" {
			return nil, "", "", false, gerror.Newf("typed guest controller request DTO name is empty: %s", method.Name)
		}
		return buildGuestTypedControllerHandler(controllerValue, method), requestType, buildGuestControllerInternalPath(method.Name), true, nil
	default:
		return nil, "", "", false, nil
	}
}

// buildGuestEnvelopeControllerHandler creates one dispatcher closure for the
// legacy envelope-based guest controller method.
func buildGuestEnvelopeControllerHandler(
	controllerValue reflect.Value,
	method reflect.Method,
) GuestHandler {
	methodFunc := method.Func
	return func(request *BridgeRequestEnvelopeV1) (*BridgeResponseEnvelopeV1, error) {
		requestValue := reflect.Zero(bridgeRequestEnvelopeType)
		if request != nil {
			requestValue = reflect.ValueOf(request)
		}
		outputs := methodFunc.Call([]reflect.Value{controllerValue, requestValue})

		var response *BridgeResponseEnvelopeV1
		if !outputs[0].IsNil() {
			response, _ = outputs[0].Interface().(*BridgeResponseEnvelopeV1)
		}

		var err error
		if !outputs[1].IsNil() {
			err, _ = outputs[1].Interface().(error)
		}
		if responseFromErr := ResponseFromError(err); responseFromErr != nil {
			return responseFromErr, nil
		}
		return response, err
	}
}

// buildGuestTypedControllerHandler creates one dispatcher closure for a typed
// guest controller method using `context.Context` plus API DTOs.
func buildGuestTypedControllerHandler(
	controllerValue reflect.Value,
	method reflect.Method,
) GuestHandler {
	methodFunc := method.Func
	requestDTOType := method.Type.In(2)
	return func(request *BridgeRequestEnvelopeV1) (*BridgeResponseEnvelopeV1, error) {
		ctx := newGuestControllerContext(request)

		requestDTOValue := reflect.New(requestDTOType.Elem())
		if err := bindGuestRequestDTO(request, requestDTOValue.Interface()); err != nil {
			if response := ClassifyBindJSONError(err); response != nil {
				return response, nil
			}
			return NewBadRequestResponse(err.Error()), nil
		}

		outputs := methodFunc.Call([]reflect.Value{
			controllerValue,
			reflect.ValueOf(ctx),
			requestDTOValue,
		})

		var payload interface{}
		if !outputs[0].IsNil() {
			payload = outputs[0].Interface()
		}

		var err error
		if !outputs[1].IsNil() {
			err, _ = outputs[1].Interface().(error)
		}
		if responseFromErr := ResponseFromError(err); responseFromErr != nil {
			return responseFromErr, nil
		}
		if err != nil {
			return nil, err
		}
		return buildGuestControllerResponse(ctx, payload)
	}
}

// buildGuestControllerInternalPath converts a controller method name to the
// kebab-case internal path used as the dispatcher fallback key.
func buildGuestControllerInternalPath(methodName string) string {
	if methodName == "" {
		return "/"
	}

	var builder strings.Builder
	builder.WriteByte('/')
	for index, r := range methodName {
		if 'A' <= r && r <= 'Z' {
			if index > 0 {
				builder.WriteByte('-')
			}
			builder.WriteByte(byte(r + ('a' - 'A')))
			continue
		}
		builder.WriteRune(r)
	}
	return builder.String()
}
