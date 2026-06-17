// This file owns the repository walking, YAML parsing, and source-pattern
// checks for plugin governance. The controlled exceptions are kept narrow:
// tests, E2E fixtures, mock-data SQL, install SQL, and migration SQL are
// skipped because they do not enter plugin production runtime paths.

package plugingovernance

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"gopkg.in/yaml.v3"
)

const (
	categoryConfig      = "ConfigGeneration"
	categoryGoAccess    = "ProductionGoAccess"
	categoryHostService = "DynamicHostService"
	categoryLegacy      = "LegacyBoundary"

	ruleConfigCoreTable          = "plugin-dao-config-core-table"
	ruleConfigLegacyBackendPath  = "plugin-dao-config-legacy-backend-path"
	ruleConfigMissingRootConfig  = "plugin-dao-config-missing-root-config"
	ruleGeneratedCoreTableFile   = "plugin-generated-core-table-file"
	ruleGoSharedCoreTable        = "plugin-go-shared-core-table"
	ruleGoModelCoreTable         = "plugin-go-model-core-table"
	ruleGoSQLCoreTable           = "plugin-go-sql-core-table"
	ruleLegacyPluginbridgeClient = "plugin-legacy-pluginbridge-client"
	ruleLegacyHostServiceHelper  = "plugin-legacy-hostservice-helper"
	ruleLegacyHostServiceMethod  = "plugin-legacy-hostservice-method"
	ruleDataCoreTable            = "plugin-data-core-table"
	ruleDataForeignPluginTable   = "plugin-data-foreign-plugin-table"
	ruleDataUnownedTable         = "plugin-data-unowned-table"
	ruleManifestLegacyMethod     = "plugin-manifest-legacy-hostservice-method"
)

var (
	generatedHostCoreTablePathPattern = regexp.MustCompile(`/backend/internal/(?:dao(?:/internal)?|model/(?:do|entity))/(?:.*/)?sys_[a-z0-9_]+\.go$`)
	goSharedCoreTablePattern          = regexp.MustCompile(`\bshared\.TableSys[A-Z][A-Za-z0-9_]*\b`)
	goModelCoreTablePattern           = regexp.MustCompile(`\.(?:Model|Table)\(\s*["']sys_[^"']*["']`)
	goSQLCoreTablePattern             = regexp.MustCompile(`(?i)[` + "`" + `"'][^` + "`" + `"']*\b(?:from|join|update|insert\s+into|delete\s+from)\s+sys_`)
	legacyPluginbridgeClientPattern   = regexp.MustCompile(`\bpluginbridge\.(?:Runtime|Storage|HTTP|Network|Data|Cache|Lock|Config|Notify|Cron|HostConfig|Manifest|Org|Tenant)\s*\(`)
	legacyHostServiceHelperPattern    = regexp.MustCompile(`\b(?:HostServicesForPlugin|ProviderEnv\.Services)\b`)
	legacyHostServiceMethodPattern    = regexp.MustCompile(`\bHostServiceMethod(?:Org|Tenant)[A-Za-z0-9_]*\b`)
)

var productionGoRules = []sourceRule{
	{
		Identifier: ruleGoSharedCoreTable,
		Category:   categoryGoAccess,
		Pattern:    goSharedCoreTablePattern,
		Message:    "plugin production Go must not use shared.TableSys* host table constants",
	},
	{
		Identifier: ruleGoModelCoreTable,
		Category:   categoryGoAccess,
		Pattern:    goModelCoreTablePattern,
		Message:    "plugin production Go must not open sys_* tables through Model/Table",
	},
	{
		Identifier: ruleGoSQLCoreTable,
		Category:   categoryGoAccess,
		Pattern:    goSQLCoreTablePattern,
		Message:    "plugin production Go must not query sys_* tables through raw SQL strings",
	},
	{
		Identifier: ruleLegacyPluginbridgeClient,
		Category:   categoryLegacy,
		Pattern:    legacyPluginbridgeClientPattern,
		Message:    "dynamic plugin code must use the governed guest/domain directory instead of old pluginbridge business clients",
	},
	{
		Identifier: ruleLegacyHostServiceHelper,
		Category:   categoryLegacy,
		Pattern:    legacyHostServiceHelperPattern,
		Message:    "plugin code must not use removed host-service escape hatches",
	},
	{
		Identifier: ruleLegacyHostServiceMethod,
		Category:   categoryLegacy,
		Pattern:    legacyHostServiceMethodPattern,
		Message:    "plugin code must not depend on removed org/tenant host-service method constants",
	},
}

var legacyManifestMethods = map[string]map[string]struct{}{
	"org": {
		"available":                  {},
		"status":                     {},
		"user_dept_assignments.list": {},
		"user_dept_info.get":         {},
		"user_dept_name.get":         {},
		"user_dept_ids.get":          {},
		"user_post_ids.get":          {},
	},
	"tenant": {
		"available":               {},
		"status":                  {},
		"current":                 {},
		"platform_bypass":         {},
		"visible.ensure":          {},
		"user_in_tenant.validate": {},
		"user_tenants.list":       {},
		"switch.validate":         {},
	},
}

type sourceRule struct {
	Identifier string
	Category   string
	Pattern    *regexp.Regexp
	Message    string
}

type tableNode struct {
	Table string
	Line  int
}

// Scan runs plugin governance checks against repoRoot.
func Scan(repoRoot string) (*Report, error) {
	root := filepath.Clean(strings.TrimSpace(repoRoot))
	if root == "" || root == "." {
		return nil, fmt.Errorf("repository root is required")
	}

	report := newReport()
	pluginRoots, err := discoverPluginRoots(root)
	if err != nil {
		return nil, err
	}
	for _, pluginRoot := range pluginRoots {
		pluginID, err := scanPluginManifest(root, pluginRoot, report)
		if err != nil {
			return nil, err
		}
		if pluginID == "" {
			pluginID = filepath.Base(pluginRoot)
		}
		if err = scanPluginDAOConfig(root, pluginRoot, report); err != nil {
			return nil, err
		}
		if err = scanPluginProductionGo(root, pluginRoot, pluginID, report); err != nil {
			return nil, err
		}
	}
	finalizeReport(report)
	return report, nil
}

// discoverPluginRoots returns plugin directories under apps/lina-plugins that
// have plugin.yaml.
func discoverPluginRoots(repoRoot string) ([]string, error) {
	pluginsRoot := filepath.Join(repoRoot, "apps", "lina-plugins")
	if _, err := os.Stat(pluginsRoot); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("stat plugin root %s: %w", pluginsRoot, err)
	}

	var roots []string
	entries, err := os.ReadDir(pluginsRoot)
	if err != nil {
		return nil, fmt.Errorf("read plugin root %s: %w", pluginsRoot, err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pluginRoot := filepath.Join(pluginsRoot, entry.Name())
		if _, statErr := os.Stat(filepath.Join(pluginRoot, "plugin.yaml")); statErr != nil {
			if os.IsNotExist(statErr) {
				continue
			}
			return nil, fmt.Errorf("stat plugin manifest %s: %w", filepath.Join(pluginRoot, "plugin.yaml"), statErr)
		}
		roots = append(roots, pluginRoot)
	}
	sort.Strings(roots)
	return roots, nil
}

// scanPluginDAOConfig blocks legacy config locations, missing reproducible DAO
// configs, and plugin DAO generation for host sys_* tables.
func scanPluginDAOConfig(repoRoot string, pluginRoot string, report *Report) error {
	legacyConfigPath := filepath.Join(pluginRoot, "backend", "hack", "config.yaml")
	if _, err := os.Stat(legacyConfigPath); err == nil {
		relPath, relErr := relSlash(repoRoot, legacyConfigPath)
		if relErr != nil {
			return relErr
		}
		addFinding(
			report,
			relPath,
			1,
			ruleConfigLegacyBackendPath,
			categoryConfig,
			"plugin DAO config must move from backend/hack/config.yaml to plugin-root hack/config.yaml",
			relPath,
		)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stat plugin legacy DAO config %s: %w", legacyConfigPath, err)
	}

	configPath := filepath.Join(pluginRoot, "hack", "config.yaml")
	if _, err := os.Stat(configPath); err != nil {
		if os.IsNotExist(err) {
			if hasGeneratedDAO(pluginRoot) {
				relPath, relErr := relSlash(repoRoot, filepath.Join(pluginRoot, "backend", "internal", "dao"))
				if relErr != nil {
					return relErr
				}
				addFinding(
					report,
					relPath,
					1,
					ruleConfigMissingRootConfig,
					categoryConfig,
					"plugin generated DAO files require plugin-root hack/config.yaml for reproducible generation",
					relPath,
				)
			}
			return nil
		}
		return fmt.Errorf("stat plugin DAO config %s: %w", configPath, err)
	}
	report.Summary.ConfigFiles++

	root, err := readYAMLDocument(configPath)
	if err != nil {
		return err
	}
	relPath, err := relSlash(repoRoot, configPath)
	if err != nil {
		return err
	}
	for _, item := range collectTables(root) {
		table := normalizeTableName(item.Table)
		if strings.HasPrefix(table, "sys_") {
			addFinding(
				report,
				relPath,
				item.Line,
				ruleConfigCoreTable,
				categoryConfig,
				"plugin root hack/config.yaml must not generate host sys_* tables",
				item.Table,
			)
		}
	}
	return nil
}

func hasGeneratedDAO(pluginRoot string) bool {
	info, err := os.Stat(filepath.Join(pluginRoot, "backend", "internal", "dao"))
	return err == nil && info.IsDir()
}

// scanPluginManifest validates dynamic hostServices data table ownership and
// legacy org/tenant method declarations.
func scanPluginManifest(repoRoot string, pluginRoot string, report *Report) (string, error) {
	manifestPath := filepath.Join(pluginRoot, "plugin.yaml")
	root, err := readYAMLDocument(manifestPath)
	if err != nil {
		return "", err
	}
	report.Summary.ManifestFiles++

	relPath, err := relSlash(repoRoot, manifestPath)
	if err != nil {
		return "", err
	}
	pluginID := scalarValue(mappingValue(root, "id"))
	if pluginID == "" {
		pluginID = filepath.Base(pluginRoot)
	}

	hostServices := mappingValue(root, "hostServices")
	if hostServices == nil || hostServices.Kind != yaml.SequenceNode {
		return pluginID, nil
	}

	for _, serviceNode := range hostServices.Content {
		serviceName := normalizeServiceName(scalarValue(mappingValue(serviceNode, "service")))
		for _, method := range scalarSequenceValues(mappingValue(serviceNode, "methods")) {
			if isLegacyManifestMethod(serviceName, method.Table) {
				addFinding(
					report,
					relPath,
					method.Line,
					ruleManifestLegacyMethod,
					categoryLegacy,
					"dynamic plugin hostServices must not declare removed org/tenant legacy methods",
					fmt.Sprintf("%s.%s", serviceName, method.Table),
				)
			}
		}
		if serviceName != "data" {
			continue
		}
		resources := mappingValue(serviceNode, "resources")
		for _, table := range scalarSequenceValues(mappingValue(resources, "tables")) {
			validateDataServiceTable(report, relPath, pluginID, table)
		}
	}
	return pluginID, nil
}

// scanPluginProductionGo blocks direct host table access in plugin production Go.
func scanPluginProductionGo(repoRoot string, pluginRoot string, pluginID string, report *Report) error {
	return filepath.WalkDir(pluginRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if shouldSkipPluginDir(entry.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" {
			return nil
		}
		relPath, err := relSlash(repoRoot, path)
		if err != nil {
			return err
		}
		if isControlledNonProductionPath(relPath) {
			return nil
		}
		report.Summary.GoFiles++
		if generatedHostCoreTablePathPattern.MatchString(relPath) {
			addFinding(
				report,
				relPath,
				1,
				ruleGeneratedCoreTableFile,
				categoryGoAccess,
				"plugin generated DAO/DO/Entity files must not include host sys_* tables",
				filepath.Base(path),
			)
		}
		return scanGoSourceFile(path, relPath, pluginID, report)
	})
}

// scanGoSourceFile applies line-oriented production Go patterns.
func scanGoSourceFile(path string, relPath string, _ string, report *Report) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read Go source %s: %w", relPath, err)
	}
	lines := strings.Split(string(content), "\n")
	for index, line := range lines {
		lineNumber := index + 1
		if isCommentLine(line) {
			continue
		}
		for _, rule := range productionGoRules {
			if !rule.Pattern.MatchString(line) {
				continue
			}
			addFinding(report, relPath, lineNumber, rule.Identifier, rule.Category, rule.Message, line)
			break
		}
	}
	return nil
}

// validateDataServiceTable enforces current-plugin table ownership for data grants.
func validateDataServiceTable(report *Report, relPath string, pluginID string, table tableNode) {
	tableName := normalizeTableName(table.Table)
	if tableName == "" {
		return
	}
	switch {
	case strings.HasPrefix(tableName, "sys_"):
		addFinding(
			report,
			relPath,
			table.Line,
			ruleDataCoreTable,
			categoryHostService,
			"dynamic data host service must not grant host sys_* tables",
			table.Table,
		)
	case !isPluginOwnedTable(pluginID, tableName):
		rule := ruleDataUnownedTable
		message := "dynamic data host service may only grant current-plugin owned tables"
		if strings.HasPrefix(tableName, "plugin_") {
			rule = ruleDataForeignPluginTable
			message = "dynamic data host service must not grant another plugin's internal table"
		}
		addFinding(report, relPath, table.Line, rule, categoryHostService, message, table.Table)
	}
}

// readYAMLDocument parses one YAML document into a node tree.
func readYAMLDocument(path string) (*yaml.Node, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read YAML %s: %w", path, err)
	}
	var doc yaml.Node
	if err = yaml.Unmarshal(content, &doc); err != nil {
		return nil, fmt.Errorf("parse YAML %s: %w", path, err)
	}
	if len(doc.Content) == 0 {
		return &doc, nil
	}
	return doc.Content[0], nil
}

// collectTables returns every YAML value under a mapping key named tables.
func collectTables(node *yaml.Node) []tableNode {
	if node == nil {
		return nil
	}
	var result []tableNode
	switch node.Kind {
	case yaml.MappingNode:
		for index := 0; index+1 < len(node.Content); index += 2 {
			key := node.Content[index]
			value := node.Content[index+1]
			if key.Value == "tables" {
				result = append(result, scalarSequenceValues(value)...)
			}
			result = append(result, collectTables(value)...)
		}
	case yaml.SequenceNode:
		for _, child := range node.Content {
			result = append(result, collectTables(child)...)
		}
	}
	return result
}

// mappingValue returns one mapping field from a YAML node.
func mappingValue(node *yaml.Node, key string) *yaml.Node {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}
	for index := 0; index+1 < len(node.Content); index += 2 {
		if node.Content[index].Value == key {
			return node.Content[index+1]
		}
	}
	return nil
}

// scalarValue returns a trimmed scalar string.
func scalarValue(node *yaml.Node) string {
	if node == nil || node.Kind != yaml.ScalarNode {
		return ""
	}
	return strings.TrimSpace(node.Value)
}

// scalarSequenceValues extracts table or method strings with source lines.
func scalarSequenceValues(node *yaml.Node) []tableNode {
	if node == nil {
		return nil
	}
	switch node.Kind {
	case yaml.SequenceNode:
		result := make([]tableNode, 0, len(node.Content))
		for _, child := range node.Content {
			result = append(result, scalarSequenceValues(child)...)
		}
		return result
	case yaml.ScalarNode:
		values := splitScalarList(node.Value)
		result := make([]tableNode, 0, len(values))
		for _, value := range values {
			result = append(result, tableNode{Table: value, Line: node.Line})
		}
		return result
	default:
		return nil
	}
}

// splitScalarList parses comma-separated gfcli table strings and scalar lists.
func splitScalarList(value string) []string {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return nil
	}
	var parts []string
	if strings.Contains(normalized, ",") {
		parts = strings.Split(normalized, ",")
	} else {
		parts = strings.Fields(normalized)
	}
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		token := normalizeScalarToken(part)
		if token != "" {
			result = append(result, token)
		}
	}
	return result
}

// normalizeScalarToken trims YAML scalar punctuation around one table or method.
func normalizeScalarToken(value string) string {
	token := strings.TrimSpace(value)
	token = strings.Trim(token, `"'`)
	token = strings.TrimRightFunc(token, func(r rune) bool {
		return r == ',' || r == ';' || unicode.IsSpace(r)
	})
	return strings.TrimSpace(token)
}

// normalizeTableName canonicalizes table names for ownership checks.
func normalizeTableName(table string) string {
	return strings.ToLower(normalizeScalarToken(table))
}

// normalizeServiceName canonicalizes host service names.
func normalizeServiceName(service string) string {
	return strings.ToLower(strings.TrimSpace(service))
}

// isPluginOwnedTable reports whether a table belongs to the current plugin.
func isPluginOwnedTable(pluginID string, table string) bool {
	normalizedPlugin := strings.ReplaceAll(strings.ToLower(strings.TrimSpace(pluginID)), "-", "_")
	prefix := "plugin_" + normalizedPlugin
	return table == prefix || strings.HasPrefix(table, prefix+"_")
}

// isLegacyManifestMethod reports old org/tenant host service methods that must
// be replaced by the new domain method naming model.
func isLegacyManifestMethod(service string, method string) bool {
	serviceMethods, ok := legacyManifestMethods[service]
	if !ok {
		return false
	}
	_, ok = serviceMethods[strings.TrimSpace(method)]
	return ok
}

// shouldSkipPluginDir skips directories that cannot contain production Go.
func shouldSkipPluginDir(name string) bool {
	switch name {
	case ".git", "node_modules", "dist", "temp", "test-results", "playwright-report":
		return true
	default:
		return false
	}
}

// isControlledNonProductionPath documents the controlled exceptions requested
// by the OpenSpec change. These paths are allowed to touch host tables only for
// tests, mock data, installation SQL, migration SQL, or generated test assets.
func isControlledNonProductionPath(relPath string) bool {
	path := filepath.ToSlash(relPath)
	if strings.HasSuffix(path, "_test.go") {
		return true
	}
	exceptions := []string{
		"/hack/tests/",
		"/manifest/sql/",
		"/manifest/sql/mock-data/",
		"/migrations/",
		"/migration/",
	}
	for _, exception := range exceptions {
		if strings.Contains(path, exception) {
			return true
		}
	}
	return false
}

// isCommentLine skips comments when applying line-oriented Go source patterns.
func isCommentLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	return trimmed == "" || strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") || strings.HasPrefix(trimmed, "*")
}

// relSlash returns a repository-relative slash-separated path.
func relSlash(root string, path string) (string, error) {
	relPath, err := filepath.Rel(root, path)
	if err != nil {
		return "", err
	}
	return filepath.ToSlash(relPath), nil
}
