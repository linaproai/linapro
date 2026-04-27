package usermsg

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/internal/service/bizctx"
	i18nsvc "lina-core/internal/service/i18n"
	notifysvc "lina-core/internal/service/notify"
)

// Stable i18n keys used to localize the user-message type label so the host
// inbox UI does not need to map message.type to free-form text on the client.
const (
	usermsgTypeNoticeI18nKey       = "usermsg.type.notice"
	usermsgTypeAnnouncementI18nKey = "usermsg.type.announcement"
)

// Service defines the usermsg service contract.
type Service interface {
	// Get returns one current-user message detail for preview consumption.
	Get(ctx context.Context, id int64) (*MessageDetail, error)
	// UnreadCount returns unread message count for current user.
	UnreadCount(ctx context.Context) (int, error)
	// List queries message list for current user with pagination.
	List(ctx context.Context, in ListInput) (*ListOutput, error)
	// MarkRead marks a single message as read.
	MarkRead(ctx context.Context, id int64) error
	// MarkReadAll marks all messages as read for current user.
	MarkReadAll(ctx context.Context) error
	// Delete deletes a single message for current user.
	Delete(ctx context.Context, id int64) error
	// Clear deletes all messages for current user.
	Clear(ctx context.Context) error
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	bizCtxSvc bizctx.Service
	notifySvc notifysvc.Service // Unified notify service
	i18nSvc   i18nsvc.Service   // Host i18n service for type label localization
}

// New creates and returns a new Service instance.
func New() Service {
	return &serviceImpl{
		bizCtxSvc: bizctx.New(),
		notifySvc: notifysvc.New(),
		i18nSvc:   i18nsvc.New(),
	}
}

// localizeType translates the legacy numeric inbox message type into the
// current request locale. The host owns this label so that the inbox UI can
// render type names without depending on plugin-installed dictionaries.
// English source text is used as the safety fallback so the inbox never
// renders an empty cell if a locale resource happens to be missing.
func (s *serviceImpl) localizeType(ctx context.Context, msgType int) string {
	if s == nil || s.i18nSvc == nil {
		return ""
	}
	switch msgType {
	case 2:
		return s.i18nSvc.Translate(ctx, usermsgTypeAnnouncementI18nKey, "Announcement")
	default:
		return s.i18nSvc.Translate(ctx, usermsgTypeNoticeI18nKey, "Notice")
	}
}

// getCurrentUserId extracts current user ID from context.
func (s *serviceImpl) getCurrentUserId(ctx context.Context) (int64, error) {
	bizCtx := s.bizCtxSvc.Get(ctx)
	if bizCtx == nil || bizCtx.UserId == 0 {
		return 0, gerror.New("未登录")
	}
	return int64(bizCtx.UserId), nil
}

// UnreadCount returns unread message count for current user.
func (s *serviceImpl) UnreadCount(ctx context.Context) (int, error) {
	userId, err := s.getCurrentUserId(ctx)
	if err != nil {
		return 0, err
	}
	return s.notifySvc.InboxUnreadCount(ctx, userId)
}

// ListInput defines input for List function.
type ListInput struct {
	PageNum  int // Page number, starting from 1
	PageSize int // Page size
}

// ListOutput defines output for List function.
type ListOutput struct {
	List  []*MessageItem // Message list
	Total int            // Total count
}

// MessageItem defines one user message facade item.
type MessageItem struct {
	Id         int64       // Message ID
	UserId     int64       // Recipient user ID
	Title      string      // Message title
	Type       int         // Message type: 1=Notice 2=Announcement
	TypeLabel  string      // Localized type label resolved at the host
	SourceType string      // Message source type
	SourceId   int64       // Message source ID
	IsRead     int         // Whether the message has been read
	ReadAt     *gtime.Time // Read time
	CreatedAt  *gtime.Time // Creation time
}

// MessageDetail defines one current-user message detail payload used by the
// inbox preview dialog.
type MessageDetail struct {
	Id            int64       // Message ID
	Title         string      // Message title
	Type          int         // Message type: 1=Notice 2=Announcement
	TypeLabel     string      // Localized type label resolved at the host
	SourceType    string      // Message source type
	SourceId      int64       // Message source ID
	Content       string      // Renderable message body content
	CreatedByName string      // Sender display name
	CreatedAt     *gtime.Time // Message creation time
}

// List queries message list for current user with pagination.
func (s *serviceImpl) List(ctx context.Context, in ListInput) (*ListOutput, error) {
	userId, err := s.getCurrentUserId(ctx)
	if err != nil {
		return nil, err
	}

	out, err := s.notifySvc.InboxList(ctx, notifysvc.InboxListInput{
		UserID:   userId,
		PageNum:  in.PageNum,
		PageSize: in.PageSize,
	})
	if err != nil {
		return nil, err
	}

	items := make([]*MessageItem, 0, len(out.List))
	for _, item := range out.List {
		if item == nil {
			continue
		}
		items = append(items, &MessageItem{
			Id:         item.Id,
			UserId:     item.UserID,
			Title:      item.Title,
			Type:       item.Type,
			TypeLabel:  s.localizeType(ctx, item.Type),
			SourceType: item.SourceType,
			SourceId:   item.SourceID,
			IsRead:     item.IsRead,
			ReadAt:     item.ReadAt,
			CreatedAt:  item.CreatedAt,
		})
	}

	return &ListOutput{List: items, Total: out.Total}, nil
}

// MarkRead marks a single message as read.
func (s *serviceImpl) MarkRead(ctx context.Context, id int64) error {
	userId, err := s.getCurrentUserId(ctx)
	if err != nil {
		return err
	}
	return s.notifySvc.InboxMarkRead(ctx, userId, id)
}

// MarkReadAll marks all messages as read for current user.
func (s *serviceImpl) MarkReadAll(ctx context.Context) error {
	userId, err := s.getCurrentUserId(ctx)
	if err != nil {
		return err
	}
	return s.notifySvc.InboxMarkAllRead(ctx, userId)
}

// Delete deletes a single message for current user.
func (s *serviceImpl) Delete(ctx context.Context, id int64) error {
	userId, err := s.getCurrentUserId(ctx)
	if err != nil {
		return err
	}
	return s.notifySvc.InboxDelete(ctx, userId, id)
}

// Clear deletes all messages for current user.
func (s *serviceImpl) Clear(ctx context.Context) error {
	userId, err := s.getCurrentUserId(ctx)
	if err != nil {
		return err
	}
	return s.notifySvc.InboxClear(ctx, userId)
}
