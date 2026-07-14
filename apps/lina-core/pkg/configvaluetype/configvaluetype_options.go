// This file validates and encodes sys_config options JSON and value payloads
// against a value type. Options text accepts either a JSON array or a simple
// one-option-per-line format for admin-friendly editing.

package configvaluetype

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
)

// Option is one selectable label/value pair for select, radio, or multi_select.
type Option struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// MultiSelectSeparator joins multi_select values in the persisted string field.
const MultiSelectSeparator = ";"

// ParseOptions decodes options text into a slice. Empty input yields nil.
// Accepted formats:
//  1. JSON array: [{"label":"Left","value":"panel-left"}]
//  2. Simple lines (one option per line):
//     - "label=value" or "label|value"
//     - plain "value" (label defaults to value)
func ParseOptions(raw string) ([]Option, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, nil
	}
	if strings.HasPrefix(trimmed, "[") {
		return parseOptionsJSON(trimmed)
	}
	return parseOptionsSimple(trimmed)
}

// parseOptionsJSON decodes a JSON array of options.
func parseOptionsJSON(raw string) ([]Option, error) {
	var options []Option
	if err := json.Unmarshal([]byte(raw), &options); err != nil {
		return nil, gerror.Wrap(err, "invalid config options JSON")
	}
	return normalizeOptions(options)
}

// parseOptionsSimple decodes one-option-per-line text.
func parseOptionsSimple(raw string) ([]Option, error) {
	lines := strings.Split(raw, "\n")
	options := make([]Option, 0, len(lines))
	for lineNo, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		label, value := splitSimpleOptionLine(line)
		if value == "" {
			return nil, gerror.Newf("config options line %d value is required", lineNo+1)
		}
		if label == "" {
			label = value
		}
		options = append(options, Option{Label: label, Value: value})
	}
	return normalizeOptions(options)
}

// splitSimpleOptionLine splits "label=value" / "label|value" / bare value.
// The first '=' or '|' is the separator; remaining text belongs to value.
func splitSimpleOptionLine(line string) (label string, value string) {
	var (
		eq   = strings.IndexByte(line, '=')
		pipe = strings.IndexByte(line, '|')
		sep  = -1
	)
	switch {
	case eq >= 0 && pipe >= 0:
		if eq < pipe {
			sep = eq
		} else {
			sep = pipe
		}
	case eq >= 0:
		sep = eq
	case pipe >= 0:
		sep = pipe
	}
	if sep < 0 {
		return "", strings.TrimSpace(line)
	}
	return strings.TrimSpace(line[:sep]), strings.TrimSpace(line[sep+1:])
}

// normalizeOptions trims and fills empty labels from values.
func normalizeOptions(options []Option) ([]Option, error) {
	if len(options) == 0 {
		return nil, nil
	}
	normalized := make([]Option, 0, len(options))
	for i, option := range options {
		value := strings.TrimSpace(option.Value)
		if value == "" {
			return nil, gerror.Newf("config options[%d].value is required", i)
		}
		label := strings.TrimSpace(option.Label)
		if label == "" {
			label = value
		}
		normalized = append(normalized, Option{Label: label, Value: value})
	}
	return normalized, nil
}

// EncodeOptions serializes options to compact JSON for durable storage.
// An empty slice encodes as an empty string for storage convenience.
func EncodeOptions(options []Option) (string, error) {
	normalized, err := normalizeOptions(options)
	if err != nil {
		return "", err
	}
	if len(normalized) == 0 {
		return "", nil
	}
	payload, err := json.Marshal(normalized)
	if err != nil {
		return "", gerror.Wrap(err, "encode config options JSON")
	}
	return string(payload), nil
}

// FormatOptionsSimple renders options as one "label=value" line per option.
// When label equals value, only the value is written.
func FormatOptionsSimple(options []Option) string {
	if len(options) == 0 {
		return ""
	}
	lines := make([]string, 0, len(options))
	for _, option := range options {
		label := strings.TrimSpace(option.Label)
		value := strings.TrimSpace(option.Value)
		if value == "" {
			continue
		}
		if label == "" || label == value {
			lines = append(lines, value)
			continue
		}
		lines = append(lines, label+"="+value)
	}
	return strings.Join(lines, "\n")
}

// ResolveCode normalizes a raw type string: empty becomes Text; unsupported
// non-empty values return an error.
func ResolveCode(raw string) (Code, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return Text, nil
	}
	code := Code(trimmed)
	if !code.IsSupported() {
		return "", gerror.Newf("unsupported config value type %q", raw)
	}
	return code, nil
}

// ValidateTypedValue checks that value matches valueType and optional options.
func ValidateTypedValue(valueType Code, optionsRaw string, value string) error {
	code, err := ResolveCode(string(valueType))
	if err != nil {
		return err
	}

	options, err := ParseOptions(optionsRaw)
	if err != nil {
		return err
	}
	if code.RequiresOptions() && len(options) == 0 {
		return gerror.Newf("config value type %q requires non-empty options", code)
	}

	switch code {
	case Boolean:
		if value != "true" && value != "false" {
			return gerror.New("boolean config value must be true or false")
		}
	case Number:
		if strings.TrimSpace(value) == "" {
			return gerror.New("number config value is required")
		}
		if _, parseErr := strconv.ParseFloat(strings.TrimSpace(value), 64); parseErr != nil {
			return gerror.New("number config value must be a decimal number")
		}
	case Select, Radio:
		if !optionValueAllowed(options, value) {
			return gerror.Newf("config value %q is not in the allowed options", value)
		}
	case MultiSelect:
		normalized := NormalizeMultiSelectValue(value)
		if normalized == "" {
			return nil
		}
		for _, part := range strings.Split(normalized, MultiSelectSeparator) {
			if !optionValueAllowed(options, part) {
				return gerror.Newf("config value %q is not in the allowed options", part)
			}
		}
	case Text, Textarea, Richtext:
		// No additional shape checks.
	}
	return nil
}

// optionValueAllowed reports whether value equals one option value exactly.
func optionValueAllowed(options []Option, value string) bool {
	for _, option := range options {
		if option.Value == value {
			return true
		}
	}
	return false
}

// NormalizeMultiSelectValue trims parts and re-joins with the multi-select separator.
func NormalizeMultiSelectValue(value string) string {
	var (
		parts      = strings.Split(value, MultiSelectSeparator)
		normalized = make([]string, 0, len(parts))
		seen       = make(map[string]struct{}, len(parts))
	)
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if _, ok := seen[part]; ok {
			continue
		}
		seen[part] = struct{}{}
		normalized = append(normalized, part)
	}
	return strings.Join(normalized, MultiSelectSeparator)
}
