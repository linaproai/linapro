// This file keeps root-package test bootstrap and shared helpers for plugin facade tests.

package plugin

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"

	"lina-core/internal/model/entity"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/testutil"
	"lina-core/pkg/pluginbridge"
)

// TestMain keeps package-level tests self-contained by generating the bundled
// dynamic sample artifact before any test scans the shared plugin workspace.
func TestMain(m *testing.M) {
	if err := ensureBundledRuntimeSampleArtifactForTests(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to prepare bundled dynamic sample: %v\n", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func ensureBundledRuntimeSampleArtifactForTests() error {
	repoRoot, err := testutil.FindRepoRoot(".")
	if err != nil {
		return err
	}

	pluginDir := filepath.Join(repoRoot, "apps", "lina-plugins", "plugin-demo-dynamic")
	if _, statErr := os.Stat(filepath.Join(pluginDir, "plugin.yaml")); statErr != nil {
		if os.IsNotExist(statErr) {
			return nil
		}
		return statErr
	}

	builderDir := filepath.Join(repoRoot, "hack", "build-wasm")
	cmd := exec.Command(
		"go",
		"run",
		".",
		"--plugin-dir",
		pluginDir,
		"--output-dir",
		testutil.TestDynamicStorageDir(),
	)
	cmd.Dir = builderDir
	cmd.Env = append(os.Environ(), "GOWORK="+filepath.Join(repoRoot, "go.work"))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("run hack/build-wasm failed: %w: %s", err, string(output))
	}
	return nil
}

func newTestService() *serviceImpl {
	return New().(*serviceImpl)
}

func newTestServiceWithTopology(topology Topology) *serviceImpl {
	return New(topology).(*serviceImpl)
}

func (s *serviceImpl) getPluginRegistry(ctx context.Context, pluginID string) (*entity.SysPlugin, error) {
	return s.catalogSvc.GetRegistry(ctx, pluginID)
}

func (s *serviceImpl) getPluginRelease(ctx context.Context, pluginID string, version string) (*entity.SysPluginRelease, error) {
	return s.catalogSvc.GetRelease(ctx, pluginID, version)
}

func (s *serviceImpl) getActivePluginManifest(ctx context.Context, pluginID string) (*catalog.Manifest, error) {
	return s.catalogSvc.GetActiveManifest(ctx, pluginID)
}

func (s *serviceImpl) buildPluginGovernanceSnapshot(
	ctx context.Context,
	pluginID string,
	version string,
	pluginType string,
	installed int,
	enabled int,
) (*catalog.GovernanceSnapshot, error) {
	return s.catalogSvc.BuildGovernanceSnapshot(ctx, pluginID, version, pluginType, installed, enabled)
}

func (s *serviceImpl) loadRuntimePluginManifestFromArtifact(artifactPath string) (*catalog.Manifest, error) {
	return s.catalogSvc.LoadManifestFromArtifactPath(artifactPath)
}

func (s *serviceImpl) syncPluginManifest(ctx context.Context, manifest *catalog.Manifest) (*entity.SysPlugin, error) {
	return s.catalogSvc.SyncManifest(ctx, manifest)
}

func (s *serviceImpl) setPluginInstalled(ctx context.Context, pluginID string, installed int) error {
	return s.catalogSvc.SetPluginInstalled(ctx, pluginID, installed)
}

func (s *serviceImpl) setPluginStatus(ctx context.Context, pluginID string, status int) error {
	return s.catalogSvc.SetPluginStatus(ctx, pluginID, status)
}

func (s *serviceImpl) executeDynamicRoute(ctx context.Context, manifest *catalog.Manifest, request *pluginbridge.BridgeRequestEnvelopeV1) (*pluginbridge.BridgeResponseEnvelopeV1, error) {
	return s.runtimeSvc.ExecuteDynamicRoute(ctx, manifest, request)
}

type testTopology struct {
	enabled bool
	primary bool
	nodeID  string
}

func (t *testTopology) IsEnabled() bool {
	return t != nil && t.enabled
}

func (t *testTopology) IsPrimary() bool {
	if t == nil {
		return true
	}
	return t.primary
}

func (t *testTopology) NodeID() string {
	if t == nil || t.nodeID == "" {
		return "test-node"
	}
	return t.nodeID
}

func buildVersionedRuntimeFrontendAssets(marker string) []*catalog.ArtifactFrontendAsset {
	return []*catalog.ArtifactFrontendAsset{
		{
			Path:          "index.html",
			ContentBase64: base64.StdEncoding.EncodeToString([]byte("<html><body>" + marker + "</body></html>")),
			ContentType:   "text/html; charset=utf-8",
		},
	}
}
