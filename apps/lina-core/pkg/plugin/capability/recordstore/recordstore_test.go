// This file tests record store capability guest-side query-plan and transaction builders.

package recordstore

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	dataplan "lina-core/pkg/plugin/capability/recordstore/internal/plan"
)

// TestTransactionRejectsMultipleTables verifies the transaction builder keeps
// the single-table contract enforced.
func TestTransactionRejectsMultipleTables(t *testing.T) {
	tx := &Tx{}
	_ = tx.Table("sys_plugin_node_state")
	_ = tx.Table("sys_plugin")
	if tx.err == nil {
		t.Fatal("expected transaction to reject multiple tables")
	}
}

// TestQueryBuilderBuildsTypedPlan verifies the fluent query builder records the
// expected typed plan state.
func TestQueryBuilderBuildsTypedPlan(t *testing.T) {
	query := Open().
		Table("sys_plugin_node_state").
		Fields("id", "pluginId", "currentState").
		WhereEq("pluginId", "plugin-demo").
		WhereIn("currentState", []string{"pending", "running"}).
		WhereLike("nodeKey", "demo-").
		OrderDesc("id").
		Page(2, 20)
	if query.err != nil {
		t.Fatalf("expected query builder to succeed, got %v", query.err)
	}
	if len(query.plan.Fields) != 3 {
		t.Fatalf("unexpected fields: %#v", query.plan.Fields)
	}
	if len(query.plan.Filters) != 3 {
		t.Fatalf("unexpected filters: %#v", query.plan.Filters)
	}
	if query.plan.Filters[1].Operator != dataplan.FilterOperatorIN {
		t.Fatalf("unexpected filter operator: %#v", query.plan.Filters[1])
	}
	if len(query.plan.Orders) != 1 || query.plan.Orders[0].Direction != dataplan.OrderDirectionDESC {
		t.Fatalf("unexpected orders: %#v", query.plan.Orders)
	}
	if query.plan.Page == nil || query.plan.Page.PageNum != 2 || query.plan.Page.PageSize != 20 {
		t.Fatalf("unexpected page: %#v", query.plan.Page)
	}
}

// TestWasip1GuestBuildDoesNotImportHostDatabaseDependencies verifies dynamic
// plugin guest builds keep host-side SQL drivers out of the wasm dependency set.
func TestWasip1GuestBuildDoesNotImportHostDatabaseDependencies(t *testing.T) {
	t.Parallel()

	moduleRoot, err := recordStoreCapabilityModuleRoot()
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("go", "list", "-deps", "-tags", "wasip1", "-json", "./pkg/plugin/capability/recordstore")
	cmd.Dir = moduleRoot
	cmd.Env = append(os.Environ(), "GOOS=wasip1", "GOARCH=wasm")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("go list wasip1 record store capability deps failed: %v", err)
	}
	packages, err := splitGoListJSONPackages(output)
	if err != nil {
		t.Fatalf("split go list package stream: %v", err)
	}
	for _, rawPackage := range packages {
		var pkg struct {
			ImportPath string
		}
		if err := json.Unmarshal(rawPackage, &pkg); err != nil {
			t.Fatalf("decode go list package: %v", err)
		}
		switch {
		case strings.Contains(pkg.ImportPath, "/internal/service/plugin/internal/datahost/internal/host"):
			t.Fatalf("wasip1 record store capability imported host-only package %s", pkg.ImportPath)
		case strings.HasPrefix(pkg.ImportPath, "github.com/lib/pq"):
			t.Fatalf("wasip1 record store capability imported PostgreSQL driver package %s", pkg.ImportPath)
		case pkg.ImportPath == "github.com/gogf/gf/contrib/drivers/pgsql/v2":
			t.Fatalf("wasip1 record store capability imported GoFrame PostgreSQL driver package %s", pkg.ImportPath)
		}
	}
}

// recordStoreCapabilityModuleRoot returns the lina-core module root for subprocess checks.
func recordStoreCapabilityModuleRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, statErr := os.Stat(filepath.Join(wd, "go.mod")); statErr == nil {
			return wd, nil
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			return "", os.ErrNotExist
		}
		wd = parent
	}
}

// splitGoListJSONPackages splits the concatenated JSON objects emitted by
// `go list -json` without relying on line-oriented formatting.
func splitGoListJSONPackages(output []byte) ([][]byte, error) {
	decoder := json.NewDecoder(bytes.NewReader(output))
	packages := make([][]byte, 0)
	for {
		var raw json.RawMessage
		if err := decoder.Decode(&raw); err != nil {
			if err == io.EOF {
				return packages, nil
			}
			return nil, err
		}
		packages = append(packages, append([]byte(nil), raw...))
	}
}
