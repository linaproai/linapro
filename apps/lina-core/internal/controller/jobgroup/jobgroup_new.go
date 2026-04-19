// This file defines the scheduled job group controller skeleton and constructor.

package jobgroup

import (
	"lina-core/api/jobgroup"
)

// ControllerV1 defines the v1 scheduled job group controller.
type ControllerV1 struct{}

// NewV1 creates and returns the v1 scheduled job group controller.
func NewV1() jobgroup.IJobgroupV1 {
	return &ControllerV1{}
}
