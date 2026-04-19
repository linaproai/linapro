// This file defines the scheduled job controller skeleton and constructor.

package job

import (
	"lina-core/api/job"
)

// ControllerV1 defines the v1 scheduled job controller.
type ControllerV1 struct{}

// NewV1 creates and returns the v1 scheduled job controller.
func NewV1() job.IJobV1 {
	return &ControllerV1{}
}
