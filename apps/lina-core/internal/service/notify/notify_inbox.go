// This file implements inbox query, read, delete, and source cleanup behaviors.

package notify

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
)

type inboxListRecord struct {
	Id           int64       `json:"id"`
	UserId       int64       `json:"userId"`
	Title        string      `json:"title"`
	CategoryCode string      `json:"categoryCode"`
	SourceType   string      `json:"sourceType"`
	SourceId     string      `json:"sourceId"`
	IsRead       int         `json:"isRead"`
	ReadAt       *gtime.Time `json:"readAt"`
	CreatedAt    *gtime.Time `json:"createdAt"`
}

// InboxUnreadCount returns the unread inbox delivery count for one user.
func (s *serviceImpl) InboxUnreadCount(ctx context.Context, userID int64) (int, error) {
	if userID <= 0 {
		return 0, gerror.New("用户不存在")
	}

	return dao.SysNotifyDelivery.Ctx(ctx).Where(do.SysNotifyDelivery{
		UserId:         userID,
		ChannelType:    ChannelTypeInbox.String(),
		DeliveryStatus: DeliveryStatusSucceeded,
		IsRead:         0,
	}).Count()
}

// InboxList returns one paged inbox list for the current user.
func (s *serviceImpl) InboxList(ctx context.Context, in InboxListInput) (*InboxListOutput, error) {
	if in.UserID <= 0 {
		return nil, gerror.New("用户不存在")
	}

	var (
		deliveryCols = dao.SysNotifyDelivery.Columns()
		messageCols  = dao.SysNotifyMessage.Columns()
		deliveryTbl  = dao.SysNotifyDelivery.Table()
		messageTbl   = dao.SysNotifyMessage.Table()
		model        = dao.SysNotifyDelivery.Ctx(ctx).
				LeftJoin(
				messageTbl,
				messageTbl+"."+messageCols.Id+"="+deliveryTbl+"."+deliveryCols.MessageId,
			).
			Where(deliveryTbl+"."+deliveryCols.UserId, in.UserID).
			Where(deliveryTbl+"."+deliveryCols.ChannelType, ChannelTypeInbox.String()).
			Where(deliveryTbl+"."+deliveryCols.DeliveryStatus, DeliveryStatusSucceeded)
	)

	total, err := model.Count()
	if err != nil {
		return nil, err
	}

	var rows []*inboxListRecord
	err = model.Fields(
		deliveryTbl+"."+deliveryCols.Id+" AS id",
		deliveryTbl+"."+deliveryCols.UserId+" AS user_id",
		messageTbl+"."+messageCols.Title+" AS title",
		messageTbl+"."+messageCols.CategoryCode+" AS category_code",
		messageTbl+"."+messageCols.SourceType+" AS source_type",
		messageTbl+"."+messageCols.SourceId+" AS source_id",
		deliveryTbl+"."+deliveryCols.IsRead+" AS is_read",
		deliveryTbl+"."+deliveryCols.ReadAt+" AS read_at",
		deliveryTbl+"."+deliveryCols.CreatedAt+" AS created_at",
	).Page(in.PageNum, in.PageSize).
		OrderDesc(deliveryTbl + "." + deliveryCols.Id).
		Scan(&rows)
	if err != nil {
		return nil, err
	}

	items := make([]*InboxListItem, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		items = append(items, &InboxListItem{
			Id:         row.Id,
			UserID:     row.UserId,
			Title:      row.Title,
			Type:       categoryCodeToLegacyMessageType(CategoryCode(row.CategoryCode)),
			SourceType: row.SourceType,
			SourceID:   gconv.Int64(row.SourceId),
			IsRead:     row.IsRead,
			ReadAt:     row.ReadAt,
			CreatedAt:  row.CreatedAt,
		})
	}

	return &InboxListOutput{
		List:  items,
		Total: total,
	}, nil
}

// InboxMarkRead marks one inbox delivery as read for the current user.
func (s *serviceImpl) InboxMarkRead(ctx context.Context, userID int64, deliveryID int64) error {
	if userID <= 0 {
		return gerror.New("用户不存在")
	}

	_, err := dao.SysNotifyDelivery.Ctx(ctx).Where(do.SysNotifyDelivery{
		Id:             deliveryID,
		UserId:         userID,
		ChannelType:    ChannelTypeInbox.String(),
		DeliveryStatus: DeliveryStatusSucceeded,
	}).Data(do.SysNotifyDelivery{
		IsRead: 1,
		ReadAt: gtime.Now(),
	}).Update()
	return err
}

// InboxMarkAllRead marks all unread inbox deliveries as read for the current user.
func (s *serviceImpl) InboxMarkAllRead(ctx context.Context, userID int64) error {
	if userID <= 0 {
		return gerror.New("用户不存在")
	}

	deliveryCols := dao.SysNotifyDelivery.Columns()
	_, err := dao.SysNotifyDelivery.Ctx(ctx).
		Where(deliveryCols.UserId, userID).
		Where(deliveryCols.ChannelType, ChannelTypeInbox.String()).
		Where(deliveryCols.DeliveryStatus, DeliveryStatusSucceeded).
		Where(deliveryCols.IsRead, 0).
		Data(do.SysNotifyDelivery{
			IsRead: 1,
			ReadAt: gtime.Now(),
		}).
		Update()
	return err
}

// InboxDelete soft-deletes one inbox delivery for the current user.
func (s *serviceImpl) InboxDelete(ctx context.Context, userID int64, deliveryID int64) error {
	if userID <= 0 {
		return gerror.New("用户不存在")
	}

	_, err := dao.SysNotifyDelivery.Ctx(ctx).Where(do.SysNotifyDelivery{
		Id:          deliveryID,
		UserId:      userID,
		ChannelType: ChannelTypeInbox.String(),
	}).Delete()
	return err
}

// InboxClear soft-deletes all inbox deliveries for the current user.
func (s *serviceImpl) InboxClear(ctx context.Context, userID int64) error {
	if userID <= 0 {
		return gerror.New("用户不存在")
	}

	_, err := dao.SysNotifyDelivery.Ctx(ctx).Where(do.SysNotifyDelivery{
		UserId:      userID,
		ChannelType: ChannelTypeInbox.String(),
	}).Delete()
	return err
}

// DeleteBySource removes notify deliveries and messages for the given business source identifiers.
func (s *serviceImpl) DeleteBySource(ctx context.Context, sourceType SourceType, sourceIDs []string) error {
	normalizedSourceIDs := normalizeSourceIDs(sourceIDs)
	if len(normalizedSourceIDs) == 0 {
		return nil
	}

	messageCols := dao.SysNotifyMessage.Columns()
	var rows []struct {
		Id int64 `json:"id"`
	}
	err := dao.SysNotifyMessage.Ctx(ctx).
		Fields(messageCols.Id).
		Where(messageCols.SourceType, sourceType.String()).
		WhereIn(messageCols.SourceId, normalizedSourceIDs).
		Scan(&rows)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}

	messageIDs := make([]int64, 0, len(rows))
	for _, row := range rows {
		if row.Id > 0 {
			messageIDs = append(messageIDs, row.Id)
		}
	}
	if len(messageIDs) == 0 {
		return nil
	}

	deliveryCols := dao.SysNotifyDelivery.Columns()
	return dao.SysNotifyDelivery.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, err = dao.SysNotifyDelivery.Ctx(ctx).WhereIn(deliveryCols.MessageId, messageIDs).Delete(); err != nil {
			return err
		}
		if _, err = dao.SysNotifyMessage.Ctx(ctx).WhereIn(messageCols.Id, messageIDs).Delete(); err != nil {
			return err
		}
		return nil
	})
}

func normalizeSourceIDs(sourceIDs []string) []string {
	if len(sourceIDs) == 0 {
		return []string{}
	}

	result := make([]string, 0, len(sourceIDs))
	seen := make(map[string]struct{}, len(sourceIDs))
	for _, sourceID := range sourceIDs {
		normalized := strings.TrimSpace(sourceID)
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	return result
}
