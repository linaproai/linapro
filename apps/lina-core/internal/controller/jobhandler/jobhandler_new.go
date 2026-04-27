// This file defines the scheduled job handler controller dependencies and constructor.

package jobhandler

import (
	"lina-core/api/jobhandler"
	i18nsvc "lina-core/internal/service/i18n"
	jobhandlersvc "lina-core/internal/service/jobhandler"
)

// ControllerV1 defines the v1 scheduled job handler controller.
type ControllerV1 struct {
	registry jobhandlersvc.Registry // registry exposes registered handler metadata.
	i18nSvc  i18nsvc.Service        // i18nSvc localizes handler display metadata.
}

// NewV1 creates and returns the v1 scheduled job handler controller.
func NewV1(registry jobhandlersvc.Registry) jobhandler.IJobhandlerV1 {
	return &ControllerV1{
		registry: registry,
		i18nSvc:  i18nsvc.New(),
	}
}
