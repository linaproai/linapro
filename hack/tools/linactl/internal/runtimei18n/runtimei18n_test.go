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

// TestValidateBizerrMessageKeysReportsHostMissingCatalogEntry verifies host
// bizerr MustDefine keys are checked against host error.json catalogs.
func TestValidateBizerrMessageKeysReportsHostMissingCatalogEntry(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	mustWriteToolTestFile(t, filepath.Join(repoRoot, "apps", "lina-core", "go.mod"), "module lina-core\n")
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-core", "internal", "service", "demo", "demo_code.go"),
		`package demo
import (
	"github.com/gogf/gf/v2/errors/gcode"
	"lina-core/pkg/bizerr"
)
var CodeDemoMissing = bizerr.MustDefine(
	"DEMO_MISSING_KEY",
	"Demo missing key",
	gcode.CodeInternalError,
)
`,
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-core", "manifest", "i18n", "zh-CN", "error.json"),
		`{"error":{"demo":{"present":"存在"}}}
`,
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-core", "manifest", "i18n", "en-US", "error.json"),
		`{"error":{"demo":{"present":"Present"}}}
`,
	)

	errors, err := validateBizerrMessageKeys(repoRoot)
	if err != nil {
		t.Fatalf("expected host bizerr validation to run, got error: %v", err)
	}
	if len(errors) != 2 {
		t.Fatalf("expected two missing-key errors (zh-CN + en-US), got %#v", errors)
	}
	joined := strings.Join(errors, "\n")
	if !strings.Contains(joined, "DEMO_MISSING_KEY") || !strings.Contains(joined, "error.demo.missing.key") {
		t.Fatalf("expected DEMO_MISSING_KEY mapping failure, got %#v", errors)
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

// TestValidateBizerrMessageKeysReportsPluginGaps verifies i18n-enabled plugins
// must map every bizerr.MustDefine code into their error catalogs.
func TestValidateBizerrMessageKeysReportsPluginGaps(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	mustWriteToolTestFile(t, filepath.Join(repoRoot, "apps", "lina-core", "go.mod"), "module lina-core\n")
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "demo-oidc", "plugin.yaml"),
		"id: demo-oidc\ni18n:\n  enabled: true\n",
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "demo-oidc", "backend", "internal", "service", "oauth", "oauth_code.go"),
		"package oauth\n\nvar CodeDiscoveryFailed = bizerr.MustDefine(\n\t\"PLUGIN_DEMO_DISCOVERY_FAILED\",\n\t\"discovery failed\",\n\tgcode.CodeInternalError,\n)\n",
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "demo-oidc", "manifest", "i18n", "zh-CN", "error.json"),
		"{\"error\":{\"plugin\":{\"demo\":{\"other\":\"其他\"}}}}\n",
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "demo-oidc", "manifest", "i18n", "en-US", "error.json"),
		"{\"error\":{\"plugin\":{\"demo\":{\"other\":\"Other\"}}}}\n",
	)

	errors, err := validateBizerrMessageKeys(repoRoot)
	if err != nil {
		t.Fatalf("expected validation to run, got error: %v", err)
	}
	if len(errors) != 2 {
		t.Fatalf("expected two locale gaps, got %#v", errors)
	}
	for _, item := range errors {
		if !strings.Contains(item, "plugin:demo-oidc") ||
			!strings.Contains(item, "PLUGIN_DEMO_DISCOVERY_FAILED") ||
			!strings.Contains(item, "error.plugin.demo.discovery.failed") {
			t.Fatalf("unexpected gap message: %s", item)
		}
	}
}

// TestValidateBizerrMessageKeysSkipsI18nDisabledPlugins verifies plugins without
// i18n.enabled=true are not required to ship error catalogs for bizerr codes.
func TestValidateBizerrMessageKeysSkipsI18nDisabledPlugins(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "single-lang", "plugin.yaml"),
		"id: single-lang\ni18n:\n  enabled: false\n",
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "single-lang", "backend", "code.go"),
		"package backend\n\nvar CodeX = bizerr.MustDefine(\"PLUGIN_SINGLE_LANG_X\", \"x\", gcode.CodeInternalError)\n",
	)

	errors, err := validateBizerrMessageKeys(repoRoot)
	if err != nil {
		t.Fatalf("expected validation to run, got error: %v", err)
	}
	if len(errors) != 0 {
		t.Fatalf("expected disabled plugin to be skipped, got %#v", errors)
	}
}

// TestValidateBizerrMessageKeysPassesWhenCatalogMatches verifies host and
// plugin catalogs that include derived messageKeys pass coverage.
func TestValidateBizerrMessageKeysPassesWhenCatalogMatches(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	mustWriteToolTestFile(t, filepath.Join(repoRoot, "apps", "lina-core", "go.mod"), "module lina-core\n")
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-core", "internal", "service", "demo", "demo_code.go"),
		"package demo\n\nvar CodeMissing = bizerr.MustDefine(\"HOST_DEMO_MISSING\", \"missing\", gcode.CodeNotFound)\n",
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-core", "manifest", "i18n", "zh-CN", "error.json"),
		"{\"error\":{\"host\":{\"demo\":{\"missing\":\"缺失\"}}}}\n",
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-core", "manifest", "i18n", "en-US", "error.json"),
		"{\"error\":{\"host\":{\"demo\":{\"missing\":\"Missing\"}}}}\n",
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "demo-oidc", "plugin.yaml"),
		"id: demo-oidc\ni18n:\n  enabled: true\n",
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "demo-oidc", "backend", "code.go"),
		"package backend\n\nvar CodeY = bizerr.MustDefine(\"PLUGIN_DEMO_Y\", \"y\", gcode.CodeInternalError)\n",
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "demo-oidc", "manifest", "i18n", "zh-CN", "error.json"),
		"{\"error\":{\"plugin\":{\"demo\":{\"y\":\"Y中文\"}}}}\n",
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "demo-oidc", "manifest", "i18n", "en-US", "error.json"),
		"{\"error\":{\"plugin\":{\"demo\":{\"y\":\"Y\"}}}}\n",
	)

	errors, err := validateBizerrMessageKeys(repoRoot)
	if err != nil {
		t.Fatalf("expected validation to run, got error: %v", err)
	}
	if len(errors) != 0 {
		t.Fatalf("expected matching catalogs to pass, got %#v", errors)
	}
}

// TestDeriveBizerrMessageKeyMatchesBizerrPackage verifies derivation matches
// lina-core/pkg/bizerr.MessageKey rules.
func TestDeriveBizerrMessageKeyMatchesBizerrPackage(t *testing.T) {
	t.Parallel()

	got := deriveBizerrMessageKey("PLUGIN_OIDC_GENERIC_DISCOVERY_FAILED")
	want := "error.plugin.oidc.generic.discovery.failed"
	if got != want {
		t.Fatalf("deriveBizerrMessageKey() = %q, want %q", got, want)
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

// TestValidatePluginDisplayMetadataKeysReportsBareKeys verifies i18n-enabled
// plugins cannot ship bare name/description keys for management-list localization.
func TestValidatePluginDisplayMetadataKeysReportsBareKeys(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "demo-mail", "plugin.yaml"),
		"id: demo-mail\ni18n:\n  enabled: true\n",
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "demo-mail", "manifest", "i18n", "zh-CN", "plugin.json"),
		"{\"name\":\"邮件演示\",\"description\":\"错误的顶层 key\"}\n",
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "demo-mail", "manifest", "i18n", "en-US", "plugin.json"),
		"{\"name\":\"Mail Demo\",\"description\":\"bare keys\"}\n",
	)

	errors, err := validatePluginDisplayMetadataKeys(repoRoot)
	if err != nil {
		t.Fatalf("expected validation to run, got error: %v", err)
	}
	if len(errors) == 0 {
		t.Fatal("expected bare metadata key failures")
	}
	joined := strings.Join(errors, "\n")
	for _, expected := range []string{
		"plugin:demo-mail",
		"plugin.demo-mail.name",
		"plugin.demo-mail.description",
		"bare name/description",
	} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("expected error to mention %q, got:\n%s", expected, joined)
		}
	}
}

// TestValidatePluginDisplayMetadataKeysPassesWhenNamespaced verifies correct
// plugin.<id>.name/description keys pass the management-list metadata check.
func TestValidatePluginDisplayMetadataKeysPassesWhenNamespaced(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "demo-mail", "plugin.yaml"),
		"id: demo-mail\ni18n:\n  enabled: true\n",
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "demo-mail", "manifest", "i18n", "zh-CN", "plugin.json"),
		"{\"plugin\":{\"demo-mail\":{\"name\":\"邮件演示\",\"description\":\"正确的命名空间\"}}}\n",
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "demo-mail", "manifest", "i18n", "en-US", "plugin.json"),
		"{\"plugin\":{\"demo-mail\":{\"name\":\"Mail Demo\",\"description\":\"correct namespace\"}}}\n",
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "single-lang", "plugin.yaml"),
		"id: single-lang\ni18n:\n  enabled: false\n",
	)

	errors, err := validatePluginDisplayMetadataKeys(repoRoot)
	if err != nil {
		t.Fatalf("expected validation to run, got error: %v", err)
	}
	if len(errors) != 0 {
		t.Fatalf("expected namespaced keys to pass, got %#v", errors)
	}
}

// TestValidateRuntimeI18NMessagesIncludesPluginDisplayMetadata verifies the
// consolidated messages check surfaces bare plugin display keys.
func TestValidateRuntimeI18NMessagesIncludesPluginDisplayMetadata(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	mustWriteToolTestFile(t, filepath.Join(repoRoot, "apps", "lina-core", "go.mod"), "module lina-core\n")
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
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "demo-mail", "plugin.yaml"),
		"id: demo-mail\ni18n:\n  enabled: true\n",
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "demo-mail", "manifest", "i18n", "zh-CN", "plugin.json"),
		"{\"name\":\"邮件演示\",\"description\":\"错误的顶层 key\"}\n",
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "demo-mail", "manifest", "i18n", "en-US", "plugin.json"),
		"{\"name\":\"Mail Demo\",\"description\":\"bare keys\"}\n",
	)

	errors, err := validateRuntimeI18NMessages(repoRoot)
	if err != nil {
		t.Fatalf("expected validation to run, got error: %v", err)
	}
	joined := strings.Join(errors, "\n")
	if !strings.Contains(joined, "plugin.demo-mail.name") {
		t.Fatalf("expected consolidated messages check to report display key gap, got:\n%s", joined)
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
	if !strings.Contains(out.String(), "Runtime i18n message coverage passed for host and plugin scopes.") {
		t.Fatalf("expected pass message, got %q", out.String())
	}
}

// TestValidateConfigDisplayMetadataKeysReportsHostSQLGaps verifies host
// sys_config seed keys require config.<key>.name/remark in every locale.
func TestValidateConfigDisplayMetadataKeysReportsHostSQLGaps(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-core", "manifest", "sql", "005-config.sql"),
		"INSERT INTO sys_config (\"name\", \"key\", \"value\") VALUES ('JWT', 'sys.jwt.expire', '24h');\n",
	)
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

	errors, err := validateConfigDisplayMetadataKeys(repoRoot)
	if err != nil {
		t.Fatalf("expected validation to run, got error: %v", err)
	}
	if len(errors) == 0 {
		t.Fatal("expected host config display gaps")
	}
	joined := strings.Join(errors, "\n")
	if !strings.Contains(joined, "host:core") ||
		!strings.Contains(joined, "sys.jwt.expire") ||
		!strings.Contains(joined, "config.sys.jwt.expire.name") {
		t.Fatalf("expected host config.sys.jwt.expire.name gap, got:\n%s", joined)
	}
}

// TestValidateConfigDisplayMetadataKeysReportsPluginSysConfigKeyGaps verifies
// i18n-enabled plugins must ship config.<SysConfigKey>.name/remark entries.
func TestValidateConfigDisplayMetadataKeysReportsPluginSysConfigKeyGaps(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "demo-storage", "plugin.yaml"),
		"id: demo-storage\ni18n:\n  enabled: true\n",
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "demo-storage", "backend", "settings.go"),
		"package settings\n\nconst ConfigKeyBucket hostconfigcap.SysConfigKey = \"plugin.demo-storage.bucket\"\n",
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "demo-storage", "manifest", "i18n", "zh-CN", "plugin.json"),
		"{\"plugin\":{\"demo-storage\":{\"name\":\"演示存储\"}}}\n",
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "demo-storage", "manifest", "i18n", "en-US", "plugin.json"),
		"{\"plugin\":{\"demo-storage\":{\"name\":\"Demo Storage\"}}}\n",
	)

	errors, err := validateConfigDisplayMetadataKeys(repoRoot)
	if err != nil {
		t.Fatalf("expected validation to run, got error: %v", err)
	}
	if len(errors) != 4 {
		t.Fatalf("expected 2 locales * 2 fields = 4 gaps, got %#v", errors)
	}
	for _, item := range errors {
		if !strings.Contains(item, "plugin:demo-storage") ||
			!strings.Contains(item, "plugin.demo-storage.bucket") {
			t.Fatalf("unexpected gap message: %s", item)
		}
	}
}

// TestValidateConfigDisplayMetadataKeysSkipsI18nDisabledPlugins verifies
// plugins without i18n.enabled=true are not required to ship config display keys.
func TestValidateConfigDisplayMetadataKeysSkipsI18nDisabledPlugins(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "single-lang", "plugin.yaml"),
		"id: single-lang\ni18n:\n  enabled: false\n",
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "single-lang", "backend", "settings.go"),
		"package settings\n\nconst ConfigKeyX hostconfigcap.SysConfigKey = \"plugin.single-lang.x\"\n",
	)

	errors, err := validateConfigDisplayMetadataKeys(repoRoot)
	if err != nil {
		t.Fatalf("expected validation to run, got error: %v", err)
	}
	if len(errors) != 0 {
		t.Fatalf("expected disabled plugin to be skipped, got %#v", errors)
	}
}

// TestValidateConfigDisplayMetadataKeysPassesWhenCatalogMatches verifies host
// SQL keys and plugin SysConfigKey constants pass when catalogs are complete.
func TestValidateConfigDisplayMetadataKeysPassesWhenCatalogMatches(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-core", "manifest", "sql", "005-config.sql"),
		"INSERT INTO sys_config (\"name\", \"key\", \"value\") VALUES ('JWT', 'sys.jwt.expire', '24h');\n",
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-core", "manifest", "i18n", "zh-CN", "config.json"),
		"{\"config\":{\"sys\":{\"jwt\":{\"expire\":{\"name\":\"JWT 有效期\",\"remark\":\"时长\"}}}}}\n",
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-core", "manifest", "i18n", "en-US", "config.json"),
		"{\"config\":{\"sys\":{\"jwt\":{\"expire\":{\"name\":\"JWT Expiration\",\"remark\":\"Duration\"}}}}}\n",
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "demo-storage", "plugin.yaml"),
		"id: demo-storage\ni18n:\n  enabled: true\n",
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "demo-storage", "backend", "settings.go"),
		"package settings\n\nconst ConfigKeyBucket hostconfigcap.SysConfigKey = \"plugin.demo-storage.bucket\"\n",
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "demo-storage", "manifest", "i18n", "zh-CN", "config.json"),
		"{\"config\":{\"plugin\":{\"demo-storage\":{\"bucket\":{\"name\":\"存储桶\",\"remark\":\"桶名\"}}}}}\n",
	)
	mustWriteToolTestFile(
		t,
		filepath.Join(repoRoot, "apps", "lina-plugins", "demo-storage", "manifest", "i18n", "en-US", "config.json"),
		"{\"config\":{\"plugin\":{\"demo-storage\":{\"bucket\":{\"name\":\"Bucket\",\"remark\":\"Bucket name\"}}}}}\n",
	)

	errors, err := validateConfigDisplayMetadataKeys(repoRoot)
	if err != nil {
		t.Fatalf("expected validation to run, got error: %v", err)
	}
	if len(errors) != 0 {
		t.Fatalf("expected complete catalogs to pass, got %#v", errors)
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
