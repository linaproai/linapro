// This file verifies source-plugin upgrade CLI planning flow, including dry-run,
// plugin=all selection, downgrade rejection, and uninstalled-plugin skipping.

package sourceupgrade

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	hostsourceupgrade "lina-core/pkg/sourceupgrade"
)

// TestRunSourcePluginUpgradeDryRunPrintsPlan verifies the command prints the
// selected dry-run plan without executing any source-plugin upgrade work.
func TestRunSourcePluginUpgradeDryRunPrintsPlan(t *testing.T) {
	repoRoot := mustResolveUpgradeRepoRoot(t)
	fakeService := &fakeSourcePluginUpgradeService{
		statuses: []*hostsourceupgrade.SourcePluginStatus{
			{
				PluginID:          "plugin-source-dry-run",
				Name:              "Dry Run Plugin",
				EffectiveVersion:  "v0.1.0",
				DiscoveredVersion: "v0.5.0",
				Installed:         hostsourceupgrade.SourcePluginInstalledYes,
				NeedsUpgrade:      true,
			},
		},
	}

	restore := swapSourcePluginUpgradeServiceFactory(fakeService)
	defer restore()

	output := captureUpgradeStdout(t, func() {
		err := runSourcePluginUpgrade(context.Background(), repoRoot, cliOptions{
			scope:  upgradeScopeSourcePlugin,
			plugin: "plugin-source-dry-run",
			dryRun: true,
		})
		if err != nil {
			t.Fatalf("expected dry-run source plugin upgrade to succeed, got error: %v", err)
		}
	})

	if fakeService.upgradeCallCount != 0 {
		t.Fatalf("expected dry-run to skip execution, got %d upgrade calls", fakeService.upgradeCallCount)
	}
	if !strings.Contains(output, "Source plugin upgrade plan") ||
		!strings.Contains(output, "plugin-source-dry-run") ||
		!strings.Contains(output, "dry-run") {
		t.Fatalf("expected dry-run output to include plan details, got %q", output)
	}
}

// TestRunSourcePluginUpgradeAllSkipsUninstalledPlugins verifies plugin=all
// executes only pending installed source plugins and skips uninstalled ones.
func TestRunSourcePluginUpgradeAllSkipsUninstalledPlugins(t *testing.T) {
	repoRoot := mustResolveUpgradeRepoRoot(t)
	fakeService := &fakeSourcePluginUpgradeService{
		statuses: []*hostsourceupgrade.SourcePluginStatus{
			{
				PluginID:          "plugin-source-upgrade-all",
				Name:              "Upgrade All Plugin",
				EffectiveVersion:  "v0.1.0",
				DiscoveredVersion: "v0.5.0",
				Installed:         hostsourceupgrade.SourcePluginInstalledYes,
				NeedsUpgrade:      true,
			},
			{
				PluginID:          "plugin-source-not-installed",
				Name:              "Not Installed Plugin",
				EffectiveVersion:  "",
				DiscoveredVersion: "v0.2.0",
				Installed:         hostsourceupgrade.SourcePluginInstalledNo,
			},
		},
		results: map[string]*hostsourceupgrade.SourcePluginUpgradeResult{
			"plugin-source-upgrade-all": {
				PluginID:    "plugin-source-upgrade-all",
				FromVersion: "v0.1.0",
				ToVersion:   "v0.5.0",
				Executed:    true,
			},
		},
	}

	restore := swapSourcePluginUpgradeServiceFactory(fakeService)
	defer restore()

	output := captureUpgradeStdout(t, func() {
		err := runSourcePluginUpgrade(context.Background(), repoRoot, cliOptions{
			scope:  upgradeScopeSourcePlugin,
			plugin: "all",
		})
		if err != nil {
			t.Fatalf("expected plugin=all source plugin upgrade to succeed, got error: %v", err)
		}
	})

	if fakeService.upgradeCallCount != 1 {
		t.Fatalf("expected only one installed plugin to execute, got %d upgrade calls", fakeService.upgradeCallCount)
	}
	if len(fakeService.upgradeCalls) != 1 || fakeService.upgradeCalls[0] != "plugin-source-upgrade-all" {
		t.Fatalf("expected only plugin-source-upgrade-all to execute, got %#v", fakeService.upgradeCalls)
	}
	if !strings.Contains(output, "plugin-source-not-installed") || !strings.Contains(output, "未安装，跳过") {
		t.Fatalf("expected output to mention skipped uninstalled plugin, got %q", output)
	}
}

// TestRunSourcePluginUpgradeRejectsDowngrade verifies the command rejects
// discovered source versions that are lower than the effective registry version.
func TestRunSourcePluginUpgradeRejectsDowngrade(t *testing.T) {
	repoRoot := mustResolveUpgradeRepoRoot(t)
	fakeService := &fakeSourcePluginUpgradeService{
		statuses: []*hostsourceupgrade.SourcePluginStatus{
			{
				PluginID:          "plugin-source-downgrade",
				Name:              "Downgrade Plugin",
				EffectiveVersion:  "v0.5.0",
				DiscoveredVersion: "v0.1.0",
				Installed:         hostsourceupgrade.SourcePluginInstalledYes,
				DowngradeDetected: true,
			},
		},
	}

	restore := swapSourcePluginUpgradeServiceFactory(fakeService)
	defer restore()

	err := runSourcePluginUpgrade(context.Background(), repoRoot, cliOptions{
		scope:  upgradeScopeSourcePlugin,
		plugin: "plugin-source-downgrade",
	})
	if err == nil {
		t.Fatal("expected source plugin downgrade to be rejected")
	}
	if !strings.Contains(err.Error(), "不支持降级或回滚") {
		t.Fatalf("expected downgrade rejection error, got %q", err.Error())
	}
	if fakeService.upgradeCallCount != 0 {
		t.Fatalf("expected downgrade rejection to skip execution, got %d upgrade calls", fakeService.upgradeCallCount)
	}
}

// fakeSourcePluginUpgradeService is a deterministic test double for the source-plugin upgrade CLI.
type fakeSourcePluginUpgradeService struct {
	// statuses is the source-plugin status list returned to the command.
	statuses []*hostsourceupgrade.SourcePluginStatus
	// results stores per-plugin upgrade execution responses.
	results map[string]*hostsourceupgrade.SourcePluginUpgradeResult
	// upgradeCalls records the plugin IDs passed to UpgradeSourcePlugin.
	upgradeCalls []string
	// upgradeCallCount stores the number of UpgradeSourcePlugin invocations.
	upgradeCallCount int
}

// ListSourcePluginStatuses returns the configured status list for CLI tests.
func (s *fakeSourcePluginUpgradeService) ListSourcePluginStatuses(ctx context.Context) ([]*hostsourceupgrade.SourcePluginStatus, error) {
	return s.statuses, nil
}

// UpgradeSourcePlugin records the plugin ID and returns the configured fake result.
func (s *fakeSourcePluginUpgradeService) UpgradeSourcePlugin(ctx context.Context, pluginID string) (*hostsourceupgrade.SourcePluginUpgradeResult, error) {
	s.upgradeCalls = append(s.upgradeCalls, pluginID)
	s.upgradeCallCount++
	if result, ok := s.results[pluginID]; ok {
		return result, nil
	}
	return &hostsourceupgrade.SourcePluginUpgradeResult{PluginID: pluginID}, nil
}

// ValidateSourcePluginUpgradeReadiness is unused by these CLI tests.
func (s *fakeSourcePluginUpgradeService) ValidateSourcePluginUpgradeReadiness(ctx context.Context) error {
	return nil
}

// swapSourcePluginUpgradeServiceFactory overrides the service factory for one test and returns a restore callback.
func swapSourcePluginUpgradeServiceFactory(fake hostsourceupgrade.Service) func() {
	previous := newSourcePluginUpgradeService
	newSourcePluginUpgradeService = func() hostsourceupgrade.Service {
		return fake
	}
	return func() {
		newSourcePluginUpgradeService = previous
	}
}

// captureUpgradeStdout captures stdout generated while fn executes.
func captureUpgradeStdout(t *testing.T, fn func()) string {
	t.Helper()

	previousStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stdout pipe: %v", err)
	}

	os.Stdout = writer
	defer func() {
		os.Stdout = previousStdout
	}()

	type capturedOutput struct {
		output string
		err    error
	}

	outputCh := make(chan capturedOutput, 1)
	go func() {
		var buffer bytes.Buffer
		_, copyErr := io.Copy(&buffer, reader)
		closeErr := reader.Close()
		if copyErr != nil && closeErr != nil {
			copyErr = errors.Join(copyErr, closeErr)
		} else if copyErr == nil {
			copyErr = closeErr
		}
		outputCh <- capturedOutput{
			output: buffer.String(),
			err:    copyErr,
		}
	}()

	fn()

	if err = writer.Close(); err != nil {
		t.Fatalf("close stdout writer: %v", err)
	}
	result := <-outputCh
	if result.err != nil {
		t.Fatalf("copy stdout: %v", result.err)
	}
	return result.output
}

// mustResolveUpgradeRepoRoot resolves the repository root from this test file location.
func mustResolveUpgradeRepoRoot(t *testing.T) string {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve current file path failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", "..", "..", ".."))
}
