// This file implements the scheduled job log cancel endpoint.

package joblog

import (
	"context"
	"encoding/json"
	"strings"

	"lina-core/api/joblog/v1"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/jobmeta"
	jobmgmtsvc "lina-core/internal/service/jobmgmt"
	"lina-core/pkg/bizerr"
)

// Cancel handles scheduled job log cancellation requests.
func (c *ControllerV1) Cancel(ctx context.Context, req *v1.CancelReq) (res *v1.CancelRes, err error) {
	logDetail, err := c.jobMgmtSvc.GetLog(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	if isShellLog(logDetail.SysJobLog) {
		if err = c.ensureShellCancelPermission(ctx); err != nil {
			return nil, err
		}
	}
	if err = c.jobMgmtSvc.CancelLog(ctx, req.Id); err != nil {
		return nil, err
	}
	return &v1.CancelRes{}, nil
}

// ensureShellCancelPermission verifies the current operator also holds system:job:shell.
func (c *ControllerV1) ensureShellCancelPermission(ctx context.Context) error {
	businessCtx := c.bizCtxSvc.Get(ctx)
	if businessCtx == nil || businessCtx.UserId <= 0 {
		return bizerr.NewCode(jobmgmtsvc.CodeJobLogCurrentUserMissing)
	}
	accessContext, err := c.roleSvc.GetUserAccessContext(ctx, businessCtx.UserId)
	if err != nil {
		return err
	}
	if accessContext == nil || accessContext.IsSuperAdmin {
		return nil
	}
	for _, permission := range accessContext.Permissions {
		if strings.TrimSpace(permission) == "system:job:shell" {
			return nil
		}
	}
	return bizerr.NewCode(
		jobmgmtsvc.CodeJobLogShellCancelPermissionDenied,
		bizerr.P("permission", "system:job:shell"),
	)
}

// isShellLog reports whether the stored job snapshot describes one shell task.
func isShellLog(logRow *entity.SysJobLog) bool {
	if logRow == nil {
		return false
	}
	var snapshot struct {
		TaskType string `json:"taskType"`
	}
	if err := json.Unmarshal([]byte(logRow.JobSnapshot), &snapshot); err != nil {
		return false
	}
	return jobmeta.NormalizeTaskType(snapshot.TaskType) == jobmeta.TaskTypeShell
}
