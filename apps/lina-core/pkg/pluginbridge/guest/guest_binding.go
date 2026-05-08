// This file binds bridge envelope inputs onto typed guest request DTOs so
// dynamic plugin controllers can use explicit req/res method signatures.

package guest

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
)

// bindJSONBody decodes the envelope request body into the supplied target
// value while preserving the public BindJSON sentinel errors.
func bindJSONBody(request *BridgeRequestEnvelopeV1, target interface{}) error {
	if request == nil || request.Request == nil || len(request.Request.Body) == 0 {
		return errGuestBindJSONEmptyBody
	}
	if err := json.Unmarshal(request.Request.Body, target); err != nil {
		return gerror.Wrap(errGuestBindJSONInvalidJSON, err.Error())
	}
	return nil
}

// bindGuestRequestDTO hydrates one typed guest request DTO from the bridge
// envelope body, path params, and query string values.
func bindGuestRequestDTO(request *BridgeRequestEnvelopeV1, target interface{}) error {
	targetValue := reflect.ValueOf(target)
	if !targetValue.IsValid() || targetValue.Kind() != reflect.Pointer || targetValue.IsNil() {
		return gerror.New("typed guest request target must be a non-nil pointer")
	}
	if targetValue.Elem().Kind() != reflect.Struct {
		return gerror.New("typed guest request target must point to a struct")
	}

	if shouldBindGuestJSONBody(request, targetValue.Elem().Type()) {
		if err := bindJSONBody(request, target); err != nil {
			return err
		}
	}

	applyGuestRouteValues(targetValue.Elem(), request)
	return nil
}

// shouldBindGuestJSONBody reports whether one typed guest request DTO should
// treat the request body as a required JSON payload.
func shouldBindGuestJSONBody(request *BridgeRequestEnvelopeV1, structType reflect.Type) bool {
	if request != nil && request.Request != nil && len(request.Request.Body) > 0 {
		return true
	}

	method := ""
	if request != nil {
		if request.Request != nil {
			method = strings.ToUpper(strings.TrimSpace(request.Request.Method))
		}
		if method == "" && request.Route != nil {
			method = strings.ToUpper(strings.TrimSpace(request.Route.Method))
		}
	}
	switch method {
	case "POST", "PUT", "PATCH":
	default:
		return false
	}

	pathKeys := map[string]struct{}{}
	queryKeys := map[string]struct{}{}
	if request != nil && request.Route != nil {
		for key := range request.Route.PathParams {
			pathKeys[key] = struct{}{}
		}
		for key := range request.Route.QueryValues {
			queryKeys[key] = struct{}{}
		}
	}

	requiredBody := false
	walkGuestRequestFields(structType, func(field reflect.StructField, jsonName string) bool {
		if _, ok := pathKeys[jsonName]; ok {
			return true
		}
		if _, ok := queryKeys[jsonName]; ok {
			return true
		}
		requiredBody = true
		return false
	})
	return requiredBody
}

// applyGuestRouteValues overlays path and query values onto the already
// decoded typed guest request DTO.
func applyGuestRouteValues(target reflect.Value, request *BridgeRequestEnvelopeV1) {
	if !target.IsValid() || target.Kind() != reflect.Struct || request == nil || request.Route == nil {
		return
	}

	walkGuestRequestFields(target.Type(), func(field reflect.StructField, jsonName string) bool {
		fieldValue := target.FieldByIndex(field.Index)
		if !fieldValue.IsValid() || !fieldValue.CanSet() {
			return true
		}

		if raw, ok := request.Route.PathParams[jsonName]; ok {
			setGuestRequestFieldValue(fieldValue, raw, true)
			return true
		}
		if values, ok := request.Route.QueryValues[jsonName]; ok && len(values) > 0 {
			setGuestRequestFieldValue(fieldValue, values[0], false)
		}
		return true
	})
}

// walkGuestRequestFields visits each exported non-anonymous field that exposes
// one usable JSON name. Returning false from visit stops iteration.
func walkGuestRequestFields(structType reflect.Type, visit func(field reflect.StructField, jsonName string) bool) {
	for index := 0; index < structType.NumField(); index++ {
		field := structType.Field(index)
		if field.PkgPath != "" || field.Anonymous {
			continue
		}

		jsonName := guestJSONFieldName(field)
		if jsonName == "" {
			continue
		}
		if !visit(field, jsonName) {
			return
		}
	}
}

// guestJSONFieldName reads one field's JSON tag name and ignores unsupported
// or explicitly skipped fields.
func guestJSONFieldName(field reflect.StructField) string {
	jsonTag := strings.TrimSpace(field.Tag.Get("json"))
	if jsonTag == "" || jsonTag == "-" {
		return ""
	}

	name := jsonTag
	if index := strings.Index(name, ","); index >= 0 {
		name = name[:index]
	}
	return strings.TrimSpace(name)
}

// setGuestRequestFieldValue parses one route string into the typed request
// field. Parse failures keep the field zero-valued to match the previous
// helper semantics used by dynamic plugin controllers.
func setGuestRequestFieldValue(fieldValue reflect.Value, raw string, isPathParam bool) {
	normalized := strings.TrimSpace(raw)
	switch fieldValue.Kind() {
	case reflect.String:
		fieldValue.SetString(normalized)
	case reflect.Bool:
		fieldValue.SetBool(parseGuestBooleanValue(normalized, isPathParam))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		parsedValue, err := strconv.ParseInt(normalized, 10, 64)
		if err != nil {
			return
		}
		fieldValue.SetInt(parsedValue)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		parsedValue, err := strconv.ParseUint(normalized, 10, 64)
		if err != nil {
			return
		}
		fieldValue.SetUint(parsedValue)
	}
}

// parseGuestBooleanValue normalizes one path/query string into the truthy
// boolean semantics already used by guest controller helpers.
func parseGuestBooleanValue(value string, isPathParam bool) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off", "":
		return false
	}
	if isPathParam {
		parsedValue, err := strconv.ParseBool(strings.TrimSpace(value))
		if err == nil {
			return parsedValue
		}
	}
	return false
}
