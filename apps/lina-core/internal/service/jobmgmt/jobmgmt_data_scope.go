// This file applies shared data-scope rules to scheduled jobs and their logs.

package jobmgmt

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/internal/dao"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/datascope"
	"lina-core/internal/service/jobmeta"
	"lina-core/pkg/bizerr"
)

// applyJobDataScope keeps built-in jobs visible and filters user-created jobs
// by created_by.
func (s *serviceImpl) applyJobDataScope(ctx context.Context, model *gdb.Model) (*gdb.Model, error) {
	model = datascope.ApplyTenantScope(ctx, model, dao.SysJob.Table()+"."+datascope.TenantColumn)
	scopedModel, _, err := s.currentScopeSvc().ApplyUserScopeWithBypass(
		ctx,
		model,
		qualifiedSysJobCreatedByColumn(),
		qualifiedSysJobIsBuiltinColumn(),
		1,
	)
	if err != nil {
		return nil, mapJobDataScopeError(err)
	}
	return scopedModel, nil
}

// applyJobLogDataScope filters logs by the visibility of their owning job.
func (s *serviceImpl) applyJobLogDataScope(ctx context.Context, model *gdb.Model) (*gdb.Model, error) {
	model = datascope.ApplyTenantScope(ctx, model, dao.SysJobLog.Table()+"."+datascope.TenantColumn)
	jobCols := dao.SysJob.Columns()
	subQuery := dao.SysJob.Ctx(ctx).
		Fields(jobCols.Id).
		Where(qualifiedSysJobIDColumn() + " = " + qualifiedSysJobLogJobIDColumn())
	scopedSubQuery, err := s.applyJobDataScope(ctx, subQuery)
	if err != nil {
		return nil, err
	}
	return model.Where("EXISTS ?", scopedSubQuery), nil
}

// ensureJobVisible verifies one job entity is inside the current data scope.
func (s *serviceImpl) ensureJobVisible(ctx context.Context, job *entity.SysJob) error {
	if job == nil {
		return bizerr.NewCode(jobmeta.CodeJobNotFound)
	}
	return s.ensureJobsVisibleByID(ctx, []int64{job.Id})
}

// ensureJobsVisibleByID verifies all selected jobs are visible.
func (s *serviceImpl) ensureJobsVisibleByID(ctx context.Context, ids []int64) error {
	normalizedIDs := normalizeJobIDs(ids)
	if len(normalizedIDs) == 0 {
		return nil
	}
	model := dao.SysJob.Ctx(ctx).WhereIn(dao.SysJob.Columns().Id, normalizedIDs)
	scopedModel, err := s.applyJobDataScope(ctx, model)
	if err != nil {
		return err
	}
	count, err := scopedModel.Count()
	if err != nil {
		return err
	}
	if count != len(normalizedIDs) {
		return bizerr.NewCode(CodeJobDataScopeDenied)
	}
	return nil
}

// ensureLogVisible verifies one log entity is inside the current data scope.
func (s *serviceImpl) ensureLogVisible(ctx context.Context, logRow *entity.SysJobLog) error {
	if logRow == nil {
		return bizerr.NewCode(CodeJobLogNotFound)
	}
	return s.ensureLogsVisible(ctx, []int64{logRow.Id})
}

// ensureLogsVisible verifies all selected logs are visible through their jobs.
func (s *serviceImpl) ensureLogsVisible(ctx context.Context, ids []int64) error {
	normalizedIDs := normalizeJobIDs(ids)
	if len(normalizedIDs) == 0 {
		return nil
	}
	model := dao.SysJobLog.Ctx(ctx).WhereIn(dao.SysJobLog.Columns().Id, normalizedIDs)
	scopedModel, err := s.applyJobLogDataScope(ctx, model)
	if err != nil {
		return err
	}
	count, err := scopedModel.Count()
	if err != nil {
		return err
	}
	if count != len(normalizedIDs) {
		return bizerr.NewCode(CodeJobDataScopeDenied)
	}
	return nil
}

// currentScopeSvc returns the shared data-scope service for job operations.
func (s *serviceImpl) currentScopeSvc() datascope.Service {
	if s != nil && s.scopeSvc != nil {
		return s.scopeSvc
	}
	return nil
}

// mapJobDataScopeError maps shared data-scope errors to scheduled-job errors.
func mapJobDataScopeError(err error) error {
	if err == nil {
		return nil
	}
	if bizerr.Is(err, datascope.CodeDataScopeDenied) {
		return bizerr.NewCode(CodeJobDataScopeDenied)
	}
	return err
}

// normalizeJobIDs removes invalid and duplicate job or log IDs.
func normalizeJobIDs(ids []int64) []int64 {
	result := make([]int64, 0, len(ids))
	seen := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

// qualifiedSysJobCreatedByColumn returns the fully qualified job owner column.
func qualifiedSysJobCreatedByColumn() string {
	return dao.SysJob.Table() + "." + dao.SysJob.Columns().CreatedBy
}

// qualifiedSysJobIsBuiltinColumn returns the fully qualified built-in flag column.
func qualifiedSysJobIsBuiltinColumn() string {
	return dao.SysJob.Table() + "." + dao.SysJob.Columns().IsBuiltin
}

// qualifiedSysJobIDColumn returns the fully qualified job ID column.
func qualifiedSysJobIDColumn() string {
	return dao.SysJob.Table() + "." + dao.SysJob.Columns().Id
}

// qualifiedSysJobLogJobIDColumn returns the fully qualified log owner job column.
func qualifiedSysJobLogJobIDColumn() string {
	return dao.SysJobLog.Table() + "." + dao.SysJobLog.Columns().JobId
}
