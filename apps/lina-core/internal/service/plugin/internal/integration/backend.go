// This file loads plugin backend declarations and dispatches generic plugin
// resource queries.

package integration

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/gtime"
	"gopkg.in/yaml.v3"

	"lina-core/internal/service/plugin/internal/catalog"
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

// LoadPluginBackendConfig loads plugin-owned hook and resource declarations into the manifest.
// It implements catalog.BackendConfigLoader.
func (s *serviceImpl) LoadPluginBackendConfig(manifest *catalog.Manifest) error {
	manifest.Hooks = make([]*catalog.HookSpec, 0)
	manifest.BackendResources = make(map[string]*catalog.ResourceSpec)

	if manifest.SourcePlugin != nil {
		return nil
	}

	if manifest.RuntimeArtifact != nil {
		manifest.Hooks = catalog.CloneHookSpecs(manifest.RuntimeArtifact.HookSpecs)
		manifest.BackendResources = catalog.CloneResourceSpecsToMap(manifest.RuntimeArtifact.ResourceSpecs)
		return nil
	}

	hookFiles, err := gfile.ScanDirFile(filepath.Join(manifest.RootDir, "backend", "hooks"), "*.yaml", false)
	if err != nil && !gfile.Exists(filepath.Join(manifest.RootDir, "backend", "hooks")) {
		err = nil
	}
	if err != nil {
		return err
	}
	for _, hookFile := range hookFiles {
		spec := &catalog.HookSpec{}
		if err = loadPluginYAMLFile(hookFile, spec); err != nil {
			return err
		}
		if err = catalog.ValidateHookSpec(manifest.ID, spec, hookFile); err != nil {
			return err
		}
		manifest.Hooks = append(manifest.Hooks, spec)
	}

	resourceFiles, err := gfile.ScanDirFile(filepath.Join(manifest.RootDir, "backend", "resources"), "*.yaml", false)
	if err != nil && !gfile.Exists(filepath.Join(manifest.RootDir, "backend", "resources")) {
		err = nil
	}
	if err != nil {
		return err
	}
	for _, resourceFile := range resourceFiles {
		spec := &catalog.ResourceSpec{}
		if err = loadPluginYAMLFile(resourceFile, spec); err != nil {
			return err
		}
		if err = catalog.ValidateResourceSpec(manifest.ID, spec, resourceFile); err != nil {
			return err
		}
		manifest.BackendResources[spec.Key] = spec
	}
	return nil
}

// loadPluginYAMLFile reads a YAML file at filePath and unmarshals it into target.
func loadPluginYAMLFile(filePath string, target interface{}) error {
	content := gfile.GetBytes(filePath)
	if len(content) == 0 {
		return gerror.Newf("插件配置文件为空: %s", filePath)
	}
	if err := yaml.Unmarshal(content, target); err != nil {
		return gerror.Wrapf(err, "解析插件配置文件失败: %s", filePath)
	}
	return nil
}

// ListResourceRecords queries plugin-owned backend resource rows using the
// generic plugin resource contract.
func (s *serviceImpl) ListResourceRecords(ctx context.Context, in ResourceListInput) (*ResourceListOutput, error) {
	manifest, err := s.catalogSvc.GetActiveManifest(ctx, in.PluginID)
	if err != nil {
		return nil, err
	}
	if !s.IsEnabled(ctx, in.PluginID) {
		return nil, gerror.New("插件未启用")
	}

	resource, ok := manifest.BackendResources[in.ResourceID]
	if !ok {
		return nil, gerror.New("插件资源不存在")
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
		switch catalog.NormalizeResourceFilterOperator(filter.Operator) {
		case catalog.ResourceFilterOperatorEQ:
			m = m.Where(filter.Column, value)
		case catalog.ResourceFilterOperatorLike:
			m = m.WhereLike(filter.Column, "%"+value+"%")
		case catalog.ResourceFilterOperatorGTEDate:
			m = m.WhereGTE(filter.Column, value+" 00:00:00")
		case catalog.ResourceFilterOperatorLTEDate:
			m = m.WhereLTE(filter.Column, value+" 23:59:59")
		default:
			return nil, gerror.Newf("插件资源过滤操作符不支持: %s", filter.Operator)
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
		fields = append(fields, fmt.Sprintf("%s AS %s", field.Column, field.Name))
	}
	fieldArgs := make([]interface{}, 0, len(fields))
	for _, field := range fields {
		fieldArgs = append(fieldArgs, field)
	}

	queryModel := m.Fields(fieldArgs...).Page(in.PageNum, in.PageSize)
	if catalog.NormalizeResourceOrderDirection(resource.OrderBy.Direction) == catalog.ResourceOrderDirectionDESC {
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
			row[field.Name] = normalizePluginResourceValue(recordMap[field.Name])
		}
		items = append(items, row)
	}
	return &ResourceListOutput{List: items, Total: total}, nil
}

// normalizePluginResourceValue converts GoFrame time values to JSON-safe strings.
func normalizePluginResourceValue(value interface{}) interface{} {
	switch typedValue := value.(type) {
	case *gtime.Time:
		if typedValue == nil {
			return ""
		}
		return typedValue.String()
	case gtime.Time:
		return typedValue.String()
	default:
		return value
	}
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
			return gerror.Wrapf(err, "解析插件%s的Hook字段失败: %s", pluginID, column)
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
		return gtime.Now(), nil
	}
	if strings.HasPrefix(expr, pluginHookEventFieldExprPrefix) {
		fieldName := strings.TrimPrefix(expr, pluginHookEventFieldExprPrefix)
		if value, ok := payload[fieldName]; ok {
			return value, nil
		}
		return nil, gerror.Newf("Hook事件字段不存在: %s", fieldName)
	}
	return nil, gerror.Newf("不支持的Hook字段表达式: %s", expr)
}

// sortStrings sorts a string slice in place.
func sortStrings(items []string) {
	for i := 1; i < len(items); i++ {
		for j := i; j > 0 && items[j] < items[j-1]; j-- {
			items[j], items[j-1] = items[j-1], items[j]
		}
	}
}
