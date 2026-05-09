// Package session implements online-session storage and activity validation.
package session

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/datascope"
)

// sessionLastActiveUpdateWindow is the minimum interval between two
// last_active_time writes for one valid session.
const sessionLastActiveUpdateWindow time.Duration = time.Minute

// Session represents an online user session.
type Session struct {
	TokenId        string      // Unique token identifier
	UserId         int         // User ID
	Username       string      // Username
	DeptName       string      // Department name
	Ip             string      // Login IP address
	Browser        string      // Browser information
	Os             string      // Operating system
	LoginTime      *gtime.Time // Login time
	LastActiveTime *gtime.Time // Last active time
}

// ListFilter defines filter options for listing sessions.
type ListFilter struct {
	Username string // Username, supports fuzzy search
	Ip       string // Login IP, supports fuzzy search
}

// ListResult defines the result for paginated session list.
type ListResult struct {
	Items []*Session // Session items
	Total int        // Total count
}

// Store defines the session storage interface for persistent online-session
// records.
type Store interface {
	// Set persists one online session record.
	Set(ctx context.Context, session *Session) error
	// Get returns one online session by token ID.
	Get(ctx context.Context, tokenId string) (*Session, error)
	// Delete removes one online session by token ID.
	Delete(ctx context.Context, tokenId string) error
	// DeleteByUserId removes all online sessions that belong to one user.
	DeleteByUserId(ctx context.Context, userId int) error
	// List returns all online sessions that match the optional filter.
	List(ctx context.Context, filter *ListFilter) ([]*Session, error)
	// ListPage returns one paginated online-session list for the optional filter.
	ListPage(ctx context.Context, filter *ListFilter, pageNum, pageSize int) (*ListResult, error)
	// ListPageScoped returns one paginated online-session list constrained by the supplied data-scope service.
	ListPageScoped(ctx context.Context, filter *ListFilter, pageNum, pageSize int, scopeSvc datascope.Service) (*ListResult, error)
	// Count returns the total number of active online sessions.
	Count(ctx context.Context) (int, error)
	// TouchOrValidate validates the session timeout and refreshes last_active_time
	// outside the short write-throttle window for the given tokenId. It returns
	// true when the session remains valid.
	TouchOrValidate(ctx context.Context, tokenId string, timeout time.Duration) (bool, error)
	// CleanupInactive deletes sessions whose last_active_time exceeds the given timeout duration.
	CleanupInactive(ctx context.Context, timeout time.Duration) (int64, error)
}

// DBStore implements Store using the persistent online-session table.
type DBStore struct{}

// NewDBStore creates a new DBStore instance.
func NewDBStore() Store {
	return &DBStore{}
}

// Set persists a session record.
func (s *DBStore) Set(ctx context.Context, session *Session) error {
	_, err := dao.SysOnlineSession.Ctx(ctx).Data(do.SysOnlineSession{
		TokenId:        session.TokenId,
		UserId:         session.UserId,
		Username:       session.Username,
		DeptName:       session.DeptName,
		Ip:             session.Ip,
		Browser:        session.Browser,
		Os:             session.Os,
		LoginTime:      session.LoginTime,
		LastActiveTime: gtime.Now(),
	}).Insert()
	return err
}

// Get returns a session by token ID.
func (s *DBStore) Get(ctx context.Context, tokenId string) (*Session, error) {
	var e *entity.SysOnlineSession
	err := dao.SysOnlineSession.Ctx(ctx).
		Where(do.SysOnlineSession{TokenId: tokenId}).
		Scan(&e)
	if err != nil {
		return nil, err
	}
	if e == nil {
		return nil, nil
	}
	return &Session{
		TokenId:        e.TokenId,
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

// Delete removes a session by token ID.
func (s *DBStore) Delete(ctx context.Context, tokenId string) error {
	_, err := dao.SysOnlineSession.Ctx(ctx).
		Where(do.SysOnlineSession{TokenId: tokenId}).
		Delete()
	return err
}

// DeleteByUserId removes all sessions belonging to a user.
func (s *DBStore) DeleteByUserId(ctx context.Context, userId int) error {
	_, err := dao.SysOnlineSession.Ctx(ctx).
		Where(do.SysOnlineSession{UserId: userId}).
		Delete()
	return err
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

// ListPageScoped returns a paginated session list constrained by user data scope.
func (s *DBStore) ListPageScoped(ctx context.Context, filter *ListFilter, pageNum, pageSize int, scopeSvc datascope.Service) (*ListResult, error) {
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

// Count returns the total number of active sessions.
func (s *DBStore) Count(ctx context.Context) (int, error) {
	return dao.SysOnlineSession.Ctx(ctx).Count()
}

// TouchOrValidate validates the session timeout and refreshes last_active_time
// only when the previous activity is outside the short write-throttle window.
func (s *DBStore) TouchOrValidate(ctx context.Context, tokenId string, timeout time.Duration) (bool, error) {
	cols := dao.SysOnlineSession.Columns()
	if timeout > 0 {
		cutoff := gtime.Now().Add(-timeout)
		expiredCount, err := dao.SysOnlineSession.Ctx(ctx).
			Where(do.SysOnlineSession{TokenId: tokenId}).
			WhereLTE(cols.LastActiveTime, cutoff).
			Count()
		if err != nil {
			return false, err
		}
		if expiredCount > 0 {
			if _, err = dao.SysOnlineSession.Ctx(ctx).
				Where(do.SysOnlineSession{TokenId: tokenId}).
				Delete(); err != nil {
				return false, err
			}
			return false, nil
		}
	}

	now := gtime.Now()
	updateCutoff := now.Add(-sessionLastActiveUpdateWindow)
	_, err := dao.SysOnlineSession.Ctx(ctx).
		Where(do.SysOnlineSession{TokenId: tokenId}).
		WhereLT(cols.LastActiveTime, updateCutoff).
		Data(do.SysOnlineSession{LastActiveTime: now}).
		Update()
	if err != nil {
		return false, err
	}

	count, err := dao.SysOnlineSession.Ctx(ctx).
		Where(do.SysOnlineSession{TokenId: tokenId}).
		Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CleanupInactive removes sessions inactive longer than the configured threshold.
func (s *DBStore) CleanupInactive(ctx context.Context, timeout time.Duration) (int64, error) {
	cutoff := gtime.Now().Add(-timeout)
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
func isSessionInactive(stored *entity.SysOnlineSession, now *gtime.Time, timeout time.Duration) bool {
	if stored == nil || timeout <= 0 || stored.LastActiveTime == nil {
		return false
	}
	return !stored.LastActiveTime.After(now.Add(-timeout))
}
