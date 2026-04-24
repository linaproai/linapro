// This file localizes config-management display metadata using stable config keys.

package sysconfig

import (
	"context"
	"strings"

	"lina-core/internal/model/entity"
)

// localizeConfigEntities localizes one config-entity list in place.
func (s *serviceImpl) localizeConfigEntities(ctx context.Context, items []*entity.SysConfig) {
	for _, item := range items {
		s.localizeConfigEntity(ctx, item)
	}
}

// localizeConfigEntity localizes one config entity in place.
func (s *serviceImpl) localizeConfigEntity(ctx context.Context, item *entity.SysConfig) {
	if s == nil || s.i18nSvc == nil || item == nil {
		return
	}
	trimmedKey := strings.TrimSpace(item.Key)
	if trimmedKey == "" {
		return
	}
	item.Name = s.i18nSvc.Translate(ctx, "config."+trimmedKey+".name", item.Name)
	item.Remark = s.i18nSvc.Translate(ctx, "config."+trimmedKey+".remark", item.Remark)
}
