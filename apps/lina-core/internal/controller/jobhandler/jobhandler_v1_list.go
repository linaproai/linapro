// This file implements the scheduled job handler list endpoint.

package jobhandler

import (
	"context"
	"strings"

	"lina-core/api/jobhandler/v1"
	"lina-core/internal/service/jobmeta"
)

// List handles scheduled job handler list requests.
func (c *ControllerV1) List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error) {
	sourceFilter := jobmeta.NormalizeHandlerSource(req.Source)
	keyword := strings.ToLower(strings.TrimSpace(req.Keyword))
	items := make([]*v1.ListItem, 0)
	for _, item := range c.registry.List() {
		if sourceFilter.IsValid() && item.Source != sourceFilter {
			continue
		}
		if keyword != "" &&
			!strings.Contains(strings.ToLower(item.Ref), keyword) &&
			!strings.Contains(strings.ToLower(item.DisplayName), keyword) {
			continue
		}
		items = append(items, &v1.ListItem{
			Ref:         item.Ref,
			DisplayName: item.DisplayName,
			Description: item.Description,
			Source:      string(item.Source),
			PluginId:    item.PluginID,
		})
	}
	return &v1.ListRes{List: items}, nil
}
