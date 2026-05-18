// This file defines the scheduled job group controller dependencies and constructor.

package jobgroup

import (
	"lina-core/api/jobgroup"
	jobmgmtsvc "lina-core/internal/service/jobmgmt"
)

// ControllerV1 defines the v1 scheduled job group controller.
type ControllerV1 struct {
	jobMgmtSvc jobmgmtsvc.Service // jobMgmtSvc handles scheduled-job group persistence.
}

// NewV1 creates and returns the v1 scheduled job group controller.
func NewV1(jobMgmtSvc jobmgmtsvc.Service) jobgroup.IJobgroupV1 {
	return &ControllerV1{jobMgmtSvc: jobMgmtSvc}
}
