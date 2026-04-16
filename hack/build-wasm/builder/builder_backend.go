// This file loads backend hook and resource declarations, extracts dynamic
// route contracts from API DTOs, and validates the collected contracts.

package builder

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"lina-core/pkg/pluginbridge"
)

func collectHookSpecs(pluginDir string, pluginID string) ([]*hookSpec, error) {
	hookDir := filepath.Join(pluginDir, "backend", "hooks")
	entries, err := os.ReadDir(hookDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*hookSpec{}, nil
		}
		return nil, err
	}

	fileNames := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}
		fileNames = append(fileNames, entry.Name())
	}
	sortStrings(fileNames)

	items := make([]*hookSpec, 0, len(fileNames))
	for _, name := range fileNames {
		filePath := filepath.Join(hookDir, name)
		spec := &hookSpec{}
		if err = loadYAMLFile(filePath, spec); err != nil {
			return nil, err
		}
		if err = validateHookSpec(pluginID, spec, filePath); err != nil {
			return nil, err
		}
		items = append(items, spec)
	}
	return items, nil
}

func collectResourceSpecs(pluginDir string, pluginID string) ([]*resourceSpec, error) {
	resourceDir := filepath.Join(pluginDir, "backend", "resources")
	entries, err := os.ReadDir(resourceDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*resourceSpec{}, nil
		}
		return nil, err
	}

	fileNames := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}
		fileNames = append(fileNames, entry.Name())
	}
	sortStrings(fileNames)

	items := make([]*resourceSpec, 0, len(fileNames))
	for _, name := range fileNames {
		filePath := filepath.Join(resourceDir, name)
		spec := &resourceSpec{}
		if err = loadYAMLFile(filePath, spec); err != nil {
			return nil, err
		}
		if err = validateResourceSpec(pluginID, spec, filePath); err != nil {
			return nil, err
		}
		items = append(items, spec)
	}
	return items, nil
}

func collectRouteContracts(pluginDir string, pluginID string) ([]*pluginbridge.RouteContract, error) {
	apiDir := filepath.Join(pluginDir, "backend", "api")
	info, err := os.Stat(apiDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*pluginbridge.RouteContract{}, nil
		}
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("runtime backend api path is not a directory: %s", apiDir)
	}

	fset := token.NewFileSet()
	contracts := make([]*pluginbridge.RouteContract, 0)
	err = filepath.WalkDir(apiDir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || filepath.Ext(path) != ".go" {
			return nil
		}
		fileNode, parseErr := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if parseErr != nil {
			return fmt.Errorf("failed to parse api file %s: %w", path, parseErr)
		}
		items, extractErr := extractRouteContractsFromFile(fileNode)
		if extractErr != nil {
			return fmt.Errorf("failed to extract route contract from %s: %w", path, extractErr)
		}
		contracts = append(contracts, items...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	if err = pluginbridge.ValidateRouteContracts(pluginID, contracts); err != nil {
		return nil, err
	}
	return contracts, nil
}

func extractRouteContractsFromFile(fileNode *ast.File) ([]*pluginbridge.RouteContract, error) {
	items := make([]*pluginbridge.RouteContract, 0)
	for _, decl := range fileNode.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok || structType.Fields == nil {
				continue
			}
			for _, field := range structType.Fields.List {
				if field == nil || field.Tag == nil {
					continue
				}
				if len(field.Names) != 0 {
					continue
				}
				tagValue := strings.Trim(field.Tag.Value, "`")
				if strings.TrimSpace(tagValue) == "" {
					continue
				}
				metaValues := parseStructTagValues(tagValue)
				if metaValues["path"] == "" || metaValues["method"] == "" {
					continue
				}
				contract := &pluginbridge.RouteContract{
					Path:        metaValues["path"],
					Method:      metaValues["method"],
					Tags:        splitTagList(metaValues["tags"]),
					Summary:     metaValues["summary"],
					Description: metaValues["dc"],
					Access:      metaValues["access"],
					Permission:  metaValues["permission"],
					RequestType: strings.TrimSpace(typeSpec.Name.Name),
				}
				if metaValues["operLog"] != "" {
					contract.OperLog = metaValues["operLog"]
				}
				items = append(items, contract)
			}
		}
	}
	return items, nil
}

func parseStructTagValues(tagValue string) map[string]string {
	values := make(map[string]string)
	cursor := 0
	for cursor < len(tagValue) {
		for cursor < len(tagValue) && tagValue[cursor] == ' ' {
			cursor++
		}
		if cursor >= len(tagValue) {
			break
		}
		keyStart := cursor
		for cursor < len(tagValue) && tagValue[cursor] != ':' {
			cursor++
		}
		if cursor >= len(tagValue) || tagValue[cursor] != ':' {
			break
		}
		key := strings.TrimSpace(tagValue[keyStart:cursor])
		cursor++
		if cursor >= len(tagValue) || tagValue[cursor] != '"' {
			break
		}
		cursor++
		valueStart := cursor
		for cursor < len(tagValue) {
			if tagValue[cursor] == '"' && tagValue[cursor-1] != '\\' {
				break
			}
			cursor++
		}
		if cursor >= len(tagValue) {
			break
		}
		values[key] = tagValue[valueStart:cursor]
		cursor++
	}
	return values
}

func splitTagList(value string) []string {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return nil
	}
	items := strings.Split(normalized, ",")
	result := make([]string, 0, len(items))
	for _, item := range items {
		tag := strings.TrimSpace(item)
		if tag == "" {
			continue
		}
		result = append(result, tag)
	}
	return result
}

func validateHookSpec(pluginID string, spec *hookSpec, filePath string) error {
	if spec == nil {
		return fmt.Errorf("plugin hook cannot be nil: %s", filePath)
	}
	if strings.TrimSpace(string(spec.Event)) == "" {
		return fmt.Errorf("plugin hook missing event: %s", filePath)
	}
	if !isHookExtensionPoint(spec.Event) {
		return fmt.Errorf("plugin hook event is not published by host: %s", filePath)
	}
	if spec.Action == "" {
		spec.Action = hookActionInsert
	}
	if !isSupportedHookAction(spec.Action) {
		return fmt.Errorf("plugin hook action is not supported: %s", filePath)
	}
	if spec.Mode == "" {
		spec.Mode = defaultCallbackExecutionMode(spec.Event)
	}
	if !isExtensionPointExecutionModeSupported(spec.Event, spec.Mode) {
		return fmt.Errorf("plugin hook execution mode is not supported: %s", filePath)
	}
	if spec.TimeoutMs < 0 {
		return fmt.Errorf("plugin hook timeoutMs cannot be negative: %s", filePath)
	}

	switch spec.Action {
	case hookActionInsert:
		if err := validateIdentifier(spec.Table); err != nil {
			return fmt.Errorf("plugin %s hook table is invalid: %s: %w", pluginID, filePath, err)
		}
		if len(spec.Fields) == 0 {
			return fmt.Errorf("plugin hook missing fields: %s", filePath)
		}
		for column := range spec.Fields {
			if err := validateIdentifier(column); err != nil {
				return fmt.Errorf("plugin %s hook field is invalid: %s: %w", pluginID, filePath, err)
			}
		}
	case hookActionSleep:
		if spec.SleepMs <= 0 {
			return fmt.Errorf("plugin hook sleep action requires sleepMs > 0: %s", filePath)
		}
	case hookActionError:
		if strings.TrimSpace(spec.ErrorMessage) == "" {
			return fmt.Errorf("plugin hook error action requires non-empty errorMessage: %s", filePath)
		}
	}

	return nil
}

func validateResourceSpec(pluginID string, spec *resourceSpec, filePath string) error {
	if spec == nil {
		return fmt.Errorf("plugin resource cannot be nil: %s", filePath)
	}
	if strings.TrimSpace(spec.Key) == "" {
		return fmt.Errorf("plugin resource missing key: %s", filePath)
	}
	if spec.Type == "" {
		spec.Type = string(resourceSpecTypeTableList)
	}
	if normalizeResourceSpecType(spec.Type) != resourceSpecTypeTableList {
		return fmt.Errorf("plugin resource type only supports table-list: %s", filePath)
	}
	if err := validateIdentifier(spec.Table); err != nil {
		return fmt.Errorf("plugin %s resource table is invalid: %s: %w", pluginID, filePath, err)
	}
	if len(spec.Fields) == 0 {
		return fmt.Errorf("plugin resource missing fields: %s", filePath)
	}
	for _, field := range spec.Fields {
		if field == nil {
			return fmt.Errorf("plugin resource field cannot be nil: %s", filePath)
		}
		if err := validateIdentifier(field.Name); err != nil {
			return fmt.Errorf("plugin %s resource field name is invalid: %s: %w", pluginID, filePath, err)
		}
		if err := validateIdentifier(field.Column); err != nil {
			return fmt.Errorf("plugin %s resource column is invalid: %s: %w", pluginID, filePath, err)
		}
	}
	for _, filter := range spec.Filters {
		if filter == nil {
			return fmt.Errorf("plugin resource filter cannot be nil: %s", filePath)
		}
		if strings.TrimSpace(filter.Param) == "" {
			return fmt.Errorf("plugin resource filter missing param: %s", filePath)
		}
		if err := validateIdentifier(filter.Column); err != nil {
			return fmt.Errorf("plugin %s resource filter column is invalid: %s: %w", pluginID, filePath, err)
		}
		if normalizeResourceFilterOperator(filter.Operator) == "" {
			return fmt.Errorf("plugin resource filter operator is not supported: %s", filePath)
		}
	}
	if err := validateIdentifier(spec.OrderBy.Column); err != nil {
		return fmt.Errorf("plugin %s resource orderBy column is invalid: %s: %w", pluginID, filePath, err)
	}
	if spec.OrderBy.Direction == "" {
		spec.OrderBy.Direction = string(resourceOrderDirectionASC)
	}
	if normalizeResourceOrderDirection(spec.OrderBy.Direction) == "" {
		return fmt.Errorf("plugin resource order direction only supports asc/desc: %s", filePath)
	}
	if spec.DataScope != nil {
		if spec.DataScope.UserColumn != "" {
			if err := validateIdentifier(spec.DataScope.UserColumn); err != nil {
				return fmt.Errorf("plugin %s resource dataScope userColumn is invalid: %s: %w", pluginID, filePath, err)
			}
		}
		if spec.DataScope.DeptColumn != "" {
			if err := validateIdentifier(spec.DataScope.DeptColumn); err != nil {
				return fmt.Errorf("plugin %s resource dataScope deptColumn is invalid: %s: %w", pluginID, filePath, err)
			}
		}
		if spec.DataScope.UserColumn == "" && spec.DataScope.DeptColumn == "" {
			return fmt.Errorf("plugin resource dataScope requires userColumn or deptColumn: %s", filePath)
		}
	}
	if len(spec.Operations) == 0 {
		spec.Operations = []string{string(resourceOperationQuery)}
	}
	operationSeen := make(map[string]struct{}, len(spec.Operations))
	for _, operation := range spec.Operations {
		normalizedOperation := normalizeResourceOperation(operation)
		if normalizedOperation == "" {
			return fmt.Errorf("plugin resource operation is not supported: %s", filePath)
		}
		operationSeen[string(normalizedOperation)] = struct{}{}
	}
	spec.Operations = normalizeResourceEnumStringSlice(spec.Operations)

	if spec.KeyField != "" {
		if err := validateIdentifier(spec.KeyField); err != nil {
			return fmt.Errorf("plugin %s resource keyField is invalid: %s: %w", pluginID, filePath, err)
		}
		if !resourceSpecHasField(spec, spec.KeyField) {
			return fmt.Errorf("plugin resource keyField is not declared in fields: %s", filePath)
		}
	}
	if _, ok := operationSeen[string(resourceOperationGet)]; ok && strings.TrimSpace(spec.KeyField) == "" {
		return fmt.Errorf("plugin resource get operation requires keyField: %s", filePath)
	}
	if _, ok := operationSeen[string(resourceOperationUpdate)]; ok && strings.TrimSpace(spec.KeyField) == "" {
		return fmt.Errorf("plugin resource update operation requires keyField: %s", filePath)
	}
	if _, ok := operationSeen[string(resourceOperationDelete)]; ok && strings.TrimSpace(spec.KeyField) == "" {
		return fmt.Errorf("plugin resource delete operation requires keyField: %s", filePath)
	}

	if len(spec.WritableFields) > 0 {
		spec.WritableFields = normalizeResourceFieldNameSlice(spec.WritableFields)
		for _, writableField := range spec.WritableFields {
			if err := validateIdentifier(writableField); err != nil {
				return fmt.Errorf("plugin %s resource writableField is invalid: %s: %w", pluginID, filePath, err)
			}
			if !resourceSpecHasField(spec, writableField) {
				return fmt.Errorf("plugin resource writableField is not declared in fields: %s", filePath)
			}
		}
	}
	if _, ok := operationSeen[string(resourceOperationCreate)]; ok && len(spec.WritableFields) == 0 {
		return fmt.Errorf("plugin resource create operation requires writableFields: %s", filePath)
	}
	if _, ok := operationSeen[string(resourceOperationUpdate)]; ok && len(spec.WritableFields) == 0 {
		return fmt.Errorf("plugin resource update operation requires writableFields: %s", filePath)
	}

	if spec.Access == "" {
		spec.Access = string(resourceAccessModeRequest)
	}
	if normalizeResourceAccessMode(spec.Access) == "" {
		return fmt.Errorf("plugin resource access is not supported: %s", filePath)
	}
	spec.Access = strings.ToLower(strings.TrimSpace(spec.Access))
	return nil
}

func validateIdentifier(value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("identifier cannot be empty")
	}
	if !safeIdentifierPattern.MatchString(value) {
		return fmt.Errorf("identifier is invalid: %s", value)
	}
	return nil
}

func defaultCallbackExecutionMode(point hookExtensionPoint) callbackExecutionMode {
	return publishedHookPoints[point]
}

func isHookExtensionPoint(point hookExtensionPoint) bool {
	_, ok := publishedHookPoints[point]
	return ok
}

func isSupportedHookAction(action hookAction) bool {
	switch action {
	case hookActionInsert, hookActionSleep, hookActionError:
		return true
	default:
		return false
	}
}

func isExtensionPointExecutionModeSupported(point hookExtensionPoint, mode callbackExecutionMode) bool {
	modes, ok := supportedHookModes[point]
	if !ok {
		return false
	}
	_, ok = modes[mode]
	return ok
}

func normalizeResourceSpecType(value string) resourceSpecType {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(resourceSpecTypeTableList):
		return resourceSpecTypeTableList
	default:
		return ""
	}
}

func normalizeResourceFilterOperator(value string) resourceFilterOperator {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(resourceFilterOperatorEQ):
		return resourceFilterOperatorEQ
	case string(resourceFilterOperatorLike):
		return resourceFilterOperatorLike
	case string(resourceFilterOperatorGTEDate):
		return resourceFilterOperatorGTEDate
	case string(resourceFilterOperatorLTEDate):
		return resourceFilterOperatorLTEDate
	default:
		return ""
	}
}

func normalizeResourceOrderDirection(value string) resourceOrderDirection {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(resourceOrderDirectionASC):
		return resourceOrderDirectionASC
	case string(resourceOrderDirectionDESC):
		return resourceOrderDirectionDESC
	default:
		return ""
	}
}

func normalizeResourceOperation(value string) resourceOperation {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(resourceOperationQuery):
		return resourceOperationQuery
	case string(resourceOperationGet):
		return resourceOperationGet
	case string(resourceOperationCreate):
		return resourceOperationCreate
	case string(resourceOperationUpdate):
		return resourceOperationUpdate
	case string(resourceOperationDelete):
		return resourceOperationDelete
	case string(resourceOperationTransaction):
		return resourceOperationTransaction
	default:
		return ""
	}
}

func normalizeResourceAccessMode(value string) resourceAccessMode {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", string(resourceAccessModeRequest):
		return resourceAccessModeRequest
	case string(resourceAccessModeSystem):
		return resourceAccessModeSystem
	case string(resourceAccessModeBoth):
		return resourceAccessModeBoth
	default:
		return ""
	}
}

func resourceSpecHasField(spec *resourceSpec, fieldName string) bool {
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

func normalizeResourceEnumStringSlice(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		normalized := strings.ToLower(strings.TrimSpace(item))
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	sort.Strings(result)
	return result
}

func normalizeResourceFieldNameSlice(items []string) []string {
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
	sort.Strings(result)
	return result
}

func sortStrings(items []string) {
	if len(items) <= 1 {
		return
	}
	sort.Strings(items)
}
