// Package testutil provides shared helpers for plugin sub-component tests.
package testutil

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"gopkg.in/yaml.v3"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/bizctx"
	configsvc "lina-core/internal/service/config"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/frontend"
	"lina-core/internal/service/plugin/internal/integration"
	"lina-core/internal/service/plugin/internal/lifecycle"
	"lina-core/internal/service/plugin/internal/openapi"
	"lina-core/internal/service/plugin/internal/runtime"
	"lina-core/pkg/pluginbridge"
)

var (
	bundledRuntimeSampleOnce sync.Once
	bundledRuntimeSampleErr  error
	testDynamicStorageDir    string
)

func init() {
	var err error
	testDynamicStorageDir, err = os.MkdirTemp("", "lina-plugin-dynamic-storage-*")
	if err != nil {
		panic(fmt.Sprintf("failed to create isolated dynamic storage dir: %v", err))
	}
	configsvc.SetPluginDynamicStoragePathOverride(testDynamicStorageDir)
}

// Services groups the wired plugin sub-services used by package-level tests.
type Services struct {
	// Catalog provides manifest discovery, registry, and release access.
	Catalog catalog.Service
	// Lifecycle provides install and uninstall orchestration.
	Lifecycle lifecycle.Service
	// Runtime provides artifact parsing, reconcile, and route execution.
	Runtime runtime.Service
	// Frontend provides in-memory frontend bundle management.
	Frontend frontend.Service
	// Integration provides menu, hook, and resource-ref integration.
	Integration integration.Service
	// OpenAPI provides dynamic route OpenAPI projection.
	OpenAPI openapi.Service
}

// RuntimeBuildOutput describes one artifact produced by the build-wasm helper in tests.
type RuntimeBuildOutput struct {
	// ArtifactPath is the on-disk path of the produced wasm artifact.
	ArtifactPath string
	// Content is the artifact byte content.
	Content []byte
}

type singleNodeTopology struct{}

func (singleNodeTopology) IsClusterModeEnabled() bool {
	return false
}

func (singleNodeTopology) IsPrimaryNode() bool {
	return true
}

func (singleNodeTopology) CurrentNodeID() string {
	return "test-node"
}

// NewServices creates a fully wired plugin sub-service set for tests.
func NewServices() *Services {
	var (
		configProvider = configsvc.New()
		bizCtxProvider = bizctx.New()
		catalogSvc     = catalog.New(configProvider)
		lifecycleSvc   = lifecycle.New(catalogSvc)
		frontendSvc    = frontend.New(catalogSvc)
		openapiSvc     = openapi.New(catalogSvc)
		runtimeSvc     = runtime.New(catalogSvc, lifecycleSvc, frontendSvc, openapiSvc)
		integrationSvc = integration.New(catalogSvc)
		topology       = singleNodeTopology{}
	)

	catalogSvc.SetBackendLoader(integrationSvc)
	catalogSvc.SetArtifactParser(runtimeSvc)
	catalogSvc.SetDynamicManifestLoader(runtimeSvc)
	catalogSvc.SetNodeStateSyncer(runtimeSvc)
	catalogSvc.SetMenuSyncer(integrationSvc)
	catalogSvc.SetResourceRefSyncer(integrationSvc)
	catalogSvc.SetReleaseStateSyncer(runtimeSvc)
	catalogSvc.SetHookDispatcher(integrationSvc)

	lifecycleSvc.SetReconciler(runtimeSvc)
	lifecycleSvc.SetTopology(topology)

	integrationSvc.SetBizCtxProvider(&bizCtxAdapter{svc: bizCtxProvider})
	integrationSvc.SetTopologyProvider(topology)

	runtimeSvc.SetMenuManager(integrationSvc)
	runtimeSvc.SetHookDispatcher(integrationSvc)
	runtimeSvc.SetAfterAuthDispatcher(integrationSvc)
	runtimeSvc.SetPermissionMenuFilter(integrationSvc)
	runtimeSvc.SetJwtConfigProvider(&jwtConfigAdapter{svc: configProvider})
	runtimeSvc.SetUserContextSetter(&userCtxAdapter{svc: bizCtxProvider})
	runtimeSvc.SetTopology(topology)

	return &Services{
		Catalog:     catalogSvc,
		Lifecycle:   lifecycleSvc,
		Runtime:     runtimeSvc,
		Frontend:    frontendSvc,
		Integration: integrationSvc,
		OpenAPI:     openapiSvc,
	}
}

// TestDynamicStorageDir returns the process-local runtime storage directory for plugin tests.
func TestDynamicStorageDir() string {
	return testDynamicStorageDir
}

type jwtConfigAdapter struct {
	svc configsvc.Service
}

func (a *jwtConfigAdapter) GetJwtSecret(ctx context.Context) string {
	return a.svc.GetJwt(ctx).Secret
}

type userCtxAdapter struct {
	svc bizctx.Service
}

func (a *userCtxAdapter) SetUser(ctx context.Context, tokenID string, userID int, username string, status int) {
	a.svc.SetUser(ctx, tokenID, userID, username, status)
}

type bizCtxAdapter struct {
	svc bizctx.Service
}

func (a *bizCtxAdapter) GetUserId(ctx context.Context) int {
	localCtx := a.svc.Get(ctx)
	if localCtx == nil {
		return 0
	}
	return localCtx.UserId
}

// FindRepoRoot walks up from startDir until it locates the repository root.
func FindRepoRoot(startDir string) (string, error) {
	abs, err := filepath.Abs(startDir)
	if err != nil {
		return "", err
	}
	dir := abs
	for {
		if _, statErr := os.Stat(filepath.Join(dir, "go.work")); statErr == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	if strings.Contains(abs, string(filepath.Separator)) {
		for dir = abs; ; dir = filepath.Dir(dir) {
			if _, statErr := os.Stat(filepath.Join(dir, "go.mod")); statErr == nil {
				if _, statErr2 := os.Stat(filepath.Join(filepath.Dir(dir), "go.work")); statErr2 == nil {
					return filepath.Dir(dir), nil
				}
			}
			if filepath.Dir(dir) == dir {
				break
			}
		}
	}
	return abs, nil
}

// EnsureBundledRuntimeSampleArtifactForTests builds the bundled dynamic sample once for package tests.
func EnsureBundledRuntimeSampleArtifactForTests(t *testing.T) {
	t.Helper()

	bundledRuntimeSampleOnce.Do(func() {
		repoRoot, err := FindRepoRoot(".")
		if err != nil {
			bundledRuntimeSampleErr = err
			return
		}

		pluginDir := filepath.Join(repoRoot, "apps", "lina-plugins", "plugin-demo-dynamic")
		if _, statErr := os.Stat(filepath.Join(pluginDir, "plugin.yaml")); statErr != nil {
			if os.IsNotExist(statErr) {
				return
			}
			bundledRuntimeSampleErr = statErr
			return
		}

		builderDir := filepath.Join(repoRoot, "hack", "build-wasm")
		cmd := exec.Command(
			"go",
			"run",
			".",
			"--plugin-dir",
			pluginDir,
			"--output-dir",
			testDynamicStorageDir,
		)
		cmd.Dir = builderDir
		cmd.Env = append(os.Environ(), "GOWORK="+filepath.Join(repoRoot, "go.work"))
		output, err := cmd.CombinedOutput()
		if err != nil {
			bundledRuntimeSampleErr = fmt.Errorf("run hack/build-wasm failed: %w output=%s", err, string(output))
		}
	})

	if bundledRuntimeSampleErr != nil {
		t.Fatalf("failed to prepare bundled dynamic sample: %v", bundledRuntimeSampleErr)
	}
}

// BuildRuntimeArtifactWithHackTool runs hack/build-wasm for one plugin source directory.
func BuildRuntimeArtifactWithHackTool(t *testing.T, pluginDir string) *RuntimeBuildOutput {
	t.Helper()

	repoRoot, err := FindRepoRoot(".")
	if err != nil {
		t.Fatalf("failed to resolve repo root: %v", err)
	}
	builderDir := filepath.Join(repoRoot, "hack", "build-wasm")
	outputDir := filepath.Join(t.TempDir(), "output")
	cmd := exec.Command("go", "run", ".", "--plugin-dir", pluginDir, "--output-dir", outputDir)
	cmd.Dir = builderDir
	cmd.Env = append(os.Environ(), "GOWORK="+filepath.Join(repoRoot, "go.work"))
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to run hack/build-wasm: %v output=%s", err, string(output))
	}

	type manifestIDHolder struct {
		ID string `yaml:"id"`
	}
	manifestContent, err := os.ReadFile(filepath.Join(pluginDir, "plugin.yaml"))
	if err != nil {
		t.Fatalf("failed to read plugin.yaml: %v", err)
	}
	var holder manifestIDHolder
	if err = yaml.Unmarshal(manifestContent, &holder); err != nil {
		t.Fatalf("failed to parse plugin.yaml: %v", err)
	}
	artifactPath := filepath.Join(outputDir, holder.ID+".wasm")
	content, err := os.ReadFile(artifactPath)
	if err != nil {
		t.Fatalf("failed to read built artifact: %v", err)
	}
	return &RuntimeBuildOutput{
		ArtifactPath: artifactPath,
		Content:      content,
	}
}

// CreateTestPluginDir creates a source plugin directory with the default file layout.
func CreateTestPluginDir(t *testing.T, pluginID string) string {
	t.Helper()

	repoRoot, err := FindRepoRoot(".")
	if err != nil {
		t.Fatalf("failed to resolve repo root: %v", err)
	}

	pluginDir := filepath.Join(repoRoot, "apps", "lina-plugins", pluginID)
	if err = os.MkdirAll(filepath.Join(pluginDir, "backend"), 0o755); err != nil {
		t.Fatalf("failed to create backend dir: %v", err)
	}
	if err = os.MkdirAll(filepath.Join(pluginDir, "frontend", "pages"), 0o755); err != nil {
		t.Fatalf("failed to create frontend pages dir: %v", err)
	}
	if err = os.MkdirAll(filepath.Join(pluginDir, "manifest", "sql", "uninstall"), 0o755); err != nil {
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
	WriteTestFile(t, filepath.Join(pluginDir, "plugin.yaml"), "id: "+pluginID+"\nname: test\nversion: 0.1.0\ntype: source\n")

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
		"id: "+pluginID+"\nname: "+pluginName+"\nversion: "+version+"\ntype: dynamic\n",
	)
	WriteRuntimeWasmArtifact(
		t,
		filepath.Join(pluginDir, runtime.BuildArtifactRelativePath(pluginID)),
		&catalog.ArtifactManifest{
			ID:      pluginID,
			Name:    pluginName,
			Version: version,
			Type:    catalog.TypeDynamic.String(),
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
	)
	return pluginDir
}

// CreateTestRuntimeStorageArtifact creates one runtime artifact in the isolated test storage directory.
func CreateTestRuntimeStorageArtifact(
	t *testing.T,
	pluginID string,
	pluginName string,
	version string,
	installSQLAssets []*catalog.ArtifactSQLAsset,
	uninstallSQLAssets []*catalog.ArtifactSQLAsset,
) string {
	return CreateTestRuntimeStorageArtifactWithFrontendAssets(
		t,
		pluginID,
		pluginName,
		version,
		DefaultTestRuntimeFrontendAssets(),
		installSQLAssets,
		uninstallSQLAssets,
	)
}

// CreateTestRuntimeStorageArtifactWithFilename creates one runtime artifact with a custom storage file name.
// This is the low-level variant used when the test needs to place two artifacts with the same plugin ID
// under different file names in order to exercise duplicate-detection logic.
func CreateTestRuntimeStorageArtifactWithFilename(
	t *testing.T,
	fileName string,
	pluginID string,
	pluginName string,
	version string,
	installSQLAssets []*catalog.ArtifactSQLAsset,
	uninstallSQLAssets []*catalog.ArtifactSQLAsset,
) string {
	t.Helper()

	storageDir := testDynamicStorageDir
	if err := os.MkdirAll(storageDir, 0o755); err != nil {
		t.Fatalf("failed to create dynamic storage dir: %v", err)
	}

	artifactPath := filepath.Join(storageDir, fileName)
	t.Cleanup(func() {
		if cleanupErr := os.Remove(artifactPath); cleanupErr != nil && !os.IsNotExist(cleanupErr) {
			t.Fatalf("failed to remove runtime storage artifact %s: %v", artifactPath, cleanupErr)
		}
	})

	WriteRuntimeWasmArtifact(
		t,
		artifactPath,
		&catalog.ArtifactManifest{
			ID:      pluginID,
			Name:    pluginName,
			Version: version,
			Type:    catalog.TypeDynamic.String(),
		},
		&catalog.ArtifactSpec{
			RuntimeKind:        pluginbridge.RuntimeKindWasm,
			ABIVersion:         pluginbridge.SupportedABIVersion,
			FrontendAssetCount: len(DefaultTestRuntimeFrontendAssets()),
			SQLAssetCount:      len(installSQLAssets) + len(uninstallSQLAssets),
		},
		DefaultTestRuntimeFrontendAssets(),
		installSQLAssets,
		uninstallSQLAssets,
		nil,
		nil,
	)
	return artifactPath
}

// BuildTestRuntimeWasmContent assembles a synthetic WASM artifact byte slice for use in tests
// that need the raw bytes (e.g. upload-path tests that call StoreUploadedPackage directly).
func BuildTestRuntimeWasmContent(
	t *testing.T,
	manifest *catalog.ArtifactManifest,
	runtimeMetadata *catalog.ArtifactSpec,
	frontendAssets []*catalog.ArtifactFrontendAsset,
	installSQLAssets []*catalog.ArtifactSQLAsset,
	uninstallSQLAssets []*catalog.ArtifactSQLAsset,
	routeContracts []*pluginbridge.RouteContract,
	bridgeSpec *pluginbridge.BridgeSpec,
) []byte {
	t.Helper()
	return buildTestRuntimeWasmArtifactContent(t, manifest, runtimeMetadata, frontendAssets, installSQLAssets, uninstallSQLAssets, routeContracts, bridgeSpec)
}

// CreateTestRuntimeStorageArtifactWithFrontendAssets creates one runtime artifact with custom frontend assets.
func CreateTestRuntimeStorageArtifactWithFrontendAssets(
	t *testing.T,
	pluginID string,
	pluginName string,
	version string,
	frontendAssets []*catalog.ArtifactFrontendAsset,
	installSQLAssets []*catalog.ArtifactSQLAsset,
	uninstallSQLAssets []*catalog.ArtifactSQLAsset,
) string {
	return CreateTestRuntimeStorageArtifactWithFrontendAssetsAndBackendContracts(
		t,
		pluginID,
		pluginName,
		version,
		frontendAssets,
		installSQLAssets,
		uninstallSQLAssets,
		nil,
		nil,
	)
}

// CreateTestRuntimeStorageArtifactWithMenus creates one runtime artifact with manifest menus.
func CreateTestRuntimeStorageArtifactWithMenus(
	t *testing.T,
	pluginID string,
	pluginName string,
	version string,
	menus []*catalog.MenuSpec,
	installSQLAssets []*catalog.ArtifactSQLAsset,
	uninstallSQLAssets []*catalog.ArtifactSQLAsset,
) string {
	t.Helper()

	storageDir := testDynamicStorageDir
	if err := os.MkdirAll(storageDir, 0o755); err != nil {
		t.Fatalf("failed to create dynamic storage dir: %v", err)
	}

	artifactPath := filepath.Join(storageDir, runtime.BuildArtifactFileName(pluginID))
	t.Cleanup(func() {
		if cleanupErr := os.Remove(artifactPath); cleanupErr != nil && !os.IsNotExist(cleanupErr) {
			t.Fatalf("failed to remove runtime menu artifact %s: %v", artifactPath, cleanupErr)
		}
	})

	WriteRuntimeWasmArtifact(
		t,
		artifactPath,
		&catalog.ArtifactManifest{
			ID:      pluginID,
			Name:    pluginName,
			Version: version,
			Type:    catalog.TypeDynamic.String(),
			Menus:   menus,
		},
		&catalog.ArtifactSpec{
			RuntimeKind:        pluginbridge.RuntimeKindWasm,
			ABIVersion:         pluginbridge.SupportedABIVersion,
			FrontendAssetCount: len(DefaultTestRuntimeFrontendAssets()),
			SQLAssetCount:      len(installSQLAssets) + len(uninstallSQLAssets),
		},
		DefaultTestRuntimeFrontendAssets(),
		installSQLAssets,
		uninstallSQLAssets,
		nil,
		nil,
	)
	return artifactPath
}

// CreateTestRuntimeStorageArtifactWithFrontendAssetsAndBackendContracts creates one runtime artifact with full contract sections.
func CreateTestRuntimeStorageArtifactWithFrontendAssetsAndBackendContracts(
	t *testing.T,
	pluginID string,
	pluginName string,
	version string,
	frontendAssets []*catalog.ArtifactFrontendAsset,
	installSQLAssets []*catalog.ArtifactSQLAsset,
	uninstallSQLAssets []*catalog.ArtifactSQLAsset,
	routeContracts []*pluginbridge.RouteContract,
	bridgeSpec *pluginbridge.BridgeSpec,
) string {
	t.Helper()

	storageDir := testDynamicStorageDir
	if err := os.MkdirAll(storageDir, 0o755); err != nil {
		t.Fatalf("failed to create dynamic storage dir: %v", err)
	}

	artifactPath := filepath.Join(storageDir, runtime.BuildArtifactFileName(pluginID))
	t.Cleanup(func() {
		if cleanupErr := os.Remove(artifactPath); cleanupErr != nil && !os.IsNotExist(cleanupErr) {
			t.Fatalf("failed to remove runtime contract artifact %s: %v", artifactPath, cleanupErr)
		}
	})

	WriteRuntimeWasmArtifact(
		t,
		artifactPath,
		&catalog.ArtifactManifest{
			ID:      pluginID,
			Name:    pluginName,
			Version: version,
			Type:    catalog.TypeDynamic.String(),
		},
		&catalog.ArtifactSpec{
			RuntimeKind:        pluginbridge.RuntimeKindWasm,
			ABIVersion:         pluginbridge.SupportedABIVersion,
			FrontendAssetCount: len(frontendAssets),
			SQLAssetCount:      len(installSQLAssets) + len(uninstallSQLAssets),
			RouteCount:         len(routeContracts),
		},
		frontendAssets,
		installSQLAssets,
		uninstallSQLAssets,
		routeContracts,
		bridgeSpec,
	)
	return artifactPath
}

// DefaultTestRuntimeFrontendAssets returns the default frontend assets used by runtime artifact fixtures.
func DefaultTestRuntimeFrontendAssets() []*catalog.ArtifactFrontendAsset {
	return []*catalog.ArtifactFrontendAsset{
		{
			Path:          "index.html",
			ContentBase64: base64.StdEncoding.EncodeToString([]byte("<html><body>dynamic frontend</body></html>")),
			ContentType:   "text/html; charset=utf-8",
		},
		{
			Path:          "assets/app.js",
			ContentBase64: base64.StdEncoding.EncodeToString([]byte("console.log('dynamic frontend');")),
			ContentType:   "application/javascript",
		},
	}
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

// WriteRuntimeWasmArtifact writes one synthetic runtime WASM artifact fixture to disk.
func WriteRuntimeWasmArtifact(
	t *testing.T,
	filePath string,
	manifest *catalog.ArtifactManifest,
	runtimeMetadata *catalog.ArtifactSpec,
	frontendAssets []*catalog.ArtifactFrontendAsset,
	installSQLAssets []*catalog.ArtifactSQLAsset,
	uninstallSQLAssets []*catalog.ArtifactSQLAsset,
	routeContracts []*pluginbridge.RouteContract,
	bridgeSpec *pluginbridge.BridgeSpec,
) {
	t.Helper()

	wasm := buildTestRuntimeWasmArtifactContent(
		t,
		manifest,
		runtimeMetadata,
		frontendAssets,
		installSQLAssets,
		uninstallSQLAssets,
		routeContracts,
		bridgeSpec,
	)
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		t.Fatalf("failed to create runtime artifact dir %s: %v", filePath, err)
	}
	tempFile, err := os.CreateTemp(filepath.Dir(filePath), filepath.Base(filePath)+".tmp-*")
	if err != nil {
		t.Fatalf("failed to create temp runtime wasm artifact %s: %v", filePath, err)
	}
	tempPath := tempFile.Name()
	defer func() {
		if cleanupErr := os.Remove(tempPath); cleanupErr != nil && !os.IsNotExist(cleanupErr) {
			t.Fatalf("failed to remove temp runtime wasm artifact %s: %v", tempPath, cleanupErr)
		}
	}()

	if _, err = tempFile.Write(wasm); err != nil {
		if closeErr := tempFile.Close(); closeErr != nil {
			t.Fatalf("failed to close temp runtime wasm artifact %s after write error: %v", filePath, closeErr)
		}
		t.Fatalf("failed to write temp runtime wasm artifact %s: %v", filePath, err)
	}
	if err = tempFile.Chmod(0o644); err != nil {
		if closeErr := tempFile.Close(); closeErr != nil {
			t.Fatalf("failed to close temp runtime wasm artifact %s after chmod error: %v", filePath, closeErr)
		}
		t.Fatalf("failed to chmod temp runtime wasm artifact %s: %v", filePath, err)
	}
	if err = tempFile.Close(); err != nil {
		t.Fatalf("failed to close temp runtime wasm artifact %s: %v", filePath, err)
	}
	if err = os.Rename(tempPath, filePath); err != nil {
		t.Fatalf("failed to move runtime wasm artifact into place %s: %v", filePath, err)
	}
}

// CleanupPluginGovernanceRowsHard removes plugin governance records created during tests.
func CleanupPluginGovernanceRowsHard(t *testing.T, ctx context.Context, pluginID string) {
	t.Helper()

	if _, err := dao.SysPluginNodeState.Ctx(ctx).
		Unscoped().
		Where(do.SysPluginNodeState{PluginId: pluginID}).
		Delete(); err != nil {
		t.Fatalf("failed to delete sys_plugin_node_state rows for %s: %v", pluginID, err)
	}
	if _, err := dao.SysPluginResourceRef.Ctx(ctx).
		Unscoped().
		Where(do.SysPluginResourceRef{PluginId: pluginID}).
		Delete(); err != nil {
		t.Fatalf("failed to delete sys_plugin_resource_ref rows for %s: %v", pluginID, err)
	}
	if _, err := dao.SysPluginMigration.Ctx(ctx).
		Unscoped().
		Where(do.SysPluginMigration{PluginId: pluginID}).
		Delete(); err != nil {
		t.Fatalf("failed to delete sys_plugin_migration rows for %s: %v", pluginID, err)
	}
	if _, err := dao.SysPluginRelease.Ctx(ctx).
		Unscoped().
		Where(do.SysPluginRelease{PluginId: pluginID}).
		Delete(); err != nil {
		t.Fatalf("failed to delete sys_plugin_release rows for %s: %v", pluginID, err)
	}
	if _, err := dao.SysPlugin.Ctx(ctx).
		Unscoped().
		Where(do.SysPlugin{PluginId: pluginID}).
		Delete(); err != nil {
		t.Fatalf("failed to delete sys_plugin rows for %s: %v", pluginID, err)
	}
}

// CleanupPluginMenuRowsHard removes plugin-owned menu rows and admin bindings created during tests.
func CleanupPluginMenuRowsHard(t *testing.T, ctx context.Context, pluginID string) {
	t.Helper()

	services := NewServices()
	menus, err := services.Integration.ListPluginMenusByPlugin(ctx, pluginID)
	if err != nil {
		t.Fatalf("expected plugin menu cleanup query to succeed, got error: %v", err)
	}
	if len(menus) == 0 {
		return
	}

	menuIDs := make([]interface{}, 0, len(menus))
	menuKeys := make([]string, 0, len(menus))
	for _, item := range menus {
		if item == nil {
			continue
		}
		menuIDs = append(menuIDs, item.Id)
		menuKeys = append(menuKeys, item.MenuKey)
	}

	if len(menuIDs) > 0 {
		if _, err := dao.SysRoleMenu.Ctx(ctx).
			WhereIn(dao.SysRoleMenu.Columns().MenuId, menuIDs).
			Delete(); err != nil {
			t.Fatalf("failed to delete sys_role_menu rows for %s: %v", pluginID, err)
		}
	}
	if len(menuKeys) > 0 {
		if _, err := dao.SysMenu.Ctx(ctx).
			Unscoped().
			WhereIn(dao.SysMenu.Columns().MenuKey, menuKeys).
			Delete(); err != nil {
			t.Fatalf("failed to delete sys_menu rows for %s: %v", pluginID, err)
		}
	}
}

// QueryMenuByKey returns one sys_menu row by menu_key.
func QueryMenuByKey(ctx context.Context, menuKey string) (*entity.SysMenu, error) {
	var menu *entity.SysMenu
	err := dao.SysMenu.Ctx(ctx).
		Where(do.SysMenu{MenuKey: menuKey}).
		Scan(&menu)
	return menu, err
}

func buildTestRuntimeWasmArtifactContent(
	t *testing.T,
	manifest *catalog.ArtifactManifest,
	runtimeMetadata *catalog.ArtifactSpec,
	frontendAssets []*catalog.ArtifactFrontendAsset,
	installSQLAssets []*catalog.ArtifactSQLAsset,
	uninstallSQLAssets []*catalog.ArtifactSQLAsset,
	routeContracts []*pluginbridge.RouteContract,
	bridgeSpec *pluginbridge.BridgeSpec,
) []byte {
	t.Helper()

	manifestContent, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("failed to marshal dynamic manifest: %v", err)
	}
	runtimeContent, err := json.Marshal(runtimeMetadata)
	if err != nil {
		t.Fatalf("failed to marshal runtime metadata: %v", err)
	}

	wasm := []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}
	wasm = appendWasmCustomSection(wasm, pluginbridge.WasmSectionManifest, manifestContent)
	wasm = appendWasmCustomSection(wasm, pluginbridge.WasmSectionRuntime, runtimeContent)
	if len(frontendAssets) > 0 {
		frontendContent, marshalErr := json.Marshal(frontendAssets)
		if marshalErr != nil {
			t.Fatalf("failed to marshal frontend assets: %v", marshalErr)
		}
		wasm = appendWasmCustomSection(wasm, pluginbridge.WasmSectionFrontendAssets, frontendContent)
	}
	if len(installSQLAssets) > 0 {
		installContent, marshalErr := json.Marshal(installSQLAssets)
		if marshalErr != nil {
			t.Fatalf("failed to marshal install sql assets: %v", marshalErr)
		}
		wasm = appendWasmCustomSection(wasm, pluginbridge.WasmSectionInstallSQL, installContent)
	}
	if len(uninstallSQLAssets) > 0 {
		uninstallContent, marshalErr := json.Marshal(uninstallSQLAssets)
		if marshalErr != nil {
			t.Fatalf("failed to marshal uninstall sql assets: %v", marshalErr)
		}
		wasm = appendWasmCustomSection(wasm, pluginbridge.WasmSectionUninstallSQL, uninstallContent)
	}
	if len(routeContracts) > 0 {
		routeContent, marshalErr := json.Marshal(routeContracts)
		if marshalErr != nil {
			t.Fatalf("failed to marshal route contracts: %v", marshalErr)
		}
		wasm = appendWasmCustomSection(wasm, pluginbridge.WasmSectionBackendRoutes, routeContent)
	}
	if bridgeSpec != nil {
		bridgeContent, marshalErr := json.Marshal(bridgeSpec)
		if marshalErr != nil {
			t.Fatalf("failed to marshal bridge spec: %v", marshalErr)
		}
		wasm = appendWasmCustomSection(wasm, pluginbridge.WasmSectionBackendBridge, bridgeContent)
	}
	if len(runtimeMetadata.HostServices) > 0 {
		hostServiceContent, marshalErr := json.Marshal(runtimeMetadata.HostServices)
		if marshalErr != nil {
			t.Fatalf("failed to marshal runtime host services: %v", marshalErr)
		}
		wasm = appendWasmCustomSection(wasm, pluginbridge.WasmSectionBackendHostServices, hostServiceContent)
	}
	return wasm
}

func appendWasmCustomSection(content []byte, name string, payload []byte) []byte {
	sectionPayload := append([]byte{}, encodeWasmULEB128(uint32(len(name)))...)
	sectionPayload = append(sectionPayload, []byte(name)...)
	sectionPayload = append(sectionPayload, payload...)

	result := append([]byte{}, content...)
	result = append(result, 0x00)
	result = append(result, encodeWasmULEB128(uint32(len(sectionPayload)))...)
	result = append(result, sectionPayload...)
	return result
}

func encodeWasmULEB128(value uint32) []byte {
	result := make([]byte, 0, 5)
	for {
		current := byte(value & 0x7f)
		value >>= 7
		if value != 0 {
			current |= 0x80
		}
		result = append(result, current)
		if value == 0 {
			return result
		}
	}
}
