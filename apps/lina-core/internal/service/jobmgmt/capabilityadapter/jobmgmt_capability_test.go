// This file verifies scheduled-job plugin capability write and execution
// methods delegate to the job-management owner instead of writing sys_job
// directly.

package capabilityadapter

import (
	"context"
	jobv1 "lina-core/api/job/v1"
	"testing"
	"time"

	"lina-core/internal/service/jobmeta"
	"lina-core/pkg/bizerr"
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
