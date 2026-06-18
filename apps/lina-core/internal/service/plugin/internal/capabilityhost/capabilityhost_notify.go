// This file adapts host notification services and storage to
// plugin-visible notification capability contracts.
package capabilityhost

import (
	"context"
	"strconv"
	"strings"
	"time"

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
func (a *notificationCapabilityAdapter) BatchGet(ctx context.Context, capCtx capmodel.CapabilityContext, ids []capabilitynotifycap.MessageID) (*capmodel.BatchResult[*capabilitynotifycap.MessageProjection, capabilitynotifycap.MessageID], error) {
	if len(ids) > capabilitynotifycap.MaxBatchGetMessages {
		return nil, bizerr.NewCode(capmodel.CodeCapabilityLimitExceeded, bizerr.P("limit", capabilitynotifycap.MaxBatchGetMessages))
	}
	if capCtx.Actor.UserID <= 0 {
		return nil, bizerr.NewCode(capmodel.CodeCapabilityActorRequired)
	}
	result := &capmodel.BatchResult[*capabilitynotifycap.MessageProjection, capabilitynotifycap.MessageID]{
		Items:      make(map[capabilitynotifycap.MessageID]*capabilitynotifycap.MessageProjection, len(ids)),
		MissingIDs: []capabilitynotifycap.MessageID{},
	}
	parsedIDs, requested := ParseInt64IDs(ids, func(id capabilitynotifycap.MessageID) {
		result.MissingIDs = append(result.MissingIDs, id)
	})
	if len(parsedIDs) == 0 {
		return result, nil
	}
	rows := make([]*entity.SysNotifyMessage, 0, len(parsedIDs))
	var (
		messageCols  = dao.SysNotifyMessage.Columns()
		deliveryCols = dao.SysNotifyDelivery.Columns()
		messageTbl   = dao.SysNotifyMessage.Table()
		deliveryTbl  = dao.SysNotifyDelivery.Table()
	)
	model := dao.SysNotifyMessage.Ctx(ctx).
		InnerJoin(deliveryTbl, deliveryTbl+"."+deliveryCols.MessageId+"="+messageTbl+"."+messageCols.Id).
		Fields(
			messageTbl+"."+messageCols.Id+" AS id",
			messageTbl+"."+messageCols.TenantId+" AS tenant_id",
			messageTbl+"."+messageCols.PluginId+" AS plugin_id",
			messageTbl+"."+messageCols.SourceType+" AS source_type",
			messageTbl+"."+messageCols.SourceId+" AS source_id",
			messageTbl+"."+messageCols.CategoryCode+" AS category_code",
			messageTbl+"."+messageCols.Title+" AS title",
			messageTbl+"."+messageCols.CreatedAt+" AS created_at",
		).
		WhereIn(messageTbl+"."+messageCols.Id, parsedIDs).
		Where(deliveryTbl+"."+deliveryCols.UserId, capCtx.Actor.UserID)
	tenantID, _ := TenantID(capCtx.TenantID)
	if tenantID > PlatformTenantID {
		model = model.WhereIn(messageTbl+"."+messageCols.TenantId, []int{PlatformTenantID, tenantID})
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
		result.Items[requestID] = projectNotifyMessage(row, requestID)
	}
	for _, id := range ids {
		if _, ok := result.Items[id]; !ok && !Contains(result.MissingIDs, id) {
			result.MissingIDs = append(result.MissingIDs, id)
		}
	}
	return result, nil
}

// BatchGetBySource returns visible notification message projections grouped by source ID.
func (a *notificationCapabilityAdapter) BatchGetBySource(
	ctx context.Context,
	capCtx capmodel.CapabilityContext,
	input capabilitynotifycap.BatchGetBySourceInput,
) (*capabilitynotifycap.BatchGetBySourceResult, error) {
	sourceType := strings.TrimSpace(string(input.SourceType))
	sourceIDs := normalizeNotifySourceIDs(input.SourceIDs)
	if len(sourceIDs) > capabilitynotifycap.MaxBatchGetBySourceIDs {
		return nil, bizerr.NewCode(capmodel.CodeCapabilityLimitExceeded, bizerr.P("limit", capabilitynotifycap.MaxBatchGetBySourceIDs))
	}
	if capCtx.Actor.UserID <= 0 {
		return nil, bizerr.NewCode(capmodel.CodeCapabilityActorRequired)
	}
	result := &capabilitynotifycap.BatchGetBySourceResult{
		Items:      map[string][]*capabilitynotifycap.MessageProjection{},
		MissingIDs: []string{},
	}
	if sourceType == "" || len(sourceIDs) == 0 {
		result.MissingIDs = append(result.MissingIDs, sourceIDs...)
		return result, nil
	}
	var (
		messageCols  = dao.SysNotifyMessage.Columns()
		deliveryCols = dao.SysNotifyDelivery.Columns()
		messageTbl   = dao.SysNotifyMessage.Table()
		deliveryTbl  = dao.SysNotifyDelivery.Table()
	)
	rows := make([]*entity.SysNotifyMessage, 0, len(sourceIDs))
	model := dao.SysNotifyMessage.Ctx(ctx).
		InnerJoin(deliveryTbl, deliveryTbl+"."+deliveryCols.MessageId+"="+messageTbl+"."+messageCols.Id).
		Fields(
			messageTbl+"."+messageCols.Id+" AS id",
			messageTbl+"."+messageCols.TenantId+" AS tenant_id",
			messageTbl+"."+messageCols.PluginId+" AS plugin_id",
			messageTbl+"."+messageCols.SourceType+" AS source_type",
			messageTbl+"."+messageCols.SourceId+" AS source_id",
			messageTbl+"."+messageCols.CategoryCode+" AS category_code",
			messageTbl+"."+messageCols.Title+" AS title",
			messageTbl+"."+messageCols.CreatedAt+" AS created_at",
		).
		Where(messageTbl+"."+messageCols.SourceType, sourceType).
		WhereIn(messageTbl+"."+messageCols.SourceId, sourceIDs).
		Where(deliveryTbl+"."+deliveryCols.UserId, capCtx.Actor.UserID).
		OrderDesc(messageTbl + "." + messageCols.Id).
		Limit(capabilitynotifycap.MaxBatchGetBySourceMessages + 1)
	tenantID, _ := TenantID(capCtx.TenantID)
	if tenantID > PlatformTenantID {
		model = model.WhereIn(messageTbl+"."+messageCols.TenantId, []int{PlatformTenantID, tenantID})
	}
	if err := model.Scan(&rows); err != nil {
		return nil, err
	}
	if len(rows) > capabilitynotifycap.MaxBatchGetBySourceMessages {
		return nil, bizerr.NewCode(capmodel.CodeCapabilityLimitExceeded, bizerr.P("limit", capabilitynotifycap.MaxBatchGetBySourceMessages))
	}
	visibleSources := make(map[string]struct{}, len(sourceIDs))
	for _, row := range rows {
		if row == nil {
			continue
		}
		sourceID := strings.TrimSpace(row.SourceId)
		if sourceID == "" {
			continue
		}
		visibleSources[sourceID] = struct{}{}
		result.Items[sourceID] = append(result.Items[sourceID], projectNotifyMessage(row, capabilitynotifycap.MessageID(strconv.FormatInt(row.Id, 10))))
	}
	for _, sourceID := range sourceIDs {
		if _, ok := visibleSources[sourceID]; !ok {
			result.MissingIDs = append(result.MissingIDs, sourceID)
		}
	}
	return result, nil
}

// EnsureVisible rejects when any requested notification message is absent or invisible.
func (a *notificationCapabilityAdapter) EnsureVisible(ctx context.Context, capCtx capmodel.CapabilityContext, ids []capabilitynotifycap.MessageID) error {
	if len(ids) > capabilitynotifycap.MaxEnsureVisibleMessages {
		return bizerr.NewCode(capmodel.CodeCapabilityLimitExceeded, bizerr.P("limit", capabilitynotifycap.MaxEnsureVisibleMessages))
	}
	result, err := a.BatchGet(ctx, capCtx, ids)
	if err != nil {
		return err
	}
	if len(result.MissingIDs) > 0 || len(result.Items) != len(parseUniqueNotifyMessageIDs(ids)) {
		return bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	return nil
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
	if err := a.EnsureVisible(ctx, capCtx, ids); err != nil {
		return err
	}
	parsedIDs, _ := ParseInt64IDs(ids, nil)
	if len(parsedIDs) == 0 {
		return bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	return dao.SysNotifyMessage.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, err := tx.Model(dao.SysNotifyDelivery.Table()).Safe().Ctx(ctx).
			WhereIn(dao.SysNotifyDelivery.Columns().MessageId, parsedIDs).
			Delete(); err != nil {
			return err
		}
		_, err := tx.Model(dao.SysNotifyMessage.Table()).Safe().Ctx(ctx).
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

// projectNotifyMessage converts one host notify row into the stable plugin projection.
func projectNotifyMessage(row *entity.SysNotifyMessage, id capabilitynotifycap.MessageID) *capabilitynotifycap.MessageProjection {
	if row == nil {
		return nil
	}
	return &capabilitynotifycap.MessageProjection{
		ID:           id,
		TenantID:     row.TenantId,
		PluginID:     row.PluginId,
		SourceType:   capabilitynotifycap.SourceType(row.SourceType),
		SourceID:     row.SourceId,
		CategoryCode: capabilitynotifycap.CategoryCode(row.CategoryCode),
		Title:        row.Title,
		CreatedAt:    unixMilli(row.CreatedAt),
	}
}

// normalizeNotifySourceIDs trims, de-duplicates, and preserves source ID order.
func normalizeNotifySourceIDs(ids []string) []string {
	seen := make(map[string]struct{}, len(ids))
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		value := strings.TrimSpace(id)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

// parseUniqueNotifyMessageIDs returns distinct positive numeric message IDs.
func parseUniqueNotifyMessageIDs(ids []capabilitynotifycap.MessageID) []int64 {
	parsedIDs, _ := ParseInt64IDs(ids, nil)
	return parsedIDs
}

// unixMilli converts nullable database timestamps into Unix milliseconds.
func unixMilli(value *time.Time) int64 {
	if value == nil {
		return 0
	}
	return value.UnixMilli()
}
