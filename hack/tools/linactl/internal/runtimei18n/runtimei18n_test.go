// This file verifies runtime i18n scanning and message coverage helpers after
// consolidation into linactl.

package runtimei18n

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestPathMatchesSupportsRecursiveGlobs verifies repository-style globs match
// direct and nested files.
func TestPathMatchesSupportsRecursiveGlobs(t *testing.T) {
	t.Parallel()

	if !pathMatches("apps/lina-core/**/*.go", "apps/lina-core/main.go") {
		t.Fatal("expected recursive glob to match direct Go file")
	}
	if !pathMatches("apps/lina-core/**/*.go", "apps/lina-core/internal/service/user/user.go") {
		t.Fatal("expected recursive glob to match nested Go file")
	}
	if pathMatches("apps/lina-core/**/*.go", "apps/lina-core/internal/service/user/user.ts") {
		t.Fatal("expected Go glob not to match TypeScript file")
	}
}

// TestScanRuntimeI18NFindsHardcodedGoErrors verifies scanner findings include
// runtime-visible Chinese gerror messages.
func TestScanRuntimeI18NFindsHardcodedGoErrors(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	mustWriteToolTestFile(t, filepath.Join(repoRoot, "apps", "lina-core", "go.mod"), "module lina-core\n")
	mustWriteToolTestFile(t, filepath.Join(repoRoot, "apps", "lina-vben", "package.json"), "{}\n")
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-core", "internal", "service", "demo", "demo.go"),
		"package demo\n\nfunc f() error { return gerror.New(\"中文错误\") }\n",
	)

	findings, err := scanRuntimeI18N(repoRoot, scanOptions{})
	if err != nil {
		t.Fatalf("expected scan to succeed, got error: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected one finding, got %#v", findings)
	}
	if findings[0].Rule != "go-caller-error-han" {
		t.Fatalf("expected go-caller-error-han finding, got %#v", findings[0])
	}
}

// TestScanRuntimeI18NFindsExpandedBackendPatterns verifies backend scanner
// coverage for common caller-visible text shapes.
func TestScanRuntimeI18NFindsExpandedBackendPatterns(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	mustWriteToolTestFile(t, filepath.Join(repoRoot, "apps", "lina-core", "go.mod"), "module lina-core\n")
	mustWriteToolTestFile(t, filepath.Join(repoRoot, "apps", "lina-vben", "package.json"), "{}\n")
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-core", "internal", "service", "demo", "demo.go"),
		"package demo\n\nfunc f() {\n_ = errors.New(\"中文错误\")\nitem.Label = \"中文标签\"\nheaders := []string{\"中文表头\"}\n}\n",
	)

	findings, err := scanRuntimeI18N(repoRoot, scanOptions{})
	if err != nil {
		t.Fatalf("expected scan to succeed, got error: %v", err)
	}
	rules := make(map[string]struct{}, len(findings))
	for _, finding := range findings {
		rules[finding.Rule] = struct{}{}
	}
	for _, expected := range []string{"go-caller-error-han", "go-message-assignment-han", "go-artifact-slice-han"} {
		if _, ok := rules[expected]; !ok {
			t.Fatalf("expected %s finding, got %#v", expected, findings)
		}
	}
}

// TestScanRuntimeI18NFindsPluginTypeScriptCopy verifies plugin frontend TS
// files are scanned for hardcoded visible copy.
func TestScanRuntimeI18NFindsPluginTypeScriptCopy(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "demo", "frontend", "pages", "data.ts"),
		"export const columns = [{ label: '中文标签' }];\n",
	)

	findings, err := scanRuntimeI18N(repoRoot, scanOptions{})
	if err != nil {
		t.Fatalf("expected scan to succeed, got error: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected one finding, got %#v", findings)
	}
	if findings[0].Rule != "frontend-property-han" {
		t.Fatalf("expected frontend-property-han finding, got %#v", findings[0])
	}
}

// TestScanRuntimeI18NReportsAllowlistAndExcludedStats verifies classified
// reports include allowlist, generated-source, and test-fixture counts.
func TestScanRuntimeI18NReportsAllowlistAndExcludedStats(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	mustWriteToolTestFile(t, filepath.Join(repoRoot, "apps", "lina-core", "go.mod"), "module lina-core\n")
	mustWriteToolTestFile(t, filepath.Join(repoRoot, "apps", "lina-vben", "package.json"), "{}\n")
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-core", "internal", "service", "demo", "demo.go"),
		"package demo\n\nfunc f() error { return errors.New(\"中文错误\") }\n",
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-core", "internal", "model", "entity", "demo.go"),
		"package entity\n\ntype Demo struct { Name string `description:\"中文字段\"` }\n",
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-core", "internal", "service", "demo", "demo_test.go"),
		"package demo\n\nconst fixture = \"中文测试样例\"\n",
	)
	allowlistPath := filepath.Join(repoRoot, "allowlist.json")
	mustWriteToolTestFile(
		t,
		allowlistPath,
		"{\"entries\":[{\"path\":\"apps/lina-core/internal/service/demo/demo.go\",\"rule\":\"go-caller-error-han\",\"category\":\"UserMessage\",\"reason\":\"temporary fixture\",\"scope\":\"test fixture\"}]}\n",
	)

	report, err := scanRuntimeI18NReport(repoRoot, scanOptions{allowlistPath: allowlistPath})
	if err != nil {
		t.Fatalf("expected scan report to succeed, got error: %v", err)
	}
	if report.Summary.Violations != 0 {
		t.Fatalf("expected allowlisted source to avoid violations, got %#v", report.Summary)
	}
	if report.Summary.AllowlistHits != 1 {
		t.Fatalf("expected one allowlist hit, got %#v", report.Summary)
	}
	if report.Summary.GeneratedFiles != 1 || report.Summary.GeneratedItems != 1 {
		t.Fatalf("expected generated stats, got %#v", report.Summary)
	}
	if report.Summary.TestFixtureFiles != 1 || report.Summary.TestFixtureItems != 1 {
		t.Fatalf("expected test fixture stats, got %#v", report.Summary)
	}
}

// TestValidateRuntimeI18NMessagesReportsMissingKeys verifies locale coverage
// validation compares non-baseline locales against zh-CN.
func TestValidateRuntimeI18NMessagesReportsMissingKeys(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	mustWriteToolTestFile(t, filepath.Join(repoRoot, "apps", "lina-core", "go.mod"), "module lina-core\n")
	mustWriteToolTestFile(t, filepath.Join(repoRoot, "apps", "lina-vben", "package.json"), "{}\n")
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-core", "manifest", "i18n", "zh-CN", "error.json"),
		"{\"error\":{\"demo\":{\"missing\":\"缺失\",\"shared\":\"共享\"}}}\n",
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-core", "manifest", "i18n", "en-US", "error.json"),
		"{\"error\":{\"demo\":{\"shared\":\"Shared\"}}}\n",
	)

	errors, err := validateRuntimeI18NMessages(repoRoot)
	if err != nil {
		t.Fatalf("expected validation to run, got error: %v", err)
	}
	if len(errors) != 1 || !strings.Contains(errors[0], "en-US missing key from zh-CN: error.demo.missing") {
		t.Fatalf("expected one missing-key error, got %#v", errors)
	}
}

// TestValidateFrontendI18NKeysReportsMissingPluginCommonKey verifies plugin
// pages are checked against the effective host and plugin catalogs.
func TestValidateFrontendI18NKeysReportsMissingPluginCommonKey(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-vben", "apps", "web-antd", "src", "locales", "langs", "zh-CN", "pages.json"),
		"{\"common\":{\"edit\":\"编辑\"}}\n",
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-vben", "apps", "web-antd", "src", "locales", "langs", "en-US", "pages.json"),
		"{\"common\":{\"edit\":\"Edit\"}}\n",
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "demo", "plugin.yaml"),
		"id: demo\ni18n:\n  enabled: true\n",
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "demo", "manifest", "i18n", "zh-CN", "plugin.json"),
		"{\"plugin\":{\"demo\":{\"title\":\"演示\"}}}\n",
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "demo", "manifest", "i18n", "en-US", "plugin.json"),
		"{\"plugin\":{\"demo\":{\"title\":\"Demo\"}}}\n",
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "demo", "frontend", "pages", "demo.vue"),
		"<template>{{ $t('plugin.demo.title') }} {{ $t('pages.common.save') }}</template>\n",
	)

	errors, err := validateFrontendI18NKeyReferences(repoRoot)
	if err != nil {
		t.Fatalf("expected validation to run, got error: %v", err)
	}
	if len(errors) != 2 {
		t.Fatalf("expected two locale errors, got %#v", errors)
	}
	for _, expected := range []string{
		"en-US missing frontend key referenced by apps/lina-plugins/demo/frontend/pages/demo.vue:1: pages.common.save",
		"zh-CN missing frontend key referenced by apps/lina-plugins/demo/frontend/pages/demo.vue:1: pages.common.save",
	} {
		found := false
		for _, item := range errors {
			if strings.Contains(item, expected) {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected frontend key error containing %q, got %#v", expected, errors)
		}
	}
}

// TestValidateFrontendI18NKeysAllowsHostSourcePluginKeys verifies host frontend
// source can use source-plugin runtime keys present in the merged runtime bundle.
func TestValidateFrontendI18NKeysAllowsHostSourcePluginKeys(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "demo", "plugin.yaml"),
		"id: demo\ni18n:\n  enabled: true\n",
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "demo", "manifest", "i18n", "zh-CN", "plugin.json"),
		"{\"plugin\":{\"demo\":{\"title\":\"演示\"}}}\n",
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "demo", "manifest", "i18n", "en-US", "plugin.json"),
		"{\"plugin\":{\"demo\":{\"title\":\"Demo\"}}}\n",
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-vben", "apps", "web-antd", "src", "views", "demo.vue"),
		"<template>{{ $t('plugin.demo.title') }}</template>\n",
	)

	errors, err := validateFrontendI18NKeyReferences(repoRoot)
	if err != nil {
		t.Fatalf("expected validation to run, got error: %v", err)
	}
	if len(errors) != 0 {
		t.Fatalf("expected no frontend key errors, got %#v", errors)
	}
}

// TestRunFrontendKeysCommandPasses verifies frontend key coverage succeeds when
// app common keys and plugin keys are both present.
func TestRunFrontendKeysCommandPasses(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-vben", "apps", "web-antd", "src", "locales", "langs", "zh-CN", "pages.json"),
		"{\"common\":{\"save\":\"保存\"}}\n",
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-vben", "apps", "web-antd", "src", "locales", "langs", "en-US", "pages.json"),
		"{\"common\":{\"save\":\"Save\"}}\n",
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "demo", "plugin.yaml"),
		"id: demo\ni18n:\n  enabled: true\n",
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "demo", "manifest", "i18n", "zh-CN", "plugin.json"),
		"{\"plugin\":{\"demo\":{\"title\":\"演示\"}}}\n",
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "demo", "manifest", "i18n", "en-US", "plugin.json"),
		"{\"plugin\":{\"demo\":{\"title\":\"Demo\"}}}\n",
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "demo", "frontend", "pages", "demo.vue"),
		"<template>{{ $t('plugin.demo.title') }} {{ $t('pages.common.save') }}</template>\n",
	)

	var out bytes.Buffer
	exitCode, err := Run(repoRoot, []string{"frontend-keys"}, &out)
	if err != nil {
		t.Fatalf("expected frontend key command to succeed, got error: %v", err)
	}
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}
	if !strings.Contains(out.String(), "Runtime i18n frontend key coverage passed") {
		t.Fatalf("expected pass message, got %q", out.String())
	}
}

// TestRunMessagesCommandPasses verifies the command prints the expected pass message.
func TestRunMessagesCommandPasses(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	mustWriteToolTestFile(t, filepath.Join(repoRoot, "apps", "lina-core", "go.mod"), "module lina-core\n")
	mustWriteToolTestFile(t, filepath.Join(repoRoot, "apps", "lina-vben", "package.json"), "{}\n")
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-core", "manifest", "i18n", "zh-CN", "framework.json"),
		"{\"framework\":{\"name\":\"LinaPro\"}}\n",
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-core", "manifest", "i18n", "en-US", "framework.json"),
		"{\"framework\":{\"name\":\"LinaPro\"}}\n",
	)

	var out bytes.Buffer
	exitCode, err := Run(repoRoot, []string{"messages"}, &out)
	if err != nil {
		t.Fatalf("expected messages command to succeed, got error: %v", err)
	}
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}
	if !strings.Contains(out.String(), "Runtime i18n message coverage passed") {
		t.Fatalf("expected pass message, got %q", out.String())
	}
}

// TestValidateModuleLevelCallsDetectsTsTopLevel verifies that $t() calls at
// TypeScript module top level (outside any function) produce a warning.
func TestValidateModuleLevelCallsDetectsTsTopLevel(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "demo", "plugin.yaml"),
		"id: demo\ni18n:\n  enabled: true\n",
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "demo", "manifest", "i18n", "zh-CN", "plugin.json"),
		"{\"plugin\":{\"demo\":{\"title\":\"演示\"}}}\n",
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "demo", "manifest", "i18n", "en-US", "plugin.json"),
		"{\"plugin\":{\"demo\":{\"title\":\"Demo\"}}}\n",
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "demo", "frontend", "pages", "data.ts"),
		"import { $t } from '#/locales';\n\nexport const options = [\n  { label: $t('plugin.demo.missing'), value: 1 },\n];\n",
	)

	warnings, err := validateModuleLevelFrontendI18NCalls(repoRoot)
	if err != nil {
		t.Fatalf("expected validation to run, got error: %v", err)
	}
	if len(warnings) != 1 {
		t.Fatalf("expected one module-level warning, got %#v", warnings)
	}
	if !strings.Contains(warnings[0], "module-level $t()") {
		t.Fatalf("expected warning about module-level $t(), got %q", warnings[0])
	}
}

// TestValidateModuleLevelCallsIgnoresTsFunctionScope verifies that $t() calls
// inside a function body do not produce warnings.
func TestValidateModuleLevelCallsIgnoresTsFunctionScope(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "demo", "plugin.yaml"),
		"id: demo\ni18n:\n  enabled: true\n",
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "demo", "manifest", "i18n", "zh-CN", "plugin.json"),
		"{\"plugin\":{\"demo\":{\"title\":\"演示\"}}}\n",
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "demo", "manifest", "i18n", "en-US", "plugin.json"),
		"{\"plugin\":{\"demo\":{\"title\":\"Demo\"}}}\n",
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "demo", "frontend", "pages", "data.ts"),
		"import { $t } from '#/locales';\n\nexport function getOptions() {\n  return [\n    { label: $t('plugin.demo.title'), value: 1 },\n  ];\n}\n",
	)

	warnings, err := validateModuleLevelFrontendI18NCalls(repoRoot)
	if err != nil {
		t.Fatalf("expected validation to run, got error: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings for function-scope $t(), got %#v", warnings)
	}
}

// TestValidateModuleLevelCallsDetectsVueScriptSetupTopLevel verifies that $t()
// calls at <script setup> top level produce a warning.
func TestValidateModuleLevelCallsDetectsVueScriptSetupTopLevel(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "demo", "plugin.yaml"),
		"id: demo\ni18n:\n  enabled: true\n",
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "demo", "manifest", "i18n", "zh-CN", "plugin.json"),
		"{\"plugin\":{\"demo\":{\"common\":{\"yes\":\"是\",\"no\":\"否\"}}}}\n",
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "demo", "manifest", "i18n", "en-US", "plugin.json"),
		"{\"plugin\":{\"demo\":{\"common\":{\"yes\":\"Yes\",\"no\":\"No\"}}}}\n",
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "demo", "frontend", "pages", "demo.vue"),
		"<script setup lang=\"ts\">\nimport { $t } from '#/locales';\n\nconst options = [\n  { label: $t('plugin.demo.common.yes'), value: 1 },\n  { label: $t('plugin.demo.common.no'), value: 0 },\n];\n</script>\n\n<template><div>{{ $t('plugin.demo.common.yes') }}</div></template>\n",
	)

	warnings, err := validateModuleLevelFrontendI18NCalls(repoRoot)
	if err != nil {
		t.Fatalf("expected validation to run, got error: %v", err)
	}
	if len(warnings) != 2 {
		t.Fatalf("expected two module-level warnings (yes + no), got %#v", warnings)
	}
}

// TestValidateModuleLevelCallsIgnoresVueTemplate verifies that $t() calls in
// <template> blocks do not produce warnings.
func TestValidateModuleLevelCallsIgnoresVueTemplate(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "demo", "plugin.yaml"),
		"id: demo\ni18n:\n  enabled: true\n",
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "demo", "manifest", "i18n", "zh-CN", "plugin.json"),
		"{\"plugin\":{\"demo\":{\"title\":\"演示\"}}}\n",
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "demo", "manifest", "i18n", "en-US", "plugin.json"),
		"{\"plugin\":{\"demo\":{\"title\":\"Demo\"}}}\n",
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "demo", "frontend", "pages", "demo.vue"),
		"<script setup lang=\"ts\">\nimport { $t } from '#/locales';\n\nfunction getTitle() {\n  return $t('plugin.demo.title');\n}\n</script>\n\n<template><div>{{ $t('plugin.demo.title') }}</div></template>\n",
	)

	warnings, err := validateModuleLevelFrontendI18NCalls(repoRoot)
	if err != nil {
		t.Fatalf("expected validation to run, got error: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings for template/function $t(), got %#v", warnings)
	}
}

// TestValidateModuleLevelCallsIgnoresVueObjectLiteral verifies that $t() calls
// inside object literals within <script setup> do not produce warnings.
func TestValidateModuleLevelCallsIgnoresVueObjectLiteral(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "demo", "plugin.yaml"),
		"id: demo\ni18n:\n  enabled: true\n",
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "demo", "manifest", "i18n", "zh-CN", "plugin.json"),
		"{\"plugin\":{\"demo\":{\"title\":\"演示\"}}}\n",
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "demo", "manifest", "i18n", "en-US", "plugin.json"),
		"{\"plugin\":{\"demo\":{\"title\":\"Demo\"}}}\n",
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "demo", "frontend", "pages", "demo.vue"),
		"<script setup lang=\"ts\">\nimport { $t } from '#/locales';\n\nfunction buildSchema() {\n  return [\n    { label: $t('plugin.demo.title'), field: 'title' },\n  ];\n}\n</script>\n",
	)

	warnings, err := validateModuleLevelFrontendI18NCalls(repoRoot)
	if err != nil {
		t.Fatalf("expected validation to run, got error: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings for function-scope object literal $t(), got %#v", warnings)
	}
}

// mustWriteToolTestFile writes one test fixture file.
func mustWriteToolTestFile(t *testing.T, path string, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create fixture dir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write fixture file: %v", err)
	}
}
