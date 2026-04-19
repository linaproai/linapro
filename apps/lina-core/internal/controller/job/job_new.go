// This file defines the scheduled job controller dependencies and constructor.

package job

import (
	"lina-core/api/job"
	jobmgmtsvc "lina-core/internal/service/jobmgmt"
)

// ControllerV1 defines the v1 scheduled job controller.
type ControllerV1 struct {
	jobMgmtSvc jobmgmtsvc.Service // jobMgmtSvc handles scheduled-job persistence and execution.
}

// NewV1 creates and returns the v1 scheduled job controller.
func NewV1(jobMgmtSvc jobmgmtsvc.Service) job.IJobV1 {
	return &ControllerV1{jobMgmtSvc: jobMgmtSvc}
}
