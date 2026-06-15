// This file verifies user capability query assembly inside the usercap component.

package capabilityhost

import (
	"context"
	"strings"
	"testing"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/internal/service/datascope"
	_ "lina-core/pkg/dbdriver"
	"lina-core/pkg/plugin/capability/capmodel"
	capabilityusercap "lina-core/pkg/plugin/capability/usercap"
)

// TestSearchCountsWithoutProjectionFields verifies the search count query
// does not reuse projection columns, which keeps PostgreSQL count SQL valid.
func TestSearchCountsWithoutProjectionFields(t *testing.T) {
	ctx := context.Background()

	sqls, err := gdb.CatchSQL(ctx, func(sqlCtx context.Context) error {
		_, searchErr := newUserCapabilityAdapter(nil, nil).Search(sqlCtx, capmodel.CapabilityContext{}, capabilityusercap.SearchInput{
			Page: capmodel.PageRequest{PageNum: 1, PageSize: 10},
		})
		return searchErr
	})
	if err != nil {
		t.Fatalf("render search users SQL: %v", err)
	}

	combinedSQL := strings.Join(sqls, "\n")
	normalizedSQL := strings.ReplaceAll(combinedSQL, "`", `"`)
	if strings.Contains(normalizedSQL, `COUNT("id","tenant_id","username","nickname","avatar","status")`) {
		t.Fatalf("count query reused projection fields: %s", combinedSQL)
	}
	if !strings.Contains(strings.ToUpper(normalizedSQL), "COUNT") {
		t.Fatalf("expected rendered SQL to include count query, got: %s", combinedSQL)
	}
}

// TestBatchGetAppliesDataScope verifies user projections are constrained
// by the shared data-scope service before rows are scanned.
func TestBatchGetAppliesDataScope(t *testing.T) {
	ctx := context.Background()
	scopeSvc := &recordingDataScope{empty: true}

	result, err := newUserCapabilityAdapter(nil, scopeSvc).BatchGet(ctx, capmodel.CapabilityContext{}, []capabilityusercap.UserID{"7", "8"})
	if err != nil {
		t.Fatalf("batch get users failed: %v", err)
	}
	if scopeSvc.applyCalls != 1 || scopeSvc.lastColumn != "sys_user.id" {
		t.Fatalf("expected data scope to apply once to sys_user.id, calls=%d column=%q", scopeSvc.applyCalls, scopeSvc.lastColumn)
	}
	if len(result.Items) != 0 || !containsUserID(result.MissingIDs, "7") || !containsUserID(result.MissingIDs, "8") {
		t.Fatalf("expected empty scope to return opaque missing IDs, got %#v", result)
	}
}

// recordingDataScope records data-scope application in usercap tests.
type recordingDataScope struct {
	empty      bool
	applyCalls int
	lastColumn string
}

// Current returns an all-data scope for unused interface paths.
func (*recordingDataScope) Current(context.Context) (*datascope.Context, error) {
	return &datascope.Context{UserID: 1, Scope: datascope.ScopeAll}, nil
}

// ApplyUserScope records the requested user owner column.
func (s *recordingDataScope) ApplyUserScope(_ context.Context, model *gdb.Model, userIDColumn string) (*gdb.Model, bool, error) {
	s.applyCalls++
	s.lastColumn = userIDColumn
	return model, s.empty, nil
}

// ApplyUserScopeWithBypass is unused by usercap tests.
func (s *recordingDataScope) ApplyUserScopeWithBypass(
	ctx context.Context,
	model *gdb.Model,
	userIDColumn string,
	_ string,
	_ any,
) (*gdb.Model, bool, error) {
	return s.ApplyUserScope(ctx, model, userIDColumn)
}

// EnsureUsersVisible accepts all users in this fixture.
func (*recordingDataScope) EnsureUsersVisible(context.Context, []int) error { return nil }

// EnsureRowsVisible accepts all rows in this fixture.
func (*recordingDataScope) EnsureRowsVisible(context.Context, *gdb.Model, string, int) error {
	return nil
}

func containsUserID(ids []capabilityusercap.UserID, target capabilityusercap.UserID) bool {
	for _, id := range ids {
		if id == target {
			return true
		}
	}
	return false
}
