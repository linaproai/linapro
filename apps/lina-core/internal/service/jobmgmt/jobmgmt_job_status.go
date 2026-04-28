// This file implements scheduled-job status transitions and manual triggering.

package jobmgmt

import (
	"context"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/service/jobmeta"
	"lina-core/pkg/bizerr"
)

// UpdateJobStatus toggles one job between enabled and disabled states.
func (s *serviceImpl) UpdateJobStatus(ctx context.Context, id uint64, status jobmeta.JobStatus) error {
	if status != jobmeta.JobStatusEnabled && status != jobmeta.JobStatusDisabled {
		return bizerr.NewCode(CodeJobStatusToggleInvalid)
	}

	job, err := s.jobByID(ctx, id)
	if err != nil {
		return err
	}
	if job == nil {
		return bizerr.NewCode(jobmeta.CodeJobNotFound)
	}
	if job.IsBuiltin == 1 {
		return bizerr.NewCode(CodeJobBuiltinStatusUpdateDenied)
	}
	if status == jobmeta.JobStatusEnabled {
		if err = s.validateExecutableJob(ctx, job); err != nil {
			return err
		}
	}

	stopReason := ""
	if status == jobmeta.JobStatusDisabled {
		stopReason = string(jobmeta.StopReasonManual)
	}
	_, err = dao.SysJob.Ctx(ctx).
		Where(do.SysJob{Id: id}).
		Data(do.SysJob{
			Status:     string(status),
			StopReason: stopReason,
			UpdatedBy:  s.currentUserID(ctx),
		}).
		Update()
	if err != nil {
		return err
	}
	if s.scheduler == nil {
		return nil
	}
	if status == jobmeta.JobStatusEnabled {
		return s.scheduler.Refresh(ctx, id)
	}
	s.scheduler.Remove(id)
	return nil
}

// ResetJob resets executed_count and stop_reason for one scheduled job.
func (s *serviceImpl) ResetJob(ctx context.Context, id uint64) error {
	job, err := s.jobByID(ctx, id)
	if err != nil {
		return err
	}
	if job == nil {
		return bizerr.NewCode(jobmeta.CodeJobNotFound)
	}
	if job.IsBuiltin == 1 {
		return bizerr.NewCode(CodeJobBuiltinResetDenied)
	}

	_, err = dao.SysJob.Ctx(ctx).
		Where(do.SysJob{Id: id}).
		Data(do.SysJob{
			ExecutedCount: 0,
			StopReason:    "",
			UpdatedBy:     s.currentUserID(ctx),
		}).
		Update()
	if err != nil {
		return err
	}
	if s.scheduler != nil && jobmeta.NormalizeJobStatus(job.Status) == jobmeta.JobStatusEnabled {
		return s.scheduler.Refresh(ctx, id)
	}
	return nil
}

// TriggerJob starts one manual execution and returns the created log ID.
func (s *serviceImpl) TriggerJob(ctx context.Context, id uint64) (uint64, error) {
	if s.scheduler == nil {
		return 0, bizerr.NewCode(CodeJobSchedulerUninitialized)
	}
	return s.scheduler.Trigger(ctx, id)
}
