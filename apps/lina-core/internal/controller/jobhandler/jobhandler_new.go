// This file defines the scheduled job handler controller skeleton and constructor.

package jobhandler

import (
	"lina-core/api/jobhandler"
)

// ControllerV1 defines the v1 scheduled job handler controller.
type ControllerV1 struct{}

// NewV1 creates and returns the v1 scheduled job handler controller.
func NewV1() jobhandler.IJobhandlerV1 {
	return &ControllerV1{}
}
