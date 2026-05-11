// This file verifies tenant isolation for notification send and notice fan-out.

package notify

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gtime"
	_ "lina-core/pkg/dbdriver"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/datascope"
	pkgtenantcap "lina-core/pkg/tenantcap"
)

const (
	// notifyTenantMembershipTable is the plugin-owned membership table used by tests.
	notifyTenantMembershipTable = "plugin_multi_tenant_user_membership"
	// notifyTenantMembershipStatusActive marks active test memberships.
	notifyTenantMembershipStatusActive = 1
)

// notifyTenantTestMembershipData is the typed payload used by tenant membership
// test setup.
type notifyTenantTestMembershipData struct {
	UserID   int64       `orm:"user_id"`
	TenantID int64       `orm:"tenant_id"`
	Status   int         `orm:"status"`
	JoinedAt *gtime.Time `orm:"joined_at"`
}

// notifyTenantTestMembershipCleanupFilter is the typed filter used to clean
// temporary membership rows.
type notifyTenantTestMembershipCleanupFilter struct {
	UserID int64 `orm:"user_id"`
}

// notifyTenantTestDeliveryRow is the delivery projection used by assertions.
type notifyTenantTestDeliveryRow struct {
	Id       int64 `json:"id" orm:"id"`
	UserID   int64 `json:"userId" orm:"user_id"`
	TenantID int   `json:"tenantId" orm:"tenant_id"`
}

// notifyTenantTestUserTenantRow is the sys_user tenant projection used by
// platform boundary assertions.
type notifyTenantTestUserTenantRow struct {
	TenantID int `json:"tenantId" orm:"tenant_id"`
}

// TestSendWritesCurrentTenantToMessageAndDelivery verifies direct inbox sends
// persist the current tenant on both notify tables.
func TestSendWritesCurrentTenantToMessageAndDelivery(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueNotifyTenantTestID()
	tenantCtx := datascope.WithTenantForTest(ctx, tenantID)
	ensureNotifyTenantTestInboxChannel(t, ctx)
	recipientID := insertNotifyTenantTestUser(t, ctx, "notify-send-recipient", tenantID, 1)
	t.Cleanup(func() { cleanupNotifyTenantTestUsers(t, ctx, []int{recipientID}) })

	out, err := New().Send(tenantCtx, SendInput{
		ChannelKey:       ChannelKeyInbox,
		SourceType:       SourceTypeSystem,
		SourceID:         uniqueNotifyTenantTestName("send-source"),
		CategoryCode:     CategoryCodeSystem,
		Title:            "Tenant send",
		Content:          "Tenant send content",
		RecipientUserIDs: []int64{int64(recipientID)},
	})
	if err != nil {
		t.Fatalf("send tenant notification: %v", err)
	}
	t.Cleanup(func() { cleanupNotifyTenantTestMessage(t, ctx, out.MessageID) })

	message := queryNotifyTenantTestMessage(t, ctx, out.MessageID)
	if message.TenantId != tenantID {
		t.Fatalf("expected message tenant %d, got %d", tenantID, message.TenantId)
	}
	deliveries := queryNotifyTenantTestDeliveries(t, ctx, out.MessageID)
	if len(deliveries) != 1 {
		t.Fatalf("expected one delivery, got %#v", deliveries)
	}
	if deliveries[0].TenantID != tenantID || deliveries[0].UserID != int64(recipientID) {
		t.Fatalf("expected delivery tenant/user %d/%d, got %#v", tenantID, recipientID, deliveries[0])
	}
}

// TestSendNoticePublicationUsesActiveMembershipBoundary verifies tenant notice
// fan-out uses active memberships instead of sys_user.tenant_id alone.
func TestSendNoticePublicationUsesActiveMembershipBoundary(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueNotifyTenantTestID()
	otherTenantID := tenantID + 1
	tenantCtx := datascope.WithTenantForTest(ctx, tenantID)
	ensureNotifyTenantTestInboxChannel(t, ctx)
	pkgtenantcap.RegisterProvider(notifyTenantTestProvider{})
	t.Cleanup(func() { pkgtenantcap.RegisterProvider(nil) })

	senderID := insertNotifyTenantTestUser(t, ctx, "notify-notice-sender", tenantID, 1)
	memberID := insertNotifyTenantTestUser(t, ctx, "notify-notice-member", otherTenantID, 1)
	sameTenantNonMemberID := insertNotifyTenantTestUser(t, ctx, "notify-notice-nonmember", tenantID, 1)
	inactiveMemberID := insertNotifyTenantTestUser(t, ctx, "notify-notice-inactive", tenantID, 1)
	otherTenantMemberID := insertNotifyTenantTestUser(t, ctx, "notify-notice-other-member", tenantID, 1)
	userIDs := []int{senderID, memberID, sameTenantNonMemberID, inactiveMemberID, otherTenantMemberID}
	t.Cleanup(func() {
		cleanupNotifyTenantTestMemberships(t, ctx, userIDs)
		cleanupNotifyTenantTestUsers(t, ctx, userIDs)
	})
	insertNotifyTenantTestMembership(t, ctx, senderID, tenantID, notifyTenantMembershipStatusActive)
	insertNotifyTenantTestMembership(t, ctx, memberID, tenantID, notifyTenantMembershipStatusActive)
	insertNotifyTenantTestMembership(t, ctx, inactiveMemberID, tenantID, 0)
	insertNotifyTenantTestMembership(t, ctx, otherTenantMemberID, otherTenantID, notifyTenantMembershipStatusActive)

	out, err := New(notifyTenantEnablementReader{}).SendNoticePublication(tenantCtx, NoticePublishInput{
		NoticeID:     int64(time.Now().UnixNano()),
		Title:        "Tenant notice",
		Content:      "Tenant notice content",
		CategoryCode: CategoryCodeSystem,
		SenderUserID: int64(senderID),
	})
	if err != nil {
		t.Fatalf("send tenant notice publication: %v", err)
	}
	t.Cleanup(func() { cleanupNotifyTenantTestMessage(t, ctx, out.MessageID) })

	if out.DeliveryCount != 1 {
		t.Fatalf("expected only one active tenant member delivery, got %d", out.DeliveryCount)
	}
	message := queryNotifyTenantTestMessage(t, ctx, out.MessageID)
	if message.TenantId != tenantID {
		t.Fatalf("expected notice message tenant %d, got %d", tenantID, message.TenantId)
	}
	deliveries := queryNotifyTenantTestDeliveries(t, ctx, out.MessageID)
	if len(deliveries) != 1 || deliveries[0].UserID != int64(memberID) || deliveries[0].TenantID != tenantID {
		t.Fatalf("expected only active membership user %d in tenant %d, got %#v", memberID, tenantID, deliveries)
	}
}

// TestSendNoticePublicationPlatformUsesPlatformUserBoundary verifies platform
// notice fan-out does not include tenant-primary users.
func TestSendNoticePublicationPlatformUsesPlatformUserBoundary(t *testing.T) {
	ctx := context.Background()
	tenantID := uniqueNotifyTenantTestID()
	platformCtx := datascope.WithTenantForTest(ctx, datascope.PlatformTenantID)
	ensureNotifyTenantTestInboxChannel(t, ctx)
	pkgtenantcap.RegisterProvider(notifyTenantTestProvider{})
	t.Cleanup(func() { pkgtenantcap.RegisterProvider(nil) })

	senderID := insertNotifyTenantTestUser(t, ctx, "notify-platform-sender", datascope.PlatformTenantID, 1)
	platformRecipientID := insertNotifyTenantTestUser(t, ctx, "notify-platform-recipient", datascope.PlatformTenantID, 1)
	tenantUserID := insertNotifyTenantTestUser(t, ctx, "notify-platform-tenant-user", tenantID, 1)
	userIDs := []int{senderID, platformRecipientID, tenantUserID}
	t.Cleanup(func() {
		cleanupNotifyTenantTestMemberships(t, ctx, userIDs)
		cleanupNotifyTenantTestUsers(t, ctx, userIDs)
	})
	insertNotifyTenantTestMembership(t, ctx, tenantUserID, tenantID, notifyTenantMembershipStatusActive)

	out, err := New(notifyTenantEnablementReader{}).SendNoticePublication(platformCtx, NoticePublishInput{
		NoticeID:     int64(time.Now().UnixNano()),
		Title:        "Platform notice",
		Content:      "Platform notice content",
		CategoryCode: CategoryCodeSystem,
		SenderUserID: int64(senderID),
	})
	if err != nil {
		t.Fatalf("send platform notice publication: %v", err)
	}
	t.Cleanup(func() { cleanupNotifyTenantTestMessage(t, ctx, out.MessageID) })

	message := queryNotifyTenantTestMessage(t, ctx, out.MessageID)
	if message.TenantId != datascope.PlatformTenantID {
		t.Fatalf("expected platform message tenant 0, got %d", message.TenantId)
	}
	deliveries := queryNotifyTenantTestDeliveries(t, ctx, out.MessageID)
	if !notifyTenantTestDeliveryContains(deliveries, int64(platformRecipientID)) {
		t.Fatalf("expected platform recipient %d in deliveries, got %#v", platformRecipientID, deliveries)
	}
	if notifyTenantTestDeliveryContains(deliveries, int64(tenantUserID)) {
		t.Fatalf("did not expect tenant user %d in platform deliveries, got %#v", tenantUserID, deliveries)
	}
	for _, delivery := range deliveries {
		if delivery.TenantID != datascope.PlatformTenantID {
			t.Fatalf("expected platform delivery tenant 0, got %#v", delivery)
		}
		userTenantID := queryNotifyTenantTestUserTenantID(t, ctx, delivery.UserID)
		if userTenantID != datascope.PlatformTenantID {
			t.Fatalf("expected platform user boundary, got user=%d tenant=%d", delivery.UserID, userTenantID)
		}
	}
}

// notifyTenantEnablementReader marks multi-tenancy enabled in notify tests.
type notifyTenantEnablementReader struct{}

// IsEnabled reports multi-tenant as enabled.
func (notifyTenantEnablementReader) IsEnabled(_ context.Context, pluginID string) bool {
	return pluginID == pkgtenantcap.ProviderPluginID
}

// notifyTenantTestProvider simulates the plugin-owned membership provider.
type notifyTenantTestProvider struct{}

// ResolveTenant returns platform for provider interface completeness.
func (notifyTenantTestProvider) ResolveTenant(context.Context, *ghttp.Request) (*pkgtenantcap.ResolverResult, error) {
	return &pkgtenantcap.ResolverResult{TenantID: pkgtenantcap.PLATFORM, Matched: true}, nil
}

// ValidateUserInTenant accepts all users for provider interface completeness.
func (notifyTenantTestProvider) ValidateUserInTenant(context.Context, int, pkgtenantcap.TenantID) error {
	return nil
}

// ListUserTenants returns no tenant rows for provider interface completeness.
func (notifyTenantTestProvider) ListUserTenants(context.Context, int) ([]pkgtenantcap.TenantInfo, error) {
	return nil, nil
}

// SwitchTenant accepts all switches for provider interface completeness.
func (notifyTenantTestProvider) SwitchTenant(context.Context, int, pkgtenantcap.TenantID) error {
	return nil
}

// ApplyUserTenantScope constrains notify fan-out by active membership.
func (notifyTenantTestProvider) ApplyUserTenantScope(
	ctx context.Context,
	model *gdb.Model,
	userIDColumn string,
) (*gdb.Model, bool, error) {
	tenantID := datascope.CurrentTenantID(ctx)
	if model == nil || tenantID == datascope.PlatformTenantID {
		return model, false, nil
	}
	return model.WhereIn(userIDColumn, notifyTenantActiveMembershipUserModel(ctx, tenantID)), false, nil
}

// ApplyUserTenantFilter is unused by notify tests.
func (notifyTenantTestProvider) ApplyUserTenantFilter(
	_ context.Context,
	model *gdb.Model,
	_ string,
	_ pkgtenantcap.TenantID,
) (*gdb.Model, bool, error) {
	return model, false, nil
}

// ListUserTenantProjections is unused by notify tests.
func (notifyTenantTestProvider) ListUserTenantProjections(
	context.Context,
	[]int,
) (map[int]*pkgtenantcap.UserTenantProjection, error) {
	return map[int]*pkgtenantcap.UserTenantProjection{}, nil
}

// ResolveUserTenantAssignment is unused by notify tests.
func (notifyTenantTestProvider) ResolveUserTenantAssignment(
	context.Context,
	[]pkgtenantcap.TenantID,
	pkgtenantcap.UserTenantAssignmentMode,
) (*pkgtenantcap.UserTenantAssignmentPlan, error) {
	return &pkgtenantcap.UserTenantAssignmentPlan{}, nil
}

// ReplaceUserTenantAssignments is unused by notify tests.
func (notifyTenantTestProvider) ReplaceUserTenantAssignments(
	context.Context,
	int,
	*pkgtenantcap.UserTenantAssignmentPlan,
) error {
	return nil
}

// EnsureUsersInTenant is unused by notify tests.
func (notifyTenantTestProvider) EnsureUsersInTenant(context.Context, []int, pkgtenantcap.TenantID) error {
	return nil
}

// ValidateStartupConsistency is unused by notify tests.
func (notifyTenantTestProvider) ValidateStartupConsistency(context.Context) ([]string, error) {
	return nil, nil
}

// notifyTenantActiveMembershipUserModel returns active membership users.
func notifyTenantActiveMembershipUserModel(ctx context.Context, tenantID int) *gdb.Model {
	return g.DB().Model(notifyTenantMembershipTable).Safe().Ctx(ctx).
		Fields("user_id").
		Where("tenant_id", tenantID).
		Where("status", notifyTenantMembershipStatusActive)
}

// ensureNotifyTenantTestInboxChannel guarantees the built-in inbox channel is
// available for notify service tests.
func ensureNotifyTenantTestInboxChannel(t *testing.T, ctx context.Context) {
	t.Helper()

	var channel *entity.SysNotifyChannel
	if err := dao.SysNotifyChannel.Ctx(ctx).Where(do.SysNotifyChannel{ChannelKey: ChannelKeyInbox}).Scan(&channel); err != nil {
		t.Fatalf("query inbox channel: %v", err)
	}
	if channel == nil {
		insertedID, err := dao.SysNotifyChannel.Ctx(ctx).Data(do.SysNotifyChannel{
			ChannelKey:  ChannelKeyInbox,
			Name:        "Inbox",
			ChannelType: ChannelTypeInbox.String(),
			Status:      ChannelStatusEnabled,
			ConfigJson:  "{}",
			Remark:      "notify tenant test",
		}).InsertAndGetId()
		if err != nil {
			t.Fatalf("insert inbox channel: %v", err)
		}
		t.Cleanup(func() {
			if _, cleanupErr := dao.SysNotifyChannel.Ctx(ctx).
				Unscoped().
				Where(do.SysNotifyChannel{Id: insertedID}).
				Delete(); cleanupErr != nil {
				t.Fatalf("cleanup inserted inbox channel: %v", cleanupErr)
			}
		})
		return
	}
	if channel.Status == ChannelStatusEnabled && channel.ChannelType == ChannelTypeInbox.String() {
		return
	}
	originalChannelType := channel.ChannelType
	originalStatus := channel.Status
	if _, err := dao.SysNotifyChannel.Ctx(ctx).
		Where(do.SysNotifyChannel{Id: channel.Id}).
		Data(do.SysNotifyChannel{
			ChannelType: ChannelTypeInbox.String(),
			Status:      ChannelStatusEnabled,
		}).
		Update(); err != nil {
		t.Fatalf("enable inbox channel: %v", err)
	}
	t.Cleanup(func() {
		if _, cleanupErr := dao.SysNotifyChannel.Ctx(ctx).
			Where(do.SysNotifyChannel{Id: channel.Id}).
			Data(do.SysNotifyChannel{
				ChannelType: originalChannelType,
				Status:      originalStatus,
			}).
			Update(); cleanupErr != nil {
			t.Fatalf("restore inbox channel: %v", cleanupErr)
		}
	})
}

// insertNotifyTenantTestUser inserts one isolated user for notification tests.
func insertNotifyTenantTestUser(t *testing.T, ctx context.Context, label string, tenantID int, status int) int {
	t.Helper()

	username := uniqueNotifyTenantTestName(label)
	id, err := dao.SysUser.Ctx(ctx).Data(do.SysUser{
		Username: username,
		Password: "test-password-hash",
		Nickname: username,
		Status:   status,
		TenantId: tenantID,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert notify tenant test user: %v", err)
	}
	return int(id)
}

// insertNotifyTenantTestMembership inserts one temporary active or inactive
// tenant membership row.
func insertNotifyTenantTestMembership(t *testing.T, ctx context.Context, userID int, tenantID int, status int) {
	t.Helper()

	ensureNotifyTenantTestMembershipTable(t, ctx)
	if _, err := g.DB().Model(notifyTenantMembershipTable).Safe().Ctx(ctx).Data(notifyTenantTestMembershipData{
		UserID:   int64(userID),
		TenantID: int64(tenantID),
		Status:   status,
		JoinedAt: gtime.Now(),
	}).Insert(); err != nil {
		t.Fatalf("insert notify tenant membership: %v", err)
	}
}

// ensureNotifyTenantTestMembershipTable creates the minimal plugin membership
// table needed by host notification tests when the multi-tenant plugin has not
// been installed in the local test database.
func ensureNotifyTenantTestMembershipTable(t *testing.T, ctx context.Context) {
	t.Helper()

	_, err := g.DB().Exec(ctx, `
CREATE TABLE IF NOT EXISTS plugin_multi_tenant_user_membership (
	    "id" BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	    "user_id" BIGINT NOT NULL,
	    "tenant_id" BIGINT NOT NULL,
	    "status" SMALLINT NOT NULL DEFAULT 1,
	    "joined_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	    "created_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	    "updated_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "deleted_at" TIMESTAMP
)`)
	if err != nil {
		t.Fatalf("ensure notify tenant membership table: %v", err)
	}
}

// queryNotifyTenantTestMessage loads one notification message by ID.
func queryNotifyTenantTestMessage(t *testing.T, ctx context.Context, messageID int64) *entity.SysNotifyMessage {
	t.Helper()

	var message *entity.SysNotifyMessage
	if err := dao.SysNotifyMessage.Ctx(ctx).Where(do.SysNotifyMessage{Id: messageID}).Scan(&message); err != nil {
		t.Fatalf("query notify message: %v", err)
	}
	if message == nil {
		t.Fatalf("expected notify message %d", messageID)
	}
	return message
}

// queryNotifyTenantTestDeliveries loads all deliveries for one message.
func queryNotifyTenantTestDeliveries(t *testing.T, ctx context.Context, messageID int64) []*notifyTenantTestDeliveryRow {
	t.Helper()

	var deliveries []*notifyTenantTestDeliveryRow
	if err := dao.SysNotifyDelivery.Ctx(ctx).
		Fields(
			dao.SysNotifyDelivery.Columns().Id,
			dao.SysNotifyDelivery.Columns().UserId,
			dao.SysNotifyDelivery.Columns().TenantId,
		).
		Where(do.SysNotifyDelivery{MessageId: messageID}).
		Scan(&deliveries); err != nil {
		t.Fatalf("query notify deliveries: %v", err)
	}
	return deliveries
}

// queryNotifyTenantTestUserTenantID returns one user's primary tenant ID.
func queryNotifyTenantTestUserTenantID(t *testing.T, ctx context.Context, userID int64) int {
	t.Helper()

	var row *notifyTenantTestUserTenantRow
	if err := dao.SysUser.Ctx(ctx).Fields(dao.SysUser.Columns().TenantId).Where(do.SysUser{Id: userID}).Scan(&row); err != nil {
		t.Fatalf("query notify user tenant: %v", err)
	}
	if row == nil {
		t.Fatalf("expected notify user %d", userID)
	}
	return row.TenantID
}

// notifyTenantTestDeliveryContains reports whether a delivery exists for one
// user.
func notifyTenantTestDeliveryContains(rows []*notifyTenantTestDeliveryRow, userID int64) bool {
	for _, row := range rows {
		if row != nil && row.UserID == userID {
			return true
		}
	}
	return false
}

// cleanupNotifyTenantTestMessage removes one temporary message and deliveries.
func cleanupNotifyTenantTestMessage(t *testing.T, ctx context.Context, messageID int64) {
	t.Helper()
	if messageID <= 0 {
		return
	}
	if _, err := dao.SysNotifyDelivery.Ctx(ctx).
		Unscoped().
		Where(do.SysNotifyDelivery{MessageId: messageID}).
		Delete(); err != nil {
		t.Fatalf("cleanup notify deliveries: %v", err)
	}
	if _, err := dao.SysNotifyMessage.Ctx(ctx).
		Where(do.SysNotifyMessage{Id: messageID}).
		Delete(); err != nil {
		t.Fatalf("cleanup notify message: %v", err)
	}
}

// cleanupNotifyTenantTestMemberships removes temporary membership rows for
// test users.
func cleanupNotifyTenantTestMemberships(t *testing.T, ctx context.Context, userIDs []int) {
	t.Helper()
	for _, userID := range userIDs {
		if _, err := g.DB().Model(notifyTenantMembershipTable).Safe().Ctx(ctx).
			Unscoped().
			Where(notifyTenantTestMembershipCleanupFilter{UserID: int64(userID)}).
			Delete(); err != nil {
			t.Fatalf("cleanup notify tenant membership for user %d: %v", userID, err)
		}
	}
}

// cleanupNotifyTenantTestUsers removes temporary users.
func cleanupNotifyTenantTestUsers(t *testing.T, ctx context.Context, userIDs []int) {
	t.Helper()
	if len(userIDs) == 0 {
		return
	}
	if _, err := dao.SysUser.Ctx(ctx).
		Unscoped().
		WhereIn(dao.SysUser.Columns().Id, userIDs).
		Delete(); err != nil {
		t.Fatalf("cleanup notify tenant users: %v", err)
	}
}

// uniqueNotifyTenantTestID returns a low-collision tenant ID for this package's
// database tests.
func uniqueNotifyTenantTestID() int {
	return int(time.Now().UnixNano()%1_000_000) + 700_000
}

// uniqueNotifyTenantTestName returns one collision-resistant label.
func uniqueNotifyTenantTestName(label string) string {
	return fmt.Sprintf("%s-%d", label, time.Now().UnixNano())
}
