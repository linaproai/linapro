// Package notice implements the notice-management, publication, and
// delete-cascade services for the content-notice source plugin. It owns the
// plugin_content_notice table access and consumes host capability seams for business
// context and notify fan-out.
package notice

import (
	"context"
	"strings"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/logger"
	plugincontract "lina-core/pkg/pluginservice/contract"
	"lina-plugin-content-notice/backend/internal/dao"
	"lina-plugin-content-notice/backend/internal/model/do"
	entitymodel "lina-plugin-content-notice/backend/internal/model/entity"
)

// Dict types used in notice
const (
	DictTypeNoticeType   = "sys_notice_type"   // Notice type dictionary
	DictTypeNoticeStatus = "sys_notice_status" // Notice status dictionary
)

// Notice type values (matching sys_notice_type dictionary)
const (
	NoticeTypeNotice       = 1 // Notice
	NoticeTypeAnnouncement = 2 // Announcement
)

// Notice status values (matching sys_notice_status dictionary)
const (
	NoticeStatusDraft     = 0 // Draft
	NoticeStatusPublished = 1 // Published
)

// Service defines the notice service contract.
type Service interface {
	// List queries the notice list with pagination and filters.
	List(ctx context.Context, in ListInput) (*ListOutput, error)
	// GetById retrieves one notice by its primary key.
	GetById(ctx context.Context, id int64) (*ListItem, error)
	// Create creates a new notice and, when the status is published, fans the
	// notice into the inbox pipeline through the notify bridge.
	Create(ctx context.Context, in CreateInput) (int64, error)
	// Update updates the notice fields. A transition from draft to published
	// triggers inbox publication through the notify bridge.
	Update(ctx context.Context, in UpdateInput) error
	// Delete soft-deletes notices by IDs and cascades the deletion into the
	// notify delivery records so inboxes stay consistent.
	Delete(ctx context.Context, ids string) error
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	bizCtxSvc    plugincontract.BizCtxService       // Business context bridge
	notifySvc    plugincontract.NotifyService       // Unified notify bridge
	tenantFilter plugincontract.TenantFilterService // Tenant query filter bridge
}

// New creates and returns a new Service instance.
func New(
	bizCtxSvc plugincontract.BizCtxService,
	notifySvc plugincontract.NotifyService,
	tenantFilter plugincontract.TenantFilterService,
) Service {
	return &serviceImpl{
		bizCtxSvc:    bizCtxSvc,
		notifySvc:    notifySvc,
		tenantFilter: tenantFilter,
	}
}

// ListInput defines input for List function.
type ListInput struct {
	PageNum   int    // Page number, starting from 1
	PageSize  int    // Page size
	Title     string // Title, supports fuzzy search
	Type      int    // Type: 1=Notice 2=Announcement (see NoticeType* constants)
	CreatedBy string // Creator username, supports fuzzy search
}

// ListItem defines a single list item.
type ListItem struct {
	*NoticeEntity        // Notice entity
	CreatedByName string `json:"createdByName"` // Creator username
}

// ListOutput defines output for List function.
type ListOutput struct {
	List  []*ListItem // List items
	Total int         // Total count
}

// List queries notice list with pagination and filters.
func (s *serviceImpl) List(ctx context.Context, in ListInput) (*ListOutput, error) {
	var (
		noticeColumns = dao.Notice.Columns()
		userColumns   = dao.SysUser.Columns()
	)

	m := s.tenantFilter.Apply(ctx, dao.Notice.Ctx(ctx))

	// Apply filters
	if in.Title != "" {
		m = m.WhereLike(noticeColumns.Title, "%"+in.Title+"%")
	}
	if in.Type > 0 {
		m = m.Where(noticeColumns.Type, in.Type)
	}
	if in.CreatedBy != "" {
		// Filter by creator username via subquery on sys_user.
		subQuery := s.tenantFilter.Apply(ctx, dao.SysUser.Ctx(ctx)).
			Fields(userColumns.Id).
			WhereLike(userColumns.Username, "%"+in.CreatedBy+"%")
		m = m.Where(noticeColumns.CreatedBy+" IN (?)", subQuery)
	}

	// Get total count
	total, err := m.Count()
	if err != nil {
		return nil, err
	}

	// Query with pagination
	list := make([]*NoticeEntity, 0)
	err = m.Page(in.PageNum, in.PageSize).
		OrderDesc(noticeColumns.Id).
		Scan(&list)
	if err != nil {
		return nil, err
	}

	// Collect unique creator IDs
	userIds := make([]int64, 0, len(list))
	seen := make(map[int64]bool)
	for _, n := range list {
		if n.CreatedBy > 0 && !seen[n.CreatedBy] {
			userIds = append(userIds, n.CreatedBy)
			seen[n.CreatedBy] = true
		}
	}

	// Resolve creator usernames
	userNameMap := make(map[int64]string)
	if len(userIds) > 0 {
		users := make([]*entitymodel.SysUser, 0)
		err = s.tenantFilter.Apply(ctx, dao.SysUser.Ctx(ctx)).
			Fields(userColumns.Id, userColumns.Username).
			WhereIn(userColumns.Id, userIds).
			Scan(&users)
		if err == nil {
			for _, u := range users {
				userNameMap[int64(u.Id)] = u.Username
			}
		}
	}

	// Build result
	items := make([]*ListItem, 0, len(list))
	for _, n := range list {
		items = append(items, &ListItem{
			NoticeEntity:  n,
			CreatedByName: userNameMap[n.CreatedBy],
		})
	}

	return &ListOutput{
		List:  items,
		Total: total,
	}, nil
}

// GetById retrieves notice by ID.
func (s *serviceImpl) GetById(ctx context.Context, id int64) (*ListItem, error) {
	var (
		noticeColumns = dao.Notice.Columns()
		userColumns   = dao.SysUser.Columns()
	)

	var notice *NoticeEntity
	err := s.tenantFilter.Apply(ctx, dao.Notice.Ctx(ctx)).
		Where(noticeColumns.Id, id).
		Scan(&notice)
	if err != nil {
		return nil, err
	}
	if notice == nil {
		return nil, bizerr.NewCode(CodeNoticeNotFound)
	}

	item := &ListItem{NoticeEntity: notice}

	// Resolve creator username
	if notice.CreatedBy > 0 {
		var user *entitymodel.SysUser
		err = s.tenantFilter.Apply(ctx, dao.SysUser.Ctx(ctx)).
			Fields(userColumns.Id, userColumns.Username).
			Where(userColumns.Id, notice.CreatedBy).
			Scan(&user)
		if err == nil && user != nil {
			item.CreatedByName = user.Username
		}
	}

	return item, nil
}

// CreateInput defines input for Create function.
type CreateInput struct {
	Title   string // Title
	Type    int    // Type: 1=Notice 2=Announcement (see NoticeType* constants)
	Content string // Content
	FileIds string // Attachment file IDs, comma-separated
	Status  int    // Status: 0=Draft 1=Published (see NoticeStatus* constants)
	Remark  string // Remark
}

// Create creates a new notice.
func (s *serviceImpl) Create(ctx context.Context, in CreateInput) (int64, error) {
	bizCtx := s.bizCtxSvc.Current(ctx)
	createdBy := int64(bizCtx.UserID)

	// Insert notice (GoFrame auto-fills created_at and updated_at).
	id, err := dao.Notice.Ctx(ctx).Data(do.Notice{
		TenantId:  s.tenantFilter.Current(ctx),
		Title:     in.Title,
		Type:      in.Type,
		Content:   in.Content,
		FileIds:   in.FileIds,
		Status:    in.Status,
		Remark:    in.Remark,
		CreatedBy: createdBy,
		UpdatedBy: createdBy,
	}).InsertAndGetId()
	if err != nil {
		return 0, err
	}

	// If published, dispatch inbox notifications through the unified notify domain.
	if in.Status == NoticeStatusPublished {
		if dispatchErr := s.dispatchPublishedNotice(ctx, id, in.Title, in.Content, in.Type, createdBy); dispatchErr != nil {
			logger.Errorf(ctx, "dispatch published notice failed for notice %d: %v", id, dispatchErr)
		}
	}

	return id, nil
}

// UpdateInput defines input for Update function.
type UpdateInput struct {
	Id      int64   // Notice ID
	Title   *string // Title
	Type    *int    // Type: 1=Notice 2=Announcement (see NoticeType* constants)
	Content *string // Content
	FileIds *string // Attachment file IDs, comma-separated
	Status  *int    // Status: 0=Draft 1=Published (see NoticeStatus* constants)
	Remark  *string // Remark
}

// Update updates notice information.
func (s *serviceImpl) Update(ctx context.Context, in UpdateInput) error {
	noticeColumns := dao.Notice.Columns()

	// Check notice exists and get old status.
	var oldNotice *NoticeEntity
	err := s.tenantFilter.Apply(ctx, dao.Notice.Ctx(ctx)).
		Where(noticeColumns.Id, in.Id).
		Scan(&oldNotice)
	if err != nil {
		return err
	}
	if oldNotice == nil {
		return bizerr.NewCode(CodeNoticeNotFound)
	}

	bizCtx := s.bizCtxSvc.Current(ctx)
	updatedBy := int64(bizCtx.UserID)

	data := do.Notice{UpdatedBy: updatedBy}
	if in.Title != nil {
		data.Title = *in.Title
	}
	if in.Type != nil {
		data.Type = *in.Type
	}
	if in.Content != nil {
		data.Content = *in.Content
	}
	if in.FileIds != nil {
		data.FileIds = *in.FileIds
	}
	if in.Status != nil {
		data.Status = *in.Status
	}
	if in.Remark != nil {
		data.Remark = *in.Remark
	}

	_, err = dao.Notice.Ctx(ctx).
		OmitNilData().
		Where(plugincontract.TenantFilterColumn, s.tenantFilter.Current(ctx)).
		Where(noticeColumns.Id, in.Id).
		Data(data).
		Update()
	if err != nil {
		return err
	}

	// If status changed from draft(0) to published(1), dispatch inbox notifications.
	if in.Status != nil && *in.Status == NoticeStatusPublished && oldNotice.Status == NoticeStatusDraft {
		title := oldNotice.Title
		if in.Title != nil {
			title = *in.Title
		}
		content := oldNotice.Content
		if in.Content != nil {
			content = *in.Content
		}
		noticeType := oldNotice.Type
		if in.Type != nil {
			noticeType = *in.Type
		}
		if dispatchErr := s.dispatchPublishedNotice(ctx, in.Id, title, content, noticeType, oldNotice.CreatedBy); dispatchErr != nil {
			logger.Errorf(ctx, "dispatch published notice failed for notice %d: %v", in.Id, dispatchErr)
		}
	}

	return nil
}

// Delete soft-deletes notices by IDs and cascades to notify deliveries.
func (s *serviceImpl) Delete(ctx context.Context, ids string) error {
	idList := normalizeNoticeDeleteIDs(ids)
	if len(idList) == 0 {
		return bizerr.NewCode(CodeNoticeDeleteRequired)
	}

	// Soft delete using GoFrame's auto soft-delete feature.
	noticeColumns := dao.Notice.Columns()
	_, err := s.tenantFilter.Apply(ctx, dao.Notice.Ctx(ctx)).
		WhereIn(noticeColumns.Id, idList).
		Delete()
	if err != nil {
		return err
	}

	if cascadeErr := s.notifySvc.DeleteBySource(ctx, plugincontract.SourceTypeNotice, idList); cascadeErr != nil {
		logger.Errorf(ctx, "cascade delete notify deliveries failed for notice ids %s: %v", ids, cascadeErr)
	}
	return nil
}

// normalizeNoticeDeleteIDs trims comma-separated notice IDs and removes empty
// entries before passing them to the DAO layer.
func normalizeNoticeDeleteIDs(ids string) []string {
	rawIDs := strings.Split(ids, ",")
	result := make([]string, 0, len(rawIDs))
	for _, id := range rawIDs {
		normalizedID := strings.TrimSpace(id)
		if normalizedID == "" {
			continue
		}
		result = append(result, normalizedID)
	}
	return result
}
