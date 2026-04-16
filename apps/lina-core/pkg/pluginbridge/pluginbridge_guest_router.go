// This file dispatches guest bridge requests to reflected controller methods
// and provides the guest-side route fallback mapping helpers.

package pluginbridge

import (
	"reflect"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
)

var (
	bridgeRequestEnvelopeType  = reflect.TypeOf(&BridgeRequestEnvelopeV1{})
	bridgeResponseEnvelopeType = reflect.TypeOf(&BridgeResponseEnvelopeV1{})
	errorInterfaceType         = reflect.TypeOf((*error)(nil)).Elem()
)

// GuestControllerRouteDispatcher dispatches guest bridge requests to controller
// methods registered by reflection.
type GuestControllerRouteDispatcher struct {
	handlersByRequestType map[string]GuestHandler
	handlersByPath        map[string]GuestHandler
}

// NewGuestControllerRouteDispatcher creates one reflection-based dispatcher for
// the provided controller object.
func NewGuestControllerRouteDispatcher(controller any) (*GuestControllerRouteDispatcher, error) {
	dispatcher := &GuestControllerRouteDispatcher{
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
func MustNewGuestControllerRouteDispatcher(controller any) *GuestControllerRouteDispatcher {
	dispatcher, err := NewGuestControllerRouteDispatcher(controller)
	if err != nil {
		panic(err)
	}
	return dispatcher
}

// RegisterController registers all exported controller methods whose
// signature matches `func(*BridgeRequestEnvelopeV1) (*BridgeResponseEnvelopeV1, error)`.
// Each matching method is exposed under `{MethodName}Req`, which aligns with
// the build-time route contract RequestType extracted from backend API DTOs.
func (d *GuestControllerRouteDispatcher) RegisterController(controller any) error {
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
		if !isGuestControllerHandlerMethod(method.Type) {
			continue
		}

		requestType := method.Name + "Req"
		if _, exists := d.handlersByRequestType[requestType]; exists {
			return gerror.Newf("guest route request type already registered: %s", requestType)
		}
		internalPath := buildGuestControllerInternalPath(method.Name)
		if _, exists := d.handlersByPath[internalPath]; exists {
			return gerror.Newf("guest route internal path already registered: %s", internalPath)
		}

		var (
			methodFunc            = method.Func
			methodControllerValue = controllerValue
			handler               = func(request *BridgeRequestEnvelopeV1) (*BridgeResponseEnvelopeV1, error) {
				requestValue := reflect.Zero(bridgeRequestEnvelopeType)
				if request != nil {
					requestValue = reflect.ValueOf(request)
				}
				outputs := methodFunc.Call([]reflect.Value{methodControllerValue, requestValue})

				var response *BridgeResponseEnvelopeV1
				if !outputs[0].IsNil() {
					response, _ = outputs[0].Interface().(*BridgeResponseEnvelopeV1)
				}

				var err error
				if !outputs[1].IsNil() {
					err, _ = outputs[1].Interface().(error)
				}
				return response, err
			}
		)
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
func (d *GuestControllerRouteDispatcher) HandleRequest(
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

func isGuestControllerHandlerMethod(methodType reflect.Type) bool {
	return methodType.NumIn() == 2 &&
		methodType.In(1) == bridgeRequestEnvelopeType &&
		methodType.NumOut() == 2 &&
		methodType.Out(0) == bridgeResponseEnvelopeType &&
		methodType.Out(1) == errorInterfaceType
}

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
