// This file synchronizes scheduled-job persistence state with handler-registry
// availability changes.

package jobmgmt

import (
	"context"
	jobv1 "lina-core/api/job/v1"
	"strings"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/jobmeta"
	"lina-core/pkg/bizerr"
)

// syncHandlerAvailability pauses or resumes handler jobs after one registry mutation.
func (s *serviceImpl) syncHandlerAvailability(
	ctx context.Context,
	ref string,
	exists bool,
) error {
	trimmedRef := strings.TrimSpace(ref)
	if s == nil || trimmedRef == "" {
		return nil
	}
	if exists {
		return s.resumeJobsForHandler(ctx, trimmedRef)
	}
	return s.pauseJobsForHandler(ctx, trimmedRef)
}

// pauseJobsForHandler marks enabled user-defined handler jobs as paused when
// their handler disappears. Built-in plugin jobs are projected and scheduled by
// the plugin lifecycle synchronization path.
func (s *serviceImpl) pauseJobsForHandler(ctx context.Context, ref string) error {
	jobIDs, err := s.matchingJobIDs(ctx, ref, jobv1.StatusEnabled, "")
	if err != nil {
		return err
	}
	if len(jobIDs) == 0 {
		return nil
	}

	_, err = dao.SysJob.Ctx(ctx).
		Where(do.SysJob{
			IsBuiltin: 0,
			TaskType:  string(jobv1.TaskTypeHandler),
			Status:    string(jobv1.StatusEnabled),
		}).
		Where(dao.SysJob.Columns().HandlerRef, ref).
		Data(do.SysJob{
			Status:     string(jobv1.StatusPausedByPlugin),
			StopReason: string(jobmeta.StopReasonPluginUnavailable),
		}).
		Update()
	if err != nil {
		return err
	}

	if s.scheduler != nil {
		for _, jobID := range jobIDs {
			s.scheduler.Remove(jobID)
		}
	}
	return nil
}

// resumeJobsForHandler restores previously paused user-defined handler jobs
// when the handler returns. Built-in plugin jobs are restored by declaration
// synchronization so registry callbacks do not register them from sys_job.
func (s *serviceImpl) resumeJobsForHandler(ctx context.Context, ref string) error {
	jobIDs, err := s.matchingJobIDs(
		ctx,
		ref,
		jobv1.StatusPausedByPlugin,
		string(jobmeta.StopReasonPluginUnavailable),
	)
	if err != nil {
		return err
	}
	if len(jobIDs) == 0 {
		return nil
	}

	_, err = dao.SysJob.Ctx(ctx).
		Where(do.SysJob{
			IsBuiltin:  0,
			TaskType:   string(jobv1.TaskTypeHandler),
			Status:     string(jobv1.StatusPausedByPlugin),
			StopReason: string(jobmeta.StopReasonPluginUnavailable),
		}).
		Where(dao.SysJob.Columns().HandlerRef, ref).
		Data(do.SysJob{
			Status:     string(jobv1.StatusEnabled),
			StopReason: "",
		}).
		Update()
	if err != nil {
		return err
	}

	if s.scheduler != nil {
		for _, jobID := range jobIDs {
			if err = s.scheduler.Refresh(ctx, jobID); err != nil {
				return err
			}
		}
	}
	return nil
}

// matchingJobIDs returns job identifiers that match one handler-ref and availability state.
func (s *serviceImpl) matchingJobIDs(
	ctx context.Context,
	ref string,
	status jobv1.Status,
	stopReason string,
) ([]int64, error) {
	if !status.IsValid() {
		return nil, bizerr.NewCode(CodeJobStatusInvalid)
	}

	cols := dao.SysJob.Columns()
	model := dao.SysJob.Ctx(ctx).
		Where(do.SysJob{
			IsBuiltin: 0,
			TaskType:  string(jobv1.TaskTypeHandler),
			Status:    string(status),
		}).
		Where(cols.HandlerRef, ref)
	if strings.TrimSpace(stopReason) != "" {
		model = model.Where(cols.StopReason, strings.TrimSpace(stopReason))
	}

	var jobs []*entity.SysJob
	if err := model.Fields(cols.Id).Scan(&jobs); err != nil {
		return nil, err
	}

	result := make([]int64, 0, len(jobs))
	for _, job := range jobs {
		if job == nil || job.Id == 0 {
			continue
		}
		result = append(result, job.Id)
	}
	return result, nil
}
