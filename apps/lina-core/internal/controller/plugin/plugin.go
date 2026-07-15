// Package plugin contains Lina core plugin management controllers and
// controller-local DTO projection helpers for plugin governance APIs.
package plugin

import (
	"strings"

	v1 "lina-core/api/plugin/v1"
	pluginsvc "lina-core/internal/service/plugin"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// buildPluginDependencyCheckResult converts service dependency results to API DTOs.
func buildPluginDependencyCheckResult(in *pluginsvc.DependencyCheckResult) *v1.PluginDependencyCheckResult {
	if in == nil {
		return nil
	}
	return &v1.PluginDependencyCheckResult{
		TargetId: in.TargetID,
		Framework: v1.PluginDependencyFrameworkCheck{
			RequiredVersion: in.Framework.RequiredVersion,
			CurrentVersion:  in.Framework.CurrentVersion,
			Status:          v1.FrameworkStatus(in.Framework.Status),
		},
		Dependencies:      buildPluginDependencyItems(in.Dependencies),
		Blockers:          buildPluginDependencyBlockers(in.Blockers),
		Cycle:             cloneAPIStringSlice(in.Cycle),
		ReverseDependents: buildPluginDependencyReverseDependents(in.ReverseDependents),
		ReverseBlockers:   buildPluginDependencyBlockers(in.ReverseBlockers),
	}
}

// buildPluginDependencyItems converts dependency edge DTOs.
func buildPluginDependencyItems(items []*pluginsvc.DependencyPluginCheck) []*v1.PluginDependencyItem {
	out := make([]*v1.PluginDependencyItem, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, &v1.PluginDependencyItem{
			OwnerId:         item.OwnerID,
			DependencyId:    item.DependencyID,
			DependencyName:  item.DependencyName,
			RequiredVersion: item.RequiredVersion,
			CurrentVersion:  item.CurrentVersion,
			Installed:       item.Installed,
			Enabled:         item.Enabled,
			Discovered:      item.Discovered,
			Status:          v1.DependencyStatus(item.Status),
			Chain:           cloneAPIStringSlice(item.Chain),
		})
	}
	return out
}

// buildPluginDependencyBlockers converts hard dependency blocker DTOs.
func buildPluginDependencyBlockers(items []*pluginsvc.DependencyBlocker) []*v1.PluginDependencyBlocker {
	out := make([]*v1.PluginDependencyBlocker, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, &v1.PluginDependencyBlocker{
			Code:            v1.BlockerCode(item.Code),
			PluginId:        item.PluginID,
			DependencyId:    item.DependencyID,
			RequiredVersion: item.RequiredVersion,
			CurrentVersion:  item.CurrentVersion,
			Chain:           cloneAPIStringSlice(item.Chain),
			Detail:          item.Detail,
		})
	}
	return out
}

// buildPluginDependencyReverseDependents converts downstream dependency DTOs.
func buildPluginDependencyReverseDependents(items []*pluginsvc.DependencyReverseDependent) []*v1.PluginDependencyReverseDependent {
	out := make([]*v1.PluginDependencyReverseDependent, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, &v1.PluginDependencyReverseDependent{
			PluginId:          item.PluginID,
			Name:              item.Name,
			Version:           item.Version,
			RequiredVersion:   item.RequiredVersion,
			Enabled:           item.Enabled,
			OwnerHostServices: buildPluginDependencyOwnerHostServices(item.OwnerHostServices),
		})
	}
	return out
}

func buildPluginDependencyOwnerHostServices(
	items []*pluginsvc.DependencyOwnerHostServiceSummary,
) []*v1.PluginDependencyOwnerHostService {
	out := make([]*v1.PluginDependencyOwnerHostService, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, &v1.PluginDependencyOwnerHostService{
			Owner:   item.Owner,
			Service: item.Service,
			Version: item.Version,
			Methods: cloneAPIStringSlice(item.Methods),
		})
	}
	return out
}

// cloneAPIStringSlice copies slices before exposing them through API DTOs.
func cloneAPIStringSlice(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	out := make([]string, len(values))
	copy(out, values)
	return out
}

// buildAuthorizationInput converts the API request payload into the service
// input model used by plugin authorization updates.
func buildAuthorizationInput(req *v1.HostServiceAuthorizationReq) *pluginsvc.HostServiceAuthorizationInput {
	if req == nil {
		return nil
	}
	input := &pluginsvc.HostServiceAuthorizationInput{
		Services: make([]*pluginsvc.HostServiceAuthorizationDecision, 0, len(req.Services)),
	}
	for _, item := range req.Services {
		if item == nil {
			continue
		}
		input.Services = append(input.Services, &pluginsvc.HostServiceAuthorizationDecision{
			Owner:        strings.TrimSpace(item.Owner),
			Service:      strings.TrimSpace(item.Service),
			Version:      strings.TrimSpace(item.Version),
			Methods:      append([]string(nil), item.Methods...),
			Paths:        append([]string(nil), item.Paths...),
			Keys:         append([]string(nil), item.Keys...),
			ResourceRefs: append([]string(nil), item.ResourceRefs...),
			Tables:       append([]string(nil), item.Tables...),
		})
	}
	return input
}

// buildHostServicePermissionItems projects host-service specs and resolved table
// comments into the API response view used by plugin detail endpoints.
func buildHostServicePermissionItems(
	specs []*protocol.HostServiceSpec,
	tableComments map[string]string,
) []*v1.HostServicePermissionItem {
	items := make([]*v1.HostServicePermissionItem, 0, len(specs))
	for _, spec := range specs {
		if spec == nil {
			continue
		}
		item := &v1.HostServicePermissionItem{
			Owner:   spec.Owner,
			Service: spec.Service,
			Version: spec.Version,
			Methods: append([]string(nil), spec.Methods...),
			Paths:   append([]string(nil), spec.Paths...),
			Keys:    append([]string(nil), spec.Keys...),
			Tables:  append([]string(nil), spec.Tables...),
			TableItems: buildHostServicePermissionTableItems(
				spec.Tables,
				tableComments,
			),
			Resources: make([]*v1.HostServicePermissionResourceItem, 0, len(spec.Resources)),
		}
		for _, resource := range spec.Resources {
			if resource == nil {
				continue
			}
			item.Resources = append(item.Resources, &v1.HostServicePermissionResourceItem{
				Ref:             resource.Ref,
				AllowMethods:    append([]string(nil), resource.AllowMethods...),
				HeaderAllowList: append([]string(nil), resource.HeaderAllowList...),
				TimeoutMs:       resource.TimeoutMs,
				MaxBodyBytes:    resource.MaxBodyBytes,
				Attributes:      cloneStringMap(resource.Attributes),
			})
		}
		items = append(items, item)
	}
	return items
}

// buildHostServicePermissionTableItems converts authorized table names into the
// table response view, enriching them with best-effort host comments.
func buildHostServicePermissionTableItems(
	tables []string,
	tableComments map[string]string,
) []*v1.HostServicePermissionTableItem {
	if len(tables) == 0 {
		return nil
	}
	items := make([]*v1.HostServicePermissionTableItem, 0, len(tables))
	for _, table := range tables {
		items = append(items, &v1.HostServicePermissionTableItem{
			Name:    table,
			Comment: tableComments[table],
		})
	}
	return items
}

// cloneStringMap copies the resource attribute map so controller responses do
// not alias service-owned state.
func cloneStringMap(source map[string]string) map[string]string {
	if len(source) == 0 {
		return nil
	}
	target := make(map[string]string, len(source))
	for key, value := range source {
		target[key] = value
	}
	return target
}

// buildPluginRouteReviewItems converts current release dynamic route contracts
// into API review items that expose the host-visible public path and key access
// metadata used by governance dialogs.
func buildPluginRouteReviewItems(
	pluginID string,
	routes []*protocol.RouteContract,
) []*v1.PluginRouteReviewItem {
	normalizedPluginID := strings.TrimSpace(pluginID)
	if normalizedPluginID == "" || len(routes) == 0 {
		return nil
	}

	items := make([]*v1.PluginRouteReviewItem, 0, len(routes))
	for _, route := range routes {
		if route == nil {
			continue
		}
		items = append(items, &v1.PluginRouteReviewItem{
			Method:      strings.ToUpper(strings.TrimSpace(route.Method)),
			PublicPath:  pluginsvc.BuildDynamicRoutePublicPath(normalizedPluginID, route.Path),
			Access:      strings.TrimSpace(route.Access),
			Permission:  strings.TrimSpace(route.Permission),
			Summary:     strings.TrimSpace(route.Summary),
			Description: strings.TrimSpace(route.Description),
		})
	}
	if len(items) == 0 {
		return nil
	}
	return items
}
