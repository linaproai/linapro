// This file defines the scheduled job log controller dependencies and constructor.

package joblog

import (
	"lina-core/api/joblog"
	"lina-core/internal/service/bizctx"
	jobmgmtsvc "lina-core/internal/service/jobmgmt"
	pluginsvc "lina-core/internal/service/plugin"
	rolesvc "lina-core/internal/service/role"
)

// ControllerV1 defines the v1 scheduled job log controller.
type ControllerV1 struct {
	bizCtxSvc  bizctx.Service     // bizCtxSvc resolves the current operator identity.
	jobMgmtSvc jobmgmtsvc.Service // jobMgmtSvc handles scheduled-job logs and execution control.
	roleSvc    rolesvc.Service    // roleSvc loads current permissions for shell-cancel checks.
}

// NewV1 creates and returns the v1 scheduled job log controller.
func NewV1(jobMgmtSvc jobmgmtsvc.Service) joblog.IJoblogV1 {
	return &ControllerV1{
		bizCtxSvc:  bizctx.New(),
		jobMgmtSvc: jobMgmtSvc,
		roleSvc:    rolesvc.New(pluginsvc.Instance()),
	}
}
