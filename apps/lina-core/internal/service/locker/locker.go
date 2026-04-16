package locker

import (
	"context"
	"time"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/pkg/logger"

	"github.com/gogf/gf/v2/os/gtime"
)

// Service defines the locker service contract.
type Service interface {
	// Lock acquires a distributed lock when it is absent or expired.
	Lock(ctx context.Context, name, holder, reason string, lease time.Duration) (*Instance, bool, error)
	// LockFunc acquires a lock and executes the given function.
	// The lock is automatically released after the function completes.
	LockFunc(ctx context.Context, name, holder, reason string, lease time.Duration, f func() error) (bool, error)
	// Unlock releases one distributed lock identified by lock ID and holder.
	Unlock(ctx context.Context, lockID int64, holder string) error
	// Renew extends one distributed lock identified by lock ID and holder.
	Renew(ctx context.Context, lockID int64, holder string, lease time.Duration) error
}

var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct{}

func New() Service {
	return &serviceImpl{}
}

// Lock acquires a distributed lock when it is absent or expired.
func (s *serviceImpl) Lock(ctx context.Context, name, holder, reason string, lease time.Duration) (*Instance, bool, error) {
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

	// Lock exists but expired, try to take over
	if now.After(locker.ExpireTime) {
		affected, err := dao.SysLocker.Ctx(ctx).Data(do.SysLocker{
			Reason:     reason,
			Holder:     holder,
			ExpireTime: expireTime,
		}).Where(do.SysLocker{
			Id: locker.Id,
		}).UpdateAndGetAffected()
		if err != nil {
			return nil, false, err
		}
		if affected <= 0 {
			return nil, false, nil
		}
		logger.Infof(ctx, "[locker] acquired expired lock '%s' (holder: %s)", name, holder)
		return &Instance{id: int64(locker.Id), holder: holder, lease: lease}, true, nil
	}

	// Lock is held by current node (same holder), extend it
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

	// Lock is held by another node
	return nil, false, nil
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
	return (&Instance{
		id:     lockID,
		holder: holder,
	}).Unlock(ctx)
}

// Renew extends one distributed lock identified by lock ID and holder.
func (s *serviceImpl) Renew(ctx context.Context, lockID int64, holder string, lease time.Duration) error {
	return (&Instance{
		id:     lockID,
		holder: holder,
		lease:  lease,
	}).Renew(ctx)
}
