// This file adapts host file metadata to plugin-visible file capability
// contracts without exposing storage paths or host file entities.
package file

import (
	"context"
	"fmt"
	"path"
	"strings"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/closeutil"
	"lina-core/pkg/plugin/capability/capmodel"
	capabilityfilecap "lina-core/pkg/plugin/capability/filecap"
	"lina-core/pkg/plugin/capability/storagecap"
	"lina-core/pkg/plugin/capability/tenantcap"
	"lina-core/pkg/plugin/capability/tenantcap/tenantspi"
)

// adapter exposes governed file projections without leaking sys_file entities.
type fileCapabilityAdapter struct {
	owner        Service
	tenantFilter tenantcap.FilterService
	storage      storagecap.Service
}

// CapabilityService is the file-owned plugin capability adapter contract. The
// extra storage binding is used only when a plugin-scoped directory can provide
// the caller's private Storage() view.
type CapabilityService interface {
	capabilityfilecap.Service
	// WithStorage returns a plugin-scoped file capability that can copy objects
	// from the supplied storage view.
	WithStorage(storage storagecap.Service) capabilityfilecap.Service
}

var _ capabilityfilecap.Service = (*fileCapabilityAdapter)(nil)
var _ CapabilityService = (*fileCapabilityAdapter)(nil)

// NewCapabilityAdapter creates the host-owned file capability adapter.
func NewCapabilityAdapter(owner Service, tenantFilter tenantcap.FilterService) CapabilityService {
	return &fileCapabilityAdapter{owner: owner, tenantFilter: tenantFilter}
}

// WithStorage returns a plugin-scoped file adapter that can copy from the
// caller's private Storage() view without making the file owner depend on it.
func (a *fileCapabilityAdapter) WithStorage(storage storagecap.Service) capabilityfilecap.Service {
	if a == nil {
		return &fileCapabilityAdapter{storage: storage}
	}
	return &fileCapabilityAdapter{
		owner:        a.owner,
		tenantFilter: a.tenantFilter,
		storage:      storage,
	}
}

// Get returns one visible file projection.
func (a *fileCapabilityAdapter) Get(ctx context.Context, id capabilityfilecap.FileID) (*capabilityfilecap.FileInfo, error) {
	result, err := a.BatchGet(ctx, []capabilityfilecap.FileID{id})
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	if projection := result.Items[id]; projection != nil {
		return projection, nil
	}
	return nil, bizerr.NewCode(capmodel.CodeCapabilityDenied)
}

// Detail returns one visible file detail projection.
func (a *fileCapabilityAdapter) Detail(ctx context.Context, id capabilityfilecap.FileID) (*capabilityfilecap.DetailInfo, error) {
	if a == nil || a.owner == nil {
		projection, err := a.Get(ctx, id)
		if err != nil {
			return nil, err
		}
		return &capabilityfilecap.DetailInfo{FileInfo: *projection}, nil
	}
	parsedID, err := parseFileID(id)
	if err != nil {
		return nil, err
	}
	if err = a.EnsureVisible(ctx, []capabilityfilecap.FileID{id}); err != nil {
		return nil, err
	}
	detail, err := a.owner.Detail(ctx, parsedID)
	if err != nil {
		return nil, err
	}
	return fileDetailProjection(detail, id), nil
}

// BatchGet returns visible file projections and opaque missing IDs.
func (a *fileCapabilityAdapter) BatchGet(ctx context.Context, ids []capabilityfilecap.FileID) (*capmodel.BatchResult[*capabilityfilecap.FileInfo, capabilityfilecap.FileID], error) {
	if len(ids) > capabilityfilecap.MaxBatchGetFiles {
		return nil, bizerr.NewCode(capmodel.CodeCapabilityLimitExceeded, bizerr.P("limit", capabilityfilecap.MaxBatchGetFiles))
	}
	var (
		result = &capmodel.BatchResult[*capabilityfilecap.FileInfo, capabilityfilecap.FileID]{
			Items:      make(map[capabilityfilecap.FileID]*capabilityfilecap.FileInfo, len(ids)),
			MissingIDs: []capabilityfilecap.FileID{},
		}
		missingSeen = make(map[capabilityfilecap.FileID]struct{}, len(ids))
		addMissing  = func(id capabilityfilecap.FileID) {
			if _, ok := missingSeen[id]; ok {
				return
			}
			missingSeen[id] = struct{}{}
			result.MissingIDs = append(result.MissingIDs, id)
		}
	)
	parsedIDs, requested := capmodel.ParseInt64IDs(ids, addMissing)
	if len(parsedIDs) == 0 {
		return result, nil
	}
	var (
		rows  = make([]*fileCapabilityRow, 0, len(parsedIDs))
		cols  = dao.SysFile.Columns()
		model = dao.SysFile.Ctx(ctx).
			Fields(cols.Id, cols.Original, cols.Name, cols.Suffix, cols.Size, cols.Scene).
			WhereIn(cols.Id, parsedIDs)
	)
	if a != nil && a.tenantFilter != nil {
		model = tenantspi.ApplyPluginTableFilter(ctx, a.tenantFilter, model, "")
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
		result.Items[requestID] = fileCapabilityInfoProjection(row, requestID)
	}
	for _, id := range ids {
		if _, ok := result.Items[id]; !ok {
			addMissing(id)
		}
	}
	return result, nil
}

// List returns one bounded page of visible file projections.
func (a *fileCapabilityAdapter) List(ctx context.Context, input capabilityfilecap.ListInput) (*capmodel.PageResult[*capabilityfilecap.FileInfo], error) {
	pageNum, pageSize := input.Page.Normalize()
	if pageSize > capabilityfilecap.MaxListPageSize {
		pageSize = capabilityfilecap.MaxListPageSize
	}
	cols := dao.SysFile.Columns()
	model := dao.SysFile.Ctx(ctx)
	if a != nil && a.tenantFilter != nil {
		model = tenantspi.ApplyPluginTableFilter(ctx, a.tenantFilter, model, "")
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
			return &capmodel.PageResult[*capabilityfilecap.FileInfo]{Items: []*capabilityfilecap.FileInfo{}, Total: 0}, nil
		}
		model = model.WhereIn(cols.Suffix, suffixes)
	}
	total, err := model.Clone().Count()
	if err != nil {
		return nil, err
	}
	rows := make([]*fileCapabilityRow, 0, pageSize)
	if err = model.Clone().
		Fields(cols.Id, cols.Original, cols.Name, cols.Suffix, cols.Size, cols.Scene).
		Page(pageNum, pageSize).
		OrderDesc(cols.Id).
		Scan(&rows); err != nil {
		return nil, err
	}
	items := make([]*capabilityfilecap.FileInfo, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		items = append(items, fileCapabilityInfoProjection(row, capabilityfilecap.FileID(fmt.Sprintf("%d", row.Id))))
	}
	return &capmodel.PageResult[*capabilityfilecap.FileInfo]{Items: items, Total: total}, nil
}

// ListScenes returns bounded governed file scene options.
func (a *fileCapabilityAdapter) ListScenes(ctx context.Context) ([]*capabilityfilecap.Option, error) {
	if a == nil || a.owner == nil {
		return nil, bizerr.NewCode(capmodel.CodeCapabilityUnavailable, bizerr.P("capability", "file-scenes"))
	}
	scenes, err := a.owner.UsageScenes(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]*capabilityfilecap.Option, 0, len(scenes))
	for _, scene := range scenes {
		if scene == nil {
			continue
		}
		items = append(items, &capabilityfilecap.Option{Value: scene.Value, Label: scene.Label})
	}
	return items, nil
}

// ListSuffixes returns bounded visible file suffix options.
func (a *fileCapabilityAdapter) ListSuffixes(ctx context.Context) ([]*capabilityfilecap.Option, error) {
	if a == nil || a.owner == nil {
		return nil, bizerr.NewCode(capmodel.CodeCapabilityUnavailable, bizerr.P("capability", "file-suffixes"))
	}
	suffixes, err := a.owner.Suffixes(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]*capabilityfilecap.Option, 0, len(suffixes))
	for _, suffix := range suffixes {
		if suffix == nil {
			continue
		}
		items = append(items, &capabilityfilecap.Option{Value: suffix.Value, Label: suffix.Label})
	}
	return items, nil
}

// fileCapabilityRow is the stable projection read from sys_file for plugin capabilities.
type fileCapabilityRow struct {
	Id       int64
	Original string
	Name     string
	Suffix   string
	Size     int64
	Scene    string
}

// fileCapabilityInfoProjection converts one sys_file projection row to plugin file info.
func fileCapabilityInfoProjection(row *fileCapabilityRow, id capabilityfilecap.FileID) *capabilityfilecap.FileInfo {
	if row == nil {
		return nil
	}
	return &capabilityfilecap.FileInfo{
		ID:            id,
		Name:          firstNonEmpty(row.Original, row.Name),
		MimeType:      mimeTypeFromSuffix(row.Suffix),
		SizeBytes:     row.Size,
		BusinessScene: row.Scene,
	}
}

// Open opens a visible file stream after target checks.
func (a *fileCapabilityAdapter) Open(ctx context.Context, id capabilityfilecap.FileID) (*capabilityfilecap.FileContent, error) {
	return a.open(ctx, id)
}

// Upload creates one host file record from a plugin-provided content stream.
func (a *fileCapabilityAdapter) Upload(ctx context.Context, input capabilityfilecap.UploadInput) (*capabilityfilecap.FileInfo, error) {
	if a == nil || a.owner == nil {
		return nil, bizerr.NewCode(capmodel.CodeCapabilityUnavailable, bizerr.P("capability", "file-upload"))
	}
	output, err := a.owner.CreateFromReader(ctx, &CreateFromReaderInput{
		Filename:  input.Filename,
		Scene:     input.BusinessScene,
		Reader:    input.Reader,
		SizeBytes: input.SizeBytes,
	})
	if err != nil {
		return nil, err
	}
	return fileUploadProjection(output, input.BusinessScene), nil
}

// CreateFromStorage copies one plugin-private storage object into the host file center.
func (a *fileCapabilityAdapter) CreateFromStorage(ctx context.Context, input capabilityfilecap.CreateFromStorageInput) (projection *capabilityfilecap.FileInfo, err error) {
	if a == nil || a.owner == nil {
		return nil, bizerr.NewCode(capmodel.CodeCapabilityUnavailable, bizerr.P("capability", "file-storage-promotion"))
	}
	if a.storage == nil {
		return nil, bizerr.NewCode(capmodel.CodeCapabilityUnavailable, bizerr.P("capability", "file-storage"))
	}
	storagePath := strings.TrimSpace(input.StoragePath)
	output, err := a.storage.Get(ctx, storagecap.GetInput{Path: storagePath})
	if err != nil {
		return nil, err
	}
	if output == nil || !output.Found || output.Body == nil {
		return nil, bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	defer closeutil.Close(ctx, output.Body, &err, "close plugin storage file source failed")

	filename := strings.TrimSpace(input.Filename)
	if filename == "" {
		filename = path.Base(strings.ReplaceAll(storagePath, "\\", "/"))
	}
	sizeBytes := input.SizeBytes
	if sizeBytes <= 0 && output.Object != nil {
		sizeBytes = output.Object.Size
	}
	uploaded, err := a.owner.CreateFromReader(ctx, &CreateFromReaderInput{
		Filename:  filename,
		Scene:     input.BusinessScene,
		Reader:    output.Body,
		SizeBytes: sizeBytes,
	})
	if err != nil {
		return nil, err
	}
	return fileUploadProjection(uploaded, input.BusinessScene), nil
}

// UpdateMetadata mutates governed visible file metadata.
func (a *fileCapabilityAdapter) UpdateMetadata(ctx context.Context, input capabilityfilecap.UpdateMetadataInput) error {
	parsedID, err := parseFileID(input.ID)
	if err != nil {
		return err
	}
	if err = a.EnsureVisible(ctx, []capabilityfilecap.FileID{input.ID}); err != nil {
		return err
	}
	data := do.SysFile{}
	hasUpdate := false
	if input.Name != nil {
		data.Original = strings.TrimSpace(*input.Name)
		hasUpdate = true
	}
	if input.BusinessScene != nil {
		data.Scene = strings.TrimSpace(*input.BusinessScene)
		hasUpdate = true
	}
	if !hasUpdate {
		return nil
	}
	model := dao.SysFile.Ctx(ctx).Where(do.SysFile{Id: parsedID})
	if a != nil && a.tenantFilter != nil {
		model = tenantspi.ApplyPluginTableFilter(ctx, a.tenantFilter, model, "")
	}
	_, err = model.Data(data).Update()
	return err
}

// EnsureVisible rejects when any requested file is absent or invisible.
func (a *fileCapabilityAdapter) EnsureVisible(ctx context.Context, ids []capabilityfilecap.FileID) error {
	if len(ids) > capabilityfilecap.MaxEnsureVisibleFiles {
		return bizerr.NewCode(capmodel.CodeCapabilityLimitExceeded, bizerr.P("limit", capabilityfilecap.MaxEnsureVisibleFiles))
	}
	result, err := a.BatchGet(ctx, ids)
	if err != nil {
		return err
	}
	if len(result.MissingIDs) > 0 {
		return bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	return nil
}

// Delete deletes one visible file.
func (a *fileCapabilityAdapter) Delete(ctx context.Context, id capabilityfilecap.FileID) error {
	return a.DeleteMany(ctx, []capabilityfilecap.FileID{id})
}

// DeleteMany deletes visible file metadata rows.
func (a *fileCapabilityAdapter) DeleteMany(ctx context.Context, ids []capabilityfilecap.FileID) error {
	if err := a.EnsureVisible(ctx, ids); err != nil {
		return err
	}
	parsedIDs, _ := capmodel.ParseInt64IDs(ids, nil)
	if len(parsedIDs) == 0 {
		return bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	if a != nil && a.owner != nil {
		return a.owner.Delete(ctx, parsedIDs)
	}
	model := dao.SysFile.Ctx(ctx).WhereIn(dao.SysFile.Columns().Id, parsedIDs)
	if a != nil && a.tenantFilter != nil {
		model = tenantspi.ApplyPluginTableFilter(ctx, a.tenantFilter, model, "")
	}
	_, err := model.Delete()
	return err
}

// open delegates content reads to the file owner so storage details stay host-owned.
func (a *fileCapabilityAdapter) open(ctx context.Context, id capabilityfilecap.FileID) (*capabilityfilecap.FileContent, error) {
	parsedID, err := parseFileID(id)
	if err != nil {
		return nil, err
	}
	if err = a.EnsureVisible(ctx, []capabilityfilecap.FileID{id}); err != nil {
		return nil, err
	}
	if a == nil || a.owner == nil {
		return nil, bizerr.NewCode(capmodel.CodeCapabilityUnavailable, bizerr.P("capability", "file-open"))
	}
	out, err := a.owner.OpenByID(ctx, parsedID)
	if err != nil {
		return nil, err
	}
	return &capabilityfilecap.FileContent{
		Reader:      out.Reader,
		Filename:    out.Original,
		ContentType: out.ContentType,
		SizeBytes:   out.Size,
	}, nil
}

// fileDetailProjection converts a host file detail into the plugin projection.
func fileDetailProjection(detail *DetailOutput, id capabilityfilecap.FileID) *capabilityfilecap.DetailInfo {
	if detail == nil || detail.SysFile == nil {
		return nil
	}
	fileInfo := detail.SysFile
	return &capabilityfilecap.DetailInfo{
		FileInfo: capabilityfilecap.FileInfo{
			ID:            id,
			Name:          firstNonEmpty(fileInfo.Original, fileInfo.Name),
			MimeType:      mimeTypeFromSuffix(fileInfo.Suffix),
			SizeBytes:     fileInfo.Size,
			BusinessScene: fileInfo.Scene,
		},
		OriginalName:  fileInfo.Original,
		URL:           fileInfo.Url,
		CreatedByName: detail.CreatedByName,
		SceneLabel:    detail.SceneLabel,
	}
}

// fileUploadProjection converts a newly created host file record into a plugin projection.
func fileUploadProjection(output *UploadOutput, scene string) *capabilityfilecap.FileInfo {
	if output == nil {
		return nil
	}
	businessScene := strings.TrimSpace(scene)
	if businessScene == "" {
		businessScene = DefaultFileSceneOther
	}
	return &capabilityfilecap.FileInfo{
		ID:            capabilityfilecap.FileID(fmt.Sprintf("%d", output.Id)),
		Name:          firstNonEmpty(output.Original, output.Name),
		MimeType:      mimeTypeFromSuffix(output.Suffix),
		SizeBytes:     output.Size,
		BusinessScene: businessScene,
	}
}

// parseFileID decodes one plugin-visible file ID into the host owner key.
func parseFileID(id capabilityfilecap.FileID) (int64, error) {
	value, parseErr := capmodel.ParsePositiveInt64Strings([]string{string(id)})
	if parseErr != nil || len(value) != 1 {
		return 0, bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	return value[0], nil
}

// firstNonEmpty returns the first non-empty value.
func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
