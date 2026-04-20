// This file converts plugin host service authorization DTOs into service inputs
// and API response views.

package plugin

import (
	"sort"
	"strings"

	"lina-core/api/plugin/v1"
	pluginsvc "lina-core/internal/service/plugin"
	"lina-core/pkg/pluginbridge"
)

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
			Service:      strings.TrimSpace(item.Service),
			Methods:      append([]string(nil), item.Methods...),
			Paths:        append([]string(nil), item.Paths...),
			ResourceRefs: append([]string(nil), item.ResourceRefs...),
			Tables:       append([]string(nil), item.Tables...),
		})
	}
	return input
}

// buildHostServicePermissionItems projects host-service specs and resolved table
// comments into the API response view used by plugin detail endpoints.
func buildHostServicePermissionItems(
	specs []*pluginbridge.HostServiceSpec,
	tableComments map[string]string,
	cronJobs []pluginsvc.ManagedCronJob,
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
			CronItems: buildHostServicePermissionCronItems(spec.Service, cronJobs),
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

// buildHostServicePermissionCronItems converts discovered managed cron jobs into
// one API response view for the cron host service block.
func buildHostServicePermissionCronItems(
	service string,
	cronJobs []pluginsvc.ManagedCronJob,
) []*v1.HostServicePermissionCronItem {
	if service != pluginbridge.HostServiceCron || len(cronJobs) == 0 {
		return nil
	}

	items := make([]*v1.HostServicePermissionCronItem, 0, len(cronJobs))
	for _, cronJob := range cronJobs {
		items = append(items, &v1.HostServicePermissionCronItem{
			Name:           cronJob.Name,
			DisplayName:    cronJob.DisplayName,
			Description:    cronJob.Description,
			Pattern:        cronJob.Pattern,
			Timezone:       cronJob.Timezone,
			Scope:          string(cronJob.Scope),
			Concurrency:    string(cronJob.Concurrency),
			MaxConcurrency: cronJob.MaxConcurrency,
		})
	}
	sort.Slice(items, func(i, j int) bool {
		leftKey := strings.ToLower(strings.TrimSpace(items[i].DisplayName))
		if leftKey == "" {
			leftKey = strings.ToLower(strings.TrimSpace(items[i].Name))
		}
		rightKey := strings.ToLower(strings.TrimSpace(items[j].DisplayName))
		if rightKey == "" {
			rightKey = strings.ToLower(strings.TrimSpace(items[j].Name))
		}
		if leftKey != rightKey {
			return leftKey < rightKey
		}
		return strings.TrimSpace(items[i].Name) < strings.TrimSpace(items[j].Name)
	})
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
