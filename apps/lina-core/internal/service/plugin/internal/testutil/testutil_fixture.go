// This file creates source and repository-backed plugin fixture directories.

package testutil

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/runtime"
	"lina-core/pkg/pluginbridge"
	"lina-core/pkg/pluginhost"
)

// CreateTestPluginDir creates a source plugin directory with the default file layout.
func CreateTestPluginDir(t *testing.T, pluginID string) string {
	t.Helper()

	pluginDir := filepath.Join(testSourcePluginRootDir, pluginID)
	if err := os.MkdirAll(filepath.Join(pluginDir, "backend"), 0o755); err != nil {
		t.Fatalf("failed to create backend dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(pluginDir, "frontend", "pages"), 0o755); err != nil {
		t.Fatalf("failed to create frontend pages dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(pluginDir, "manifest", "sql", "uninstall"), 0o755); err != nil {
		t.Fatalf("failed to create sql dir: %v", err)
	}

	t.Cleanup(func() {
		if cleanupErr := os.RemoveAll(pluginDir); cleanupErr != nil && !os.IsNotExist(cleanupErr) {
			t.Fatalf("failed to remove test plugin dir %s: %v", pluginDir, cleanupErr)
		}
	})

	WriteTestFile(t, filepath.Join(pluginDir, "go.mod"), "module "+strings.ReplaceAll(pluginID, "-", "_")+"\n\ngo 1.25.0\n")
	WriteTestFile(t, filepath.Join(pluginDir, "backend", "plugin.go"), "package backend\n")
	WriteTestFile(t, filepath.Join(pluginDir, "frontend", "pages", "main-entry.vue"), "<template><div /></template>\n")
	WriteTestFile(t, filepath.Join(pluginDir, "manifest", "sql", "001-"+pluginID+".sql"), "SELECT 1;\n")
	WriteTestFile(t, filepath.Join(pluginDir, "manifest", "sql", "uninstall", "001-"+pluginID+".sql"), "SELECT 1;\n")
	WriteTestFile(
		t,
		filepath.Join(pluginDir, "plugin.yaml"),
		"id: "+pluginID+"\nname: test\nversion: 0.1.0\ntype: source\nscope_nature: tenant_aware\nsupports_multi_tenant: true\ndefault_install_mode: tenant_scoped\n",
	)

	sourcePlugin := pluginhost.NewSourcePlugin(pluginID)
	sourcePlugin.Assets().UseEmbeddedFiles(os.DirFS(pluginDir))
	if err := sourcePlugin.Cron().RegisterCron(
		pluginhost.ExtensionPointCronRegister,
		pluginhost.CallbackExecutionModeBlocking,
		func(ctx context.Context, registrar pluginhost.CronRegistrar) error {
			hostServices := registrar.HostServices()
			if hostServices == nil || hostServices.Config() == nil {
				return gerror.New("test source plugin cron requires host config service")
			}
			if _, err := hostServices.Config().Exists(ctx, "monitor.interval"); err != nil {
				return err
			}
			return registrar.AddWithMetadata(
				ctx,
				"# * * * * *",
				pluginID+"-test-source-fixture-cron",
				"Test Source Fixture Cron",
				"Verifies source-plugin cron collection receives host services.",
				func(ctx context.Context) error {
					return nil
				},
			)
		},
	); err != nil {
		t.Fatalf("failed to register source plugin fixture cron %s: %v", pluginID, err)
	}
	cleanup, err := pluginhost.RegisterSourcePluginForTest(sourcePlugin)
	if err != nil {
		t.Fatalf("failed to register source plugin fixture %s: %v", pluginID, err)
	}
	t.Cleanup(cleanup)

	return pluginDir
}

// CreateTestRuntimePluginDir creates a runtime plugin source directory with a default frontend bundle.
func CreateTestRuntimePluginDir(
	t *testing.T,
	pluginID string,
	pluginName string,
	version string,
	installSQLAssets []*catalog.ArtifactSQLAsset,
	uninstallSQLAssets []*catalog.ArtifactSQLAsset,
) string {
	return CreateTestRuntimePluginDirWithFrontendAssets(
		t,
		pluginID,
		pluginName,
		version,
		DefaultTestRuntimeFrontendAssets(),
		installSQLAssets,
		uninstallSQLAssets,
	)
}

// CreateTestRuntimePluginDirWithFrontendAssets creates a runtime plugin source directory with one embedded artifact.
func CreateTestRuntimePluginDirWithFrontendAssets(
	t *testing.T,
	pluginID string,
	pluginName string,
	version string,
	frontendAssets []*catalog.ArtifactFrontendAsset,
	installSQLAssets []*catalog.ArtifactSQLAsset,
	uninstallSQLAssets []*catalog.ArtifactSQLAsset,
) string {
	t.Helper()

	repoRoot, err := FindRepoRoot(".")
	if err != nil {
		t.Fatalf("failed to resolve repo root: %v", err)
	}

	pluginDir := filepath.Join(repoRoot, "apps", "lina-plugins", pluginID)
	if err = os.MkdirAll(filepath.Join(pluginDir, "runtime"), 0o755); err != nil {
		t.Fatalf("failed to create runtime dir: %v", err)
	}

	t.Cleanup(func() {
		if cleanupErr := os.RemoveAll(pluginDir); cleanupErr != nil && !os.IsNotExist(cleanupErr) {
			t.Fatalf("failed to remove runtime test plugin dir %s: %v", pluginDir, cleanupErr)
		}
	})

	WriteTestFile(
		t,
		filepath.Join(pluginDir, "plugin.yaml"),
		"id: "+pluginID+"\nname: "+pluginName+"\nversion: "+version+"\ntype: dynamic\nscope_nature: tenant_aware\nsupports_multi_tenant: true\ndefault_install_mode: tenant_scoped\n",
	)
	supportsMultiTenant := true
	WriteRuntimeWasmArtifact(
		t,
		filepath.Join(pluginDir, runtime.BuildArtifactRelativePath(pluginID)),
		&catalog.ArtifactManifest{
			ID:                  pluginID,
			Name:                pluginName,
			Version:             version,
			Type:                catalog.TypeDynamic.String(),
			ScopeNature:         catalog.ScopeNatureTenantAware.String(),
			SupportsMultiTenant: &supportsMultiTenant,
			DefaultInstallMode:  catalog.InstallModeTenantScoped.String(),
		},
		&catalog.ArtifactSpec{
			RuntimeKind:        pluginbridge.RuntimeKindWasm,
			ABIVersion:         pluginbridge.SupportedABIVersion,
			FrontendAssetCount: len(frontendAssets),
			SQLAssetCount:      len(installSQLAssets) + len(uninstallSQLAssets),
		},
		frontendAssets,
		installSQLAssets,
		uninstallSQLAssets,
		nil,
		nil,
		nil,
	)
	return pluginDir
}

// WriteTestFile writes one UTF-8 fixture file to disk for the current test.
func WriteTestFile(t *testing.T, filePath string, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		t.Fatalf("failed to create test file dir %s: %v", filePath, err)
	}
	tempFile, err := os.CreateTemp(filepath.Dir(filePath), filepath.Base(filePath)+".tmp-*")
	if err != nil {
		t.Fatalf("failed to create temp test file %s: %v", filePath, err)
	}
	tempPath := tempFile.Name()
	defer func() {
		if cleanupErr := os.Remove(tempPath); cleanupErr != nil && !os.IsNotExist(cleanupErr) {
			t.Fatalf("failed to remove temp test file %s: %v", tempPath, cleanupErr)
		}
	}()

	if _, err = tempFile.Write([]byte(content)); err != nil {
		if closeErr := tempFile.Close(); closeErr != nil {
			t.Fatalf("failed to close temp test file %s after write error: %v", filePath, closeErr)
		}
		t.Fatalf("failed to write temp test file %s: %v", filePath, err)
	}
	if err = tempFile.Chmod(0o644); err != nil {
		if closeErr := tempFile.Close(); closeErr != nil {
			t.Fatalf("failed to close temp test file %s after chmod error: %v", filePath, closeErr)
		}
		t.Fatalf("failed to chmod temp test file %s: %v", filePath, err)
	}
	if err = tempFile.Close(); err != nil {
		t.Fatalf("failed to close temp test file %s: %v", filePath, err)
	}
	if err = os.Rename(tempPath, filePath); err != nil {
		t.Fatalf("failed to move test file into place %s: %v", filePath, err)
	}
}
