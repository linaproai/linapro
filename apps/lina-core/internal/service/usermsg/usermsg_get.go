// usermsg_get.go implements current-user message detail loading for the inbox
// preview dialog without requiring additional notice-management permissions.

package usermsg

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"

	"lina-core/internal/dao"
	notifysvc "lina-core/internal/service/notify"
)

// messageDetailRecord is the joined database projection used to build one
// current-user message detail response.
type messageDetailRecord struct {
	Id            int64       `json:"id"`
	Title         string      `json:"title"`
	CategoryCode  string      `json:"categoryCode"`
	SourceType    string      `json:"sourceType"`
	SourceId      string      `json:"sourceId"`
	Content       string      `json:"content"`
	CreatedByName string      `json:"createdByName"`
	CreatedAt     *gtime.Time `json:"createdAt"`
}

// Get returns one current-user message detail for preview consumption.
func (s *serviceImpl) Get(ctx context.Context, id int64) (*MessageDetail, error) {
	userId, err := s.getCurrentUserId(ctx)
	if err != nil {
		return nil, err
	}
	if id <= 0 {
		return nil, gerror.New("消息不存在")
	}

	var (
		deliveryCols = dao.SysNotifyDelivery.Columns()
		messageCols  = dao.SysNotifyMessage.Columns()
		userCols     = dao.SysUser.Columns()
		deliveryTbl  = dao.SysNotifyDelivery.Table()
		messageTbl   = dao.SysNotifyMessage.Table()
		userTbl      = dao.SysUser.Table()
		record       *messageDetailRecord
	)

	err = dao.SysNotifyDelivery.Ctx(ctx).
		LeftJoin(
			messageTbl,
			messageTbl+"."+messageCols.Id+"="+deliveryTbl+"."+deliveryCols.MessageId,
		).
		LeftJoin(
			userTbl,
			userTbl+"."+userCols.Id+"="+messageTbl+"."+messageCols.SenderUserId,
		).
		Fields(
			deliveryTbl+"."+deliveryCols.Id+" AS id",
			messageTbl+"."+messageCols.Title+" AS title",
			messageTbl+"."+messageCols.CategoryCode+" AS category_code",
			messageTbl+"."+messageCols.SourceType+" AS source_type",
			messageTbl+"."+messageCols.SourceId+" AS source_id",
			messageTbl+"."+messageCols.Content+" AS content",
			userTbl+"."+userCols.Username+" AS created_by_name",
			messageTbl+"."+messageCols.CreatedAt+" AS created_at",
		).
		Where(deliveryTbl+"."+deliveryCols.Id, id).
		Where(deliveryTbl+"."+deliveryCols.UserId, userId).
		Where(deliveryTbl+"."+deliveryCols.ChannelType, notifysvc.ChannelTypeInbox.String()).
		Where(deliveryTbl+"."+deliveryCols.DeliveryStatus, notifysvc.DeliveryStatusSucceeded).
		Scan(&record)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, gerror.New("消息不存在")
	}

	categoryCode := resolveCategoryCode(record.CategoryCode)
	return &MessageDetail{
		Id:            record.Id,
		Title:         record.Title,
		CategoryCode:  categoryCode,
		TypeLabel:     s.localizeCategoryLabel(ctx, categoryCode),
		TypeColor:     s.localizeCategoryColor(ctx, categoryCode),
		SourceType:    record.SourceType,
		SourceId:      gconv.Int64(record.SourceId),
		Content:       record.Content,
		CreatedByName: record.CreatedByName,
		CreatedAt:     record.CreatedAt,
	}, nil
}
