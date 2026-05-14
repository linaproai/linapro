// This file implements operations on one acquired distributed lock instance.

package locker

import (
	"context"
	"database/sql"
	"time"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/service/coordination"
	"lina-core/pkg/bizerr"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gtime"
)

// Instance represents a distributed lock instance.
type Instance struct {
	id     int64                    // Lock record ID
	name   string                   // Name is the logical lock name.
	holder string                   // Node identifier that holds this lock
	lease  time.Duration            // Lease duration used when this lock was acquired
	handle *coordination.LockHandle // Handle stores coordination owner token metadata.
}

// ID returns the persistent lock record ID.
func (i *Instance) ID() int64 {
	return i.id
}

// Holder returns the current lock holder token.
func (i *Instance) Holder() string {
	if i != nil && i.handle != nil {
		return i.handle.Token
	}
	return i.holder
}

// Name returns the logical lock name when it is known by the backend.
func (i *Instance) Name() string {
	return i.name
}

// Unlock releases the lock by setting its expire_time to the past.
// This effectively releases the lock for other nodes to acquire.
func (i *Instance) Unlock(ctx context.Context) error {
	if i != nil && i.handle != nil {
		if lockStore := currentCoordinationLockStore(); lockStore != nil {
			return mapCoordinationLockError(lockStore.Release(ctx, i.handle))
		}
	}
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
	if i != nil && i.handle != nil {
		if lockStore := currentCoordinationLockStore(); lockStore != nil {
			return mapCoordinationLockError(lockStore.Renew(ctx, i.handle, i.lease))
		}
	}

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
	if i != nil && i.handle != nil {
		lockStore := currentCoordinationLockStore()
		if lockStore == nil {
			return false, nil
		}
		held, err := lockStore.IsHeld(ctx, i.handle)
		return held, mapCoordinationLockError(err)
	}
	count, err := dao.SysLocker.Ctx(ctx).
		Where(do.SysLocker{Id: i.id}).
		WhereGT("expire_time", gtime.Now()).
		Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// lockWithCoordination acquires one distributed lock through coordination.
func lockWithCoordination(
	ctx context.Context,
	lockStore coordination.LockStore,
	name string,
	holder string,
	reason string,
	lease time.Duration,
) (*Instance, bool, error) {
	handle, ok, err := lockStore.Acquire(ctx, name, holder, reason, lease)
	if err != nil {
		return nil, false, mapCoordinationLockError(err)
	}
	if !ok || handle == nil {
		return nil, false, nil
	}
	return &Instance{
		id:     handle.FencingToken,
		name:   name,
		holder: handle.Owner,
		lease:  lease,
		handle: handle,
	}, true, nil
}

// renewCoordinationByName renews one coordination lock using the holder token.
func renewCoordinationByName(
	ctx context.Context,
	lockStore coordination.LockStore,
	name string,
	holder string,
	lease time.Duration,
) error {
	handle := &coordination.LockHandle{Name: name, Owner: holder, Token: holder, Lease: lease}
	return mapCoordinationLockError(lockStore.Renew(ctx, handle, lease))
}

// unlockCoordinationByName releases one coordination lock using the holder token.
func unlockCoordinationByName(ctx context.Context, lockStore coordination.LockStore, name string, holder string) error {
	handle := &coordination.LockHandle{Name: name, Owner: holder, Token: holder}
	return mapCoordinationLockError(lockStore.Release(ctx, handle))
}

// mapCoordinationLockError maps coordination ownership errors to locker errors.
func mapCoordinationLockError(err error) error {
	if err == nil {
		return nil
	}
	if bizerr.Is(err, coordination.CodeCoordinationLockNotHeld) {
		return ErrLockNotHeld
	}
	return err
}
