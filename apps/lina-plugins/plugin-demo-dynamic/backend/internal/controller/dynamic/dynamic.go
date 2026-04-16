// Package dynamic implements the dynamic plugin route controllers.

package dynamic

import dynamicservice "lina-plugin-demo-dynamic/backend/internal/service/dynamic"

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
