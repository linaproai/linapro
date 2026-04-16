// This file provides JSON conversion helpers used by plugindb plans and
// builders.

package shared

import (
	"encoding/json"
	"reflect"

	"github.com/gogf/gf/v2/errors/gerror"
)

// MarshalValueJSON encodes one value into JSON bytes.
func MarshalValueJSON(value any) ([]byte, error) {
	if value == nil {
		return nil, nil
	}
	return json.Marshal(value)
}

// MarshalValuesJSON encodes one slice or array of values into JSON bytes.
func MarshalValuesJSON(values any) ([][]byte, error) {
	if values == nil {
		return nil, nil
	}
	refValue := reflect.ValueOf(values)
	if refValue.Kind() != reflect.Slice && refValue.Kind() != reflect.Array {
		return nil, gerror.New("plugindb in operator expects a slice or array")
	}
	encoded := make([][]byte, 0, refValue.Len())
	for index := 0; index < refValue.Len(); index++ {
		itemJSON, err := MarshalValueJSON(refValue.Index(index).Interface())
		if err != nil {
			return nil, err
		}
		encoded = append(encoded, itemJSON)
	}
	return encoded, nil
}

// UnmarshalValueJSON decodes one JSON-encoded value.
func UnmarshalValueJSON(data []byte) (any, error) {
	if len(data) == 0 {
		return nil, nil
	}
	var value any
	if err := json.Unmarshal(data, &value); err != nil {
		return nil, err
	}
	return value, nil
}

// UnmarshalValuesJSON decodes one list of JSON-encoded values.
func UnmarshalValuesJSON(items [][]byte) ([]any, error) {
	if len(items) == 0 {
		return nil, nil
	}
	decoded := make([]any, 0, len(items))
	for _, item := range items {
		value, err := UnmarshalValueJSON(item)
		if err != nil {
			return nil, err
		}
		decoded = append(decoded, value)
	}
	return decoded, nil
}

// NewEQFilter builds one equality filter.
func NewEQFilter(field string, value any) (*DataFilter, error) {
	valueJSON, err := MarshalValueJSON(value)
	if err != nil {
		return nil, err
	}
	return &DataFilter{Field: field, Operator: DataFilterOperatorEQ, ValueJSON: valueJSON}, nil
}

// NewINFilter builds one list-membership filter.
func NewINFilter(field string, values any) (*DataFilter, error) {
	valuesJSON, err := MarshalValuesJSON(values)
	if err != nil {
		return nil, err
	}
	return &DataFilter{Field: field, Operator: DataFilterOperatorIN, ValuesJSON: valuesJSON}, nil
}

// NewLikeFilter builds one wildcard filter.
func NewLikeFilter(field string, value any) (*DataFilter, error) {
	valueJSON, err := MarshalValueJSON(value)
	if err != nil {
		return nil, err
	}
	return &DataFilter{Field: field, Operator: DataFilterOperatorLike, ValueJSON: valueJSON}, nil
}

// NewASCOrder builds one ascending order clause.
func NewASCOrder(field string) *DataOrder {
	return &DataOrder{Field: field, Direction: DataOrderDirectionASC}
}

// NewDESCOrder builds one descending order clause.
func NewDESCOrder(field string) *DataOrder {
	return &DataOrder{Field: field, Direction: DataOrderDirectionDESC}
}

// MarshalQueryPlanJSON encodes one typed query plan into JSON bytes.
func MarshalQueryPlanJSON(plan *DataQueryPlan) ([]byte, error) {
	if err := ValidateDataQueryPlan(plan); err != nil {
		return nil, err
	}
	return json.Marshal(plan)
}

// UnmarshalQueryPlanJSON decodes one typed query plan from JSON bytes.
func UnmarshalQueryPlanJSON(data []byte) (*DataQueryPlan, error) {
	if len(data) == 0 {
		return nil, nil
	}
	plan := &DataQueryPlan{}
	if err := json.Unmarshal(data, plan); err != nil {
		return nil, err
	}
	if err := ValidateDataQueryPlan(plan); err != nil {
		return nil, err
	}
	return plan, nil
}
