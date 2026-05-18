// This file implements notification send orchestration and notice publication fan-out.

package notify

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/datascope"
	"lina-core/pkg/bizerr"
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
		return nil, bizerr.NewCode(
			CodeNotifyChannelTypeUnsupported,
			bizerr.P("channelType", channel.ChannelType),
		)
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

// sendInbox validates inbox recipients and persists one notify message with
// corresponding inbox delivery rows.
func (s *serviceImpl) sendInbox(
	ctx context.Context,
	channel *entity.SysNotifyChannel,
	in SendInput,
) (*SendOutput, error) {
	normalizedTitle := strings.TrimSpace(in.Title)
	if normalizedTitle == "" {
		return nil, bizerr.NewCode(CodeNotifyTitleRequired)
	}

	recipientUserIDs := uniquePositiveUserIDs(in.RecipientUserIDs)
	if len(recipientUserIDs) == 0 {
		return nil, bizerr.NewCode(CodeNotifyInboxRecipientRequired)
	}

	payloadJSON, err := marshalNotifyPayload(in.Payload)
	if err != nil {
		return nil, err
	}

	var (
		now           = gtime.Now()
		tenantID      = datascope.CurrentTenantID(ctx)
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
			TenantId:     tenantID,
		}).InsertAndGetId()
		if err != nil {
			return bizerr.WrapCode(err, CodeNotifyMessageCreateFailed)
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
				TenantId:       tenantID,
			}).Insert(); err != nil {
				return bizerr.WrapCode(err, CodeNotifyDeliveryCreateFailed)
			}
			deliveryCount++
		}

		return nil
	})
	if err != nil {
		if _, ok := bizerr.As(err); ok {
			return nil, err
		}
		return nil, bizerr.WrapCode(err, CodeNotifyMessageCreateFailed)
	}

	return &SendOutput{
		MessageID:     messageID,
		DeliveryCount: deliveryCount,
	}, nil
}

// getChannel loads one enabled global notify channel by key.
func (s *serviceImpl) getChannel(ctx context.Context, channelKey string) (*entity.SysNotifyChannel, error) {
	normalizedChannelKey := strings.TrimSpace(channelKey)
	if normalizedChannelKey == "" {
		return nil, bizerr.NewCode(CodeNotifyChannelKeyRequired)
	}

	var channel *entity.SysNotifyChannel
	err := dao.SysNotifyChannel.Ctx(ctx).Where(do.SysNotifyChannel{
		ChannelKey: normalizedChannelKey,
		Status:     ChannelStatusEnabled,
	}).Scan(&channel)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeNotifyChannelQueryFailed)
	}
	if channel == nil {
		return nil, bizerr.NewCode(CodeNotifyChannelUnavailable)
	}
	return channel, nil
}

// listActiveInboxUserIDs returns deliverable enabled users except the optional
// excluded sender. Tenant contexts use active membership as the delivery
// boundary, while platform context stays constrained to platform users.
func (s *serviceImpl) listActiveInboxUserIDs(ctx context.Context, excludedUserID int64) ([]int64, error) {
	var (
		userCols = dao.SysUser.Columns()
		tenantID = datascope.CurrentTenantID(ctx)
		model    = dao.SysUser.Ctx(ctx).Fields(userCols.Id).Where(do.SysUser{Status: 1})
	)
	if tenantID == datascope.PlatformTenantID {
		model = model.Where(do.SysUser{TenantId: datascope.PlatformTenantID})
	} else if s == nil || s.tenantSvc == nil || !s.tenantSvc.Enabled(ctx) {
		model = model.Where(do.SysUser{TenantId: tenantID})
	} else {
		var err error
		model, _, err = s.tenantSvc.ApplyUserTenantScope(ctx, model, userCols.Id)
		if err != nil {
			return nil, bizerr.WrapCode(err, CodeNotifyRecipientQueryFailed)
		}
	}
	if excludedUserID > 0 {
		model = model.WhereNot(userCols.Id, excludedUserID)
	}

	var users []*entity.SysUser
	if err := model.Scan(&users); err != nil {
		return nil, bizerr.WrapCode(err, CodeNotifyRecipientQueryFailed)
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

// marshalNotifyPayload serializes optional notify payload metadata into the
// persisted JSON string form.
func marshalNotifyPayload(payload map[string]any) (string, error) {
	if len(payload) == 0 {
		return "{}", nil
	}

	content, err := json.Marshal(payload)
	if err != nil {
		return "", bizerr.WrapCode(err, CodeNotifyPayloadMarshalFailed)
	}
	return string(content), nil
}

// uniquePositiveUserIDs trims invalid IDs and de-duplicates the remaining
// positive user identifiers while preserving order.
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

// normalizeSourceType falls back to the system source type when the input is
// empty.
func normalizeSourceType(sourceType SourceType) SourceType {
	if strings.TrimSpace(sourceType.String()) == "" {
		return SourceTypeSystem
	}
	return sourceType
}

// normalizeCategoryCode falls back to the generic category when the input is
// empty.
func normalizeCategoryCode(categoryCode CategoryCode) CategoryCode {
	if strings.TrimSpace(categoryCode.String()) == "" {
		return CategoryCodeOther
	}
	return categoryCode
}
