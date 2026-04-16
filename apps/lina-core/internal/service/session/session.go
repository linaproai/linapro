// This file implements the online-session storage abstraction backed by MySQL.

package session

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
)

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

// Store defines the session storage interface.
// Current implementation uses MySQL MEMORY engine.
// Future implementations may use gcache + Redis.
type Store interface {
	Set(ctx context.Context, session *Session) error
	Get(ctx context.Context, tokenId string) (*Session, error)
	Delete(ctx context.Context, tokenId string) error
	DeleteByUserId(ctx context.Context, userId int) error
	List(ctx context.Context, filter *ListFilter) ([]*Session, error)
	ListPage(ctx context.Context, filter *ListFilter, pageNum, pageSize int) (*ListResult, error)
	Count(ctx context.Context) (int, error)
	// TouchOrValidate updates last_active_time for the given tokenId.
	// Returns true if the session exists (affected rows > 0), false otherwise.
	TouchOrValidate(ctx context.Context, tokenId string) (bool, error)
	// CleanupInactive deletes sessions whose last_active_time exceeds the given timeout duration.
	CleanupInactive(ctx context.Context, timeout time.Duration) (int64, error)
}

// DBStore implements Store using MySQL MEMORY engine table.
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

// Count returns the total number of active sessions.
func (s *DBStore) Count(ctx context.Context) (int, error) {
	return dao.SysOnlineSession.Ctx(ctx).Count()
}

// TouchOrValidate validates a session and refreshes its last active time.
func (s *DBStore) TouchOrValidate(ctx context.Context, tokenId string) (bool, error) {
	result, err := dao.SysOnlineSession.Ctx(ctx).
		Where(do.SysOnlineSession{TokenId: tokenId}).
		Data(do.SysOnlineSession{LastActiveTime: gtime.Now()}).
		Update()
	if err != nil {
		return false, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	if affected > 0 {
		return true, nil
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
