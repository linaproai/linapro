// Package capabilityadapter adapts host scheduled-job rows to plugin-visible
// job capability contracts without exposing sys_job entities.
package capabilityadapter

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"

	jobv1 "lina-core/api/job/v1"
	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/service/datascope"
	"lina-core/internal/service/jobmeta"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/capmodel"
	capabilityjobcap "lina-core/pkg/plugin/capability/jobcap"
	"lina-core/pkg/plugin/capability/tenantcap"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
)

// adapter exposes scheduled-job projections without leaking sys_job entities.
type jobCapabilityAdapter struct {
	owner        jobmeta.Owner
	scopeSvc     datascope.Service
	tenantFilter tenantcap.FilterService
}

var _ capabilityjobcap.Service = (*jobCapabilityAdapter)(nil)

// jobInfoRow stores the minimal sys_job projection returned to plugins.
type jobInfoRow struct {
	Id                   int64
	Name                 string
	GroupId              int64
	Status               string
	LogRetentionOverride string
}

// NewCapabilityAdapter creates the host-owned scheduled-job capability adapter.
func NewCapabilityAdapter(owner jobmeta.Owner, tenantFilter tenantcap.FilterService, scopeSvc datascope.Service) capabilityjobcap.Service {
	return &jobCapabilityAdapter{owner: owner, scopeSvc: scopeSvc, tenantFilter: tenantFilter}
}

// Get returns one visible scheduled-job projection.
func (a *jobCapabilityAdapter) Get(ctx context.Context, id capabilityjobcap.JobID) (*capabilityjobcap.JobInfo, error) {
	result, err := a.BatchGet(ctx, []capabilityjobcap.JobID{id})
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	if projection := result.Items[id]; projection != nil {
		return projection, nil
	}
	return nil, bizerr.NewCode(capmodel.CodeCapabilityDenied)
}

// BatchGet returns visible scheduled-job projections and opaque missing IDs.
func (a *jobCapabilityAdapter) BatchGet(ctx context.Context, ids []capabilityjobcap.JobID) (*capmodel.BatchResult[*capabilityjobcap.JobInfo, capabilityjobcap.JobID], error) {
	result := &capmodel.BatchResult[*capabilityjobcap.JobInfo, capabilityjobcap.JobID]{
		Items:      make(map[capabilityjobcap.JobID]*capabilityjobcap.JobInfo, len(ids)),
		MissingIDs: []capabilityjobcap.JobID{},
	}
	parsedIDs, requested := capmodel.ParseInt64IDs(ids, func(id capabilityjobcap.JobID) {
		result.MissingIDs = append(result.MissingIDs, id)
	})
	if len(parsedIDs) == 0 {
		return result, nil
	}
	rows := make([]*jobInfoRow, 0, len(parsedIDs))
	cols := dao.SysJob.Columns()
	model := dao.SysJob.Ctx(ctx).
		Fields(cols.Id, cols.Name, cols.GroupId, cols.Status, cols.LogRetentionOverride).
		WhereIn(cols.Id, parsedIDs)
	if a != nil && a.tenantFilter != nil {
		model = tenantspi.ApplyPluginTableFilter(ctx, a.tenantFilter, model, "")
	}
	model, err := a.applyReadScope(ctx, model)
	if err != nil {
		return nil, err
	}
	if err = model.Scan(&rows); err != nil {
		return nil, err
	}
	for _, row := range rows {
		if row == nil {
			continue
		}
		requestID, ok := requested[row.Id]
		if !ok {
			continue
		}
		info, err := jobInfoFromRow(row, requestID)
		if err != nil {
			return nil, err
		}
		result.Items[requestID] = info
	}
	for _, id := range ids {
		if _, ok := result.Items[id]; !ok && !slices.Contains(result.MissingIDs, id) {
			result.MissingIDs = append(result.MissingIDs, id)
		}
	}
	return result, nil
}

// List returns one bounded page of visible scheduled-job projections.
func (a *jobCapabilityAdapter) List(ctx context.Context, input capabilityjobcap.ListInput) (*capmodel.PageResult[*capabilityjobcap.JobInfo], error) {
	pageNum, pageSize := input.Page.Normalize()
	if pageSize > capabilityjobcap.MaxListPageSize {
		pageSize = capabilityjobcap.MaxListPageSize
	}
	cols := dao.SysJob.Columns()
	model := dao.SysJob.Ctx(ctx)
	if a != nil && a.tenantFilter != nil {
		model = tenantspi.ApplyPluginTableFilter(ctx, a.tenantFilter, model, "")
	}
	model, err := a.applyReadScope(ctx, model)
	if err != nil {
		return nil, err
	}
	if keyword := strings.TrimSpace(input.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		model = model.Where(
			fmt.Sprintf("(%s LIKE ? OR %s LIKE ?)", cols.Name, cols.HandlerRef),
			like,
			like,
		)
	}
	if group := strings.TrimSpace(input.Group); group != "" {
		groupID, err := strconv.ParseInt(group, 10, 64)
		if err != nil || groupID <= 0 {
			return &capmodel.PageResult[*capabilityjobcap.JobInfo]{Items: []*capabilityjobcap.JobInfo{}, Total: 0}, nil
		}
		model = model.Where(do.SysJob{GroupId: groupID})
	}
	if status := strings.TrimSpace(string(input.Status)); status != "" {
		model = model.Where(do.SysJob{Status: status})
	}
	total, err := model.Clone().Count()
	if err != nil {
		return nil, err
	}
	rows := make([]*jobInfoRow, 0, pageSize)
	if err = model.Clone().
		Fields(cols.Id, cols.Name, cols.GroupId, cols.Status, cols.LogRetentionOverride).
		Page(pageNum, pageSize).
		OrderDesc(cols.Id).
		Scan(&rows); err != nil {
		return nil, err
	}
	items := make([]*capabilityjobcap.JobInfo, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		info, err := jobInfoFromRow(row, capabilityjobcap.JobID(strconv.FormatInt(row.Id, 10)))
		if err != nil {
			return nil, err
		}
		items = append(items, info)
	}
	return &capmodel.PageResult[*capabilityjobcap.JobInfo]{Items: items, Total: total}, nil
}

// EnsureVisible rejects when any requested scheduled job is absent or invisible.
func (a *jobCapabilityAdapter) EnsureVisible(ctx context.Context, ids []capabilityjobcap.JobID) error {
	if len(ids) > capabilityjobcap.MaxEnsureVisible {
		return bizerr.NewCode(capmodel.CodeCapabilityLimitExceeded, bizerr.P("limit", capabilityjobcap.MaxEnsureVisible))
	}
	result, err := a.BatchGet(ctx, ids)
	if err != nil {
		return err
	}
	if result == nil || len(result.MissingIDs) > 0 {
		return bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	return nil
}

// Create creates one governed scheduled job through the job owner.
func (a *jobCapabilityAdapter) Create(ctx context.Context, input capabilityjobcap.SaveInput) (capabilityjobcap.JobID, error) {
	save, err := jobSaveInput(input)
	if err != nil {
		return "", err
	}
	owner, err := a.requireOwner()
	if err != nil {
		return "", err
	}
	id, err := owner.CreateJob(ctx, save)
	if err != nil {
		return "", err
	}
	return capabilityjobcap.JobID(strconv.FormatInt(id, 10)), nil
}

// Update mutates one visible scheduled job through the job owner.
func (a *jobCapabilityAdapter) Update(ctx context.Context, input capabilityjobcap.UpdateInput) error {
	id, err := parseJobID(input.ID)
	if err != nil {
		return err
	}
	save, err := jobSaveInput(input.SaveInput)
	if err != nil {
		return err
	}
	owner, err := a.requireOwner()
	if err != nil {
		return err
	}
	return owner.UpdateJob(ctx, jobmeta.UpdateJobInput{
		ID:           id,
		SaveJobInput: save,
	})
}

// Delete deletes one visible scheduled job through the job owner.
func (a *jobCapabilityAdapter) Delete(ctx context.Context, id capabilityjobcap.JobID) error {
	parsedID, err := parseJobID(id)
	if err != nil {
		return err
	}
	owner, err := a.requireOwner()
	if err != nil {
		return err
	}
	return owner.DeleteJobs(ctx, strconv.FormatInt(parsedID, 10))
}

// Run triggers one visible scheduled job through the job owner.
func (a *jobCapabilityAdapter) Run(ctx context.Context, id capabilityjobcap.JobID) error {
	parsedID, err := parseJobID(id)
	if err != nil {
		return err
	}
	owner, err := a.requireOwner()
	if err != nil {
		return err
	}
	_, err = owner.TriggerJob(ctx, parsedID)
	return err
}

// SetStatus changes a visible scheduled job status.
func (a *jobCapabilityAdapter) SetStatus(ctx context.Context, id capabilityjobcap.JobID, status jobv1.Status) error {
	parsedID, err := parseJobID(id)
	if err != nil {
		return err
	}
	owner, err := a.requireOwner()
	if err != nil {
		return err
	}
	return owner.UpdateJobStatus(ctx, parsedID, jobmeta.NormalizeJobStatus(string(status)))
}

// jobSaveInput converts plugin job inputs into the host job owner model.
func jobSaveInput(input capabilityjobcap.SaveInput) (jobmeta.SaveJobInput, error) {
	groupID, err := strconv.ParseInt(strings.TrimSpace(input.GroupID), 10, 64)
	if err != nil || groupID <= 0 {
		return jobmeta.SaveJobInput{}, bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	timeout := input.Timeout
	if timeout <= 0 {
		timeout = 300 * time.Second
	}
	scope := jobmeta.NormalizeJobScope(string(input.Scope))
	if scope == "" {
		scope = jobv1.ScopeMasterOnly
	}
	concurrency := jobmeta.NormalizeJobConcurrency(string(input.Concurrency))
	if concurrency == "" {
		concurrency = jobv1.ConcurrencySingleton
	}
	maxConcurrency := input.MaxConcurrency
	if maxConcurrency <= 0 {
		maxConcurrency = 1
	}
	status := jobmeta.NormalizeJobStatus(string(input.Status))
	if status == "" {
		status = jobv1.StatusDisabled
	}
	var retentionOverride *jobmeta.RetentionOption
	if input.LogRetentionOverride != nil {
		retentionOverride = &jobmeta.RetentionOption{
			Mode:  jobmeta.NormalizeRetentionMode(string(input.LogRetentionOverride.Mode)),
			Value: input.LogRetentionOverride.Value,
		}
	}
	return jobmeta.SaveJobInput{
		GroupID:              groupID,
		Name:                 input.Name,
		Description:          input.Description,
		TaskType:             jobv1.TaskTypeShell,
		Timeout:              timeout,
		ShellCmd:             input.ShellCmd,
		WorkDir:              input.WorkDir,
		Env:                  input.Env,
		CronExpr:             input.CronExpr,
		Timezone:             input.Timezone,
		Scope:                scope,
		Concurrency:          concurrency,
		MaxConcurrency:       maxConcurrency,
		MaxExecutions:        input.MaxExecutions,
		Status:               status,
		LogRetentionOverride: retentionOverride,
	}, nil
}

// parseJobID decodes one plugin-visible job ID into the host owner key.
func parseJobID(id capabilityjobcap.JobID) (int64, error) {
	parsedID, err := strconv.ParseInt(strings.TrimSpace(string(id)), 10, 64)
	if err != nil || parsedID <= 0 {
		return 0, bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	return parsedID, nil
}

// jobInfoFromRow converts one sys_job projection into the plugin contract.
func jobInfoFromRow(row *jobInfoRow, id capabilityjobcap.JobID) (*capabilityjobcap.JobInfo, error) {
	retentionOverride, err := jobLogRetentionOption(row.LogRetentionOverride)
	if err != nil {
		return nil, err
	}
	return &capabilityjobcap.JobInfo{
		ID:                   id,
		Name:                 row.Name,
		Group:                strconv.FormatInt(row.GroupId, 10),
		Status:               jobv1.Status(row.Status),
		LogRetentionOverride: retentionOverride,
	}, nil
}

// jobLogRetentionOption parses the persisted job policy without leaking
// jobmeta internals through the plugin contract.
func jobLogRetentionOption(raw string) (*capabilityjobcap.LogRetentionOption, error) {
	option, err := jobmeta.ParseRetentionOption(raw)
	if err != nil {
		return nil, err
	}
	if option == nil {
		return nil, nil
	}
	return &capabilityjobcap.LogRetentionOption{
		Mode:  option.Mode,
		Value: option.Value,
	}, nil
}

// requireOwner returns the injected job owner or a structured unavailable error.
func (a *jobCapabilityAdapter) requireOwner() (jobmeta.Owner, error) {
	if a == nil || a.owner == nil {
		return nil, bizerr.NewCode(capmodel.CodeCapabilityUnavailable, bizerr.P("capability", "jobs"))
	}
	return a.owner, nil
}

// applyReadScope applies tenant and user data-scope constraints to job reads
// while keeping built-in jobs visible.
func (a *jobCapabilityAdapter) applyReadScope(ctx context.Context, model *gdb.Model) (*gdb.Model, error) {
	if model == nil || a == nil || a.scopeSvc == nil {
		return model, nil
	}
	cols := dao.SysJob.Columns()
	scopedModel, _, err := a.scopeSvc.ApplyUserScopeWithBypass(
		ctx,
		model,
		dao.SysJob.Table()+"."+cols.CreatedBy,
		dao.SysJob.Table()+"."+cols.IsBuiltin,
		1,
	)
	if err != nil {
		if bizerr.Is(err, datascope.CodeDataScopeDenied) {
			return nil, bizerr.NewCode(capmodel.CodeCapabilityDenied)
		}
		return nil, err
	}
	return scopedModel, nil
}
