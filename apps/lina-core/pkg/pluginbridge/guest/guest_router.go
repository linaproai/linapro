// This file dispatches guest bridge requests to reflected controller methods
// and provides the guest-side route fallback mapping helpers.

package guest

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

// GuestControllerHandlerKind identifies the bridge signature shape used by one
// reflected guest controller handler.
type GuestControllerHandlerKind string

const (
	// GuestControllerHandlerKindEnvelope identifies handlers that receive the
	// full bridge request envelope and return a bridge response envelope.
	GuestControllerHandlerKindEnvelope GuestControllerHandlerKind = "envelope"
	// GuestControllerHandlerKindTyped identifies handlers that receive
	// context.Context plus a typed request DTO and return a typed response DTO.
	GuestControllerHandlerKindTyped GuestControllerHandlerKind = "typed"
)

// GuestControllerHandlerMetadata describes one controller method accepted by
// the reflected guest dispatcher.
type GuestControllerHandlerMetadata struct {
	// MethodName is the exported Go method name on the guest controller.
	MethodName string
	// RequestType is the dispatcher request type used for primary lookup.
	RequestType string
	// InternalPath is the dispatcher fallback path derived from the method name.
	InternalPath string
	// Kind records which supported handler signature matched the method.
	Kind GuestControllerHandlerKind
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

// DiscoverGuestControllerHandlers returns the reflected metadata for all
// controller methods that the guest dispatcher would register.
func DiscoverGuestControllerHandlers(controller any) ([]GuestControllerHandlerMetadata, error) {
	metadata, err := collectGuestControllerHandlerMetadata(controller)
	if err != nil {
		return nil, err
	}
	items := make([]GuestControllerHandlerMetadata, len(metadata))
	copy(items, metadata)
	return items, nil
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
	metadata, err := collectGuestControllerHandlerMetadata(controller)
	if err != nil {
		return err
	}
	controllerValue := reflect.ValueOf(controller)
	if d.handlersByRequestType == nil {
		d.handlersByRequestType = make(map[string]GuestHandler)
	}
	if d.handlersByPath == nil {
		d.handlersByPath = make(map[string]GuestHandler)
	}

	controllerType := controllerValue.Type()
	methodsByName := make(map[string]reflect.Method, controllerType.NumMethod())
	for index := 0; index < controllerType.NumMethod(); index++ {
		methodsByName[controllerType.Method(index).Name] = controllerType.Method(index)
	}
	for _, item := range metadata {
		method, ok := methodsByName[item.MethodName]
		if !ok {
			return gerror.Newf("guest route controller method metadata has no method: %s", item.MethodName)
		}
		handler, err := buildGuestControllerHandler(controllerValue, method, item)
		if err != nil {
			return err
		}
		if _, exists := d.handlersByRequestType[item.RequestType]; exists {
			return gerror.Newf("guest route request type already registered: %s", item.RequestType)
		}
		if _, exists := d.handlersByPath[item.InternalPath]; exists {
			return gerror.Newf("guest route internal path already registered: %s", item.InternalPath)
		}
		d.handlersByRequestType[item.RequestType] = handler
		d.handlersByPath[item.InternalPath] = handler
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

// BuildGuestControllerInternalPath converts a controller method name to the
// kebab-case internal path used as the dispatcher fallback key.
func BuildGuestControllerInternalPath(methodName string) string {
	return buildGuestControllerInternalPath(methodName)
}

// collectGuestControllerHandlerMetadata discovers all methods accepted by the
// reflected guest dispatcher and validates lookup-key uniqueness.
func collectGuestControllerHandlerMetadata(controller any) ([]GuestControllerHandlerMetadata, error) {
	controllerValue := reflect.ValueOf(controller)
	if !controllerValue.IsValid() {
		return nil, gerror.New("guest route controller cannot be nil")
	}
	if controllerValue.Kind() == reflect.Pointer && controllerValue.IsNil() {
		return nil, gerror.New("guest route controller cannot be nil")
	}

	controllerType := controllerValue.Type()
	items := make([]GuestControllerHandlerMetadata, 0, controllerType.NumMethod())
	seenRequestTypes := make(map[string]struct{}, controllerType.NumMethod())
	seenInternalPaths := make(map[string]struct{}, controllerType.NumMethod())
	for index := 0; index < controllerType.NumMethod(); index++ {
		method := controllerType.Method(index)
		item, ok, err := buildGuestControllerHandlerMetadata(method)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		if _, exists := seenRequestTypes[item.RequestType]; exists {
			return nil, gerror.Newf("guest route request type already registered: %s", item.RequestType)
		}
		seenRequestTypes[item.RequestType] = struct{}{}
		if _, exists := seenInternalPaths[item.InternalPath]; exists {
			return nil, gerror.Newf("guest route internal path already registered: %s", item.InternalPath)
		}
		seenInternalPaths[item.InternalPath] = struct{}{}
		items = append(items, item)
	}
	if len(items) == 0 {
		return nil, gerror.Newf(
			"guest route controller %T does not expose any bridge handler methods",
			controller,
		)
	}
	return items, nil
}

// buildGuestControllerHandlerMetadata creates dispatcher metadata for one
// reflected method when it matches a supported guest bridge signature.
func buildGuestControllerHandlerMetadata(method reflect.Method) (GuestControllerHandlerMetadata, bool, error) {
	switch {
	case isGuestEnvelopeControllerHandlerMethod(method.Type):
		return GuestControllerHandlerMetadata{
			MethodName:   method.Name,
			RequestType:  method.Name + "Req",
			InternalPath: buildGuestControllerInternalPath(method.Name),
			Kind:         GuestControllerHandlerKindEnvelope,
		}, true, nil
	case isGuestTypedControllerHandlerMethod(method.Type):
		requestType := strings.TrimSpace(method.Type.In(2).Elem().Name())
		if requestType == "" {
			return GuestControllerHandlerMetadata{}, false, gerror.Newf("typed guest controller request DTO name is empty: %s", method.Name)
		}
		return GuestControllerHandlerMetadata{
			MethodName:   method.Name,
			RequestType:  requestType,
			InternalPath: buildGuestControllerInternalPath(method.Name),
			Kind:         GuestControllerHandlerKindTyped,
		}, true, nil
	default:
		return GuestControllerHandlerMetadata{}, false, nil
	}
}

// buildGuestControllerHandler creates one dispatch handler for either the
// legacy envelope signature or the typed guest controller signature.
func buildGuestControllerHandler(
	controllerValue reflect.Value,
	method reflect.Method,
	metadata GuestControllerHandlerMetadata,
) (GuestHandler, error) {
	switch metadata.Kind {
	case GuestControllerHandlerKindEnvelope:
		return buildGuestEnvelopeControllerHandler(controllerValue, method), nil
	case GuestControllerHandlerKindTyped:
		return buildGuestTypedControllerHandler(controllerValue, method), nil
	default:
		return nil, gerror.Newf("guest route handler kind is unsupported: %s", metadata.Kind)
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
