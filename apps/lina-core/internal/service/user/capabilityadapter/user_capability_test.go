// This file verifies user capability query assembly inside the usercap component.

package capabilityadapter

import (
	"context"
	"strings"
	"testing"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/internal/service/datascope"
	"lina-core/pkg/bizerr"
	_ "lina-core/pkg/dbdriver"
	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/capmodel"
	capabilityusercap "lina-core/pkg/plugin/capability/usercap"
)

// TestListCountsWithoutProjectionFields verifies the list count query
// does not reuse projection columns, which keeps PostgreSQL count SQL valid.
func TestListCountsWithoutProjectionFields(t *testing.T) {
	ctx := context.Background()

	sqls, err := gdb.CatchSQL(ctx, func(sqlCtx context.Context) error {
		_, listErr := NewCapabilityAdapter(nil, nil, nil, nil).List(sqlCtx, capabilityusercap.ListInput{
			Page: capmodel.PageRequest{PageNum: 1, PageSize: 10},
		})
		return listErr
	})
	if err != nil {
		t.Fatalf("render list users SQL: %v", err)
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

	result, err := NewCapabilityAdapter(nil, nil, scopeSvc, nil).BatchGet(ctx, []capabilityusercap.UserID{"7", "8"})
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

// TestBatchGetAllowsHostSystemOrchestrationWithoutDataScope verifies startup
// and lifecycle orchestrations can read stable user projections without an HTTP
// request data-scope snapshot.
func TestBatchGetAllowsHostSystemOrchestrationWithoutDataScope(t *testing.T) {
	ctx := context.Background()
	scopeSvc := &recordingDataScope{empty: true}

	systemCtx := bizctxcap.WithCurrentContext(ctx, bizctxcap.CurrentContext{PlatformBypass: true})
	result, err := NewCapabilityAdapter(nil, nil, scopeSvc, nil).BatchGet(systemCtx, []capabilityusercap.UserID{"7"})
	if err != nil {
		t.Fatalf("batch get users failed: %v", err)
	}
	if scopeSvc.applyCalls != 0 {
		t.Fatalf("expected host system orchestration to bypass request data scope, calls=%d", scopeSvc.applyCalls)
	}
	if !containsUserID(result.MissingIDs, "7") {
		t.Fatalf("expected missing row to stay opaque, got %#v", result)
	}
}

// TestBatchGetScopesRegularRequests verifies requests without the platform
// bypass marker do not bypass data-scope rules.
func TestBatchGetScopesRegularRequests(t *testing.T) {
	ctx := context.Background()
	scopeSvc := &recordingDataScope{empty: true}

	_, err := NewCapabilityAdapter(nil, nil, scopeSvc, nil).BatchGet(ctx, []capabilityusercap.UserID{"7"})
	if err != nil {
		t.Fatalf("batch get users failed: %v", err)
	}
	if scopeSvc.applyCalls != 1 || scopeSvc.lastColumn != "sys_user.id" {
		t.Fatalf("expected HTTP system call to apply data scope, calls=%d column=%q", scopeSvc.applyCalls, scopeSvc.lastColumn)
	}
}

// TestCurrentRequiresCurrentUser verifies current user projection calls fail
// closed without a current user.
func TestCurrentRequiresCurrentUser(t *testing.T) {
	ctx := context.Background()
	_, err := NewCapabilityAdapter(nil, nil, nil, nil).Current(ctx)
	if !bizerr.Is(err, capmodel.CodeCapabilityCurrentUserRequired) {
		t.Fatalf("expected current user required error, got %v", err)
	}
}

// TestBatchResolveAppliesDataScope verifies user resolution is scoped before rows are scanned.
func TestBatchResolveAppliesDataScope(t *testing.T) {
	ctx := context.Background()
	scopeSvc := &recordingDataScope{empty: true}

	result, err := NewCapabilityAdapter(nil, nil, scopeSvc, nil).BatchResolve(ctx, capabilityusercap.BatchResolveInput{
		IDs:       []capabilityusercap.UserID{"7"},
		Usernames: []string{"alice"},
		Contacts:  []string{"alice@example.test"},
	})
	if err != nil {
		t.Fatalf("batch resolve users failed: %v", err)
	}
	if scopeSvc.applyCalls != 1 || scopeSvc.lastColumn != "sys_user.id" {
		t.Fatalf("expected data scope to apply once to sys_user.id, calls=%d column=%q", scopeSvc.applyCalls, scopeSvc.lastColumn)
	}
	for _, key := range []capabilityusercap.ResolveKey{"id:7", "username:alice", "contact:alice@example.test"} {
		if !containsResolveKey(result.MissingIDs, key) {
			t.Fatalf("expected %s to be opaque missing, got %#v", key, result.MissingIDs)
		}
	}
}

// TestBatchResolveRejectsLimit verifies user resolution input is bounded before SQL assembly.
func TestBatchResolveRejectsLimit(t *testing.T) {
	ids := make([]capabilityusercap.UserID, capabilityusercap.MaxBatchResolveIDs+1)
	_, err := NewCapabilityAdapter(nil, nil, nil, nil).BatchResolve(context.Background(), capabilityusercap.BatchResolveInput{IDs: ids})
	if !bizerr.Is(err, capmodel.CodeCapabilityLimitExceeded) {
		t.Fatalf("expected limit error, got %v", err)
	}
}

// TestNormalizeUserResolveInputDeduplicates verifies repeated resolve keys do
// not inflate database lookup dimensions.
func TestNormalizeUserResolveInputDeduplicates(t *testing.T) {
	result := normalizeUserResolveInput(capabilityusercap.BatchResolveInput{
		IDs:       []capabilityusercap.UserID{"7", " 7 "},
		Usernames: []string{"alice", "alice"},
		Contacts:  []string{"alice@example.test", "alice@example.test"},
	})
	if len(result.ids) != 1 || len(result.usernames) != 1 || len(result.contacts) != 1 {
		t.Fatalf("expected lookup dimensions to be deduplicated, got ids=%#v usernames=%#v contacts=%#v", result.ids, result.usernames, result.contacts)
	}
	for _, key := range []capabilityusercap.ResolveKey{"id:7", "username:alice", "contact:alice@example.test"} {
		if !containsResolveKey(result.keys, key) {
			t.Fatalf("expected normalized key %s in %#v", key, result.keys)
		}
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

func containsResolveKey(ids []capabilityusercap.ResolveKey, target capabilityusercap.ResolveKey) bool {
	for _, id := range ids {
		if id == target {
			return true
		}
	}
	return false
}
