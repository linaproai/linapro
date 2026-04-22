// Package dynamic implements the dynamic plugin route controllers.

package dynamic

import (
	dynamicapi "lina-plugin-demo-dynamic/backend/api/dynamic"
	dynamicservice "lina-plugin-demo-dynamic/backend/internal/service/dynamic"
)

// Interface compliance assertion for the typed dynamic plugin route controller.
var _ dynamicapi.IDynamicV1 = (*Controller)(nil)

// Controller handles dynamic plugin route requests.
type Controller struct {
	dynamicSvc dynamicservice.Service
}

// New creates and returns a new dynamic plugin controller instance.
func New() *Controller {
	return &Controller{
		dynamicSvc: dynamicservice.New(),
	}
}
