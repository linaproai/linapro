// This file adapts host file metadata to plugin-visible file capability
// contracts without exposing storage paths or host file entities.
package capabilityhost

import (
	"context"

	"lina-core/internal/dao"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/capmodel"
	capabilityfilecap "lina-core/pkg/plugin/capability/filecap"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
)

// Service exposes the file domain service and management commands.
type fileCapabilityService interface {
	capabilityfilecap.Service
	capabilityfilecap.AdminService
}

// adapter exposes governed file projections without leaking sys_file entities.
type fileCapabilityAdapter struct {
	tenantFilter tenantspi.PluginTableFilterService
}

var (
	_ capabilityfilecap.Service      = (*fileCapabilityAdapter)(nil)
	_ capabilityfilecap.AdminService = (*fileCapabilityAdapter)(nil)
)

// New creates the host-owned file capability adapter.
func newFileCapabilityAdapter(tenantFilter tenantspi.PluginTableFilterService) fileCapabilityService {
	return &fileCapabilityAdapter{tenantFilter: tenantFilter}
}

// BatchGet returns visible file projections and opaque missing IDs.
func (a *fileCapabilityAdapter) BatchGet(ctx context.Context, _ capmodel.CapabilityContext, ids []capabilityfilecap.FileID) (*capmodel.BatchResult[*capabilityfilecap.FileProjection, capabilityfilecap.FileID], error) {
	result := &capmodel.BatchResult[*capabilityfilecap.FileProjection, capabilityfilecap.FileID]{
		Items:      make(map[capabilityfilecap.FileID]*capabilityfilecap.FileProjection, len(ids)),
		MissingIDs: []capabilityfilecap.FileID{},
	}
	parsedIDs, requested := ParseInt64IDs(ids, func(id capabilityfilecap.FileID) {
		result.MissingIDs = append(result.MissingIDs, id)
	})
	if len(parsedIDs) == 0 {
		return result, nil
	}
	rows := make([]*struct {
		Id       int64
		Original string
		Name     string
		Suffix   string
		Size     int64
		Scene    string
	}, 0, len(parsedIDs))
	cols := dao.SysFile.Columns()
	model := dao.SysFile.Ctx(ctx).
		Fields(cols.Id, cols.Original, cols.Name, cols.Suffix, cols.Size, cols.Scene).
		WhereIn(cols.Id, parsedIDs)
	if a != nil && a.tenantFilter != nil {
		model = a.tenantFilter.Apply(ctx, model, "")
	}
	if err := model.Scan(&rows); err != nil {
		return nil, err
	}
	for _, row := range rows {
		if row == nil {
			continue
		}
		requestID, ok := requested[row.Id]
		if !ok {
			continue
		}
		result.Items[requestID] = &capabilityfilecap.FileProjection{
			ID:            requestID,
			Name:          FirstNonEmpty(row.Original, row.Name),
			MimeType:      MimeTypeFromSuffix(row.Suffix),
			SizeBytes:     row.Size,
			BusinessScene: row.Scene,
		}
	}
	for _, id := range ids {
		if _, ok := result.Items[id]; !ok && !Contains(result.MissingIDs, id) {
			result.MissingIDs = append(result.MissingIDs, id)
		}
	}
	return result, nil
}

// EnsureVisible rejects when any requested file is absent or invisible.
func (a *fileCapabilityAdapter) EnsureVisible(ctx context.Context, capCtx capmodel.CapabilityContext, ids []capabilityfilecap.FileID) error {
	result, err := a.BatchGet(ctx, capCtx, ids)
	if err != nil {
		return err
	}
	if len(result.MissingIDs) > 0 {
		return bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	return nil
}

// Delete soft-deletes visible file metadata rows.
func (a *fileCapabilityAdapter) Delete(ctx context.Context, capCtx capmodel.CapabilityContext, ids []capabilityfilecap.FileID) error {
	if err := a.EnsureVisible(ctx, capCtx, ids); err != nil {
		return err
	}
	parsedIDs, _ := ParseInt64IDs(ids, nil)
	if len(parsedIDs) == 0 {
		return bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	model := dao.SysFile.Ctx(ctx).WhereIn(dao.SysFile.Columns().Id, parsedIDs)
	if a != nil && a.tenantFilter != nil {
		model = a.tenantFilter.Apply(ctx, model, "")
	}
	_, err := model.Delete()
	return err
}
