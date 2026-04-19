// This file defines the scheduled job log controller skeleton and constructor.

package joblog

import (
	"lina-core/api/joblog"
)

// ControllerV1 defines the v1 scheduled job log controller.
type ControllerV1 struct{}

// NewV1 creates and returns the v1 scheduled job log controller.
func NewV1() joblog.IJoblogV1 {
	return &ControllerV1{}
}
