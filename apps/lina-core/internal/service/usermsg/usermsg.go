package usermsg

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/internal/service/bizctx"
	notifysvc "lina-core/internal/service/notify"
)

// Service defines the usermsg service contract.
type Service interface {
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

var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	bizCtxSvc bizctx.Service
	notifySvc notifysvc.Service // Unified notify service
}

// New creates and returns a new Service instance.
func New() Service {
	return &serviceImpl{
		bizCtxSvc: bizctx.New(),
		notifySvc: notifysvc.New(),
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
	SourceType string      // Message source type
	SourceId   int64       // Message source ID
	IsRead     int         // Whether the message has been read
	ReadAt     *gtime.Time // Read time
	CreatedAt  *gtime.Time // Creation time
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
