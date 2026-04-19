// This file validates the restricted JSON Schema draft-07 subset accepted by
// scheduled-job handler parameters.

package jobhandler

import (
	"bytes"
	"encoding/json"
	"io"
	"reflect"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
)

// handlerSchemaRoot stores the supported root-level schema fields.
type handlerSchemaRoot struct {
	Type        string                        `json:"type"`
	Properties  map[string]handlerSchemaField `json:"properties"`
	Required    []string                      `json:"required"`
	Description string                        `json:"description"`
}

// handlerSchemaField stores the supported field-level schema fields.
type handlerSchemaField struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Default     any    `json:"default"`
	Enum        []any  `json:"enum"`
	Format      string `json:"format"`
}

// normalizeSchema validates one handler schema definition and returns the
// trimmed text that should be stored in the registry.
func normalizeSchema(schemaText string) (string, error) {
	trimmed := strings.TrimSpace(schemaText)
	if trimmed == "" {
		trimmed = `{"type":"object","properties":{}}`
	}
	if _, err := parseSchema(trimmed); err != nil {
		return "", err
	}
	return trimmed, nil
}

// ValidateParams validates one parameter JSON payload against the supported
// handler-schema subset.
func ValidateParams(schemaText string, paramsJSON json.RawMessage) error {
	schema, err := parseSchema(schemaText)
	if err != nil {
		return err
	}

	payload, err := parseParams(paramsJSON)
	if err != nil {
		return err
	}
	for _, key := range schema.Required {
		value, ok := payload[key]
		if !ok || value == nil {
			return gerror.Newf("任务处理器参数缺少必填字段 %s", key)
		}
	}
	for key, value := range payload {
		field, ok := schema.Properties[key]
		if !ok {
			continue
		}
		if err = validateFieldValue(key, field, value); err != nil {
			return err
		}
	}
	return nil
}

// parseSchema decodes and validates one supported handler parameter schema.
func parseSchema(schemaText string) (*handlerSchemaRoot, error) {
	decoder := json.NewDecoder(strings.NewReader(strings.TrimSpace(schemaText)))
	decoder.DisallowUnknownFields()
	decoder.UseNumber()

	var schema handlerSchemaRoot
	if err := decoder.Decode(&schema); err != nil {
		return nil, gerror.Wrap(err, "解析任务处理器参数 Schema 失败")
	}
	if err := ensureDecoderEOF(decoder); err != nil {
		return nil, gerror.Wrap(err, "任务处理器参数 Schema 只能包含一个 JSON 对象")
	}
	if strings.TrimSpace(schema.Type) != "object" {
		return nil, gerror.New("任务处理器参数 Schema 根节点必须声明 type=object")
	}
	if schema.Properties == nil {
		schema.Properties = map[string]handlerSchemaField{}
	}

	for key, field := range schema.Properties {
		if strings.TrimSpace(key) == "" {
			return nil, gerror.New("任务处理器参数 Schema 不允许存在空属性名")
		}
		if err := validateSchemaField(key, field); err != nil {
			return nil, err
		}
	}
	for _, required := range schema.Required {
		if _, ok := schema.Properties[required]; !ok {
			return nil, gerror.Newf("任务处理器参数 Schema 的 required 字段 %s 未在 properties 中声明", required)
		}
	}
	return &schema, nil
}

// validateSchemaField validates one supported schema field definition.
func validateSchemaField(key string, field handlerSchemaField) error {
	fieldType := strings.TrimSpace(field.Type)
	switch fieldType {
	case "string", "integer", "number", "boolean":
	default:
		return gerror.Newf("任务处理器参数 %s 的类型 %s 不在受支持范围内", key, field.Type)
	}

	format := strings.TrimSpace(field.Format)
	switch format {
	case "", "date", "date-time", "textarea":
	default:
		return gerror.Newf("任务处理器参数 %s 的 format=%s 不受支持", key, field.Format)
	}
	if format != "" && fieldType != "string" {
		return gerror.Newf("任务处理器参数 %s 只有 string 类型才允许声明 format", key)
	}
	if len(field.Enum) > 0 {
		for _, item := range field.Enum {
			if err := validateSchemaEnumValue(key, fieldType, item); err != nil {
				return err
			}
		}
	}
	return nil
}

// validateSchemaEnumValue validates one enum literal against the declared field type.
func validateSchemaEnumValue(key string, fieldType string, value any) error {
	if err := validateType(fieldType, value); err != nil {
		return gerror.Wrapf(err, "任务处理器参数 %s 的枚举值不合法", key)
	}
	return nil
}

// parseParams decodes one params payload into a dynamic object while preserving numeric precision.
func parseParams(paramsJSON json.RawMessage) (map[string]any, error) {
	trimmed := bytes.TrimSpace(paramsJSON)
	if len(trimmed) == 0 {
		return map[string]any{}, nil
	}

	decoder := json.NewDecoder(bytes.NewReader(trimmed))
	decoder.UseNumber()

	var payload map[string]any
	if err := decoder.Decode(&payload); err != nil {
		return nil, gerror.Wrap(err, "解析任务处理器参数失败")
	}
	if err := ensureDecoderEOF(decoder); err != nil {
		return nil, gerror.Wrap(err, "任务处理器参数只能包含一个 JSON 对象")
	}
	if payload == nil {
		payload = map[string]any{}
	}
	return payload, nil
}

// validateFieldValue validates one input field value against its schema definition.
func validateFieldValue(key string, field handlerSchemaField, value any) error {
	if err := validateType(strings.TrimSpace(field.Type), value); err != nil {
		return gerror.Wrapf(err, "任务处理器参数 %s 类型不匹配", key)
	}
	if len(field.Enum) > 0 && !enumContains(field.Enum, value) {
		return gerror.Newf("任务处理器参数 %s 不在允许的枚举值范围内", key)
	}

	format := strings.TrimSpace(field.Format)
	if format == "" {
		return nil
	}
	stringValue, _ := value.(string)
	switch format {
	case "textarea":
		return nil
	case "date":
		if _, err := time.Parse("2006-01-02", stringValue); err != nil {
			return gerror.Newf("任务处理器参数 %s 必须为 YYYY-MM-DD 日期格式", key)
		}
	case "date-time":
		if _, err := time.Parse(time.RFC3339, stringValue); err != nil {
			return gerror.Newf("任务处理器参数 %s 必须为 RFC3339 时间格式", key)
		}
	}
	return nil
}

// validateType validates one dynamic JSON value against the supported field type.
func validateType(fieldType string, value any) error {
	switch fieldType {
	case "string":
		if _, ok := value.(string); ok {
			return nil
		}
	case "boolean":
		if _, ok := value.(bool); ok {
			return nil
		}
	case "integer":
		if number, ok := value.(json.Number); ok {
			if _, err := number.Int64(); err == nil {
				return nil
			}
		}
	case "number":
		if number, ok := value.(json.Number); ok {
			if _, err := number.Float64(); err == nil {
				return nil
			}
		}
	}
	return gerror.Newf("不支持的值 %v", value)
}

// enumContains reports whether the input value matches one declared enum literal.
func enumContains(values []any, target any) bool {
	for _, value := range values {
		if enumValueEqual(value, target) {
			return true
		}
	}
	return false
}

// enumValueEqual compares two enum values while preserving json.Number semantics.
func enumValueEqual(left any, right any) bool {
	leftNumber, leftIsNumber := left.(json.Number)
	rightNumber, rightIsNumber := right.(json.Number)
	if leftIsNumber && rightIsNumber {
		return leftNumber.String() == rightNumber.String()
	}
	return reflect.DeepEqual(left, right)
}

// ensureDecoderEOF verifies that the decoder consumed exactly one JSON value.
func ensureDecoderEOF(decoder *json.Decoder) error {
	var trailing struct{}
	if err := decoder.Decode(&trailing); err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}
	return gerror.New("检测到多余 JSON 内容")
}
