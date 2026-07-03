// This file adapts host notification services and storage to
// plugin-visible notification capability contracts.
package notify

import (
	"context"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"

	usermsgv1 "lina-core/api/usermsg/v1"
	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/datascope"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/capmodel"
	capabilitynotifycap "lina-core/pkg/plugin/capability/notifycap"
)

// adapter exposes notification projections and governed sends.
type notificationCapabilityAdapter struct {
	publisher Service
	bizCtx    bizctxcap.Service
	pluginID  string
}

// CapabilityService is the notify-owned plugin capability adapter contract.
// The plugin binding keeps source and dynamic notifications attributed to the
// caller without exposing the adapter implementation type.
type CapabilityService interface {
	capabilitynotifycap.Service
	// ForPlugin returns a notification capability bound to one plugin ID.
	ForPlugin(pluginID string) capabilitynotifycap.Service
}

var _ capabilitynotifycap.Service = (*notificationCapabilityAdapter)(nil)
var _ CapabilityService = (*notificationCapabilityAdapter)(nil)

// NewCapabilityAdapter creates the host-owned notification capability adapter.
func NewCapabilityAdapter(publisher Service, bizCtx bizctxcap.Service) CapabilityService {
	return &notificationCapabilityAdapter{publisher: publisher, bizCtx: bizCtx}
}

// ForPlugin returns one notification adapter bound to the source or dynamic plugin ID.
func (a *notificationCapabilityAdapter) ForPlugin(pluginID string) capabilitynotifycap.Service {
	if a == nil {
		return nil
	}
	clone := *a
	clone.pluginID = strings.TrimSpace(pluginID)
	return &clone
}

// Get returns one visible notification message projection.
func (a *notificationCapabilityAdapter) Get(ctx context.Context, id capabilitynotifycap.MessageID) (*capabilitynotifycap.MessageInfo, error) {
	result, err := a.BatchGet(ctx, []capabilitynotifycap.MessageID{id})
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	if projection := result.Items[id]; projection != nil {
		return projection, nil
	}
	return nil, bizerr.NewCode(capmodel.CodeCapabilityDenied)
}

// BatchGet returns visible notification message projections.
func (a *notificationCapabilityAdapter) BatchGet(ctx context.Context, ids []capabilitynotifycap.MessageID) (*capmodel.BatchResult[*capabilitynotifycap.MessageInfo, capabilitynotifycap.MessageID], error) {
	if len(ids) > capabilitynotifycap.MaxBatchGetMessages {
		return nil, bizerr.NewCode(capmodel.CodeCapabilityLimitExceeded, bizerr.P("limit", capabilitynotifycap.MaxBatchGetMessages))
	}
	current, err := a.notificationActorContext(ctx)
	if err != nil {
		return nil, err
	}
	result := &capmodel.BatchResult[*capabilitynotifycap.MessageInfo, capabilitynotifycap.MessageID]{
		Items:      make(map[capabilitynotifycap.MessageID]*capabilitynotifycap.MessageInfo, len(ids)),
		MissingIDs: []capabilitynotifycap.MessageID{},
	}
	parsedIDs, requested := capmodel.ParseInt64IDs(ids, func(id capabilitynotifycap.MessageID) {
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
		Where(deliveryTbl+"."+deliveryCols.UserId, current.UserID)
	tenantID := current.TenantID
	if tenantID > datascope.PlatformTenantID {
		model = model.WhereIn(messageTbl+"."+messageCols.TenantId, []int{datascope.PlatformTenantID, tenantID})
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
		if _, ok := result.Items[id]; !ok && !slices.Contains(result.MissingIDs, id) {
			result.MissingIDs = append(result.MissingIDs, id)
		}
	}
	return result, nil
}

// List returns one bounded page of visible notification messages.
func (a *notificationCapabilityAdapter) List(ctx context.Context, input capabilitynotifycap.ListInput) (*capmodel.PageResult[*capabilitynotifycap.MessageInfo], error) {
	current, err := a.notificationActorContext(ctx)
	if err != nil {
		return nil, err
	}
	pageNum, pageSize := input.Page.Normalize()
	if pageSize > capabilitynotifycap.MaxListMessagesPageSize {
		pageSize = capabilitynotifycap.MaxListMessagesPageSize
	}
	var (
		messageCols  = dao.SysNotifyMessage.Columns()
		deliveryCols = dao.SysNotifyDelivery.Columns()
		messageTbl   = dao.SysNotifyMessage.Table()
		deliveryTbl  = dao.SysNotifyDelivery.Table()
	)
	model := dao.SysNotifyMessage.Ctx(ctx).
		InnerJoin(deliveryTbl, deliveryTbl+"."+deliveryCols.MessageId+"="+messageTbl+"."+messageCols.Id).
		Where(deliveryTbl+"."+deliveryCols.UserId, current.UserID)
	if sourceType := strings.TrimSpace(string(input.SourceType)); sourceType != "" {
		model = model.Where(messageTbl+"."+messageCols.SourceType, sourceType)
	}
	if sourceID := strings.TrimSpace(input.SourceID); sourceID != "" {
		model = model.Where(messageTbl+"."+messageCols.SourceId, sourceID)
	}
	if category := strings.TrimSpace(string(input.CategoryCode)); category != "" {
		model = model.Where(messageTbl+"."+messageCols.CategoryCode, category)
	}
	tenantID := current.TenantID
	if tenantID > datascope.PlatformTenantID {
		model = model.WhereIn(messageTbl+"."+messageCols.TenantId, []int{datascope.PlatformTenantID, tenantID})
	}
	total, err := model.Clone().Count()
	if err != nil {
		return nil, err
	}
	rows := make([]*entity.SysNotifyMessage, 0, pageSize)
	if err = model.Clone().
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
		Page(pageNum, pageSize).
		OrderDesc(messageTbl + "." + messageCols.Id).
		Scan(&rows); err != nil {
		return nil, err
	}
	items := make([]*capabilitynotifycap.MessageInfo, 0, len(rows))
	for _, row := range rows {
		if row == nil {
			continue
		}
		items = append(items, projectNotifyMessage(row, capabilitynotifycap.MessageID(strconv.FormatInt(row.Id, 10))))
	}
	return &capmodel.PageResult[*capabilitynotifycap.MessageInfo]{Items: items, Total: total}, nil
}

// BatchGetBySource returns visible notification message projections grouped by source ID.
func (a *notificationCapabilityAdapter) BatchGetBySource(
	ctx context.Context,
	input capabilitynotifycap.BatchGetBySourceInput,
) (*capabilitynotifycap.BatchGetBySourceResult, error) {
	sourceType := strings.TrimSpace(string(input.SourceType))
	sourceIDs := normalizeNotifySourceIDs(input.SourceIDs)
	if len(sourceIDs) > capabilitynotifycap.MaxBatchGetBySourceIDs {
		return nil, bizerr.NewCode(capmodel.CodeCapabilityLimitExceeded, bizerr.P("limit", capabilitynotifycap.MaxBatchGetBySourceIDs))
	}
	current, err := a.notificationActorContext(ctx)
	if err != nil {
		return nil, err
	}
	result := &capabilitynotifycap.BatchGetBySourceResult{
		Items:      map[string][]*capabilitynotifycap.MessageInfo{},
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
		Where(deliveryTbl+"."+deliveryCols.UserId, current.UserID).
		OrderDesc(messageTbl + "." + messageCols.Id).
		Limit(capabilitynotifycap.MaxBatchGetBySourceMessages + 1)
	tenantID := current.TenantID
	if tenantID > datascope.PlatformTenantID {
		model = model.WhereIn(messageTbl+"."+messageCols.TenantId, []int{datascope.PlatformTenantID, tenantID})
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
func (a *notificationCapabilityAdapter) EnsureVisible(ctx context.Context, ids []capabilitynotifycap.MessageID) error {
	if len(ids) > capabilitynotifycap.MaxEnsureVisibleMessages {
		return bizerr.NewCode(capmodel.CodeCapabilityLimitExceeded, bizerr.P("limit", capabilitynotifycap.MaxEnsureVisibleMessages))
	}
	result, err := a.BatchGet(ctx, ids)
	if err != nil {
		return err
	}
	if len(result.MissingIDs) > 0 || len(result.Items) != len(parseUniqueNotifyMessageIDs(ids)) {
		return bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	return nil
}

// Send sends one governed notification message through the shared notification service.
func (a *notificationCapabilityAdapter) Send(ctx context.Context, input capabilitynotifycap.SendInput) (*capabilitynotifycap.SendResult, error) {
	if a == nil || a.publisher == nil {
		return nil, bizerr.NewCode(capmodel.CodeCapabilityUnavailable, bizerr.P("capability", "notification"))
	}
	current, err := a.notificationActorContext(ctx)
	if err != nil {
		return nil, err
	}
	var (
		sourceType = SourceType(strings.TrimSpace(string(input.SourceType)))
		sourceID   = strings.TrimSpace(input.SourceID)
		channelKey = strings.TrimSpace(input.ChannelKey)
	)
	if channelKey == "" {
		channelKey = ChannelKeyInbox
	}
	senderUserID := input.SenderUserID
	if senderUserID == 0 {
		senderUserID = int64(current.UserID)
	}
	var (
		out *SendOutput
	)
	if sourceType == usermsgv1.SourceTypeNotice && len(input.Recipients) == 0 {
		noticeID, parseErr := strconv.ParseInt(sourceID, 10, 64)
		if parseErr != nil || noticeID <= 0 {
			return nil, bizerr.NewCode(capmodel.CodeCapabilityDenied)
		}
		out, err = a.publisher.SendNoticePublication(ctx, NoticePublishInput{
			NoticeID:     noticeID,
			Title:        input.Title,
			Content:      input.Content,
			CategoryCode: CategoryCode(input.Category),
			SenderUserID: senderUserID,
		})
	} else {
		recipientIDs, parseErr := capmodel.ParsePositiveInt64Strings(input.Recipients)
		if parseErr != nil {
			return nil, bizerr.NewCode(capmodel.CodeCapabilityDenied)
		}
		out, err = a.publisher.Send(ctx, SendInput{
			ChannelKey:       channelKey,
			PluginID:         a.pluginID,
			SourceType:       sourceType,
			SourceID:         sourceID,
			CategoryCode:     CategoryCode(input.Category),
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
func (a *notificationCapabilityAdapter) Delete(ctx context.Context, ids []capabilitynotifycap.MessageID) error {
	if err := a.EnsureVisible(ctx, ids); err != nil {
		return err
	}
	parsedIDs, _ := capmodel.ParseInt64IDs(ids, nil)
	if len(parsedIDs) == 0 {
		return bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	return dao.SysNotifyMessage.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		if _, err := dao.SysNotifyDelivery.Ctx(ctx).
			WhereIn(dao.SysNotifyDelivery.Columns().MessageId, parsedIDs).
			Delete(); err != nil {
			return err
		}
		_, err := dao.SysNotifyMessage.Ctx(ctx).
			WhereIn(dao.SysNotifyMessage.Columns().Id, parsedIDs).
			Delete()
		return err
	})
}

// DeleteBySource removes visible notification messages by business source IDs.
func (a *notificationCapabilityAdapter) DeleteBySource(ctx context.Context, sourceType usermsgv1.SourceType, sourceIDs []string) error {
	if a == nil || a.publisher == nil {
		return bizerr.NewCode(capmodel.CodeCapabilityUnavailable, bizerr.P("capability", "notification"))
	}
	return a.publisher.DeleteBySource(ctx, sourceType, sourceIDs)
}

// MarkRead marks visible notifications as read for the current actor.
func (a *notificationCapabilityAdapter) MarkRead(ctx context.Context, ids []capabilitynotifycap.MessageID) error {
	now := time.Now()
	return a.setReadState(ctx, ids, 1, &now)
}

// MarkUnread marks visible notifications as unread for the current actor.
func (a *notificationCapabilityAdapter) MarkUnread(ctx context.Context, ids []capabilitynotifycap.MessageID) error {
	return a.setReadState(ctx, ids, 0, nil)
}

// setReadState updates the current actor's delivery rows after message visibility checks.
func (a *notificationCapabilityAdapter) setReadState(ctx context.Context, ids []capabilitynotifycap.MessageID, isRead int, readAt *time.Time) error {
	current, err := a.notificationActorContext(ctx)
	if err != nil {
		return err
	}
	if err := a.EnsureVisible(ctx, ids); err != nil {
		return err
	}
	parsedIDs, _ := capmodel.ParseInt64IDs(ids, nil)
	if len(parsedIDs) == 0 {
		return bizerr.NewCode(capmodel.CodeCapabilityDenied)
	}
	data := do.SysNotifyDelivery{IsRead: isRead}
	if readAt != nil {
		data.ReadAt = readAt
	}
	model := dao.SysNotifyDelivery.Ctx(ctx).
		WhereIn(dao.SysNotifyDelivery.Columns().MessageId, parsedIDs).
		Where(do.SysNotifyDelivery{UserId: current.UserID})
	_, err = model.Data(data).Update()
	return err
}

// notificationActorContext requires a current user for notification read and
// delivery-state operations.
func (a *notificationCapabilityAdapter) notificationActorContext(ctx context.Context) (bizctxcap.CurrentContext, error) {
	current := bizctxcap.CurrentFromContext(ctx)
	if a != nil && a.bizCtx != nil {
		current = a.bizCtx.Current(ctx)
	}
	if current.UserID <= 0 {
		return bizctxcap.CurrentContext{}, bizerr.NewCode(capmodel.CodeCapabilityCurrentUserRequired)
	}
	return current, nil
}

// projectNotifyMessage converts one host notify row into the stable plugin projection.
func projectNotifyMessage(row *entity.SysNotifyMessage, id capabilitynotifycap.MessageID) *capabilitynotifycap.MessageInfo {
	if row == nil {
		return nil
	}
	return &capabilitynotifycap.MessageInfo{
		ID:           id,
		TenantID:     row.TenantId,
		PluginID:     row.PluginId,
		SourceType:   usermsgv1.SourceType(row.SourceType),
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
	parsedIDs, _ := capmodel.ParseInt64IDs(ids, nil)
	return parsedIDs
}

// unixMilli converts nullable database timestamps into Unix milliseconds.
func unixMilli(value *time.Time) int64 {
	if value == nil {
		return 0
	}
	return value.UnixMilli()
}
