// This file clones and projects dynamic-route review metadata so management
// responses can expose current release route declarations without aliasing
// manifest-owned state.

package runtime

import bridgecontract "lina-core/pkg/pluginbridge/contract"

// cloneRouteContracts deep-copies dynamic route contracts for management-list
// projections so response models do not alias catalog manifest slices.
func cloneRouteContracts(routes []*bridgecontract.RouteContract) []*bridgecontract.RouteContract {
	if len(routes) == 0 {
		return nil
	}

	items := make([]*bridgecontract.RouteContract, 0, len(routes))
	for _, route := range routes {
		if route == nil {
			continue
		}
		items = append(items, &bridgecontract.RouteContract{
			Path:        route.Path,
			Method:      route.Method,
			Tags:        append([]string(nil), route.Tags...),
			Summary:     route.Summary,
			Description: route.Description,
			Access:      route.Access,
			Permission:  route.Permission,
			Meta:        cloneStringMap(route.Meta),
			RequestType: route.RequestType,
		})
	}
	if len(items) == 0 {
		return nil
	}
	return items
}
