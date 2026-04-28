// This file verifies runtime i18n tool scanning and message coverage helpers.

package main

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
	if findings[0].Rule != "go-error-han" {
		t.Fatalf("expected go-error-han finding, got %#v", findings[0])
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

// TestRunMessagesCommandPasses verifies the command prints the expected pass message.
func TestRunMessagesCommandPasses(t *testing.T) {
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

	previousWorkingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("resolve working directory: %v", err)
	}
	t.Cleanup(func() {
		if chdirErr := os.Chdir(previousWorkingDir); chdirErr != nil {
			t.Fatalf("restore working directory: %v", chdirErr)
		}
	})
	if err = os.Chdir(repoRoot); err != nil {
		t.Fatalf("switch working directory: %v", err)
	}

	var out bytes.Buffer
	exitCode, err := run([]string{"messages"}, &out)
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
