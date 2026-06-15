// This file adapts host scheduled-job rows to plugin-visible job
// capability contracts without exposing sys_job entities.
package capabilityhost

import (
	"context"
	"strconv"
	"strings"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/capmodel"
	capabilityjobcap "lina-core/pkg/plugin/capability/jobcap"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
)

// Service exposes the scheduled-job domain service and management commands.
type jobCapabilityService interface {
	capabilityjobcap.Service
	capabilityjobcap.AdminService
}

// adapter exposes scheduled-job projections without leaking sys_job entities.
type jobCapabilityAdapter struct {
	tenantFilter tenantspi.PluginTableFilterService
}

var (
	_ capabilityjobcap.Service      = (*jobCapabilityAdapter)(nil)
	_ capabilityjobcap.AdminService = (*jobCapabilityAdapter)(nil)
)

// New creates the host-owned scheduled-job capability adapter.
func newJobCapabilityAdapter(tenantFilter tenantspi.PluginTableFilterService) jobCapabilityService {
	return &jobCapabilityAdapter{tenantFilter: tenantFilter}
}

// BatchGet returns visible scheduled-job projections and opaque missing IDs.
func (a *jobCapabilityAdapter) BatchGet(ctx context.Context, _ capmodel.CapabilityContext, ids []capabilityjobcap.JobID) (*capmodel.BatchResult[*capabilityjobcap.Projection, capabilityjobcap.JobID], error) {
	result := &capmodel.BatchResult[*capabilityjobcap.Projection, capabilityjobcap.JobID]{
		Items:      make(map[capabilityjobcap.JobID]*capabilityjobcap.Projection, len(ids)),
		MissingIDs: []capabilityjobcap.JobID{},
	}
	parsedIDs, requested := ParseInt64IDs(ids, func(id capabilityjobcap.JobID) {
		result.MissingIDs = append(result.MissingIDs, id)
	})
	if len(parsedIDs) == 0 {
		return result, nil
	}
	rows := make([]*struct {
		Id      int64
		Name    string
		GroupId int64
		Status  string
	}, 0, len(parsedIDs))
	cols := dao.SysJob.Columns()
	model := dao.SysJob.Ctx(ctx).
		Fields(cols.Id, cols.Name, cols.GroupId, cols.Status).
		WhereIn(cols.Id, parsedIDs)
	if a != nil && a.tenantFilter != nil {
		model = a.tenantFilter.Apply(ctx, model, "")
	}
	if err := model.Scan(&rows); err != nil {
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
		result.Items[requestID] = &capabilityjobcap.Projection{
			ID:     requestID,
			Name:   row.Name,
			Group:  strconv.FormatInt(row.GroupId, 10),
			Status: row.Status,
		}
	}
	for _, id := range ids {
		if _, ok := result.Items[id]; !ok && !Contains(result.MissingIDs, id) {
			result.MissingIDs = append(result.MissingIDs, id)
		}
	}
	return result, nil
}

// Run reports unavailable because executing a scheduled job requires the
// scheduler owner service, which is not part of the current source-plugin
// directory construction path.
func (a *jobCapabilityAdapter) Run(context.Context, capmodel.CapabilityContext, capabilityjobcap.JobID) error {
	return bizerr.NewCode(capmodel.CodeCapabilityUnavailable, bizerr.P("capability", "job-run"))
}

// SetStatus changes a visible scheduled job status.
func (a *jobCapabilityAdapter) SetStatus(ctx context.Context, capCtx capmodel.CapabilityContext, id capabilityjobcap.JobID, status string) error {
	result, err := a.BatchGet(ctx, capCtx, []capabilityjobcap.JobID{id})
	if err != nil {
		return err
	}
	if len(result.MissingIDs) > 0 {
		return bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	parsedID, err := strconv.ParseInt(strings.TrimSpace(string(id)), 10, 64)
	if err != nil || parsedID <= 0 {
		return bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	model := dao.SysJob.Ctx(ctx).Where(do.SysJob{Id: parsedID})
	if a != nil && a.tenantFilter != nil {
		model = a.tenantFilter.Apply(ctx, model, "")
	}
	_, err = model.Data(do.SysJob{Status: strings.TrimSpace(status)}).Update()
	return err
}
