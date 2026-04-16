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
		return gerror.Newf("插件Hook不能为空: %s", filePath)
	}
	if spec.Event == "" {
		return gerror.Newf("插件Hook缺少event: %s", filePath)
	}
	if !pluginhost.IsHookExtensionPoint(spec.Event) {
		return gerror.Newf("插件Hook插槽未发布: %s", filePath)
	}
	if spec.Action == "" {
		spec.Action = pluginhost.HookActionInsert
	}
	if !pluginhost.IsSupportedHookAction(spec.Action) {
		return gerror.Newf("插件Hook动作不受宿主支持: %s", filePath)
	}
	if spec.Mode == "" {
		spec.Mode = pluginhost.DefaultCallbackExecutionMode(spec.Event)
	}
	if !pluginhost.IsExtensionPointExecutionModeSupported(spec.Event, spec.Mode) {
		return gerror.Newf("插件Hook执行模式不受当前插槽支持: %s", filePath)
	}
	if spec.TimeoutMs < 0 {
		return gerror.Newf("插件Hook timeoutMs 不能小于0: %s", filePath)
	}
	switch spec.Action {
	case pluginhost.HookActionInsert:
		if err := validatePluginIdentifier(spec.Table); err != nil {
			return gerror.Wrapf(err, "插件%s的Hook表名非法: %s", pluginID, filePath)
		}
		if len(spec.Fields) == 0 {
			return gerror.Newf("插件Hook缺少fields映射: %s", filePath)
		}
		for column := range spec.Fields {
			if err := validatePluginIdentifier(column); err != nil {
				return gerror.Wrapf(err, "插件%s的Hook字段非法: %s", pluginID, filePath)
			}
		}
	case pluginhost.HookActionSleep:
		if spec.SleepMs <= 0 {
			return gerror.Newf("插件Hook sleep 动作要求 sleepMs > 0: %s", filePath)
		}
	case pluginhost.HookActionError:
		if strings.TrimSpace(spec.ErrorMessage) == "" {
			return gerror.Newf("插件Hook error 动作要求 errorMessage 非空: %s", filePath)
		}
	}
	return nil
}

// ValidateResourceSpec validates a plugin-declared backend resource specification.
func ValidateResourceSpec(pluginID string, spec *ResourceSpec, filePath string) error {
	if spec == nil {
		return gerror.Newf("插件资源不能为空: %s", filePath)
	}
	if spec.Key == "" {
		return gerror.Newf("插件资源缺少key: %s", filePath)
	}
	if spec.Type == "" {
		spec.Type = ResourceSpecTypeTableList.String()
	}
	if NormalizeResourceSpecType(spec.Type) != ResourceSpecTypeTableList {
		return gerror.Newf("插件资源类型仅支持table-list: %s", filePath)
	}
	if err := validatePluginIdentifier(spec.Table); err != nil {
		return gerror.Wrapf(err, "插件%s资源表名非法: %s", pluginID, filePath)
	}
	if len(spec.Fields) == 0 {
		return gerror.Newf("插件资源缺少fields定义: %s", filePath)
	}
	for _, field := range spec.Fields {
		if field == nil {
			return gerror.Newf("插件资源字段不能为空: %s", filePath)
		}
		if err := validatePluginIdentifier(field.Name); err != nil {
			return gerror.Wrapf(err, "插件%s资源字段名称非法: %s", pluginID, filePath)
		}
		if err := validatePluginIdentifier(field.Column); err != nil {
			return gerror.Wrapf(err, "插件%s资源列名非法: %s", pluginID, filePath)
		}
	}
	for _, filter := range spec.Filters {
		if filter == nil {
			return gerror.Newf("插件资源过滤器不能为空: %s", filePath)
		}
		if filter.Param == "" {
			return gerror.Newf("插件资源过滤器缺少param: %s", filePath)
		}
		if err := validatePluginIdentifier(filter.Column); err != nil {
			return gerror.Wrapf(err, "插件%s资源过滤列非法: %s", pluginID, filePath)
		}
		if NormalizeResourceFilterOperator(filter.Operator) == "" {
			return gerror.Newf("插件资源过滤操作符不支持: %s", filePath)
		}
	}
	if err := validatePluginIdentifier(spec.OrderBy.Column); err != nil {
		return gerror.Wrapf(err, "插件%s资源排序列非法: %s", pluginID, filePath)
	}
	if spec.OrderBy.Direction == "" {
		spec.OrderBy.Direction = ResourceOrderDirectionASC.String()
	}
	if NormalizeResourceOrderDirection(spec.OrderBy.Direction) == "" {
		return gerror.Newf("插件资源排序方向仅支持 asc/desc: %s", filePath)
	}
	if spec.DataScope != nil {
		if spec.DataScope.UserColumn != "" {
			if err := validatePluginIdentifier(spec.DataScope.UserColumn); err != nil {
				return gerror.Wrapf(err, "插件%s资源数据权限 userColumn 非法: %s", pluginID, filePath)
			}
		}
		if spec.DataScope.DeptColumn != "" {
			if err := validatePluginIdentifier(spec.DataScope.DeptColumn); err != nil {
				return gerror.Wrapf(err, "插件%s资源数据权限 deptColumn 非法: %s", pluginID, filePath)
			}
		}
		if spec.DataScope.UserColumn == "" && spec.DataScope.DeptColumn == "" {
			return gerror.Newf("插件资源 dataScope 至少需要声明 userColumn 或 deptColumn: %s", filePath)
		}
	}
	if len(spec.Operations) == 0 {
		spec.Operations = []string{ResourceOperationQuery.String()}
	}
	operationSeen := make(map[string]struct{}, len(spec.Operations))
	for _, operation := range spec.Operations {
		normalizedOperation := NormalizeResourceOperation(operation)
		if normalizedOperation == "" {
			return gerror.Newf("插件资源操作不受支持: %s", filePath)
		}
		operationSeen[normalizedOperation.String()] = struct{}{}
	}
	spec.Operations = normalizeEnumStringSliceForResourceSpec(spec.Operations)

	if spec.KeyField != "" {
		if err := validatePluginIdentifier(spec.KeyField); err != nil {
			return gerror.Wrapf(err, "插件%s资源 keyField 非法: %s", pluginID, filePath)
		}
		if !resourceHasField(spec, spec.KeyField) {
			return gerror.Newf("插件资源 keyField 未出现在 fields 中: %s", filePath)
		}
	}
	if _, needsKeyField := operationSeen[ResourceOperationGet.String()]; needsKeyField && spec.KeyField == "" {
		return gerror.Newf("插件资源 get 操作要求声明 keyField: %s", filePath)
	}
	if _, needsKeyField := operationSeen[ResourceOperationUpdate.String()]; needsKeyField && spec.KeyField == "" {
		return gerror.Newf("插件资源 update 操作要求声明 keyField: %s", filePath)
	}
	if _, needsKeyField := operationSeen[ResourceOperationDelete.String()]; needsKeyField && spec.KeyField == "" {
		return gerror.Newf("插件资源 delete 操作要求声明 keyField: %s", filePath)
	}

	if len(spec.WritableFields) > 0 {
		spec.WritableFields = normalizeFieldNameSliceForResourceSpec(spec.WritableFields)
		for _, writableField := range spec.WritableFields {
			if err := validatePluginIdentifier(writableField); err != nil {
				return gerror.Wrapf(err, "插件%s资源 writableField 非法: %s", pluginID, filePath)
			}
			if !resourceHasField(spec, writableField) {
				return gerror.Newf("插件资源 writableField 未出现在 fields 中: %s", filePath)
			}
		}
	}
	if _, needsWritableFields := operationSeen[ResourceOperationCreate.String()]; needsWritableFields && len(spec.WritableFields) == 0 {
		return gerror.Newf("插件资源 create 操作要求声明 writableFields: %s", filePath)
	}
	if _, needsWritableFields := operationSeen[ResourceOperationUpdate.String()]; needsWritableFields && len(spec.WritableFields) == 0 {
		return gerror.Newf("插件资源 update 操作要求声明 writableFields: %s", filePath)
	}

	if spec.Access == "" {
		spec.Access = ResourceAccessModeRequest.String()
	}
	if NormalizeResourceAccessMode(spec.Access) == "" {
		return gerror.Newf("插件资源 access 仅支持 request/system/both: %s", filePath)
	}
	return nil
}

// validatePluginIdentifier validates that a table or column name contains only safe characters.
func validatePluginIdentifier(value string) error {
	if value == "" {
		return gerror.New("插件标识不能为空")
	}
	if !safePluginIdentifierPattern.MatchString(value) {
		return gerror.Newf("插件标识非法: %s", value)
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
