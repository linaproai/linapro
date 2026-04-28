// This file implements CRUD and attachment storage behavior for the dynamic
// plugin demo-record sample.

package dynamicservice

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"lina-core/pkg/bizerr"
)

// Demo-record constants define the sample table name, paging defaults, and
// attachment limits used by the dynamic plugin sample.
const (
	demoRecordTable                  = "plugin_demo_dynamic_record"
	demoRecordAttachmentPrefix       = "demo-record-files"
	demoRecordDefaultPageNum         = 1
	demoRecordDefaultPageSize        = 20
	demoRecordMaxPageSize            = 100
	demoRecordMaxAttachmentBytes int = 10 * 1024 * 1024
)

// IsDemoRecordInvalidInput reports whether the error should be exposed as one bad request.
func IsDemoRecordInvalidInput(err error) bool {
	return bizerr.Is(err, CodeDynamicDemoRecordInvalidInput) ||
		bizerr.Is(err, CodeDynamicDemoRecordRequestBodyRequired) ||
		bizerr.Is(err, CodeDynamicDemoRecordTitleRequired) ||
		bizerr.Is(err, CodeDynamicDemoRecordTitleTooLong) ||
		bizerr.Is(err, CodeDynamicDemoRecordContentTooLong) ||
		bizerr.Is(err, CodeDynamicDemoRecordAttachmentNameRequired) ||
		bizerr.Is(err, CodeDynamicDemoRecordAttachmentBase64Invalid) ||
		bizerr.Is(err, CodeDynamicDemoRecordAttachmentContentRequired) ||
		bizerr.Is(err, CodeDynamicDemoRecordAttachmentSizeTooLarge) ||
		bizerr.Is(err, CodeDynamicDemoRecordIDRequired)
}

// NewDemoRecordInvalidInputError creates one invalid-input business error that
// should be translated into a bad-request bridge response.
func NewDemoRecordInvalidInputError(message string) error {
	return bizerr.NewCode(
		CodeDynamicDemoRecordInvalidInput,
		bizerr.P("reason", strings.TrimSpace(message)),
	)
}

// IsDemoRecordNotFound reports whether the error should be exposed as one not-found response.
func IsDemoRecordNotFound(err error) bool {
	return bizerr.Is(err, CodeDynamicDemoRecordNotFound) ||
		bizerr.Is(err, CodeDynamicDemoRecordAttachmentNotFound)
}

// ListDemoRecordsPayload returns one paged demo-record list backed by the plugin-owned SQL table.
func (s *serviceImpl) ListDemoRecordsPayload(input *DemoRecordListInput) (*demoRecordListPayload, error) {
	pageNum, pageSize, keyword := normalizeDemoRecordListInput(input)

	query := s.dataSvc.Table(demoRecordTable).
		Fields("id", "title", "content", "attachmentName", "attachmentPath", "createdAt", "updatedAt").
		OrderDesc("updatedAt").
		OrderDesc("id").
		Page(int32(pageNum), int32(pageSize))
	if keyword != "" {
		query = query.WhereLike("title", keyword)
	}

	records, total, err := query.All()
	if err != nil {
		return nil, err
	}

	items := make([]*demoRecordPayload, 0, len(records))
	for _, record := range records {
		item, mapErr := mapDemoRecordPayload(record)
		if mapErr != nil {
			return nil, mapErr
		}
		items = append(items, item)
	}
	return &demoRecordListPayload{
		List:  items,
		Total: int(total),
	}, nil
}

// GetDemoRecordPayload returns one demo-record detail by ID.
func (s *serviceImpl) GetDemoRecordPayload(recordID string) (*demoRecordPayload, error) {
	recordID, err := validateDemoRecordID(recordID)
	if err != nil {
		return nil, err
	}
	record, err := s.loadDemoRecord(recordID)
	if err != nil {
		return nil, err
	}
	return demoRecordEntityToPayload(record), nil
}

// CreateDemoRecordPayload creates one demo record and stores its optional attachment.
func (s *serviceImpl) CreateDemoRecordPayload(input *DemoRecordMutationInput) (payload *demoRecordPayload, err error) {
	normalizedInput, err := normalizeDemoRecordMutationInput(input)
	if err != nil {
		return nil, err
	}
	recordID, err := s.runtimeSvc.UUID()
	if err != nil {
		return nil, err
	}
	attachmentName, attachmentPath, err := s.saveDemoRecordAttachment(normalizedInput)
	if err != nil {
		return nil, err
	}
	if attachmentPath != "" {
		defer func() {
			if err == nil {
				return
			}
			if cleanupErr := s.deleteDemoRecordAttachment(attachmentPath); cleanupErr != nil {
				err = bizerr.WrapCode(cleanupErr, CodeDynamicDemoRecordAttachmentRollbackFailed)
			}
		}()
	}

	recordMap := map[string]any{
		"id":             recordID,
		"title":          normalizedInput.Title,
		"content":        normalizedInput.Content,
		"attachmentName": attachmentName,
		"attachmentPath": attachmentPath,
	}
	if _, err = s.dataSvc.Table(demoRecordTable).Insert(recordMap); err != nil {
		return nil, err
	}
	record, err := s.loadDemoRecord(recordID)
	if err != nil {
		return nil, err
	}
	payload = demoRecordEntityToPayload(record)
	return payload, nil
}

// UpdateDemoRecordPayload updates one demo record and replaces or removes its optional attachment.
func (s *serviceImpl) UpdateDemoRecordPayload(recordID string, input *DemoRecordMutationInput) (payload *demoRecordPayload, err error) {
	validatedRecordID, validateErr := validateDemoRecordID(recordID)
	if validateErr != nil {
		return nil, validateErr
	}
	existingRecord, err := s.loadDemoRecord(validatedRecordID)
	if err != nil {
		return nil, err
	}

	normalizedInput, err := normalizeDemoRecordMutationInput(input)
	if err != nil {
		return nil, err
	}
	newAttachmentName := existingRecord.AttachmentName
	newAttachmentPath := existingRecord.AttachmentPath
	replacedAttachmentPath := ""
	if normalizedInput.RemoveAttachment {
		newAttachmentName = ""
		newAttachmentPath = ""
		replacedAttachmentPath = existingRecord.AttachmentPath
	}

	if strings.TrimSpace(normalizedInput.AttachmentContentBase64) != "" {
		uploadedName, uploadedPath, uploadErr := s.saveDemoRecordAttachment(normalizedInput)
		if uploadErr != nil {
			return nil, uploadErr
		}
		newAttachmentName = uploadedName
		newAttachmentPath = uploadedPath
		replacedAttachmentPath = existingRecord.AttachmentPath
		defer func() {
			if err == nil || uploadedPath == "" {
				return
			}
			if cleanupErr := s.deleteDemoRecordAttachment(uploadedPath); cleanupErr != nil {
				err = bizerr.WrapCode(cleanupErr, CodeDynamicDemoRecordAttachmentRollbackFailed)
			}
		}()
	}

	recordMap := map[string]any{
		"title":          normalizedInput.Title,
		"content":        normalizedInput.Content,
		"attachmentName": newAttachmentName,
		"attachmentPath": newAttachmentPath,
	}
	if _, err = s.dataSvc.Table(demoRecordTable).WhereKey(validatedRecordID).Update(recordMap); err != nil {
		return nil, err
	}
	if replacedAttachmentPath != "" && replacedAttachmentPath != newAttachmentPath {
		if err = s.deleteDemoRecordAttachment(replacedAttachmentPath); err != nil {
			return nil, err
		}
	}

	record, err := s.loadDemoRecord(validatedRecordID)
	if err != nil {
		return nil, err
	}
	payload = demoRecordEntityToPayload(record)
	return payload, nil
}

// DeleteDemoRecordPayload deletes one demo record and its optional attachment.
func (s *serviceImpl) DeleteDemoRecordPayload(recordID string) (*demoRecordDeletePayload, error) {
	recordID, err := validateDemoRecordID(recordID)
	if err != nil {
		return nil, err
	}
	record, err := s.loadDemoRecord(recordID)
	if err != nil {
		return nil, err
	}
	if _, err = s.dataSvc.Table(demoRecordTable).WhereKey(recordID).Delete(); err != nil {
		return nil, err
	}
	if err = s.deleteDemoRecordAttachment(record.AttachmentPath); err != nil {
		return nil, err
	}
	return &demoRecordDeletePayload{
		Id:      recordID,
		Deleted: true,
	}, nil
}

// BuildDemoRecordAttachmentDownload returns one attachment download descriptor.
func (s *serviceImpl) BuildDemoRecordAttachmentDownload(recordID string) (*demoRecordAttachmentDownloadPayload, error) {
	recordID, err := validateDemoRecordID(recordID)
	if err != nil {
		return nil, err
	}
	record, err := s.loadDemoRecord(recordID)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(record.AttachmentPath) == "" {
		return nil, bizerr.NewCode(CodeDynamicDemoRecordAttachmentNotFound)
	}

	body, object, found, err := s.storageSvc.Get(record.AttachmentPath)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, bizerr.NewCode(CodeDynamicDemoRecordAttachmentNotFound)
	}

	contentType := "application/octet-stream"
	if object != nil && strings.TrimSpace(object.ContentType) != "" {
		contentType = strings.TrimSpace(object.ContentType)
	}
	return &demoRecordAttachmentDownloadPayload{
		OriginalName: record.AttachmentName,
		ContentType:  contentType,
		Body:         body,
	}, nil
}

// loadDemoRecord loads one sample record entity by its logical ID.
func (s *serviceImpl) loadDemoRecord(recordID string) (*demoRecordEntity, error) {
	recordMap, found, err := s.dataSvc.Table(demoRecordTable).
		Fields("id", "title", "content", "attachmentName", "attachmentPath", "createdAt", "updatedAt").
		WhereKey(recordID).
		One()
	if err != nil {
		return nil, err
	}
	if !found || recordMap == nil {
		return nil, bizerr.NewCode(CodeDynamicDemoRecordNotFound)
	}
	return parseDemoRecordEntity(recordMap)
}

// mapDemoRecordPayload converts one raw structured-data row into the response
// payload shape.
func mapDemoRecordPayload(record map[string]any) (*demoRecordPayload, error) {
	entity, err := parseDemoRecordEntity(record)
	if err != nil {
		return nil, err
	}
	return demoRecordEntityToPayload(entity), nil
}

// demoRecordEntityToPayload converts one internal record entity into the JSON
// payload returned by the bridge controller.
func demoRecordEntityToPayload(record *demoRecordEntity) *demoRecordPayload {
	if record == nil {
		return &demoRecordPayload{}
	}
	return &demoRecordPayload{
		Id:             record.Id,
		Title:          record.Title,
		Content:        record.Content,
		AttachmentName: record.AttachmentName,
		HasAttachment:  strings.TrimSpace(record.AttachmentPath) != "",
		CreatedAt:      record.CreatedAt,
		UpdatedAt:      record.UpdatedAt,
	}
}

// normalizeDemoRecordListInput applies paging defaults and trims the keyword
// filter for list requests.
func normalizeDemoRecordListInput(input *DemoRecordListInput) (int, int, string) {
	pageNum := demoRecordDefaultPageNum
	pageSize := demoRecordDefaultPageSize
	keyword := ""
	if input == nil {
		return pageNum, pageSize, keyword
	}
	if input.PageNum > 0 {
		pageNum = input.PageNum
	}
	if input.PageSize > 0 {
		pageSize = input.PageSize
	}
	if pageSize > demoRecordMaxPageSize {
		pageSize = demoRecordMaxPageSize
	}
	keyword = strings.TrimSpace(input.Keyword)
	return pageNum, pageSize, keyword
}

// normalizeDemoRecordMutationInput validates and trims create/update request
// data for sample records.
func normalizeDemoRecordMutationInput(input *DemoRecordMutationInput) (*DemoRecordMutationInput, error) {
	if input == nil {
		return nil, bizerr.NewCode(CodeDynamicDemoRecordRequestBodyRequired)
	}

	normalizedInput := &DemoRecordMutationInput{
		Title:                   strings.TrimSpace(input.Title),
		Content:                 strings.TrimSpace(input.Content),
		AttachmentName:          strings.TrimSpace(input.AttachmentName),
		AttachmentContentBase64: strings.TrimSpace(input.AttachmentContentBase64),
		AttachmentContentType:   strings.TrimSpace(input.AttachmentContentType),
		RemoveAttachment:        input.RemoveAttachment,
	}

	if normalizedInput.Title == "" {
		return nil, bizerr.NewCode(CodeDynamicDemoRecordTitleRequired)
	}
	if len([]rune(normalizedInput.Title)) > 128 {
		return nil, bizerr.NewCode(CodeDynamicDemoRecordTitleTooLong, bizerr.P("maxChars", 128))
	}
	if len([]rune(normalizedInput.Content)) > 1000 {
		return nil, bizerr.NewCode(CodeDynamicDemoRecordContentTooLong, bizerr.P("maxChars", 1000))
	}
	if normalizedInput.AttachmentContentBase64 != "" && normalizedInput.AttachmentName == "" {
		return nil, bizerr.NewCode(CodeDynamicDemoRecordAttachmentNameRequired)
	}
	return normalizedInput, nil
}

// saveDemoRecordAttachment decodes and stores one optional Base64 attachment
// into the governed plugin storage area.
func (s *serviceImpl) saveDemoRecordAttachment(input *DemoRecordMutationInput) (string, string, error) {
	if input == nil || strings.TrimSpace(input.AttachmentContentBase64) == "" {
		return "", "", nil
	}

	body, err := base64.StdEncoding.DecodeString(input.AttachmentContentBase64)
	if err != nil {
		return "", "", bizerr.NewCode(CodeDynamicDemoRecordAttachmentBase64Invalid)
	}
	if len(body) == 0 {
		return "", "", bizerr.NewCode(CodeDynamicDemoRecordAttachmentContentRequired)
	}
	if len(body) > demoRecordMaxAttachmentBytes {
		return "", "", bizerr.NewCode(
			CodeDynamicDemoRecordAttachmentSizeTooLarge,
			bizerr.P("maxSizeMB", demoRecordMaxAttachmentBytes/1024/1024),
		)
	}

	attachmentName := sanitizeDemoRecordAttachmentName(input.AttachmentName)
	attachmentID, err := s.runtimeSvc.UUID()
	if err != nil {
		return "", "", err
	}

	objectPath := fmt.Sprintf(
		"%s/%s_%s",
		demoRecordAttachmentPrefix,
		attachmentID,
		attachmentName,
	)
	contentType := strings.TrimSpace(input.AttachmentContentType)
	if contentType == "" {
		contentType = http.DetectContentType(body)
	}
	if _, err = s.storageSvc.Put(objectPath, body, contentType, true); err != nil {
		return "", "", err
	}
	return attachmentName, objectPath, nil
}

// deleteDemoRecordAttachment removes one governed attachment object when its
// logical path is present.
func (s *serviceImpl) deleteDemoRecordAttachment(objectPath string) error {
	if strings.TrimSpace(objectPath) == "" {
		return nil
	}
	return s.storageSvc.Delete(objectPath)
}

// validateDemoRecordID validates the logical record identifier used by sample
// CRUD operations.
func validateDemoRecordID(value string) (string, error) {
	normalizedValue := strings.TrimSpace(value)
	if normalizedValue == "" {
		return "", bizerr.NewCode(CodeDynamicDemoRecordIDRequired)
	}
	return normalizedValue, nil
}

// sanitizeDemoRecordAttachmentName strips unsafe path characters from uploaded
// attachment names.
func sanitizeDemoRecordAttachmentName(filename string) string {
	sanitizedName := filepath.Base(strings.ReplaceAll(strings.TrimSpace(filename), "\x00", ""))
	if sanitizedName == "." || sanitizedName == "" {
		return "attachment.bin"
	}

	disallowed := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, item := range disallowed {
		sanitizedName = strings.ReplaceAll(sanitizedName, item, "_")
	}
	return sanitizedName
}
