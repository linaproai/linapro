// This file defines the scheduled job handler controller dependencies and constructor.

package jobhandler

import (
	"lina-core/api/jobhandler"
	jobhandlersvc "lina-core/internal/service/jobhandler"
)

// ControllerV1 defines the v1 scheduled job handler controller.
type ControllerV1 struct {
	registry jobhandlersvc.Registry // registry exposes registered handler metadata.
}

// NewV1 creates and returns the v1 scheduled job handler controller.
func NewV1(registry jobhandlersvc.Registry) jobhandler.IJobhandlerV1 {
	return &ControllerV1{registry: registry}
}
