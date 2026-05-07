// This file provides request-scoped startup snapshots for small scheduled-job
// governance tables so built-in reconciliation avoids repeated point queries.

package jobmgmt

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/util/gconv"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/startupstats"
)

// startupDataSnapshotContextKey stores scheduled-job startup snapshots in context.
type startupDataSnapshotContextKey struct{}

// startupDataSnapshot contains full-table snapshots for job groups and
// built-in scheduled jobs used during startup reconciliation.
type startupDataSnapshot struct {
	groupsByCode             map[string]*entity.SysJobGroup
	builtinJobsByID          map[uint64]*entity.SysJob
	builtinJobsByHandlerRef  map[string]*entity.SysJob
	builtinJobsByGroupName   map[string]*entity.SysJob
	startupSnapshotAvailable bool
}

// withStartupDataSnapshot returns a child context containing full-table
// snapshots for scheduled-job startup reconciliation.
func (s *serviceImpl) withStartupDataSnapshot(ctx context.Context) (context.Context, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if startupDataSnapshotFromContext(ctx) != nil {
		return ctx, nil
	}

	snapshot, err := s.buildStartupDataSnapshot(ctx)
	if err != nil {
		return ctx, err
	}
	startupstats.Add(ctx, startupstats.CounterJobSnapshotBuilds, 1)
	return context.WithValue(ctx, startupDataSnapshotContextKey{}, snapshot), nil
}

// WithStartupDataSnapshot returns a child context containing scheduled-job
// startup snapshots for host startup orchestration.
func (s *serviceImpl) WithStartupDataSnapshot(ctx context.Context) (context.Context, error) {
	return s.withStartupDataSnapshot(ctx)
}

// buildStartupDataSnapshot loads scheduled-job governance rows in bulk.
func (s *serviceImpl) buildStartupDataSnapshot(ctx context.Context) (*startupDataSnapshot, error) {
	var groups []*entity.SysJobGroup
	if err := dao.SysJobGroup.Ctx(ctx).Scan(&groups); err != nil {
		return nil, err
	}

	var jobs []*entity.SysJob
	if err := dao.SysJob.Ctx(ctx).
		Where(do.SysJob{IsBuiltin: 1}).
		Scan(&jobs); err != nil {
		return nil, err
	}

	snapshot := &startupDataSnapshot{
		groupsByCode:             make(map[string]*entity.SysJobGroup, len(groups)),
		builtinJobsByID:          make(map[uint64]*entity.SysJob, len(jobs)),
		builtinJobsByHandlerRef:  make(map[string]*entity.SysJob, len(jobs)),
		builtinJobsByGroupName:   make(map[string]*entity.SysJob, len(jobs)),
		startupSnapshotAvailable: true,
	}
	for _, group := range groups {
		if group == nil || strings.TrimSpace(group.Code) == "" {
			continue
		}
		snapshot.groupsByCode[strings.TrimSpace(group.Code)] = group
	}
	for _, job := range jobs {
		snapshot.storeBuiltinJob(job)
	}
	return snapshot, nil
}

// startupDataSnapshotFromContext returns the scheduled-job startup snapshot
// stored on the context, if present.
func startupDataSnapshotFromContext(ctx context.Context) *startupDataSnapshot {
	if ctx == nil {
		return nil
	}
	snapshot, ok := ctx.Value(startupDataSnapshotContextKey{}).(*startupDataSnapshot)
	if !ok || snapshot == nil || !snapshot.startupSnapshotAvailable {
		return nil
	}
	return snapshot
}

// groupByCode returns one scheduled-job group from the startup snapshot.
func (s *startupDataSnapshot) groupByCode(code string) *entity.SysJobGroup {
	if s == nil {
		return nil
	}
	return s.groupsByCode[strings.TrimSpace(code)]
}

// builtinJobByHandlerRef returns one built-in job by handler reference.
func (s *startupDataSnapshot) builtinJobByHandlerRef(handlerRef string) *entity.SysJob {
	if s == nil {
		return nil
	}
	return s.builtinJobsByHandlerRef[strings.TrimSpace(handlerRef)]
}

// builtinJobByGroupAndName returns one built-in job by group ID and display name.
func (s *startupDataSnapshot) builtinJobByGroupAndName(groupID uint64, name string) *entity.SysJob {
	if s == nil {
		return nil
	}
	return s.builtinJobsByGroupName[buildBuiltinJobNameKey(groupID, name)]
}

// listBuiltinJobs returns all built-in scheduled jobs from the startup snapshot.
func (s *startupDataSnapshot) listBuiltinJobs() []*entity.SysJob {
	if s == nil {
		return nil
	}
	items := make([]*entity.SysJob, 0, len(s.builtinJobsByID))
	for _, job := range s.builtinJobsByID {
		if job == nil {
			continue
		}
		items = append(items, job)
	}
	return items
}

// storeBuiltinJob records one built-in scheduled job in every startup index.
func (s *startupDataSnapshot) storeBuiltinJob(job *entity.SysJob) {
	if s == nil || job == nil || job.Id == 0 {
		return
	}
	if existing := s.builtinJobsByID[job.Id]; existing != nil {
		s.deleteBuiltinJob(job.Id)
	}
	s.builtinJobsByID[job.Id] = job
	if strings.TrimSpace(job.HandlerRef) != "" {
		s.builtinJobsByHandlerRef[strings.TrimSpace(job.HandlerRef)] = job
	}
	s.builtinJobsByGroupName[buildBuiltinJobNameKey(job.GroupId, job.Name)] = job
}

// deleteBuiltinJob removes one built-in scheduled job from startup indexes.
func (s *startupDataSnapshot) deleteBuiltinJob(jobID uint64) {
	if s == nil || jobID == 0 {
		return
	}
	existing := s.builtinJobsByID[jobID]
	if existing == nil {
		return
	}
	delete(s.builtinJobsByID, jobID)
	if strings.TrimSpace(existing.HandlerRef) != "" {
		delete(s.builtinJobsByHandlerRef, strings.TrimSpace(existing.HandlerRef))
	}
	delete(s.builtinJobsByGroupName, buildBuiltinJobNameKey(existing.GroupId, existing.Name))
}

// buildBuiltinJobEntity creates the startup snapshot projection for one
// scheduled-job row after an insert or update.
func buildBuiltinJobEntity(jobID uint64, record do.SysJob) *entity.SysJob {
	return &entity.SysJob{
		Id:                   jobID,
		GroupId:              gconv.Uint64(record.GroupId),
		Name:                 gconv.String(record.Name),
		Description:          gconv.String(record.Description),
		TaskType:             gconv.String(record.TaskType),
		HandlerRef:           gconv.String(record.HandlerRef),
		Params:               gconv.String(record.Params),
		TimeoutSeconds:       gconv.Int(record.TimeoutSeconds),
		ShellCmd:             gconv.String(record.ShellCmd),
		WorkDir:              gconv.String(record.WorkDir),
		Env:                  gconv.String(record.Env),
		CronExpr:             gconv.String(record.CronExpr),
		Timezone:             gconv.String(record.Timezone),
		Scope:                gconv.String(record.Scope),
		Concurrency:          gconv.String(record.Concurrency),
		MaxConcurrency:       gconv.Int(record.MaxConcurrency),
		MaxExecutions:        gconv.Int(record.MaxExecutions),
		ExecutedCount:        gconv.Int64(record.ExecutedCount),
		StopReason:           gconv.String(record.StopReason),
		LogRetentionOverride: gconv.String(record.LogRetentionOverride),
		Status:               gconv.String(record.Status),
		IsBuiltin:            gconv.Int(record.IsBuiltin),
		SeedVersion:          gconv.Int(record.SeedVersion),
		CreatedBy:            gconv.Int64(record.CreatedBy),
		UpdatedBy:            gconv.Int64(record.UpdatedBy),
	}
}
