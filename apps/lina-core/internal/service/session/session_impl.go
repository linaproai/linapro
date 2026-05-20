// session_impl.go implements persistent online-session storage, paging, scoped
// listing, validation, and cleanup. It writes both canonical session rows and
// projection fields, and applies tenant/data-scope filters before exposing
// online-session records to callers.

package session

import (
	"context"
	"time"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/datascope"
	tenantcapsvc "lina-core/internal/service/tenantcap"

	"github.com/gogf/gf/v2/database/gdb"
)

// Set persists a session record.
func (s *DBStore) Set(ctx context.Context, session *Session) error {
	_, err := dao.SysOnlineSession.Ctx(ctx).
		Data(do.SysOnlineSession{
			TokenId:        session.TokenId,
			TenantId:       session.TenantId,
			UserId:         session.UserId,
			Username:       session.Username,
			DeptName:       session.DeptName,
			Ip:             session.Ip,
			Browser:        session.Browser,
			Os:             session.Os,
			LoginTime:      session.LoginTime,
			LastActiveTime: normalizeSessionLastActive(session),
		}).
		OnConflict(dao.SysOnlineSession.Columns().TokenId).
		OnDuplicate(
			dao.SysOnlineSession.Columns().TenantId,
			dao.SysOnlineSession.Columns().UserId,
			dao.SysOnlineSession.Columns().Username,
			dao.SysOnlineSession.Columns().DeptName,
			dao.SysOnlineSession.Columns().Ip,
			dao.SysOnlineSession.Columns().Browser,
			dao.SysOnlineSession.Columns().Os,
			dao.SysOnlineSession.Columns().LoginTime,
			dao.SysOnlineSession.Columns().LastActiveTime,
		).
		Save()
	return err
}

// normalizeSessionLastActive returns the caller-provided activity time or the
// current time for newly created online-session projections.
func normalizeSessionLastActive(session *Session) *time.Time {
	if session != nil && session.LastActiveTime != nil {
		return session.LastActiveTime
	}
	now := time.Now()
	return &now
}

// setProjection persists or refreshes a session projection in PostgreSQL.
func (s *DBStore) setProjection(ctx context.Context, session *Session) error {
	_, err := dao.SysOnlineSession.Ctx(ctx).Data(do.SysOnlineSession{
		TokenId:        session.TokenId,
		TenantId:       session.TenantId,
		UserId:         session.UserId,
		Username:       session.Username,
		DeptName:       session.DeptName,
		Ip:             session.Ip,
		Browser:        session.Browser,
		Os:             session.Os,
		LoginTime:      session.LoginTime,
		LastActiveTime: normalizeSessionLastActive(session),
	}).Insert()
	return err
}

// tokenSessionModel builds the session lookup model for one globally unique token.
func tokenSessionModel(ctx context.Context, tokenId string) *gdb.Model {
	return dao.SysOnlineSession.Ctx(ctx).
		Where(do.SysOnlineSession{TokenId: tokenId})
}

// tenantSessionModel builds the session lookup model for a tenant/token pair.
func tenantSessionModel(ctx context.Context, tenantId int, tokenId string) *gdb.Model {
	return tokenSessionModel(ctx, tokenId).
		Where(do.SysOnlineSession{TenantId: tenantId})
}

// Get returns a session by globally unique token ID.
func (s *DBStore) Get(ctx context.Context, tokenId string) (*Session, error) {
	var e *entity.SysOnlineSession
	err := tokenSessionModel(ctx, tokenId).Scan(&e)
	if err != nil {
		return nil, err
	}
	if e == nil {
		return nil, nil
	}
	return &Session{
		TokenId:        e.TokenId,
		TenantId:       e.TenantId,
		UserId:         e.UserId,
		Username:       e.Username,
		DeptName:       e.DeptName,
		Ip:             e.Ip,
		Browser:        e.Browser,
		Os:             e.Os,
		LoginTime:      e.LoginTime,
		LastActiveTime: e.LastActiveTime,
	}, nil
}

// Delete removes a session by globally unique token ID.
func (s *DBStore) Delete(ctx context.Context, tokenId string) error {
	_, err := tokenSessionModel(ctx, tokenId).Delete()
	return err
}

// DeleteByUserId removes all sessions belonging to a user in one tenant.
func (s *DBStore) DeleteByUserId(ctx context.Context, tenantId int, userId int) error {
	_, err := tenantUserSessionModel(ctx, tenantId, userId).Delete()
	return err
}

// tenantUserSessionModel builds the session lookup model for one tenant/user.
func tenantUserSessionModel(ctx context.Context, tenantId int, userId int) *gdb.Model {
	return dao.SysOnlineSession.Ctx(ctx).
		Where(do.SysOnlineSession{TenantId: tenantId, UserId: userId})
}

// List returns all sessions matching the filter.
func (s *DBStore) List(ctx context.Context, filter *ListFilter) ([]*Session, error) {
	m := dao.SysOnlineSession.Ctx(ctx)
	if filter != nil {
		cols := dao.SysOnlineSession.Columns()
		if filter.Username != "" {
			m = m.WhereLike(cols.Username, "%"+filter.Username+"%")
		}
		if filter.Ip != "" {
			m = m.WhereLike(cols.Ip, "%"+filter.Ip+"%")
		}
	}
	var entities []*entity.SysOnlineSession
	err := m.OrderDesc(dao.SysOnlineSession.Columns().LoginTime).Scan(&entities)
	if err != nil {
		return nil, err
	}
	sessions := make([]*Session, len(entities))
	for i, e := range entities {
		sessions[i] = &Session{
			TokenId:        e.TokenId,
			TenantId:       e.TenantId,
			UserId:         e.UserId,
			Username:       e.Username,
			DeptName:       e.DeptName,
			Ip:             e.Ip,
			Browser:        e.Browser,
			Os:             e.Os,
			LoginTime:      e.LoginTime,
			LastActiveTime: e.LastActiveTime,
		}
	}
	return sessions, nil
}

// ListPage returns a paginated session list.
func (s *DBStore) ListPage(ctx context.Context, filter *ListFilter, pageNum, pageSize int) (*ListResult, error) {
	m := dao.SysOnlineSession.Ctx(ctx)
	if filter != nil {
		cols := dao.SysOnlineSession.Columns()
		if filter.Username != "" {
			m = m.WhereLike(cols.Username, "%"+filter.Username+"%")
		}
		if filter.Ip != "" {
			m = m.WhereLike(cols.Ip, "%"+filter.Ip+"%")
		}
	}

	// Get total count
	total, err := m.Count()
	if err != nil {
		return nil, err
	}

	// Get paginated items
	var entities []*entity.SysOnlineSession
	err = m.OrderDesc(dao.SysOnlineSession.Columns().LoginTime).
		Page(pageNum, pageSize).
		Scan(&entities)
	if err != nil {
		return nil, err
	}

	sessions := make([]*Session, len(entities))
	for i, e := range entities {
		sessions[i] = &Session{
			TokenId:        e.TokenId,
			TenantId:       e.TenantId,
			UserId:         e.UserId,
			Username:       e.Username,
			DeptName:       e.DeptName,
			Ip:             e.Ip,
			Browser:        e.Browser,
			Os:             e.Os,
			LoginTime:      e.LoginTime,
			LastActiveTime: e.LastActiveTime,
		}
	}

	return &ListResult{
		Items: sessions,
		Total: total,
	}, nil
}

// ListPageScoped returns a paginated session list constrained by tenant and
// user data scope.
func (s *DBStore) ListPageScoped(
	ctx context.Context,
	filter *ListFilter,
	pageNum, pageSize int,
	scopeSvc datascope.Service,
	tenantSvc tenantcapsvc.Service,
) (*ListResult, error) {
	m := dao.SysOnlineSession.Ctx(ctx)
	if filter != nil {
		cols := dao.SysOnlineSession.Columns()
		if filter.Username != "" {
			m = m.WhereLike(cols.Username, "%"+filter.Username+"%")
		}
		if filter.Ip != "" {
			m = m.WhereLike(cols.Ip, "%"+filter.Ip+"%")
		}
	}
	if tenantSvc != nil {
		var err error
		m, err = tenantSvc.Apply(ctx, m, qualifiedOnlineSessionTenantIDColumn())
		if err != nil {
			return nil, err
		}
	}
	if scopeSvc != nil {
		var err error
		var empty bool
		m, empty, err = scopeSvc.ApplyUserScope(ctx, m, qualifiedOnlineSessionUserIDColumn())
		if err != nil {
			return nil, err
		}
		if empty {
			return &ListResult{Items: []*Session{}, Total: 0}, nil
		}
	}

	total, err := m.Count()
	if err != nil {
		return nil, err
	}

	var entities []*entity.SysOnlineSession
	err = m.OrderDesc(dao.SysOnlineSession.Columns().LoginTime).
		Page(pageNum, pageSize).
		Scan(&entities)
	if err != nil {
		return nil, err
	}

	sessions := make([]*Session, len(entities))
	for i, e := range entities {
		sessions[i] = &Session{
			TokenId:        e.TokenId,
			TenantId:       e.TenantId,
			UserId:         e.UserId,
			Username:       e.Username,
			DeptName:       e.DeptName,
			Ip:             e.Ip,
			Browser:        e.Browser,
			Os:             e.Os,
			LoginTime:      e.LoginTime,
			LastActiveTime: e.LastActiveTime,
		}
	}

	return &ListResult{
		Items: sessions,
		Total: total,
	}, nil
}

// qualifiedOnlineSessionUserIDColumn returns the fully qualified session owner column.
func qualifiedOnlineSessionUserIDColumn() string {
	return dao.SysOnlineSession.Table() + "." + dao.SysOnlineSession.Columns().UserId
}

// qualifiedOnlineSessionTenantIDColumn returns the fully qualified session tenant column.
func qualifiedOnlineSessionTenantIDColumn() string {
	return dao.SysOnlineSession.Table() + "." + dao.SysOnlineSession.Columns().TenantId
}

// Count returns the total number of active sessions.
func (s *DBStore) Count(ctx context.Context) (int, error) {
	return dao.SysOnlineSession.Ctx(ctx).Count()
}

// TouchOrValidate validates tenant ownership and the session timeout, then
// refreshes last_active_time only when the previous activity is outside the
// short write-throttle window.
func (s *DBStore) TouchOrValidate(ctx context.Context, tenantId int, tokenId string, timeout time.Duration) (bool, error) {
	var stored *entity.SysOnlineSession
	err := tenantSessionModel(ctx, tenantId, tokenId).Scan(&stored)
	if err != nil {
		return false, err
	}
	if stored == nil {
		return false, nil
	}

	now := time.Now()
	if isSessionInactive(stored, now, timeout) {
		if _, err = tenantSessionModel(ctx, tenantId, tokenId).Delete(); err != nil {
			return false, err
		}
		return false, nil
	}

	updateCutoff := now.Add(-sessionLastActiveUpdateWindow)
	if stored.LastActiveTime != nil && sessionWallClockBefore(*stored.LastActiveTime, updateCutoff) {
		_, err = tenantSessionModel(ctx, tenantId, tokenId).
			WhereLT(dao.SysOnlineSession.Columns().LastActiveTime, updateCutoff).
			Data(do.SysOnlineSession{LastActiveTime: &now}).
			Update()
		if err != nil {
			return false, err
		}
	}
	return true, nil
}

// CleanupInactive removes sessions inactive longer than the configured threshold.
func (s *DBStore) CleanupInactive(ctx context.Context, timeout time.Duration) (int64, error) {
	cutoff := time.Now().Add(-timeout)
	result, err := dao.SysOnlineSession.Ctx(ctx).
		WhereLT(dao.SysOnlineSession.Columns().LastActiveTime, cutoff).
		Delete()
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// isSessionInactive reports whether one stored session is already expired by
// the configured inactivity timeout before the caller uses it as valid state.
func isSessionInactive(stored *entity.SysOnlineSession, now time.Time, timeout time.Duration) bool {
	if stored == nil || timeout <= 0 || stored.LastActiveTime == nil {
		return false
	}
	return sessionWallClockBeforeOrEqual(*stored.LastActiveTime, now.Add(-timeout))
}

// sessionWallClockBefore reports whether left is before right using database
// timestamp-without-time-zone wall-clock fields instead of absolute instants.
func sessionWallClockBefore(left time.Time, right time.Time) bool {
	return sessionWallClockTime(left).Before(sessionWallClockTime(right))
}

// sessionWallClockBeforeOrEqual reports whether left is before or equal to
// right using the same wall-clock semantics as PostgreSQL timestamp columns.
func sessionWallClockBeforeOrEqual(left time.Time, right time.Time) bool {
	return !sessionWallClockTime(left).After(sessionWallClockTime(right))
}

// sessionWallClockTime normalizes a timestamp into UTC while preserving its
// displayed calendar fields, matching database timestamp comparison semantics.
func sessionWallClockTime(value time.Time) time.Time {
	return time.Date(
		value.Year(),
		value.Month(),
		value.Day(),
		value.Hour(),
		value.Minute(),
		value.Second(),
		value.Nanosecond(),
		time.UTC,
	)
}
