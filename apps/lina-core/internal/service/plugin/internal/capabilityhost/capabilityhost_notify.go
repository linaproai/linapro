// This file adapts host notification services and storage to
// plugin-visible notification capability contracts.
package capabilityhost

import (
	"context"
	"strconv"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/internal/dao"
	"lina-core/internal/model/entity"
	notifysvc "lina-core/internal/service/notify"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/capmodel"
	capabilitynotifycap "lina-core/pkg/plugin/capability/notifycap"
)

// Publisher defines the host notification slice required by this adapter.
type notificationPublisher interface {
	// Send validates the notify channel and creates message and delivery records.
	Send(ctx context.Context, in notifysvc.SendInput) (*notifysvc.SendOutput, error)
	// SendNoticePublication sends one published notice through the inbox channel.
	SendNoticePublication(ctx context.Context, in notifysvc.NoticePublishInput) (*notifysvc.SendOutput, error)
	// DeleteBySource removes notify records for the given business source identifiers.
	DeleteBySource(ctx context.Context, sourceType notifysvc.SourceType, sourceIDs []string) error
}

// Service exposes the notification domain service and management commands.
type notificationCapabilityService interface {
	capabilitynotifycap.Service
	capabilitynotifycap.AdminService
}

// adapter exposes notification projections and governed sends.
type notificationCapabilityAdapter struct {
	publisher notificationPublisher
}

var (
	_ capabilitynotifycap.Service      = (*notificationCapabilityAdapter)(nil)
	_ capabilitynotifycap.AdminService = (*notificationCapabilityAdapter)(nil)
)

// New creates the host-owned notification capability adapter.
func newNotificationCapabilityAdapter(publisher notificationPublisher) notificationCapabilityService {
	return &notificationCapabilityAdapter{publisher: publisher}
}

// BatchGet returns visible notification message projections.
func (a *notificationCapabilityAdapter) BatchGet(ctx context.Context, capCtx capmodel.CapabilityContext, ids []capabilitynotifycap.MessageID) (*capmodel.BatchResult[map[string]any, capabilitynotifycap.MessageID], error) {
	result := &capmodel.BatchResult[map[string]any, capabilitynotifycap.MessageID]{
		Items:      make(map[capabilitynotifycap.MessageID]map[string]any, len(ids)),
		MissingIDs: []capabilitynotifycap.MessageID{},
	}
	parsedIDs, requested := ParseInt64IDs(ids, func(id capabilitynotifycap.MessageID) {
		result.MissingIDs = append(result.MissingIDs, id)
	})
	if len(parsedIDs) == 0 {
		return result, nil
	}
	rows := make([]*entity.SysNotifyMessage, 0, len(parsedIDs))
	cols := dao.SysNotifyMessage.Columns()
	model := dao.SysNotifyMessage.Ctx(ctx).
		Fields(cols.Id, cols.TenantId, cols.PluginId, cols.SourceType, cols.SourceId, cols.CategoryCode, cols.Title, cols.CreatedAt).
		WhereIn(cols.Id, parsedIDs)
	tenantID, _ := TenantID(capCtx.TenantID)
	if tenantID > PlatformTenantID {
		model = model.WhereIn(cols.TenantId, []int{PlatformTenantID, tenantID})
	}
	if err := model.Scan(&rows); err != nil {
		return nil, err
	}
	for _, row := range rows {
		if row == nil {
			continue
		}
		requestID, ok := requested[row.Id]
		if !ok {
			continue
		}
		result.Items[requestID] = map[string]any{
			"id":           requestID,
			"tenantId":     row.TenantId,
			"pluginId":     row.PluginId,
			"sourceType":   row.SourceType,
			"sourceId":     row.SourceId,
			"categoryCode": row.CategoryCode,
			"title":        row.Title,
			"createdAt":    row.CreatedAt,
		}
	}
	for _, id := range ids {
		if _, ok := result.Items[id]; !ok && !Contains(result.MissingIDs, id) {
			result.MissingIDs = append(result.MissingIDs, id)
		}
	}
	return result, nil
}

// Send sends one governed notification message through the shared notification service.
func (a *notificationCapabilityAdapter) Send(ctx context.Context, capCtx capmodel.CapabilityContext, input capabilitynotifycap.SendInput) (*capabilitynotifycap.SendResult, error) {
	if a == nil || a.publisher == nil {
		return nil, bizerr.NewCode(capmodel.CodeCapabilityUnavailable, bizerr.P("capability", "notification"))
	}
	sourceType := notifysvc.SourceType(strings.TrimSpace(string(input.SourceType)))
	sourceID := strings.TrimSpace(input.SourceID)
	channelKey := strings.TrimSpace(input.ChannelKey)
	if channelKey == "" {
		channelKey = notifysvc.ChannelKeyInbox
	}
	senderUserID := input.SenderUserID
	if senderUserID == 0 {
		senderUserID = capCtx.Actor.UserID
	}
	var (
		out *notifysvc.SendOutput
		err error
	)
	if sourceType == notifysvc.SourceTypeNotice && len(input.Recipients) == 0 {
		noticeID, parseErr := strconv.ParseInt(sourceID, 10, 64)
		if parseErr != nil || noticeID <= 0 {
			return nil, bizerr.NewCode(capmodel.CodeCapabilityDenied)
		}
		out, err = a.publisher.SendNoticePublication(ctx, notifysvc.NoticePublishInput{
			NoticeID:     noticeID,
			Title:        input.Title,
			Content:      input.Content,
			CategoryCode: notifysvc.CategoryCode(input.Category),
			SenderUserID: senderUserID,
		})
	} else {
		recipientIDs, parseErr := ParsePositiveInt64Strings(input.Recipients)
		if parseErr != nil {
			return nil, bizerr.NewCode(capmodel.CodeCapabilityDenied)
		}
		out, err = a.publisher.Send(ctx, notifysvc.SendInput{
			ChannelKey:       channelKey,
			PluginID:         capCtx.PluginID,
			SourceType:       sourceType,
			SourceID:         sourceID,
			CategoryCode:     notifysvc.CategoryCode(input.Category),
			Title:            input.Title,
			Content:          input.Content,
			Payload:          input.Payload,
			SenderUserID:     senderUserID,
			RecipientUserIDs: recipientIDs,
		})
	}
	if err != nil || out == nil {
		return nil, err
	}
	return &capabilitynotifycap.SendResult{
		MessageID:     capabilitynotifycap.MessageID(strconv.FormatInt(out.MessageID, 10)),
		DeliveryCount: out.DeliveryCount,
	}, nil
}

// Delete removes visible notification messages.
func (a *notificationCapabilityAdapter) Delete(ctx context.Context, capCtx capmodel.CapabilityContext, ids []capabilitynotifycap.MessageID) error {
	result, err := a.BatchGet(ctx, capCtx, ids)
	if err != nil {
		return err
	}
	if len(result.MissingIDs) > 0 {
		return bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	parsedIDs, _ := ParseInt64IDs(ids, nil)
	if len(parsedIDs) == 0 {
		return bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	return dao.SysNotifyMessage.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, err = tx.Model(dao.SysNotifyDelivery.Table()).Safe().Ctx(ctx).
			WhereIn(dao.SysNotifyDelivery.Columns().MessageId, parsedIDs).
			Delete(); err != nil {
			return err
		}
		_, err = tx.Model(dao.SysNotifyMessage.Table()).Safe().Ctx(ctx).
			WhereIn(dao.SysNotifyMessage.Columns().Id, parsedIDs).
			Delete()
		return err
	})
}

// DeleteBySource removes visible notification messages by business source IDs.
func (a *notificationCapabilityAdapter) DeleteBySource(ctx context.Context, _ capmodel.CapabilityContext, sourceType capabilitynotifycap.SourceType, sourceIDs []string) error {
	if a == nil || a.publisher == nil {
		return bizerr.NewCode(capmodel.CodeCapabilityUnavailable, bizerr.P("capability", "notification"))
	}
	return a.publisher.DeleteBySource(ctx, notifysvc.SourceType(sourceType), sourceIDs)
}
