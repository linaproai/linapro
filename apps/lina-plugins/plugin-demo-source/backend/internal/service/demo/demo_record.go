// demo_record.go implements CRUD, paging, and attachment download behavior for
// the plugin-demo-source record sample.

package demo

import (
	"context"
	"mime"
	"net/http"
	"os"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/logger"
	plugincontract "lina-core/pkg/pluginservice/contract"
	"lina-plugin-demo-source/backend/internal/dao"
	"lina-plugin-demo-source/backend/internal/model/do"
	entitymodel "lina-plugin-demo-source/backend/internal/model/entity"
)

// Demo-record constants define the table schema fields and paging defaults
// used by the source-plugin sample service.
const (
	defaultPageNum  = 1
	defaultPageSize = 10
	maxPageSize     = 100
)

// ListRecordsInput defines the demo record list query.
type ListRecordsInput struct {
	// Keyword is the optional fuzzy-match keyword applied to title.
	Keyword string
	// PageNum is the requested page number.
	PageNum int
	// PageSize is the requested page size.
	PageSize int
}

// ListRecordsOutput defines the demo record list result.
type ListRecordsOutput struct {
	// List contains the current page of records.
	List []*RecordListItemOutput
	// Total is the total matched row count.
	Total int
}

// RecordListItemOutput defines one demo record row.
type RecordListItemOutput struct {
	// Id is the record ID.
	Id int64
	// Title is the record title.
	Title string
	// Content is the record content summary.
	Content string
	// AttachmentName is the original attachment filename.
	AttachmentName string
	// HasAttachment reports whether the record owns one attachment.
	HasAttachment bool
	// CreatedAt is the formatted creation time.
	CreatedAt string
	// UpdatedAt is the formatted update time.
	UpdatedAt string
}

// RecordDetailOutput defines one demo record detail result.
type RecordDetailOutput struct {
	// Id is the record ID.
	Id int64
	// Title is the record title.
	Title string
	// Content is the record content body.
	Content string
	// AttachmentName is the original attachment filename.
	AttachmentName string
	// HasAttachment reports whether the record owns one attachment.
	HasAttachment bool
}

// CreateRecordInput defines the create-record input.
type CreateRecordInput struct {
	// Title is the required record title.
	Title string
	// Content is the optional record content.
	Content string
	// File is the optional uploaded attachment.
	File *ghttp.UploadFile
}

// UpdateRecordInput defines the update-record input.
type UpdateRecordInput struct {
	// Id is the record ID.
	Id int64
	// Title is the required record title.
	Title string
	// Content is the optional record content.
	Content string
	// File is the optional new uploaded attachment.
	File *ghttp.UploadFile
	// RemoveAttachment reports whether the current attachment should be removed.
	RemoveAttachment bool
}

// RecordMutationOutput defines the record create/update result.
type RecordMutationOutput struct {
	// Id is the affected record ID.
	Id int64
}

// AttachmentDownloadOutput defines one attachment download descriptor.
type AttachmentDownloadOutput struct {
	// OriginalName is the original attachment filename.
	OriginalName string
	// FullPath is the absolute storage path for the attachment.
	FullPath string
	// ContentType is the detected content type.
	ContentType string
}

// demoRecordEntity reuses the plugin-local generated record entity.
type demoRecordEntity = entitymodel.PluginDemoSourceRecord

// ListRecords returns the paged demo records rendered by the source-plugin CRUD page.
func (s *serviceImpl) ListRecords(ctx context.Context, in *ListRecordsInput) (out *ListRecordsOutput, err error) {
	if err = ensureDemoRecordTableReady(ctx); err != nil {
		return nil, err
	}

	pageNum, pageSize := normalizeListPagination(in)
	model := s.tenantFilter.Apply(ctx, dao.PluginDemoSourceRecord.Ctx(ctx))
	keyword := strings.TrimSpace(in.Keyword)
	if keyword != "" {
		model = model.WhereLike(dao.PluginDemoSourceRecord.Columns().Title, "%"+keyword+"%")
	}

	total, err := model.Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeDemoRecordCountQueryFailed)
	}

	items := make([]*demoRecordEntity, 0)
	err = model.
		OrderDesc(dao.PluginDemoSourceRecord.Columns().UpdatedAt).
		OrderDesc(dao.PluginDemoSourceRecord.Columns().Id).
		Page(pageNum, pageSize).
		Scan(&items)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeDemoRecordListQueryFailed)
	}

	list := make([]*RecordListItemOutput, 0, len(items))
	for _, item := range items {
		list = append(list, buildRecordListItemOutput(item))
	}
	return &ListRecordsOutput{List: list, Total: total}, nil
}

// GetRecord returns one demo record detail for edit forms.
func (s *serviceImpl) GetRecord(ctx context.Context, id int64) (out *RecordDetailOutput, err error) {
	record, err := s.getRecordEntity(ctx, id)
	if err != nil {
		return nil, err
	}
	return &RecordDetailOutput{
		Id:             record.Id,
		Title:          record.Title,
		Content:        record.Content,
		AttachmentName: record.AttachmentName,
		HasAttachment:  record.AttachmentPath != "",
	}, nil
}

// CreateRecord creates one demo record and stores its optional attachment file.
func (s *serviceImpl) CreateRecord(ctx context.Context, in *CreateRecordInput) (out *RecordMutationOutput, err error) {
	if err = ensureDemoRecordTableReady(ctx); err != nil {
		return nil, err
	}
	if err = validateRecordTitle(in.Title); err != nil {
		return nil, err
	}

	attachmentName, attachmentPath, err := saveDemoAttachmentFile(ctx, in.File)
	if err != nil {
		return nil, err
	}
	if attachmentPath != "" {
		defer func() {
			if err != nil {
				cleanupDemoAttachmentAfterMutationFailure(ctx, attachmentPath)
			}
		}()
	}

	recordID, err := dao.PluginDemoSourceRecord.Ctx(ctx).Data(do.PluginDemoSourceRecord{
		TenantId:       s.tenantFilter.Current(ctx),
		Title:          strings.TrimSpace(in.Title),
		Content:        strings.TrimSpace(in.Content),
		AttachmentName: stringPointer(attachmentName),
		AttachmentPath: stringPointer(attachmentPath),
	}).InsertAndGetId()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeDemoRecordCreateFailed)
	}
	return &RecordMutationOutput{Id: recordID}, nil
}

// UpdateRecord updates one demo record and replaces or removes its optional attachment.
func (s *serviceImpl) UpdateRecord(ctx context.Context, in *UpdateRecordInput) (out *RecordMutationOutput, err error) {
	if err = ensureDemoRecordTableReady(ctx); err != nil {
		return nil, err
	}
	if err = validateRecordTitle(in.Title); err != nil {
		return nil, err
	}

	record, err := s.getRecordEntity(ctx, in.Id)
	if err != nil {
		return nil, err
	}

	updateData := do.PluginDemoSourceRecord{
		Title:          strings.TrimSpace(in.Title),
		Content:        strings.TrimSpace(in.Content),
		AttachmentName: stringPointer(record.AttachmentName),
		AttachmentPath: stringPointer(record.AttachmentPath),
	}
	oldAttachmentPath := strings.TrimSpace(record.AttachmentPath)

	if in.RemoveAttachment {
		updateData.AttachmentName = stringPointer("")
		updateData.AttachmentPath = stringPointer("")
	}

	newAttachmentName := ""
	newAttachmentPath := ""
	if in.File != nil {
		newAttachmentName, newAttachmentPath, err = saveDemoAttachmentFile(ctx, in.File)
		if err != nil {
			return nil, err
		}
		updateData.AttachmentName = stringPointer(newAttachmentName)
		updateData.AttachmentPath = stringPointer(newAttachmentPath)
		defer func() {
			if err != nil && newAttachmentPath != "" {
				cleanupDemoAttachmentAfterMutationFailure(ctx, newAttachmentPath)
			}
		}()
	}

	_, err = dao.PluginDemoSourceRecord.Ctx(ctx).
		Where(plugincontract.TenantFilterColumn, s.tenantFilter.Current(ctx)).
		Where(do.PluginDemoSourceRecord{Id: in.Id}).
		Data(updateData).
		Update()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeDemoRecordUpdateFailed)
	}

	if (in.RemoveAttachment || newAttachmentPath != "") && oldAttachmentPath != "" {
		if err = deleteDemoAttachmentFile(ctx, oldAttachmentPath); err != nil {
			return nil, err
		}
	}
	return &RecordMutationOutput{Id: in.Id}, nil
}

// DeleteRecord deletes one demo record and cleans its attachment file.
func (s *serviceImpl) DeleteRecord(ctx context.Context, id int64) error {
	record, err := s.getRecordEntity(ctx, id)
	if err != nil {
		return err
	}

	_, err = dao.PluginDemoSourceRecord.Ctx(ctx).
		Where(plugincontract.TenantFilterColumn, s.tenantFilter.Current(ctx)).
		Where(do.PluginDemoSourceRecord{Id: id}).
		Delete()
	if err != nil {
		return bizerr.WrapCode(err, CodeDemoRecordDeleteFailed)
	}
	if record.AttachmentPath != "" {
		if err = deleteDemoAttachmentFile(ctx, record.AttachmentPath); err != nil {
			return err
		}
	}
	return nil
}

// BuildAttachmentDownload returns one attachment download descriptor for the given record.
func (s *serviceImpl) BuildAttachmentDownload(
	ctx context.Context,
	id int64,
) (out *AttachmentDownloadOutput, err error) {
	record, err := s.getRecordEntity(ctx, id)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(record.AttachmentPath) == "" {
		return nil, bizerr.NewCode(CodeDemoRecordAttachmentRequired)
	}

	fullPath, err := buildDemoAttachmentFullPath(ctx, record.AttachmentPath)
	if err != nil {
		return nil, err
	}
	if !gfile.Exists(fullPath) {
		return nil, bizerr.NewCode(CodeDemoRecordAttachmentFileNotFound)
	}

	contentType := mime.TypeByExtension("." + gfile.ExtName(record.AttachmentName))
	if contentType == "" {
		contentType = http.DetectContentType(nil)
	}
	return &AttachmentDownloadOutput{
		OriginalName: record.AttachmentName,
		FullPath:     fullPath,
		ContentType:  contentType,
	}, nil
}

// getRecordEntity loads one sample record entity by primary key.
func (s *serviceImpl) getRecordEntity(ctx context.Context, id int64) (*demoRecordEntity, error) {
	if err := ensureDemoRecordTableReady(ctx); err != nil {
		return nil, err
	}
	if id <= 0 {
		return nil, bizerr.NewCode(CodeDemoRecordIDRequired)
	}

	var record *demoRecordEntity
	err := s.tenantFilter.Apply(ctx, dao.PluginDemoSourceRecord.Ctx(ctx)).
		Where(do.PluginDemoSourceRecord{Id: id}).
		Scan(&record)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeDemoRecordDetailQueryFailed)
	}
	if record == nil {
		return nil, bizerr.NewCode(CodeDemoRecordNotFound)
	}
	return record, nil
}

// ensureDemoRecordTableReady verifies the sample table exists before CRUD work
// continues.
func ensureDemoRecordTableReady(ctx context.Context) error {
	fields, err := dao.PluginDemoSourceRecord.DB().TableFields(ctx, dao.PluginDemoSourceRecord.Table())
	if err != nil {
		return bizerr.WrapCode(err, CodeDemoRecordTableCheckFailed)
	}
	if len(fields) == 0 {
		return bizerr.NewCode(CodeDemoRecordTableNotInstalled)
	}
	return nil
}

// normalizeListPagination applies paging defaults and max-page-size limits.
func normalizeListPagination(in *ListRecordsInput) (int, int) {
	if in == nil {
		return defaultPageNum, defaultPageSize
	}

	pageNum := in.PageNum
	if pageNum <= 0 {
		pageNum = defaultPageNum
	}
	pageSize := in.PageSize
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	return pageNum, pageSize
}

// validateRecordTitle validates the required sample record title field.
func validateRecordTitle(title string) error {
	if strings.TrimSpace(title) == "" {
		return bizerr.NewCode(CodeDemoRecordTitleRequired)
	}
	return nil
}

// buildRecordListItemOutput converts one internal entity into the list item
// response shape.
func buildRecordListItemOutput(item *demoRecordEntity) *RecordListItemOutput {
	if item == nil {
		return &RecordListItemOutput{}
	}
	return &RecordListItemOutput{
		Id:             item.Id,
		Title:          item.Title,
		Content:        item.Content,
		AttachmentName: item.AttachmentName,
		HasAttachment:  strings.TrimSpace(item.AttachmentPath) != "",
		CreatedAt:      formatRecordTime(item.CreatedAt),
		UpdatedAt:      formatRecordTime(item.UpdatedAt),
	}
}

// formatRecordTime formats one optional GoFrame time value for API output.
func formatRecordTime(value *gtime.Time) string {
	if value == nil {
		return ""
	}
	return value.String()
}

// stringPointer allocates one string pointer for optional DB mutation fields.
func stringPointer(value string) *string {
	return &value
}

// listAllAttachmentPaths returns all persisted attachment paths stored by the
// sample records table.
func listAllAttachmentPaths(ctx context.Context) ([]string, error) {
	fields, err := dao.PluginDemoSourceRecord.DB().TableFields(ctx, dao.PluginDemoSourceRecord.Table())
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeDemoRecordTableCheckFailed)
	}
	if len(fields) == 0 {
		return []string{}, nil
	}

	rows, err := dao.PluginDemoSourceRecord.Ctx(ctx).
		Fields(dao.PluginDemoSourceRecord.Columns().AttachmentPath).
		All()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeDemoRecordAttachmentPathQueryFailed)
	}

	paths := make([]string, 0, len(rows))
	for _, row := range rows {
		value := strings.TrimSpace(row[dao.PluginDemoSourceRecord.Columns().AttachmentPath].String())
		if value != "" {
			paths = append(paths, value)
		}
	}
	return paths, nil
}

// withRecordTransaction runs one handler inside the shared source-plugin record
// transaction boundary.
func withRecordTransaction(ctx context.Context, handler func(ctx context.Context, tx gdb.TX) error) error {
	return dao.PluginDemoSourceRecord.Transaction(ctx, handler)
}

// fileExists reports whether the path exists and points to a regular
// non-directory file.
func fileExists(path string) bool {
	fileInfo, err := os.Stat(path)
	return err == nil && !fileInfo.IsDir()
}

// cleanupDemoAttachmentAfterMutationFailure removes an attachment created by a
// failed mutation and logs cleanup failures without hiding the primary error.
func cleanupDemoAttachmentAfterMutationFailure(ctx context.Context, attachmentPath string) {
	if cleanupErr := deleteDemoAttachmentFile(ctx, attachmentPath); cleanupErr != nil {
		logger.Warningf(
			ctx,
			"cleanup demo attachment after failed mutation failed path=%s err=%v",
			attachmentPath,
			cleanupErr,
		)
	}
}
