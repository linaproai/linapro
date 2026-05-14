// This file verifies user messages remain self-isolated regardless of role data scope.

package usermsg

import (
	"context"
	"fmt"
	"testing"
	"time"

	_ "lina-core/pkg/dbdriver"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/gconv"

	"lina-core/internal/dao"
	"lina-core/internal/model"
	"lina-core/internal/model/do"
	"lina-core/internal/service/notify"
	"lina-core/internal/service/tenantcap"
	"lina-core/pkg/bizerr"
)

// TestUserMessagesRemainSelfIsolatedForAllDataScope verifies all-data role
// scope never broadens inbox message reads, marks, or deletes to other users.
func TestUserMessagesRemainSelfIsolatedForAllDataScope(t *testing.T) {
	ctx := context.Background()
	currentUserID := insertUserMsgScopeUser(t, ctx, "usermsg-current")
	otherUserID := insertUserMsgScopeUser(t, ctx, "usermsg-other")
	roleID := insertUserMsgScopeRole(t, ctx, "usermsg-all", 1)
	t.Cleanup(func() {
		cleanupUserMsgScopeUsers(t, ctx, []int{currentUserID, otherUserID})
		cleanupUserMsgScopeRoles(t, ctx, []int{roleID})
	})
	insertUserMsgScopeUserRole(t, ctx, currentUserID, roleID)

	currentDeliveryID := insertUserMsgScopeDelivery(t, ctx, currentUserID, "current-message")
	otherDeliveryID := insertUserMsgScopeDelivery(t, ctx, otherUserID, "other-message")
	t.Cleanup(func() { cleanupUserMsgScopeDeliveries(t, ctx, []int64{currentDeliveryID, otherDeliveryID}) })

	svc := New(nil, notify.New(tenantcap.New(nil, nil)), nil).(*serviceImpl)
	svc.bizCtxSvc = userMsgScopeStaticBizCtx{ctx: &model.Context{UserId: currentUserID}}

	out, err := svc.List(ctx, ListInput{PageNum: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("list user messages: %v", err)
	}
	if out.Total != 1 || len(out.List) != 1 || out.List[0].Id != currentDeliveryID {
		t.Fatalf("expected only current user's message, got total=%d list=%#v", out.Total, out.List)
	}

	if _, err = svc.Get(ctx, otherDeliveryID); !bizerr.Is(err, CodeUserMsgNotFound) {
		t.Fatalf("expected other message get to be hidden as not found, got %v", err)
	}
	if err = svc.MarkRead(ctx, otherDeliveryID); err != nil {
		t.Fatalf("mark other message should be a no-op update scoped to current user: %v", err)
	}
	if read := queryUserMsgDeliveryRead(t, ctx, otherDeliveryID); read != 0 {
		t.Fatalf("expected other message to remain unread, got %d", read)
	}
	if err = svc.Delete(ctx, otherDeliveryID); err != nil {
		t.Fatalf("delete other message should be a no-op delete scoped to current user: %v", err)
	}
	if count := countUserMsgDelivery(t, ctx, otherDeliveryID); count != 1 {
		t.Fatalf("expected other message to remain after scoped delete, count=%d", count)
	}
}

// userMsgScopeStaticBizCtx returns a fixed business context.
type userMsgScopeStaticBizCtx struct {
	ctx *model.Context
}

// Init is unused by user-message tests.
func (s userMsgScopeStaticBizCtx) Init(_ *ghttp.Request, _ *model.Context) {}

// Get returns the configured business context.
func (s userMsgScopeStaticBizCtx) Get(context.Context) *model.Context { return s.ctx }

// SetLocale is unused by user-message tests.
func (s userMsgScopeStaticBizCtx) SetLocale(context.Context, string) {}

// SetUser is unused by user-message tests.
func (s userMsgScopeStaticBizCtx) SetUser(context.Context, string, int, string, int) {}

// SetTenant is unused by user-message tests.
func (s userMsgScopeStaticBizCtx) SetTenant(context.Context, int) {}

// SetImpersonation is unused by user-message tests.
func (s userMsgScopeStaticBizCtx) SetImpersonation(context.Context, int, int, bool, bool) {}


// SetUserAccess is unused by user-message tests.
func (s userMsgScopeStaticBizCtx) SetUserAccess(context.Context, int, bool, int) {}

// insertUserMsgScopeUser inserts one temporary user.
func insertUserMsgScopeUser(t *testing.T, ctx context.Context, prefix string) int {
	t.Helper()
	id, err := dao.SysUser.Ctx(ctx).Data(do.SysUser{
		Username: uniqueUserMsgScopeName(prefix),
		Password: "hashed",
		Nickname: prefix,
		Status:   1,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert usermsg-scope user: %v", err)
	}
	return int(id)
}

// uniqueUserMsgScopeName returns one collision-resistant identifier.
func uniqueUserMsgScopeName(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

// insertUserMsgScopeRole inserts one temporary role.
func insertUserMsgScopeRole(t *testing.T, ctx context.Context, prefix string, scope int) int {
	t.Helper()
	id, err := dao.SysRole.Ctx(ctx).Data(do.SysRole{
		Name:      uniqueUserMsgScopeName(prefix),
		Key:       uniqueUserMsgScopeName(prefix + "-key"),
		Sort:      99,
		DataScope: scope,
		Status:    1,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert usermsg-scope role: %v", err)
	}
	return int(id)
}

// insertUserMsgScopeUserRole binds one user to one role.
func insertUserMsgScopeUserRole(t *testing.T, ctx context.Context, userID int, roleID int) {
	t.Helper()
	if _, err := dao.SysUserRole.Ctx(ctx).Data(do.SysUserRole{UserId: userID, RoleId: roleID}).Insert(); err != nil {
		t.Fatalf("insert usermsg-scope user role: %v", err)
	}
}

// insertUserMsgScopeDelivery inserts one message and one inbox delivery.
func insertUserMsgScopeDelivery(t *testing.T, ctx context.Context, userID int, title string) int64 {
	t.Helper()
	messageID, err := dao.SysNotifyMessage.Ctx(ctx).Data(do.SysNotifyMessage{
		SourceType:   notify.SourceTypeSystem.String(),
		SourceId:     uniqueUserMsgScopeName("source"),
		CategoryCode: notify.CategoryCodeSystem.String(),
		Title:        title,
		Content:      title + " content",
		PayloadJson:  "{}",
		SenderUserId: userID,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert usermsg-scope message: %v", err)
	}
	deliveryID, err := dao.SysNotifyDelivery.Ctx(ctx).Data(do.SysNotifyDelivery{
		MessageId:      messageID,
		ChannelKey:     notify.ChannelKeyInbox,
		ChannelType:    notify.ChannelTypeInbox.String(),
		RecipientType:  notify.RecipientTypeUser.String(),
		RecipientKey:   gconv.String(userID),
		UserId:         userID,
		DeliveryStatus: notify.DeliveryStatusSucceeded,
		IsRead:         0,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert usermsg-scope delivery: %v", err)
	}
	return deliveryID
}

// cleanupUserMsgScopeUsers removes temporary users.
func cleanupUserMsgScopeUsers(t *testing.T, ctx context.Context, ids []int) {
	t.Helper()
	if len(ids) == 0 {
		return
	}
	if _, err := dao.SysUserRole.Ctx(ctx).WhereIn(dao.SysUserRole.Columns().UserId, ids).Delete(); err != nil {
		t.Fatalf("cleanup usermsg user roles: %v", err)
	}
	if _, err := dao.SysUser.Ctx(ctx).Unscoped().WhereIn(dao.SysUser.Columns().Id, ids).Delete(); err != nil {
		t.Fatalf("cleanup usermsg users: %v", err)
	}
}

// cleanupUserMsgScopeRoles removes temporary roles.
func cleanupUserMsgScopeRoles(t *testing.T, ctx context.Context, ids []int) {
	t.Helper()
	if len(ids) == 0 {
		return
	}
	if _, err := dao.SysRole.Ctx(ctx).Unscoped().WhereIn(dao.SysRole.Columns().Id, ids).Delete(); err != nil {
		t.Fatalf("cleanup usermsg roles: %v", err)
	}
}

// cleanupUserMsgScopeDeliveries removes temporary deliveries and messages.
func cleanupUserMsgScopeDeliveries(t *testing.T, ctx context.Context, ids []int64) {
	t.Helper()
	if len(ids) == 0 {
		return
	}
	var rows []struct {
		MessageId int64 `json:"messageId"`
	}
	if err := dao.SysNotifyDelivery.Ctx(ctx).Unscoped().Fields(dao.SysNotifyDelivery.Columns().MessageId).WhereIn(dao.SysNotifyDelivery.Columns().Id, ids).Scan(&rows); err != nil {
		t.Fatalf("query usermsg deliveries for cleanup: %v", err)
	}
	if _, err := dao.SysNotifyDelivery.Ctx(ctx).Unscoped().WhereIn(dao.SysNotifyDelivery.Columns().Id, ids).Delete(); err != nil {
		t.Fatalf("cleanup usermsg deliveries: %v", err)
	}
	messageIDs := make([]int64, 0, len(rows))
	for _, row := range rows {
		messageIDs = append(messageIDs, row.MessageId)
	}
	if len(messageIDs) > 0 {
		if _, err := dao.SysNotifyMessage.Ctx(ctx).WhereIn(dao.SysNotifyMessage.Columns().Id, messageIDs).Delete(); err != nil {
			t.Fatalf("cleanup usermsg messages: %v", err)
		}
	}
}

// queryUserMsgDeliveryRead returns one delivery read flag.
func queryUserMsgDeliveryRead(t *testing.T, ctx context.Context, id int64) int {
	t.Helper()
	var row *struct {
		IsRead int `json:"isRead"`
	}
	if err := dao.SysNotifyDelivery.Ctx(ctx).Unscoped().Fields(dao.SysNotifyDelivery.Columns().IsRead).Where(do.SysNotifyDelivery{Id: id}).Scan(&row); err != nil {
		t.Fatalf("query usermsg read flag: %v", err)
	}
	if row == nil {
		t.Fatalf("expected usermsg delivery %d", id)
	}
	return row.IsRead
}

// countUserMsgDelivery counts visible delivery rows by ID.
func countUserMsgDelivery(t *testing.T, ctx context.Context, id int64) int {
	t.Helper()
	count, err := dao.SysNotifyDelivery.Ctx(ctx).Where(do.SysNotifyDelivery{Id: id}).Count()
	if err != nil {
		t.Fatalf("count usermsg delivery: %v", err)
	}
	return count
}
