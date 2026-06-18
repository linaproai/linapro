// This file adapts host file metadata to plugin-visible file capability
// contracts without exposing storage paths or host file entities.
package capabilityhost

import (
	"context"
	"fmt"
	"strings"

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

// Search returns one bounded page of visible file projections.
func (a *fileCapabilityAdapter) Search(ctx context.Context, _ capmodel.CapabilityContext, input capabilityfilecap.SearchInput) (*capmodel.PageResult[*capabilityfilecap.FileProjection], error) {
	pageNum, pageSize := NormalizePage(input.Page)
	if pageSize > capabilityfilecap.MaxSearchPageSize {
		pageSize = capabilityfilecap.MaxSearchPageSize
	}
	cols := dao.SysFile.Columns()
	model := dao.SysFile.Ctx(ctx)
	if a != nil && a.tenantFilter != nil {
		model = a.tenantFilter.Apply(ctx, model, "")
	}
	if scene := strings.TrimSpace(input.BusinessScene); scene != "" {
		model = model.Where(cols.Scene, scene)
	}
	if keyword := strings.TrimSpace(input.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		model = model.Where(
			fmt.Sprintf("(%s LIKE ? OR %s LIKE ?)", cols.Original, cols.Name),
			like,
			like,
		)
	}
	if mimeType := strings.TrimSpace(input.MimeType); mimeType != "" {
		suffixes := suffixesForMimeType(mimeType)
		if len(suffixes) == 0 {
			return &capmodel.PageResult[*capabilityfilecap.FileProjection]{Items: []*capabilityfilecap.FileProjection{}, Total: 0}, nil
		}
		model = model.WhereIn(cols.Suffix, suffixes)
	}
	total, err := model.Clone().Count()
	if err != nil {
		return nil, err
	}
	rows := make([]*struct {
		Id       int64
		Original string
		Name     string
		Suffix   string
		Size     int64
		Scene    string
	}, 0, pageSize)
	if err = model.Clone().
		Fields(cols.Id, cols.Original, cols.Name, cols.Suffix, cols.Size, cols.Scene).
		Page(pageNum, pageSize).
		OrderDesc(cols.Id).
		Scan(&rows); err != nil {
		return nil, err
	}
	items := make([]*capabilityfilecap.FileProjection, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		items = append(items, &capabilityfilecap.FileProjection{
			ID:            capabilityfilecap.FileID(fmt.Sprintf("%d", row.Id)),
			Name:          FirstNonEmpty(row.Original, row.Name),
			MimeType:      MimeTypeFromSuffix(row.Suffix),
			SizeBytes:     row.Size,
			BusinessScene: row.Scene,
		})
	}
	return &capmodel.PageResult[*capabilityfilecap.FileProjection]{Items: items, Total: total}, nil
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

// suffixesForMimeType maps stable coarse MIME filters to file suffixes.
func suffixesForMimeType(mimeType string) []string {
	switch strings.ToLower(strings.TrimSpace(mimeType)) {
	case "image/jpeg":
		return []string{"jpg", "jpeg"}
	case "image/png":
		return []string{"png"}
	case "image/gif":
		return []string{"gif"}
	case "application/pdf":
		return []string{"pdf"}
	case "text/plain":
		return []string{"txt", "log"}
	case "application/json":
		return []string{"json"}
	default:
		return nil
	}
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
