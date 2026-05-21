// This file exercises lifecycle callbacks from the bundled dynamic sample
// artifact so install-time bridge regressions are caught before E2E runs.

package runtime_test

import (
	"context"
	"path/filepath"
	"testing"

	"lina-core/internal/service/plugin/internal/runtime"
	"lina-core/internal/service/plugin/internal/testutil"
	"lina-core/pkg/pluginhost"
)

// TestRunBundledDynamicSampleBeforeInstallLifecycleAllowsRuntimeLog verifies
// the bundled dynamic sample can run its BeforeInstall callback, including the
// runtime.log.write host service used by the callback implementation.
func TestRunBundledDynamicSampleBeforeInstallLifecycleAllowsRuntimeLog(t *testing.T) {
	testutil.EnsureBundledRuntimeSampleArtifactForTests(t)

	services := testutil.NewServices()
	artifactPath := filepath.Join(testutil.TestDynamicStorageDir(), runtime.BuildArtifactFileName("linapro-demo-dynamic"))
	manifest, err := services.Catalog.LoadManifestFromArtifactPath(artifactPath)
	if err != nil {
		t.Fatalf("expected bundled dynamic manifest to load, got error: %v", err)
	}

	decision, err := services.Runtime.RunDynamicLifecyclePrecondition(context.Background(), manifest, runtime.DynamicLifecycleInput{
		PluginID:  manifest.ID,
		Operation: pluginhost.LifecycleHookBeforeInstall,
	})
	if err != nil {
		t.Fatalf("expected bundled BeforeInstall lifecycle to succeed, got error: %v decision=%#v", err, decision)
	}
	if decision == nil || !decision.OK {
		t.Fatalf("expected bundled BeforeInstall lifecycle to allow install, got %#v", decision)
	}
}
