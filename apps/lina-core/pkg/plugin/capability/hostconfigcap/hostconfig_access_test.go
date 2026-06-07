// This file verifies source-plugin host config reads are not constrained by a
// hard-coded allowlist.

package hostconfigcap

import (
	"context"
	"strings"
	"testing"

	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"
)

// TestHostConfigReadsAnyNonRootHostConfigKey verifies source plugins can read
// host config keys that are not present in a public-key allowlist.
func TestHostConfigReadsAnyNonRootHostConfigKey(t *testing.T) {
	setTestHostConfigAdapter(t, `
database:
  default:
    link: "pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable"
plugin:
  dynamic:
    storagePath: "temp/dynamic"
`)

	svc := New(testRawHostConfigReader{})
	ctx := context.Background()

	link, err := svc.String(ctx, "database.default.link", "")
	if err != nil {
		t.Fatalf("read database.default.link: %v", err)
	}
	if link != "pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable" {
		t.Fatalf("expected database.default.link to be readable, got %q", link)
	}

	path, err := svc.String(ctx, "plugin.dynamic.storagePath", "")
	if err != nil {
		t.Fatalf("read plugin.dynamic.storagePath: %v", err)
	}
	if path != "temp/dynamic" {
		t.Fatalf("expected plugin.dynamic.storagePath to be readable, got %q", path)
	}
}

// TestHostConfigMissingKeyReturnsAbsent verifies unknown keys are treated as
// absent instead of being rejected as non-public.
func TestHostConfigMissingKeyReturnsAbsent(t *testing.T) {
	setTestHostConfigAdapter(t, `
workspace:
  basePath: "/admin"
`)

	svc := New(testRawHostConfigReader{})
	found, err := svc.Exists(context.Background(), "database.default.link")
	if err != nil {
		t.Fatalf("check missing host config key: %v", err)
	}
	if found {
		t.Fatal("expected missing host config key to report absent")
	}
}

// TestHostConfigAllowsRootReads verifies source plugins can read the host
// config root when they explicitly ask for it.
func TestHostConfigAllowsRootReads(t *testing.T) {
	setTestHostConfigAdapter(t, `
workspace:
  basePath: "/admin"
`)

	svc := New(testRawHostConfigReader{})
	value, err := svc.Get(context.Background(), ".")
	if err != nil {
		t.Fatalf("read host config root: %v", err)
	}
	if value == nil || value.IsNil() {
		t.Fatal("expected host config root to be readable")
	}

	root := value.MapStrAny()
	workspace, ok := root["workspace"].(map[string]any)
	if !ok {
		t.Fatalf("expected workspace section in root config, got %#v", root["workspace"])
	}
	if workspace["basePath"] != "/admin" {
		t.Fatalf("expected workspace.basePath in root config, got %#v", workspace["basePath"])
	}
}

// TestHostConfigRequiresInjectedRawReader verifies HostConfig uses the
// startup-injected host config service instead of silently constructing one.
func TestHostConfigRequiresInjectedRawReader(t *testing.T) {
	svc := New(nil)

	if _, err := svc.Get(context.Background(), "workspace.basePath"); err == nil ||
		!strings.Contains(err.Error(), "not configured") {
		t.Fatalf("expected missing host config service to fail explicitly, got %v", err)
	}
}

// setTestHostConfigAdapter swaps the process config adapter for one test case
// and restores the original adapter afterward.
func setTestHostConfigAdapter(t *testing.T, content string) {
	t.Helper()

	adapter, err := gcfg.NewAdapterContent(content)
	if err != nil {
		t.Fatalf("create content adapter: %v", err)
	}

	originalAdapter := g.Cfg().GetAdapter()
	g.Cfg().SetAdapter(adapter)
	t.Cleanup(func() {
		g.Cfg().SetAdapter(originalAdapter)
	})
}

// testRawHostConfigReader reads from the test-scoped GoFrame config adapter.
type testRawHostConfigReader struct{}

// GetRaw returns one raw test config value through the active GoFrame adapter.
func (testRawHostConfigReader) GetRaw(ctx context.Context, key string) (*gvar.Var, error) {
	return g.Cfg().Get(ctx, key)
}
