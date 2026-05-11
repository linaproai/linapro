// This file verifies the public operation-log filter DTO contract.

package v1

import (
	"reflect"
	"strings"
	"testing"
)

// TestOperLogFilterRequestsExposeOnlyOperType verifies list and export filters
// do not publish a duplicate action-kind query parameter.
func TestOperLogFilterRequestsExposeOnlyOperType(t *testing.T) {
	requestTypes := []reflect.Type{
		reflect.TypeOf(ListReq{}),
		reflect.TypeOf(ExportReq{}),
	}
	duplicateAlias := "action" + "Kind"

	for _, requestType := range requestTypes {
		if !hasJSONField(requestType, "operType") {
			t.Fatalf("%s must expose operType filter", requestType.Name())
		}
		if hasJSONField(requestType, duplicateAlias) {
			t.Fatalf("%s must not expose duplicate action-kind filter", requestType.Name())
		}
	}
}

// hasJSONField reports whether a request DTO publishes one json-tagged field.
func hasJSONField(requestType reflect.Type, fieldName string) bool {
	for i := 0; i < requestType.NumField(); i++ {
		tag := requestType.Field(i).Tag.Get("json")
		name, _, _ := strings.Cut(tag, ",")
		if name == fieldName {
			return true
		}
	}
	return false
}
