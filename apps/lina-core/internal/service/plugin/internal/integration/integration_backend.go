// This file loads plugin backend declarations and dispatches generic plugin
// resource queries.

package integration

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	pluginv1 "lina-core/api/plugin/v1"
	"lina-core/internal/model"
	"lina-core/internal/service/datascope"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/plugintypes"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/statusflag"
)

// pluginHookEventFieldExprPrefix is the prefix for event-field expressions in hook specs.
const pluginHookEventFieldExprPrefix = "event."

// ResourceListInput defines input for querying a plugin-owned backend resource.
type ResourceListInput struct {
	// PluginID is the plugin identifier.
	PluginID string
	// ResourceID is the plugin-declared resource key.
	ResourceID string
	// Filters contains query-string filter values.
	Filters map[string]string
	// PageNum is the requested page number.
	PageNum int
	// PageSize is the requested page size.
	PageSize int
}

// ResourceListOutput defines output for querying a plugin-owned backend resource.
type ResourceListOutput struct {
	// List contains the queried resource rows.
	List []map[string]interface{}
	// Total is the total row count.
	Total int
}

// pluginResourceDataScopeMode classifies host role scopes for generic plugin resource filtering.
type pluginResourceDataScopeMode int

// Internal plugin resource data-scope filter modes derived from sys_role.data_scope.
const (
	// pluginResourceDataScopeDeny denies access to governed plugin resource rows.
	pluginResourceDataScopeDeny pluginResourceDataScopeMode = iota
	// pluginResourceDataScopeAll grants all rows visible in the current tenant boundary.
	pluginResourceDataScopeAll
	// pluginResourceDataScopeDept restricts rows by the resource department-owner column.
	pluginResourceDataScopeDept
	// pluginResourceDataScopeSelf restricts rows by the resource user-owner column.
	pluginResourceDataScopeSelf
)

// ResolveResourcePermission resolves the plugin-scoped permission for one
// plugin-owned backend resource exposed by the generic resource endpoint.
func (s *serviceImpl) ResolveResourcePermission(
	ctx context.Context,
	pluginID string,
	resourceID string,
) (string, error) {
	manifest, err := s.resolveActiveOrDesiredManifest(ctx, pluginID)
	if err != nil {
		return "", err
	}
	resource, ok := manifest.BackendResources[resourceID]
	if !ok {
		return "", gerror.New("plugin resource does not exist")
	}
	if permission := strings.TrimSpace(resource.Permission); permission != "" {
		return permission, nil
	}
	return buildDefaultResourcePermission(pluginID, resourceID), nil
}

// resolveActiveOrDesiredManifest loads the active release manifest for installed
// dynamic plugins and falls back to discovery for source or inactive plugins.
func (s *serviceImpl) resolveActiveOrDesiredManifest(ctx context.Context, pluginID string) (*catalog.Manifest, error) {
	registry, err := s.storeSvc.GetRegistry(ctx, pluginID)
	if err != nil {
		return nil, err
	}
	if registry != nil &&
		plugintypes.NormalizeType(registry.Type) == pluginv1.PluginTypeDynamic &&
		registry.Installed == statusflag.Installed.Int() &&
		registry.ReleaseId > 0 {
		release, releaseErr := s.storeSvc.GetRegistryRelease(ctx, registry)
		if releaseErr != nil || release == nil {
			return nil, releaseErr
		}
		return s.storeSvc.LoadReleaseManifest(ctx, release)
	}
	return s.catalogSvc.GetDesiredManifest(pluginID)
}

// LoadPluginBackendConfig loads plugin-owned hook and resource declarations into the manifest.
// It implements catalog.BackendConfigLoader.
func (s *serviceImpl) LoadPluginBackendConfig(manifest *catalog.Manifest) error {
	return catalog.LoadPluginBackendConfig(manifest)
}

// ListResourceRecords queries plugin-owned backend resource rows using the
// generic plugin resource contract.
func (s *serviceImpl) ListResourceRecords(ctx context.Context, in ResourceListInput) (*ResourceListOutput, error) {
	manifest, err := s.resolveActiveOrDesiredManifest(ctx, in.PluginID)
	if err != nil {
		return nil, err
	}
	if !s.CanExposeBusinessEntries(ctx, in.PluginID) {
		return nil, gerror.New("plugin is not enabled")
	}

	resource, ok := manifest.BackendResources[in.ResourceID]
	if !ok {
		return nil, gerror.New("plugin resource does not exist")
	}
	if in.PageNum <= 0 {
		in.PageNum = 1
	}
	if in.PageSize <= 0 {
		in.PageSize = 10
	}
	if in.PageSize > 100 {
		in.PageSize = 100
	}

	m := g.DB().Model(resource.Table).Safe().Ctx(ctx)
	for _, filter := range resource.Filters {
		value := strings.TrimSpace(in.Filters[filter.Param])
		if value == "" {
			continue
		}
		switch plugintypes.NormalizeResourceFilterOperator(filter.Operator) {
		case plugintypes.ResourceFilterOperatorEQ:
			m = m.Where(filter.Column, value)
		case plugintypes.ResourceFilterOperatorLike:
			m = m.WhereLike(filter.Column, "%"+value+"%")
		case plugintypes.ResourceFilterOperatorGTEDate:
			m = m.WhereGTE(filter.Column, value+" 00:00:00")
		case plugintypes.ResourceFilterOperatorLTEDate:
			m = m.WhereLTE(filter.Column, value+" 23:59:59")
		default:
			return nil, gerror.Newf("plugin resource filter operator is not supported: %s", filter.Operator)
		}
	}
	m, err = s.applyPluginResourceDataScope(ctx, m, resource)
	if err != nil {
		return nil, err
	}

	total, err := m.Count()
	if err != nil {
		return nil, err
	}

	fields := make([]string, 0, len(resource.Fields))
	for _, field := range resource.Fields {
		fields = append(fields, fmt.Sprintf("%s AS %s", field.Column, quotePluginResourceAlias(field.Name)))
	}
	fieldArgs := make([]interface{}, 0, len(fields))
	for _, field := range fields {
		fieldArgs = append(fieldArgs, field)
	}

	queryModel := m.Fields(fieldArgs...).Page(in.PageNum, in.PageSize)
	if plugintypes.NormalizeResourceOrderDirection(resource.OrderBy.Direction) == plugintypes.ResourceOrderDirectionDESC {
		queryModel = queryModel.OrderDesc(resource.OrderBy.Column)
	} else {
		queryModel = queryModel.OrderAsc(resource.OrderBy.Column)
	}
	records, err := queryModel.All()
	if err != nil {
		return nil, err
	}
	items := make([]map[string]interface{}, 0, len(records))
	for _, record := range records {
		recordMap := record.Map()
		row := make(map[string]interface{}, len(resource.Fields))
		for _, field := range resource.Fields {
			row[field.Name] = normalizePluginResourceValue(resolvePluginResourceRecordValue(recordMap, field))
		}
		items = append(items, row)
	}
	return &ResourceListOutput{List: items, Total: total}, nil
}

// applyPluginResourceDataScope injects host role data-scope constraints into one plugin resource query.
func (s *serviceImpl) applyPluginResourceDataScope(
	ctx context.Context,
	model *gdb.Model,
	resource *catalog.ResourceSpec,
) (*gdb.Model, error) {
	if model == nil || resource == nil || resource.DataScope == nil {
		return model, nil
	}

	currentUserID := s.getCurrentPluginResourceUserID(ctx)
	if currentUserID <= 0 {
		return model.Where("1 = 0"), nil
	}

	bizUser := s.currentPluginResourceBizContext(ctx)
	if bizUser != nil {
		if bizUser.DataScopeUnsupported {
			return nil, bizerr.NewCode(
				datascope.CodeDataScopeUnsupported,
				bizerr.P("scope", bizUser.UnsupportedDataScope),
			)
		}
	}

	switch resolvePluginResourceDataScopeMode(currentPluginResourceDataScope(bizUser)) {
	case pluginResourceDataScopeAll:
		return model, nil
	case pluginResourceDataScopeDept:
		if resource.DataScope.DeptColumn == "" {
			return model.Where("1 = 0"), nil
		}
		deptIDs, deptErr := s.getCurrentPluginResourceDeptIDs(ctx, currentUserID)
		if deptErr != nil {
			return nil, deptErr
		}
		if len(deptIDs) == 0 {
			return model.Where("1 = 0"), nil
		}
		return model.WhereIn(resource.DataScope.DeptColumn, deptIDs), nil
	case pluginResourceDataScopeSelf:
		if resource.DataScope.UserColumn == "" {
			return model.Where("1 = 0"), nil
		}
		return model.Where(resource.DataScope.UserColumn, currentUserID), nil
	default:
		return model.Where("1 = 0"), nil
	}
}

// resolvePluginResourceDataScopeMode maps host role data-scope values to plugin resource filter modes.
func resolvePluginResourceDataScopeMode(scope int) pluginResourceDataScopeMode {
	switch datascope.Scope(scope) {
	case datascope.ScopeAll, datascope.ScopeTenant:
		return pluginResourceDataScopeAll
	case datascope.ScopeDept:
		return pluginResourceDataScopeDept
	case datascope.ScopeSelf:
		return pluginResourceDataScopeSelf
	default:
		return pluginResourceDataScopeDeny
	}
}

// getCurrentPluginResourceUserID returns the current request user ID from the business context.
func (s *serviceImpl) getCurrentPluginResourceUserID(ctx context.Context) int {
	bizUser := s.currentPluginResourceBizContext(ctx)
	if bizUser == nil {
		return 0
	}
	return bizUser.UserId
}

// getCurrentPluginResourceDeptIDs returns the deduplicated department IDs for the given user.
func (s *serviceImpl) getCurrentPluginResourceDeptIDs(ctx context.Context, userID int) ([]int, error) {
	if s == nil || s.orgSvc == nil || s.orgSvc.Assignment() == nil {
		return []int{}, nil
	}
	return s.orgSvc.Assignment().GetUserDeptIDs(ctx, userID)
}

// currentPluginResourceBizContext returns the current request business context.
func (s *serviceImpl) currentPluginResourceBizContext(ctx context.Context) *model.Context {
	if s == nil || s.bizCtxSvc == nil {
		return nil
	}
	return s.bizCtxSvc.Get(ctx)
}

// currentPluginResourceDataScope returns the data-scope value from the current request context.
func currentPluginResourceDataScope(bizUser *model.Context) int {
	if bizUser == nil {
		return 0
	}
	return bizUser.DataScope
}

// quotePluginResourceAlias preserves logical camelCase aliases across database
// engines that fold unquoted select aliases.
func quotePluginResourceAlias(alias string) string {
	return `"` + strings.ReplaceAll(alias, `"`, `""`) + `"`
}

// resolvePluginResourceRecordValue reads a projected row by logical alias first
// and by physical column name as a fallback.
func resolvePluginResourceRecordValue(recordMap map[string]interface{}, field *catalog.ResourceField) interface{} {
	if recordMap == nil || field == nil {
		return nil
	}
	if value, ok := recordMap[field.Name]; ok {
		return value
	}
	if value, ok := recordMap[field.Column]; ok {
		return value
	}
	return nil
}

// normalizePluginResourceValue converts time values to JSON-safe strings.
func normalizePluginResourceValue(value interface{}) interface{} {
	switch typedValue := value.(type) {
	case *time.Time:
		if typedValue == nil {
			return ""
		}
		return typedValue.Format(time.RFC3339Nano)
	case time.Time:
		return typedValue.Format(time.RFC3339Nano)
	case interface{ String() string }:
		return typedValue.String()
	default:
		return value
	}
}

// buildDefaultResourcePermission derives the fallback permission when a plugin
// resource does not declare one explicitly.
func buildDefaultResourcePermission(pluginID string, resourceID string) string {
	return strings.TrimSpace(pluginID) + ":" + strings.TrimSpace(resourceID) + ":list"
}

// executePluginInsertHook executes a generic insert hook declared by a source plugin.
func (s *serviceImpl) executePluginInsertHook(ctx context.Context, pluginID string, hook *catalog.HookSpec, payload map[string]interface{}) error {
	columns := make([]string, 0, len(hook.Fields))
	for column := range hook.Fields {
		columns = append(columns, column)
	}
	// Sort for deterministic SQL generation.
	sortStrings(columns)

	values := make([]interface{}, 0, len(columns))
	placeholders := make([]string, 0, len(columns))
	for _, column := range columns {
		expr := hook.Fields[column]
		value, err := resolvePluginHookValue(expr, payload)
		if err != nil {
			return gerror.Wrapf(err, "resolve plugin %s hook field failed: %s", pluginID, column)
		}
		values = append(values, value)
		placeholders = append(placeholders, "?")
	}

	sql := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		hook.Table,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)
	_, err := g.DB().Exec(ctx, sql, values...)
	return err
}

// executePluginSleepHook sleeps for the duration specified in the hook spec.
func executePluginSleepHook(ctx context.Context, hook *catalog.HookSpec) error {
	if hook == nil || hook.SleepMs <= 0 {
		return nil
	}
	timer := time.NewTimer(time.Duration(hook.SleepMs) * time.Millisecond)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// executePluginErrorHook returns the configured error message.
func executePluginErrorHook(hook *catalog.HookSpec) error {
	if hook == nil {
		return nil
	}
	return gerror.New(strings.TrimSpace(hook.ErrorMessage))
}

// resolvePluginHookValue evaluates one hook field expression against the hook payload.
func resolvePluginHookValue(expr string, payload map[string]interface{}) (interface{}, error) {
	if expr == "now" {
		return time.Now(), nil
	}
	if strings.HasPrefix(expr, pluginHookEventFieldExprPrefix) {
		fieldName := strings.TrimPrefix(expr, pluginHookEventFieldExprPrefix)
		if value, ok := payload[fieldName]; ok {
			return value, nil
		}
		return nil, gerror.Newf("hook event field does not exist: %s", fieldName)
	}
	return nil, gerror.Newf("unsupported hook field expression: %s", expr)
}

// sortStrings sorts a string slice in place.
func sortStrings(items []string) {
	for i := 1; i < len(items); i++ {
		for j := i; j > 0 && items[j] < items[j-1]; j-- {
			items[j], items[j-1] = items[j-1], items[j]
		}
	}
}
