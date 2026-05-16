// This file contains current-user inbox operations that delegate ownership and
// persistence checks to the unified notification service.

package usermsg

import (
	"context"

	notifysvc "lina-core/internal/service/notify"
	"lina-core/pkg/bizerr"
)

// getCurrentUserId extracts current user ID from context.
func (s *serviceImpl) getCurrentUserId(ctx context.Context) (int64, error) {
	bizCtx := s.bizCtxSvc.Get(ctx)
	if bizCtx == nil || bizCtx.UserId == 0 {
		return 0, bizerr.NewCode(CodeUserMsgNotAuthenticated)
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
		categoryCode := resolveCategoryCode(item.CategoryCode)
		items = append(items, &MessageItem{
			Id:           item.Id,
			UserId:       item.UserID,
			Title:        item.Title,
			CategoryCode: categoryCode,
			TypeLabel:    s.localizeCategoryLabel(ctx, categoryCode),
			TypeColor:    s.localizeCategoryColor(ctx, categoryCode),
			SourceType:   item.SourceType,
			SourceId:     item.SourceID,
			IsRead:       item.IsRead,
			ReadAt:       item.ReadAt,
			CreatedAt:    item.CreatedAt,
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
