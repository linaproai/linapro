package locker

import (
	"context"
	"database/sql"
	"time"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gtime"
)

// Instance represents a distributed lock instance.
type Instance struct {
	id     int64         // Lock record ID
	holder string        // Node identifier that holds this lock
	lease  time.Duration // Lease duration used when this lock was acquired
}

// ID returns the persistent lock record ID.
func (i *Instance) ID() int64 {
	return i.id
}

// Holder returns the current lock holder token.
func (i *Instance) Holder() string {
	return i.holder
}

// Unlock releases the lock by setting its expire_time to the past.
// This effectively releases the lock for other nodes to acquire.
func (i *Instance) Unlock(ctx context.Context) error {
	_, err := dao.SysLocker.Ctx(ctx).Data(do.SysLocker{
		ExpireTime: gtime.Now().Add(-1 * time.Second),
	}).Where(do.SysLocker{
		Id:     i.id,
		Holder: i.holder,
	}).Update()
	return err
}

// Renew extends the lock's expiration time.
// It only succeeds if the lock is still held by the current node and hasn't expired.
// Returns ErrLockNotHeld if the lock was lost or expired.
func (i *Instance) Renew(ctx context.Context) error {
	now := gtime.Now()
	expireTime := now.Add(i.lease)

	// First check if lock is still valid
	var locker struct {
		Id int64
	}
	err := dao.SysLocker.Ctx(ctx).
		Where(do.SysLocker{
			Id:     i.id,
			Holder: i.holder,
		}).
		WhereGT("expire_time", now).
		Scan(&locker)
	if err != nil {
		if gerror.Is(err, sql.ErrNoRows) {
			return ErrLockNotHeld
		}
		return err
	}
	if locker.Id == 0 {
		return ErrLockNotHeld
	}

	// Lock is valid, extend it
	_, err = dao.SysLocker.Ctx(ctx).Data(do.SysLocker{
		ExpireTime: expireTime,
	}).Where(do.SysLocker{
		Id: locker.Id,
	}).Update()
	return err
}

// IsHeld checks if the lock is still held by the current node.
// A lock is considered held if its expire_time is in the future.
func (i *Instance) IsHeld(ctx context.Context) (bool, error) {
	count, err := dao.SysLocker.Ctx(ctx).
		Where(do.SysLocker{Id: i.id}).
		WhereGT("expire_time", gtime.Now()).
		Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
