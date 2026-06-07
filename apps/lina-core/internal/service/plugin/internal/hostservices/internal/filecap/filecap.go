// Package filecap adapts host file metadata to plugin-visible file capability
// contracts without exposing storage paths or host file entities.
package filecap

import (
	"context"

	"lina-core/internal/dao"
	"lina-core/internal/service/plugin/internal/hostservices/internal/domaincap"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/capmodel"
	capabilityfilecap "lina-core/pkg/plugin/capability/filecap"
	"lina-core/pkg/plugin/capability/tenantcap"
)

// Service exposes the file domain service and management commands.
type Service interface {
	capabilityfilecap.Service
	capabilityfilecap.AdminService
}

// adapter exposes governed file projections without leaking sys_file entities.
type adapter struct {
	tenantFilter tenantcap.PluginTableFilterService
}

var (
	_ capabilityfilecap.Service      = (*adapter)(nil)
	_ capabilityfilecap.AdminService = (*adapter)(nil)
)

// New creates the host-owned file capability adapter.
func New(tenantFilter tenantcap.PluginTableFilterService) Service {
	return &adapter{tenantFilter: tenantFilter}
}

// BatchGetFiles returns visible file projections and opaque missing IDs.
func (a *adapter) BatchGetFiles(ctx context.Context, _ capmodel.CapabilityContext, ids []capabilityfilecap.FileID) (*capmodel.BatchResult[*capabilityfilecap.FileProjection, capabilityfilecap.FileID], error) {
	result := &capmodel.BatchResult[*capabilityfilecap.FileProjection, capabilityfilecap.FileID]{
		Items:      make(map[capabilityfilecap.FileID]*capabilityfilecap.FileProjection, len(ids)),
		MissingIDs: []capabilityfilecap.FileID{},
	}
	parsedIDs, requested := domaincap.ParseInt64IDs(ids, func(id capabilityfilecap.FileID) {
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
			Name:          domaincap.FirstNonEmpty(row.Original, row.Name),
			MimeType:      domaincap.MimeTypeFromSuffix(row.Suffix),
			SizeBytes:     row.Size,
			BusinessScene: row.Scene,
		}
	}
	for _, id := range ids {
		if _, ok := result.Items[id]; !ok && !domaincap.Contains(result.MissingIDs, id) {
			result.MissingIDs = append(result.MissingIDs, id)
		}
	}
	return result, nil
}

// EnsureFilesVisible rejects when any requested file is absent or invisible.
func (a *adapter) EnsureFilesVisible(ctx context.Context, capCtx capmodel.CapabilityContext, ids []capabilityfilecap.FileID) error {
	result, err := a.BatchGetFiles(ctx, capCtx, ids)
	if err != nil {
		return err
	}
	if len(result.MissingIDs) > 0 {
		return bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	return nil
}

// DeleteFiles soft-deletes visible file metadata rows.
func (a *adapter) DeleteFiles(ctx context.Context, capCtx capmodel.CapabilityContext, ids []capabilityfilecap.FileID) error {
	if err := a.EnsureFilesVisible(ctx, capCtx, ids); err != nil {
		return err
	}
	parsedIDs, _ := domaincap.ParseInt64IDs(ids, nil)
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
