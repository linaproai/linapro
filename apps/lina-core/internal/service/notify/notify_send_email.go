// This file implements email channel send orchestration through the optional
// mail-core email delivery bridge registered on notifycap.

package notify

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/datascope"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/notifycap"
)

// emailChannelConfig is the optional JSON config for email notify channels.
type emailChannelConfig struct {
	// AccountID selects a linapro-mail-core Account; zero means default.
	AccountID int64 `json:"accountId"`
}

// sendEmail creates message/delivery rows and delivers through mail-core.
func (s *serviceImpl) sendEmail(
	ctx context.Context,
	channel *entity.SysNotifyChannel,
	in SendInput,
) (*SendOutput, error) {
	delivery := notifycap.EmailDeliveryOrNil()
	if delivery == nil {
		return nil, bizerr.NewCode(CodeNotifyEmailDeliveryUnavailable)
	}

	normalizedTitle := strings.TrimSpace(in.Title)
	if normalizedTitle == "" {
		return nil, bizerr.NewCode(CodeNotifyTitleRequired)
	}

	recipientEmails, err := s.resolveEmailRecipients(ctx, in.RecipientUserIDs)
	if err != nil {
		return nil, err
	}
	if len(recipientEmails) == 0 {
		return nil, bizerr.NewCode(CodeNotifyEmailRecipientRequired)
	}

	accountID, err := parseEmailChannelAccountID(channel.ConfigJson)
	if err != nil {
		return nil, err
	}

	payloadJSON, err := marshalNotifyPayload(in.Payload)
	if err != nil {
		return nil, err
	}

	var (
		now           = time.Now()
		tenantID      = datascope.CurrentTenantID(ctx)
		sourceType    = normalizeSourceType(in.SourceType)
		categoryCode  = normalizeCategoryCode(in.CategoryCode)
		messageID     int64
		deliveryCount int
	)

	err = dao.SysNotifyMessage.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		messageID, err = dao.SysNotifyMessage.Ctx(ctx).Data(do.SysNotifyMessage{
			PluginId:     strings.TrimSpace(in.PluginID),
			SourceType:   string(sourceType),
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

		for _, item := range recipientEmails {
			status := DeliveryStatusPending
			if _, err = dao.SysNotifyDelivery.Ctx(ctx).Data(do.SysNotifyDelivery{
				MessageId:      messageID,
				ChannelKey:     channel.ChannelKey,
				ChannelType:    channel.ChannelType,
				RecipientType:  RecipientTypeEmail.String(),
				RecipientKey:   item.email,
				UserId:         item.userID,
				DeliveryStatus: status,
				IsRead:         0,
				ReadAt:         nil,
				SentAt:         nil,
				TenantId:       tenantID,
			}).Insert(); err != nil {
				return bizerr.WrapCode(err, CodeNotifyDeliveryCreateFailed)
			}
			deliveryCount++
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	result, deliverErr := delivery.Deliver(ctx, notifycap.EmailDeliveryInput{
		AccountID: accountID,
		To:        emailAddresses(recipientEmails),
		Subject:   normalizedTitle,
		Content:   in.Content,
	})
	if deliverErr != nil {
		_ = s.markEmailDeliveriesFailed(ctx, messageID)
		return nil, bizerr.WrapCode(deliverErr, CodeNotifyEmailDeliveryFailed)
	}
	_ = result
	_ = s.markEmailDeliveriesSucceeded(ctx, messageID, now)

	return &SendOutput{
		MessageID:     messageID,
		DeliveryCount: deliveryCount,
	}, nil
}

type emailRecipient struct {
	userID int64
	email  string
}

func (s *serviceImpl) resolveEmailRecipients(ctx context.Context, userIDs []int64) ([]emailRecipient, error) {
	ids := uniquePositiveUserIDs(userIDs)
	if len(ids) == 0 {
		return nil, nil
	}
	var users []entity.SysUser
	if err := dao.SysUser.Ctx(ctx).
		Fields(dao.SysUser.Columns().Id, dao.SysUser.Columns().Email).
		WhereIn(dao.SysUser.Columns().Id, ids).
		Scan(&users); err != nil {
		return nil, bizerr.WrapCode(err, CodeNotifyRecipientQueryFailed)
	}
	out := make([]emailRecipient, 0, len(users))
	for _, user := range users {
		email := strings.TrimSpace(user.Email)
		if email == "" {
			continue
		}
		out = append(out, emailRecipient{userID: int64(user.Id), email: email})
	}
	return out, nil
}

func emailAddresses(items []emailRecipient) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, item.email)
	}
	return out
}

func parseEmailChannelAccountID(configJSON string) (int64, error) {
	configJSON = strings.TrimSpace(configJSON)
	if configJSON == "" || configJSON == "{}" {
		return 0, nil
	}
	var cfg emailChannelConfig
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		return 0, bizerr.NewCode(CodeNotifyChannelTypeUnsupported, bizerr.P("channelType", ChannelTypeEmail.String()))
	}
	if cfg.AccountID < 0 {
		return 0, nil
	}
	return cfg.AccountID, nil
}

func (s *serviceImpl) markEmailDeliveriesSucceeded(ctx context.Context, messageID int64, at time.Time) error {
	_, err := dao.SysNotifyDelivery.Ctx(ctx).
		Where(dao.SysNotifyDelivery.Columns().MessageId, messageID).
		Where(dao.SysNotifyDelivery.Columns().ChannelType, ChannelTypeEmail.String()).
		Data(do.SysNotifyDelivery{
			DeliveryStatus: DeliveryStatusSucceeded,
			SentAt:         &at,
		}).Update()
	return err
}

func (s *serviceImpl) markEmailDeliveriesFailed(ctx context.Context, messageID int64) error {
	_, err := dao.SysNotifyDelivery.Ctx(ctx).
		Where(dao.SysNotifyDelivery.Columns().MessageId, messageID).
		Where(dao.SysNotifyDelivery.Columns().ChannelType, ChannelTypeEmail.String()).
		Data(do.SysNotifyDelivery{
			DeliveryStatus: DeliveryStatusFailed,
		}).Update()
	return err
}
