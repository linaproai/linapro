// This file contains the locker service methods that select between the SQL
// lock table and the configured coordination lock backend.

package locker

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/internal/service/coordination"
	"lina-core/pkg/dialect"
	"lina-core/pkg/logger"
)

// Lock acquires a distributed lock when it is absent or expired.
func (s *serviceImpl) Lock(ctx context.Context, name, holder, reason string, lease time.Duration) (*Instance, bool, error) {
	if lockStore := currentCoordinationLockStore(); lockStore != nil {
		return lockWithCoordination(ctx, lockStore, name, holder, reason, lease)
	}

	var locker *entity.SysLocker
	err := dao.SysLocker.Ctx(ctx).Where(do.SysLocker{
		Name: name,
	}).Scan(&locker)
	if err != nil {
		return nil, false, err
	}

	now := gtime.Now()
	expireTime := now.Add(lease)

	// Lock doesn't exist, try to create it
	if locker == nil {
		result, err := dao.SysLocker.Ctx(ctx).Data(do.SysLocker{
			Name:       name,
			Reason:     reason,
			Holder:     holder,
			ExpireTime: expireTime,
		}).Insert()
		if err != nil {
			if dialect.IsUniqueConstraintViolation(err) {
				return nil, false, nil
			}
			return nil, false, err
		}
		insertId, err := result.LastInsertId()
		if err != nil {
			return nil, false, err
		}
		if insertId <= 0 {
			return nil, false, nil
		}
		logger.Infof(ctx, "[locker] acquired lock '%s' (holder: %s)", name, holder)
		return &Instance{id: insertId, holder: holder, lease: lease}, true, nil
	}

	// Lock exists and is held by current node, extend it without requiring
	// client-side timestamp interpretation.
	if locker.Holder == holder {
		_, err := dao.SysLocker.Ctx(ctx).Data(do.SysLocker{
			ExpireTime: expireTime,
		}).Where(do.SysLocker{
			Id: locker.Id,
		}).Update()
		if err != nil {
			return nil, false, err
		}
		return &Instance{id: int64(locker.Id), holder: holder, lease: lease}, true, nil
	}

	// Lock exists but may be expired, try to take over atomically using the
	// database predicate so TIMESTAMP location handling cannot affect expiry.
	{
		cols := dao.SysLocker.Columns()
		affected, err := dao.SysLocker.Ctx(ctx).Data(do.SysLocker{
			Reason:     reason,
			Holder:     holder,
			ExpireTime: expireTime,
		}).Where(do.SysLocker{
			Id: locker.Id,
		}).Wheref("(%s IS NULL OR %s < ?)", cols.ExpireTime, cols.ExpireTime, now).UpdateAndGetAffected()
		if err != nil {
			return nil, false, err
		}
		if affected <= 0 {
			return nil, false, nil
		}
		logger.Infof(ctx, "[locker] acquired expired lock '%s' (holder: %s)", name, holder)
		return &Instance{id: int64(locker.Id), holder: holder, lease: lease}, true, nil
	}
}

// isExpiredLock reports whether one lock row is available for takeover before
// any holder data is reused.
func isExpiredLock(expireTime *gtime.Time, now *gtime.Time) bool {
	if expireTime == nil {
		return true
	}
	return now.After(expireTime)
}

// LockFunc acquires a lock and executes the given function.
// The lock is automatically released after the function completes.
func (s *serviceImpl) LockFunc(ctx context.Context, name, holder, reason string, lease time.Duration, f func() error) (bool, error) {
	instance, ok, err := s.Lock(ctx, name, holder, reason, lease)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}
	defer func() {
		if err := instance.Unlock(ctx); err != nil {
			logger.Warningf(ctx, "[locker] failed to unlock '%s': %v", name, err)
		}
	}()
	if err = f(); err != nil {
		return true, err
	}
	return true, nil
}

// Unlock releases one distributed lock identified by lock ID and holder.
func (s *serviceImpl) Unlock(ctx context.Context, lockID int64, holder string) error {
	if currentCoordinationLockStore() != nil {
		return ErrLockNotHeld
	}
	return (&Instance{
		id:     lockID,
		holder: holder,
	}).Unlock(ctx)
}

// Renew extends one distributed lock identified by lock ID and holder.
func (s *serviceImpl) Renew(ctx context.Context, lockID int64, holder string, lease time.Duration) error {
	if currentCoordinationLockStore() != nil {
		return ErrLockNotHeld
	}
	return (&Instance{
		id:     lockID,
		holder: holder,
		lease:  lease,
	}).Renew(ctx)
}

// UnlockByName releases one distributed lock identified by lock name and holder.
func (s *serviceImpl) UnlockByName(ctx context.Context, name string, holder string) error {
	if lockStore := currentCoordinationLockStore(); lockStore != nil {
		return unlockCoordinationByName(ctx, lockStore, name, holder)
	}
	_, err := dao.SysLocker.Ctx(ctx).Data(do.SysLocker{
		ExpireTime: gtime.Now().Add(-1 * time.Second),
	}).Where(do.SysLocker{
		Name:   name,
		Holder: holder,
	}).Update()
	return err
}

// RenewByName extends one distributed lock identified by lock name and holder.
func (s *serviceImpl) RenewByName(ctx context.Context, name string, holder string, lease time.Duration) error {
	if lockStore := currentCoordinationLockStore(); lockStore != nil {
		return renewCoordinationByName(ctx, lockStore, name, holder, lease)
	}

	now := gtime.Now()
	expireTime := now.Add(lease)
	var locker struct {
		Id int64
	}
	err := dao.SysLocker.Ctx(ctx).
		Where(do.SysLocker{
			Name:   name,
			Holder: holder,
		}).
		WhereGT("expire_time", now).
		Scan(&locker)
	if err != nil {
		return err
	}
	if locker.Id == 0 {
		return ErrLockNotHeld
	}

	_, err = dao.SysLocker.Ctx(ctx).Data(do.SysLocker{
		ExpireTime: expireTime,
	}).Where(do.SysLocker{
		Id: locker.Id,
	}).Update()
	return err
}

// ConfigureCoordination switches all locker service instances to a
// coordination-backed implementation. Passing nil restores the SQL table
// implementation used by single-node deployments and tests.
func ConfigureCoordination(coordinationSvc coordination.Service) {
	processCoordinationLockStore.Lock()
	if coordinationSvc == nil {
		processCoordinationLockStore.store = nil
	} else {
		processCoordinationLockStore.store = coordinationSvc.Lock()
	}
	processCoordinationLockStore.Unlock()
}

// currentCoordinationLockStore returns the active coordination lock backend.
func currentCoordinationLockStore() coordination.LockStore {
	processCoordinationLockStore.RLock()
	store := processCoordinationLockStore.store
	processCoordinationLockStore.RUnlock()
	return store
}
