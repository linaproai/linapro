// This file implements notification send orchestration and notice publication fan-out.

package notify

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
)

// Send validates the notify channel and creates unified notify message and delivery records.
func (s *serviceImpl) Send(ctx context.Context, in SendInput) (*SendOutput, error) {
	channel, err := s.getChannel(ctx, in.ChannelKey)
	if err != nil {
		return nil, err
	}

	switch ChannelType(channel.ChannelType) {
	case ChannelTypeInbox:
		return s.sendInbox(ctx, channel, in)
	default:
		return nil, gerror.Newf("通知通道类型暂不支持: %s", channel.ChannelType)
	}
}

// SendNoticePublication sends one published notice through the built-in inbox channel.
func (s *serviceImpl) SendNoticePublication(ctx context.Context, in NoticePublishInput) (*SendOutput, error) {
	recipientUserIDs, err := s.listActiveInboxUserIDs(ctx, in.SenderUserID)
	if err != nil {
		return nil, err
	}
	if len(recipientUserIDs) == 0 {
		return &SendOutput{}, nil
	}

	return s.Send(ctx, SendInput{
		ChannelKey:       ChannelKeyInbox,
		SourceType:       SourceTypeNotice,
		SourceID:         gconv.String(in.NoticeID),
		CategoryCode:     in.CategoryCode,
		Title:            in.Title,
		Content:          in.Content,
		Payload:          map[string]any{},
		SenderUserID:     in.SenderUserID,
		RecipientUserIDs: recipientUserIDs,
	})
}

func (s *serviceImpl) sendInbox(
	ctx context.Context,
	channel *entity.SysNotifyChannel,
	in SendInput,
) (*SendOutput, error) {
	normalizedTitle := strings.TrimSpace(in.Title)
	if normalizedTitle == "" {
		return nil, gerror.New("通知标题不能为空")
	}

	recipientUserIDs := uniquePositiveUserIDs(in.RecipientUserIDs)
	if len(recipientUserIDs) == 0 {
		return nil, gerror.New("站内信接收用户不能为空")
	}

	payloadJSON, err := marshalNotifyPayload(in.Payload)
	if err != nil {
		return nil, err
	}

	var (
		now           = gtime.Now()
		sourceType    = normalizeSourceType(in.SourceType)
		categoryCode  = normalizeCategoryCode(in.CategoryCode)
		messageID     int64
		deliveryCount int
	)

	err = dao.SysNotifyMessage.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		messageID, err = dao.SysNotifyMessage.Ctx(ctx).Data(do.SysNotifyMessage{
			PluginId:     strings.TrimSpace(in.PluginID),
			SourceType:   sourceType.String(),
			SourceId:     strings.TrimSpace(in.SourceID),
			CategoryCode: categoryCode.String(),
			Title:        normalizedTitle,
			Content:      in.Content,
			PayloadJson:  payloadJSON,
			SenderUserId: in.SenderUserID,
		}).InsertAndGetId()
		if err != nil {
			return err
		}

		for _, userID := range recipientUserIDs {
			if _, err = dao.SysNotifyDelivery.Ctx(ctx).Data(do.SysNotifyDelivery{
				MessageId:      messageID,
				ChannelKey:     channel.ChannelKey,
				ChannelType:    channel.ChannelType,
				RecipientType:  RecipientTypeUser.String(),
				RecipientKey:   gconv.String(userID),
				UserId:         userID,
				DeliveryStatus: DeliveryStatusSucceeded,
				IsRead:         0,
				SentAt:         now,
			}).Insert(); err != nil {
				return err
			}
			deliveryCount++
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &SendOutput{
		MessageID:     messageID,
		DeliveryCount: deliveryCount,
	}, nil
}

func (s *serviceImpl) getChannel(ctx context.Context, channelKey string) (*entity.SysNotifyChannel, error) {
	normalizedChannelKey := strings.TrimSpace(channelKey)
	if normalizedChannelKey == "" {
		return nil, gerror.New("通知通道标识不能为空")
	}

	var channel *entity.SysNotifyChannel
	err := dao.SysNotifyChannel.Ctx(ctx).Where(do.SysNotifyChannel{
		ChannelKey: normalizedChannelKey,
		Status:     ChannelStatusEnabled,
	}).Scan(&channel)
	if err != nil {
		return nil, err
	}
	if channel == nil {
		return nil, gerror.New("通知通道不存在或已停用")
	}
	return channel, nil
}

func (s *serviceImpl) listActiveInboxUserIDs(ctx context.Context, excludedUserID int64) ([]int64, error) {
	userCols := dao.SysUser.Columns()
	model := dao.SysUser.Ctx(ctx).Fields(userCols.Id).Where(do.SysUser{Status: 1})
	if excludedUserID > 0 {
		model = model.WhereNot(userCols.Id, excludedUserID)
	}

	var users []*entity.SysUser
	if err := model.Scan(&users); err != nil {
		return nil, err
	}

	userIDs := make([]int64, 0, len(users))
	for _, user := range users {
		if user == nil || user.Id <= 0 {
			continue
		}
		userIDs = append(userIDs, int64(user.Id))
	}
	return userIDs, nil
}

func marshalNotifyPayload(payload map[string]any) (string, error) {
	if len(payload) == 0 {
		return "{}", nil
	}

	content, err := json.Marshal(payload)
	if err != nil {
		return "", gerror.Wrap(err, "序列化通知扩展载荷失败")
	}
	return string(content), nil
}

func uniquePositiveUserIDs(userIDs []int64) []int64 {
	if len(userIDs) == 0 {
		return []int64{}
	}

	result := make([]int64, 0, len(userIDs))
	seen := make(map[int64]struct{}, len(userIDs))
	for _, userID := range userIDs {
		if userID <= 0 {
			continue
		}
		if _, ok := seen[userID]; ok {
			continue
		}
		seen[userID] = struct{}{}
		result = append(result, userID)
	}
	return result
}

func normalizeSourceType(sourceType SourceType) SourceType {
	if strings.TrimSpace(sourceType.String()) == "" {
		return SourceTypeSystem
	}
	return sourceType
}

func normalizeCategoryCode(categoryCode CategoryCode) CategoryCode {
	if strings.TrimSpace(categoryCode.String()) == "" {
		return CategoryCodeOther
	}
	return categoryCode
}
