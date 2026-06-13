// This file centralizes defensive deep-copy helpers for management list cache
// entries. The cache path must stay explicit and avoid JSON or gob round trips.

package management

import (
	plugindep "lina-core/internal/service/plugin/internal/dependency"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// CloneListOutput copies one list output and the plugin rows it owns.
func CloneListOutput(in *ListOutput) *ListOutput {
	if in == nil {
		return nil
	}
	out := &ListOutput{
		List:  make([]*PluginItem, 0, len(in.List)),
		Total: in.Total,
	}
	for _, item := range in.List {
		out.List = append(out.List, ClonePluginItem(item))
	}
	return out
}

// ClonePluginItem copies one plugin item while preserving immutable nested
// projections by value where callers may otherwise mutate list rows.
func ClonePluginItem(in *PluginItem) *PluginItem {
	if in == nil {
		return nil
	}
	out := *in
	if in.LastUpgradeFailure != nil {
		lastUpgradeFailure := *in.LastUpgradeFailure
		out.LastUpgradeFailure = &lastUpgradeFailure
	}
	out.RequestedHostServices = cloneHostServiceSpecs(in.RequestedHostServices)
	out.AuthorizedHostServices = cloneHostServiceSpecs(in.AuthorizedHostServices)
	out.DeclaredRoutes = cloneRouteContracts(in.DeclaredRoutes)
	out.DependencyCheck = plugindep.CloneCheckProjection(in.DependencyCheck)
	return &out
}

// cloneHostServiceSpecs deep-copies host-service declarations because list
// consumers may reuse rows while action modals are open.
func cloneHostServiceSpecs(in []*protocol.HostServiceSpec) []*protocol.HostServiceSpec {
	if len(in) == 0 {
		return nil
	}
	out := make([]*protocol.HostServiceSpec, 0, len(in))
	for _, item := range in {
		if item == nil {
			continue
		}
		out = append(out, &protocol.HostServiceSpec{
			Service:   item.Service,
			Methods:   append([]string(nil), item.Methods...),
			Paths:     append([]string(nil), item.Paths...),
			Tables:    append([]string(nil), item.Tables...),
			Keys:      append([]string(nil), item.Keys...),
			Resources: cloneHostServiceResources(item.Resources),
		})
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// cloneHostServiceResources deep-copies governed host-service resource specs.
func cloneHostServiceResources(in []*protocol.HostServiceResourceSpec) []*protocol.HostServiceResourceSpec {
	if len(in) == 0 {
		return nil
	}
	out := make([]*protocol.HostServiceResourceSpec, 0, len(in))
	for _, item := range in {
		if item == nil {
			continue
		}
		out = append(out, &protocol.HostServiceResourceSpec{
			Ref:             item.Ref,
			AllowMethods:    append([]string(nil), item.AllowMethods...),
			HeaderAllowList: append([]string(nil), item.HeaderAllowList...),
			TimeoutMs:       item.TimeoutMs,
			MaxBodyBytes:    item.MaxBodyBytes,
			Attributes:      cloneStringMap(item.Attributes),
		})
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// cloneRouteContracts deep-copies route review declarations.
func cloneRouteContracts(in []*protocol.RouteContract) []*protocol.RouteContract {
	if len(in) == 0 {
		return nil
	}
	out := make([]*protocol.RouteContract, 0, len(in))
	for _, item := range in {
		if item == nil {
			continue
		}
		out = append(out, &protocol.RouteContract{
			Path:        item.Path,
			Method:      item.Method,
			Tags:        append([]string(nil), item.Tags...),
			Summary:     item.Summary,
			Description: item.Description,
			Access:      item.Access,
			Permission:  item.Permission,
			Meta:        cloneStringMap(item.Meta),
			RequestType: item.RequestType,
		})
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// cloneStringMap copies string maps used by cached list projections.
func cloneStringMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}
