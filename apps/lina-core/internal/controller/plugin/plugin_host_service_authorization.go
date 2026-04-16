// This file converts plugin host service authorization DTOs into service inputs
// and API response views.

package plugin

import (
	"strings"

	"lina-core/api/plugin/v1"
	pluginsvc "lina-core/internal/service/plugin"
	"lina-core/pkg/pluginbridge"
)

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
			Service:      strings.TrimSpace(item.Service),
			Methods:      append([]string(nil), item.Methods...),
			Paths:        append([]string(nil), item.Paths...),
			ResourceRefs: append([]string(nil), item.ResourceRefs...),
			Tables:       append([]string(nil), item.Tables...),
		})
	}
	return input
}

func buildHostServicePermissionItems(
	specs []*pluginbridge.HostServiceSpec,
	tableComments map[string]string,
) []*v1.HostServicePermissionItem {
	items := make([]*v1.HostServicePermissionItem, 0, len(specs))
	for _, spec := range specs {
		if spec == nil {
			continue
		}
		item := &v1.HostServicePermissionItem{
			Service: spec.Service,
			Methods: append([]string(nil), spec.Methods...),
			Paths:   append([]string(nil), spec.Paths...),
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
