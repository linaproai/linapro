// This file verifies scheduled-job plugin capability write and execution
// methods delegate to the job-management owner instead of writing sys_job
// directly.

package capabilityadapter

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	jobv1 "lina-core/api/job/v1"
	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/service/jobmeta"
	"lina-core/pkg/bizerr"
	_ "lina-core/pkg/dbdriver"
	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/jobcap"
)

// TestJobCapabilityAdapterCreateDelegatesToOwner verifies runtime job creation
// uses the owner contract with normalized shell-job defaults.
func TestJobCapabilityAdapterCreateDelegatesToOwner(t *testing.T) {
	var (
		ctx     = context.Background()
		owner   = &recordingJobOwner{createID: 42}
		adapter = NewCapabilityAdapter(owner, nil, nil)
	)

	id, err := adapter.Create(ctx, jobcap.SaveInput{
		GroupID:       "7",
		Name:          "Plugin shell job",
		Description:   "runs a plugin-owned shell command",
		ShellCmd:      "echo plugin",
		CronExpr:      "*/5 * * * *",
		Timezone:      "Asia/Shanghai",
		MaxExecutions: 3,
		LogRetentionOverride: &jobcap.LogRetentionOption{
			Mode:  jobv1.RetentionModeDays,
			Value: 60,
		},
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if id != jobcap.JobID("42") {
		t.Fatalf("Create returned id %q, want 42", id)
	}

	got := owner.createInput
	if got.GroupID != 7 || got.Name != "Plugin shell job" || got.TaskType != jobv1.TaskTypeShell {
		t.Fatalf("Create owner input mismatch: %#v", got)
	}
	if got.Timeout != 300*time.Second {
		t.Fatalf("Create timeout = %s, want 300s", got.Timeout)
	}
	if got.Scope != jobv1.ScopeMasterOnly || got.Concurrency != jobv1.ConcurrencySingleton {
		t.Fatalf("Create scheduling defaults mismatch: scope=%s concurrency=%s", got.Scope, got.Concurrency)
	}
	if got.MaxConcurrency != 1 || got.MaxExecutions != 3 || got.Status != jobv1.StatusDisabled {
		t.Fatalf("Create limit/status defaults mismatch: %#v", got)
	}
	if got.ShellCmd != "echo plugin" || got.CronExpr != "*/5 * * * *" || got.Timezone != "Asia/Shanghai" {
		t.Fatalf("Create command schedule mismatch: %#v", got)
	}
	if got.LogRetentionOverride == nil ||
		got.LogRetentionOverride.Mode != jobv1.RetentionModeDays ||
		got.LogRetentionOverride.Value != 60 {
		t.Fatalf("Create retention override mismatch: %#v", got.LogRetentionOverride)
	}
}

// TestJobCapabilityAdapterMutationsDelegateToOwner verifies visible job
// mutations and manual execution go through the job-management owner.
func TestJobCapabilityAdapterMutationsDelegateToOwner(t *testing.T) {
	var (
		ctx     = context.Background()
		owner   = &recordingJobOwner{}
		adapter = NewCapabilityAdapter(owner, nil, nil)
	)

	err := adapter.Update(ctx, jobcap.UpdateInput{
		ID: jobcap.JobID("8"),
		SaveInput: jobcap.SaveInput{
			GroupID:        "9",
			Name:           "Updated",
			ShellCmd:       "echo updated",
			CronExpr:       "0 * * * *",
			Timeout:        30 * time.Second,
			Scope:          jobv1.ScopeAllNode,
			Concurrency:    jobv1.ConcurrencyParallel,
			MaxConcurrency: 4,
			Status:         jobv1.StatusEnabled,
			LogRetentionOverride: &jobcap.LogRetentionOption{
				Mode: jobv1.RetentionModeNone,
			},
		},
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if owner.updateInput.ID != 8 || owner.updateInput.GroupID != 9 {
		t.Fatalf("Update owner input mismatch: %#v", owner.updateInput)
	}
	if owner.updateInput.Scope != jobv1.ScopeAllNode || owner.updateInput.Concurrency != jobv1.ConcurrencyParallel {
		t.Fatalf("Update owner scheduling mismatch: %#v", owner.updateInput)
	}
	if owner.updateInput.MaxConcurrency != 4 || owner.updateInput.Status != jobv1.StatusEnabled {
		t.Fatalf("Update owner status mismatch: %#v", owner.updateInput)
	}
	if owner.updateInput.LogRetentionOverride == nil ||
		owner.updateInput.LogRetentionOverride.Mode != jobv1.RetentionModeNone ||
		owner.updateInput.LogRetentionOverride.Value != 0 {
		t.Fatalf("Update retention override mismatch: %#v", owner.updateInput.LogRetentionOverride)
	}

	if err = adapter.Delete(ctx, jobcap.JobID("8")); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	if owner.deletedIDs != "8" {
		t.Fatalf("Delete owner ids = %q, want 8", owner.deletedIDs)
	}

	if err = adapter.Run(ctx, jobcap.JobID("8")); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if owner.triggeredID != 8 {
		t.Fatalf("Run owner id = %d, want 8", owner.triggeredID)
	}

	if err = adapter.SetStatus(ctx, jobcap.JobID("8"), jobv1.StatusDisabled); err != nil {
		t.Fatalf("SetStatus returned error: %v", err)
	}
	if owner.statusID != 8 || owner.status != jobv1.StatusDisabled {
		t.Fatalf("SetStatus owner state = id:%d status:%s, want 8 disabled", owner.statusID, owner.status)
	}
}

// TestJobCapabilityAdapterRejectsInvalidInputBeforeOwner verifies invalid
// identifiers and missing owner dependencies are reported without direct DB use.
func TestJobCapabilityAdapterRejectsInvalidInputBeforeOwner(t *testing.T) {
	var (
		ctx     = context.Background()
		owner   = &recordingJobOwner{}
		adapter = NewCapabilityAdapter(owner, nil, nil)
	)
	if err := adapter.Delete(ctx, jobcap.JobID("bad")); !bizerr.Is(err, capmodel.CodeCapabilityDenied) {
		t.Fatalf("Delete invalid id error = %v, want capability denied", err)
	}
	if owner.deletedIDs != "" {
		t.Fatalf("Delete invalid id still reached owner with ids %q", owner.deletedIDs)
	}

	adapter = NewCapabilityAdapter(nil, nil, nil)
	_, err := adapter.Create(ctx, jobcap.SaveInput{GroupID: "1", Name: "missing owner", ShellCmd: "echo x", CronExpr: "* * * * *"})
	if !bizerr.Is(err, capmodel.CodeCapabilityUnavailable) {
		t.Fatalf("Create nil owner error = %v, want capability unavailable", err)
	}
}

// TestJobCapabilityAdapterQueriesReturnLogRetentionOverride verifies read
// methods project the persisted per-job log cleanup policy for plugins.
func TestJobCapabilityAdapterQueriesReturnLogRetentionOverride(t *testing.T) {
	ctx := context.Background()
	groupID := createJobCapabilityTestGroup(t, ctx)
	prefix := fmt.Sprintf("jobcap-retention-query-%d", time.Now().UnixNano())
	retainedJobID := createJobCapabilityTestJob(
		t,
		ctx,
		groupID,
		prefix+"-retained",
		`{"mode":"days","value":60}`,
	)
	defaultJobID := createJobCapabilityTestJob(t, ctx, groupID, prefix+"-default", "")
	adapter := NewCapabilityAdapter(nil, nil, nil)

	got, err := adapter.Get(ctx, jobcap.JobID(strconv.FormatInt(retainedJobID, 10)))
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	assertJobRetentionOverride(t, got, jobv1.RetentionModeDays, 60)

	batch, err := adapter.BatchGet(ctx, []jobcap.JobID{
		jobcap.JobID(strconv.FormatInt(retainedJobID, 10)),
		jobcap.JobID(strconv.FormatInt(defaultJobID, 10)),
	})
	if err != nil {
		t.Fatalf("BatchGet returned error: %v", err)
	}
	assertJobRetentionOverride(
		t,
		batch.Items[jobcap.JobID(strconv.FormatInt(retainedJobID, 10))],
		jobv1.RetentionModeDays,
		60,
	)
	if got := batch.Items[jobcap.JobID(strconv.FormatInt(defaultJobID, 10))]; got == nil || got.LogRetentionOverride != nil {
		t.Fatalf("BatchGet default retention item = %#v, want nil override", got)
	}

	list, err := adapter.List(ctx, jobcap.ListInput{
		Keyword: prefix,
		Group:   strconv.FormatInt(groupID, 10),
		Page:    capmodel.PageRequest{PageNum: 1, PageSize: 10},
	})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	listed := jobInfoByID(list.Items, jobcap.JobID(strconv.FormatInt(retainedJobID, 10)))
	assertJobRetentionOverride(t, listed, jobv1.RetentionModeDays, 60)
	listedDefault := jobInfoByID(list.Items, jobcap.JobID(strconv.FormatInt(defaultJobID, 10)))
	if listedDefault == nil || listedDefault.LogRetentionOverride != nil {
		t.Fatalf("List default retention item = %#v, want nil override", listedDefault)
	}
}

// recordingJobOwner captures calls made by the capability adapter.
type recordingJobOwner struct {
	jobmeta.Owner

	createID    int64
	createInput jobmeta.SaveJobInput
	updateInput jobmeta.UpdateJobInput
	deletedIDs  string
	triggeredID int64
	statusID    int64
	status      jobv1.Status
}

// CreateJob records one create request and returns the configured ID.
func (o *recordingJobOwner) CreateJob(_ context.Context, in jobmeta.SaveJobInput) (int64, error) {
	o.createInput = in
	return o.createID, nil
}

// UpdateJob records one update request.
func (o *recordingJobOwner) UpdateJob(_ context.Context, in jobmeta.UpdateJobInput) error {
	o.updateInput = in
	return nil
}

// DeleteJobs records one delete request.
func (o *recordingJobOwner) DeleteJobs(_ context.Context, ids string) error {
	o.deletedIDs = ids
	return nil
}

// UpdateJobStatus records one status change request.
func (o *recordingJobOwner) UpdateJobStatus(_ context.Context, id int64, status jobv1.Status) error {
	o.statusID = id
	o.status = status
	return nil
}

// TriggerJob records one manual trigger request.
func (o *recordingJobOwner) TriggerJob(_ context.Context, id int64) (int64, error) {
	o.triggeredID = id
	return 99, nil
}

// createJobCapabilityTestGroup inserts one isolated group for query tests.
func createJobCapabilityTestGroup(t *testing.T, ctx context.Context) int64 {
	t.Helper()

	code := fmt.Sprintf("jobcap-retention-%d", time.Now().UnixNano())
	groupID, err := dao.SysJobGroup.Ctx(ctx).Data(do.SysJobGroup{
		TenantId:  0,
		Code:      code,
		Name:      "Jobcap retention query",
		Remark:    "temporary group for jobcap query tests",
		SortOrder: 0,
		IsDefault: 0,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("expected test job group insert to succeed, got error: %v", err)
	}
	t.Cleanup(func() {
		if _, err := dao.SysJobGroup.Ctx(ctx).Unscoped().Where(do.SysJobGroup{Id: groupID}).Delete(); err != nil {
			t.Fatalf("expected test job group cleanup to succeed, got error: %v", err)
		}
	})
	return groupID
}

// createJobCapabilityTestJob inserts one isolated sys_job row for query tests.
func createJobCapabilityTestJob(t *testing.T, ctx context.Context, groupID int64, name string, retentionOverride string) int64 {
	t.Helper()

	jobID, err := dao.SysJob.Ctx(ctx).Data(do.SysJob{
		TenantId:             0,
		GroupId:              groupID,
		Name:                 name,
		Description:          "temporary job for jobcap query tests",
		TaskType:             jobv1.TaskTypeShell,
		TimeoutSeconds:       300,
		ShellCmd:             "echo jobcap",
		CronExpr:             "*/5 * * * *",
		Timezone:             "Asia/Shanghai",
		Scope:                jobv1.ScopeMasterOnly,
		Concurrency:          jobv1.ConcurrencySingleton,
		MaxConcurrency:       1,
		MaxExecutions:        0,
		LogRetentionOverride: retentionOverride,
		Status:               jobv1.StatusEnabled,
		IsBuiltin:            0,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("expected test job insert to succeed, got error: %v", err)
	}
	t.Cleanup(func() {
		if _, err := dao.SysJob.Ctx(ctx).Unscoped().Where(do.SysJob{Id: jobID}).Delete(); err != nil {
			t.Fatalf("expected test job cleanup to succeed, got error: %v", err)
		}
	})
	return jobID
}

// assertJobRetentionOverride verifies one plugin job projection policy.
func assertJobRetentionOverride(t *testing.T, got *jobcap.JobInfo, mode jobv1.RetentionMode, value int64) {
	t.Helper()

	if got == nil || got.LogRetentionOverride == nil ||
		got.LogRetentionOverride.Mode != mode ||
		got.LogRetentionOverride.Value != value {
		t.Fatalf("retention override = %#v for item %#v, want mode=%s value=%d", got, got, mode, value)
	}
}

// jobInfoByID finds one job projection by ID.
func jobInfoByID(items []*jobcap.JobInfo, id jobcap.JobID) *jobcap.JobInfo {
	for _, item := range items {
		if item != nil && item.ID == id {
			return item
		}
	}
	return nil
}
