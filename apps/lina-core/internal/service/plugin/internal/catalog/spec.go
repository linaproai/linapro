// This file provides hook and resource specification validation, normalizer
// utilities, and deep-clone helpers shared by the runtime artifact parser and
// integration service loader.

package catalog

import (
	"regexp"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/pluginhost"
)

// safePluginIdentifierPattern validates identifiers embedded into generated SQL
// or query fragments for hook and resource specs.
var safePluginIdentifierPattern = regexp.MustCompile(`^[A-Za-z0-9_]+$`)

// NormalizeResourceSpecType maps a raw string to the canonical ResourceSpecType constant.
func NormalizeResourceSpecType(value string) ResourceSpecType {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case ResourceSpecTypeTableList.String():
		return ResourceSpecTypeTableList
	default:
		return ResourceSpecType("")
	}
}

// NormalizeResourceFilterOperator maps a raw string to the canonical ResourceFilterOperator constant.
func NormalizeResourceFilterOperator(value string) ResourceFilterOperator {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case ResourceFilterOperatorEQ.String():
		return ResourceFilterOperatorEQ
	case ResourceFilterOperatorLike.String():
		return ResourceFilterOperatorLike
	case ResourceFilterOperatorGTEDate.String():
		return ResourceFilterOperatorGTEDate
	case ResourceFilterOperatorLTEDate.String():
		return ResourceFilterOperatorLTEDate
	default:
		return ResourceFilterOperator("")
	}
}

// NormalizeResourceOrderDirection maps a raw string to the canonical ResourceOrderDirection constant.
func NormalizeResourceOrderDirection(value string) ResourceOrderDirection {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case ResourceOrderDirectionASC.String():
		return ResourceOrderDirectionASC
	case ResourceOrderDirectionDESC.String():
		return ResourceOrderDirectionDESC
	default:
		return ResourceOrderDirection("")
	}
}

// NormalizeResourceOperation maps a raw string to the canonical ResourceOperation constant.
func NormalizeResourceOperation(value string) ResourceOperation {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case ResourceOperationQuery.String():
		return ResourceOperationQuery
	case ResourceOperationGet.String():
		return ResourceOperationGet
	case ResourceOperationCreate.String():
		return ResourceOperationCreate
	case ResourceOperationUpdate.String():
		return ResourceOperationUpdate
	case ResourceOperationDelete.String():
		return ResourceOperationDelete
	case ResourceOperationTransaction.String():
		return ResourceOperationTransaction
	default:
		return ResourceOperation("")
	}
}

// NormalizeResourceAccessMode maps a raw string to the canonical ResourceAccessMode constant.
func NormalizeResourceAccessMode(value string) ResourceAccessMode {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "", ResourceAccessModeRequest.String():
		return ResourceAccessModeRequest
	case ResourceAccessModeSystem.String():
		return ResourceAccessModeSystem
	case ResourceAccessModeBoth.String():
		return ResourceAccessModeBoth
	default:
		return ResourceAccessMode("")
	}
}

// ValidateHookSpec validates a plugin-declared hook handler specification.
func ValidateHookSpec(pluginID string, spec *HookSpec, filePath string) error {
	if spec == nil {
		return gerror.Newf("plugin hook cannot be nil: %s", filePath)
	}
	if spec.Event == "" {
		return gerror.Newf("plugin hook is missing event: %s", filePath)
	}
	if !pluginhost.IsHookExtensionPoint(spec.Event) {
		return gerror.Newf("plugin hook extension point is not published: %s", filePath)
	}
	if spec.Action == "" {
		spec.Action = pluginhost.HookActionInsert
	}
	if !pluginhost.IsSupportedHookAction(spec.Action) {
		return gerror.Newf("plugin hook action is not supported by the host: %s", filePath)
	}
	if spec.Mode == "" {
		spec.Mode = pluginhost.DefaultCallbackExecutionMode(spec.Event)
	}
	if !pluginhost.IsExtensionPointExecutionModeSupported(spec.Event, spec.Mode) {
		return gerror.Newf("plugin hook execution mode is not supported by the current extension point: %s", filePath)
	}
	if spec.TimeoutMs < 0 {
		return gerror.Newf("plugin hook timeoutMs cannot be less than 0: %s", filePath)
	}
	switch spec.Action {
	case pluginhost.HookActionInsert:
		if err := validatePluginIdentifier(spec.Table); err != nil {
			return gerror.Wrapf(err, "plugin %s hook table name is invalid: %s", pluginID, filePath)
		}
		if len(spec.Fields) == 0 {
			return gerror.Newf("plugin hook is missing fields mapping: %s", filePath)
		}
		for column := range spec.Fields {
			if err := validatePluginIdentifier(column); err != nil {
				return gerror.Wrapf(err, "plugin %s hook field is invalid: %s", pluginID, filePath)
			}
		}
	case pluginhost.HookActionSleep:
		if spec.SleepMs <= 0 {
			return gerror.Newf("plugin hook sleep action requires sleepMs > 0: %s", filePath)
		}
	case pluginhost.HookActionError:
		if strings.TrimSpace(spec.ErrorMessage) == "" {
			return gerror.Newf("plugin hook error action requires a non-empty errorMessage: %s", filePath)
		}
	}
	return nil
}

// ValidateResourceSpec validates a plugin-declared backend resource specification.
func ValidateResourceSpec(pluginID string, spec *ResourceSpec, filePath string) error {
	if spec == nil {
		return gerror.Newf("plugin resource cannot be nil: %s", filePath)
	}
	if spec.Key == "" {
		return gerror.Newf("plugin resource is missing key: %s", filePath)
	}
	if spec.Type == "" {
		spec.Type = ResourceSpecTypeTableList.String()
	}
	if NormalizeResourceSpecType(spec.Type) != ResourceSpecTypeTableList {
		return gerror.Newf("plugin resource type only supports table-list: %s", filePath)
	}
	if err := validatePluginIdentifier(spec.Table); err != nil {
		return gerror.Wrapf(err, "plugin %s resource table name is invalid: %s", pluginID, filePath)
	}
	if len(spec.Fields) == 0 {
		return gerror.Newf("plugin resource is missing fields definition: %s", filePath)
	}
	for _, field := range spec.Fields {
		if field == nil {
			return gerror.Newf("plugin resource field cannot be nil: %s", filePath)
		}
		if err := validatePluginIdentifier(field.Name); err != nil {
			return gerror.Wrapf(err, "plugin %s resource field name is invalid: %s", pluginID, filePath)
		}
		if err := validatePluginIdentifier(field.Column); err != nil {
			return gerror.Wrapf(err, "plugin %s resource column name is invalid: %s", pluginID, filePath)
		}
	}
	for _, filter := range spec.Filters {
		if filter == nil {
			return gerror.Newf("plugin resource filter cannot be nil: %s", filePath)
		}
		if filter.Param == "" {
			return gerror.Newf("plugin resource filter is missing param: %s", filePath)
		}
		if err := validatePluginIdentifier(filter.Column); err != nil {
			return gerror.Wrapf(err, "plugin %s resource filter column is invalid: %s", pluginID, filePath)
		}
		if NormalizeResourceFilterOperator(filter.Operator) == "" {
			return gerror.Newf("plugin resource filter operator is unsupported: %s", filePath)
		}
	}
	if err := validatePluginIdentifier(spec.OrderBy.Column); err != nil {
		return gerror.Wrapf(err, "plugin %s resource order column is invalid: %s", pluginID, filePath)
	}
	if spec.OrderBy.Direction == "" {
		spec.OrderBy.Direction = ResourceOrderDirectionASC.String()
	}
	if NormalizeResourceOrderDirection(spec.OrderBy.Direction) == "" {
		return gerror.Newf("plugin resource order direction only supports asc/desc: %s", filePath)
	}
	if spec.DataScope != nil {
		if spec.DataScope.UserColumn != "" {
			if err := validatePluginIdentifier(spec.DataScope.UserColumn); err != nil {
				return gerror.Wrapf(err, "plugin %s resource dataScope userColumn is invalid: %s", pluginID, filePath)
			}
		}
		if spec.DataScope.DeptColumn != "" {
			if err := validatePluginIdentifier(spec.DataScope.DeptColumn); err != nil {
				return gerror.Wrapf(err, "plugin %s resource dataScope deptColumn is invalid: %s", pluginID, filePath)
			}
		}
		if spec.DataScope.UserColumn == "" && spec.DataScope.DeptColumn == "" {
			return gerror.Newf("plugin resource dataScope must declare userColumn or deptColumn: %s", filePath)
		}
	}
	if len(spec.Operations) == 0 {
		spec.Operations = []string{ResourceOperationQuery.String()}
	}
	operationSeen := make(map[string]struct{}, len(spec.Operations))
	for _, operation := range spec.Operations {
		normalizedOperation := NormalizeResourceOperation(operation)
		if normalizedOperation == "" {
			return gerror.Newf("plugin resource operation is unsupported: %s", filePath)
		}
		operationSeen[normalizedOperation.String()] = struct{}{}
	}
	spec.Operations = normalizeEnumStringSliceForResourceSpec(spec.Operations)

	if spec.KeyField != "" {
		if err := validatePluginIdentifier(spec.KeyField); err != nil {
			return gerror.Wrapf(err, "plugin %s resource keyField is invalid: %s", pluginID, filePath)
		}
		if !resourceHasField(spec, spec.KeyField) {
			return gerror.Newf("plugin resource keyField is not declared in fields: %s", filePath)
		}
	}
	if _, needsKeyField := operationSeen[ResourceOperationGet.String()]; needsKeyField && spec.KeyField == "" {
		return gerror.Newf("plugin resource get operation requires keyField: %s", filePath)
	}
	if _, needsKeyField := operationSeen[ResourceOperationUpdate.String()]; needsKeyField && spec.KeyField == "" {
		return gerror.Newf("plugin resource update operation requires keyField: %s", filePath)
	}
	if _, needsKeyField := operationSeen[ResourceOperationDelete.String()]; needsKeyField && spec.KeyField == "" {
		return gerror.Newf("plugin resource delete operation requires keyField: %s", filePath)
	}

	if len(spec.WritableFields) > 0 {
		spec.WritableFields = normalizeFieldNameSliceForResourceSpec(spec.WritableFields)
		for _, writableField := range spec.WritableFields {
			if err := validatePluginIdentifier(writableField); err != nil {
				return gerror.Wrapf(err, "plugin %s resource writableField is invalid: %s", pluginID, filePath)
			}
			if !resourceHasField(spec, writableField) {
				return gerror.Newf("plugin resource writableField is not declared in fields: %s", filePath)
			}
		}
	}
	if _, needsWritableFields := operationSeen[ResourceOperationCreate.String()]; needsWritableFields && len(spec.WritableFields) == 0 {
		return gerror.Newf("plugin resource create operation requires writableFields: %s", filePath)
	}
	if _, needsWritableFields := operationSeen[ResourceOperationUpdate.String()]; needsWritableFields && len(spec.WritableFields) == 0 {
		return gerror.Newf("plugin resource update operation requires writableFields: %s", filePath)
	}

	if spec.Access == "" {
		spec.Access = ResourceAccessModeRequest.String()
	}
	if NormalizeResourceAccessMode(spec.Access) == "" {
		return gerror.Newf("plugin resource access only supports request/system/both: %s", filePath)
	}
	if spec.Permission != "" {
		spec.Permission = strings.TrimSpace(spec.Permission)
		parts := strings.Split(spec.Permission, ":")
		if len(parts) != 3 || strings.TrimSpace(parts[0]) != strings.TrimSpace(pluginID) {
			return gerror.Newf("plugin resource permission must use %s:{resource}:{action} format: %s", pluginID, filePath)
		}
		if strings.TrimSpace(parts[1]) == "" || strings.TrimSpace(parts[2]) == "" {
			return gerror.Newf("plugin resource permission resource and action cannot be empty: %s", filePath)
		}
	}
	return nil
}

// validatePluginIdentifier validates that a table or column name contains only safe characters.
func validatePluginIdentifier(value string) error {
	if value == "" {
		return gerror.New("plugin identifier cannot be empty")
	}
	if !safePluginIdentifierPattern.MatchString(value) {
		return gerror.Newf("plugin identifier is invalid: %s", value)
	}
	return nil
}

// CloneHookSpecs returns a deep copy of the given hook spec slice.
func CloneHookSpecs(items []*HookSpec) []*HookSpec {
	if len(items) == 0 {
		return []*HookSpec{}
	}
	cloned := make([]*HookSpec, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		next := *item
		if len(item.Fields) > 0 {
			next.Fields = make(map[string]string, len(item.Fields))
			for key, value := range item.Fields {
				next.Fields[key] = value
			}
		}
		cloned = append(cloned, &next)
	}
	return cloned
}

// CloneResourceSpecsToMap returns a deep copy of the resource spec slice keyed by resource Key.
func CloneResourceSpecsToMap(items []*ResourceSpec) map[string]*ResourceSpec {
	if len(items) == 0 {
		return map[string]*ResourceSpec{}
	}
	cloned := make(map[string]*ResourceSpec, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		next := CloneResourceSpec(item)
		cloned[next.Key] = next
	}
	return cloned
}

// CloneResourceSpec returns a deep copy of one resource spec.
func CloneResourceSpec(item *ResourceSpec) *ResourceSpec {
	if item == nil {
		return nil
	}
	next := *item
	if len(item.Fields) > 0 {
		next.Fields = make([]*ResourceField, 0, len(item.Fields))
		for _, field := range item.Fields {
			if field == nil {
				continue
			}
			fieldCopy := *field
			next.Fields = append(next.Fields, &fieldCopy)
		}
	}
	if len(item.Filters) > 0 {
		next.Filters = make([]*ResourceQuery, 0, len(item.Filters))
		for _, filter := range item.Filters {
			if filter == nil {
				continue
			}
			filterCopy := *filter
			next.Filters = append(next.Filters, &filterCopy)
		}
	}
	if item.DataScope != nil {
		dataScopeCopy := *item.DataScope
		next.DataScope = &dataScopeCopy
	}
	if len(item.Operations) > 0 {
		next.Operations = append([]string(nil), item.Operations...)
	}
	if len(item.WritableFields) > 0 {
		next.WritableFields = append([]string(nil), item.WritableFields...)
	}
	return &next
}

// resourceHasField reports whether the resource declares the given logical field name.
func resourceHasField(spec *ResourceSpec, fieldName string) bool {
	if spec == nil {
		return false
	}
	targetFieldName := strings.TrimSpace(fieldName)
	if targetFieldName == "" {
		return false
	}
	for _, field := range spec.Fields {
		if field != nil && field.Name == targetFieldName {
			return true
		}
	}
	return false
}

// normalizeEnumStringSliceForResourceSpec trims, lowercases, and de-duplicates
// enum-like string slices used by resource specs.
func normalizeEnumStringSliceForResourceSpec(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		normalized := strings.TrimSpace(strings.ToLower(item))
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	return result
}

// normalizeFieldNameSliceForResourceSpec trims and de-duplicates field names
// while preserving their original casing for output.
func normalizeFieldNameSliceForResourceSpec(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		lookupKey := strings.ToLower(trimmed)
		if _, ok := seen[lookupKey]; ok {
			continue
		}
		seen[lookupKey] = struct{}{}
		result = append(result, trimmed)
	}
	return result
}
