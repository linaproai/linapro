// This file provides JSON conversion helpers used by record store capability plans and
// builders.

package plan

import (
	"bytes"
	"encoding/json"
	"reflect"
	"sort"

	"github.com/gogf/gf/v2/errors/gerror"
)

// MarshalValueJSON encodes one value into JSON bytes.
func MarshalValueJSON(value any) ([]byte, error) {
	if value == nil {
		return nil, nil
	}
	if record, ok := value.(map[string]any); ok {
		return marshalStringAnyMapJSON(record)
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
		return nil, gerror.New("record store capability in operator expects a slice or array")
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
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	return value, nil
}

// marshalStringAnyMapJSON normalizes generic map values through the JSON
// decoder before the final encode. The extra round trip keeps numeric values as
// json.Number and avoids wasip1-specific float formatting panics observed when
// Go's JSON encoder receives directly constructed map[string]any mutation
// records from dynamic plugins.
func marshalStringAnyMapJSON(record map[string]any) ([]byte, error) {
	if record == nil {
		return nil, nil
	}
	content, err := marshalOrderedMapJSON(record)
	if err != nil {
		return nil, err
	}
	normalized := make(map[string]any)
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.UseNumber()
	if err = decoder.Decode(&normalized); err != nil {
		return nil, err
	}
	return marshalOrderedMapJSON(normalized)
}

// marshalOrderedMapJSON encodes map entries in a deterministic order.
func marshalOrderedMapJSON(record map[string]any) ([]byte, error) {
	keys := make([]string, 0, len(record))
	for key := range record {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var builder bytes.Buffer
	builder.WriteByte('{')
	for index, key := range keys {
		if index > 0 {
			builder.WriteByte(',')
		}
		keyJSON, err := json.Marshal(key)
		if err != nil {
			return nil, err
		}
		valueJSON, err := json.Marshal(record[key])
		if err != nil {
			return nil, err
		}
		builder.Write(keyJSON)
		builder.WriteByte(':')
		builder.Write(valueJSON)
	}
	builder.WriteByte('}')
	return builder.Bytes(), nil
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
func NewEQFilter(field string, value any) (*Filter, error) {
	valueJSON, err := MarshalValueJSON(value)
	if err != nil {
		return nil, err
	}
	return &Filter{Field: field, Operator: FilterOperatorEQ, ValueJSON: valueJSON}, nil
}

// NewINFilter builds one list-membership filter.
func NewINFilter(field string, values any) (*Filter, error) {
	valuesJSON, err := MarshalValuesJSON(values)
	if err != nil {
		return nil, err
	}
	return &Filter{Field: field, Operator: FilterOperatorIN, ValuesJSON: valuesJSON}, nil
}

// NewLikeFilter builds one wildcard filter.
func NewLikeFilter(field string, value any) (*Filter, error) {
	valueJSON, err := MarshalValueJSON(value)
	if err != nil {
		return nil, err
	}
	return &Filter{Field: field, Operator: FilterOperatorLike, ValueJSON: valueJSON}, nil
}

// NewASCOrder builds one ascending order clause.
func NewASCOrder(field string) *Order {
	return &Order{Field: field, Direction: OrderDirectionASC}
}

// NewDESCOrder builds one descending order clause.
func NewDESCOrder(field string) *Order {
	return &Order{Field: field, Direction: OrderDirectionDESC}
}

// MarshalQueryPlanJSON encodes one typed query plan into JSON bytes.
func MarshalQueryPlanJSON(plan *QueryPlan) ([]byte, error) {
	if err := ValidateQueryPlan(plan); err != nil {
		return nil, err
	}
	return json.Marshal(plan)
}

// UnmarshalQueryPlanJSON decodes one typed query plan from JSON bytes.
func UnmarshalQueryPlanJSON(data []byte) (*QueryPlan, error) {
	if len(data) == 0 {
		return nil, nil
	}
	plan := &QueryPlan{}
	if err := json.Unmarshal(data, plan); err != nil {
		return nil, err
	}
	if err := ValidateQueryPlan(plan); err != nil {
		return nil, err
	}
	return plan, nil
}
