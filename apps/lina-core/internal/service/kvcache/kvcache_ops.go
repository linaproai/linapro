// This file implements distributed KV cache CRUD, increment, expire, and cleanup behaviors.

package kvcache

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
)

// Get returns the current cache entry snapshot identified by ownerType, ownerKey,
// namespace, and cacheKey.
//
// Parameters:
//   - ctx: request-scoped context used for database access, tracing, and cancellation.
//   - ownerType: cache owner category, used to isolate entries across different business scopes.
//   - ownerKey: concrete owner identifier within ownerType, such as a module key or plugin key.
//   - namespace: logical group name used to organize related cache entries for the same owner.
//   - cacheKey: concrete key to read inside the namespace.
//
// Returns:
//   - *Item: the cache entry snapshot when the entry exists, including value kind, value, and expiration time.
//   - bool: whether the cache entry exists after expired data has been cleaned up.
//   - error: returned when identity parameters are invalid, expired-entry cleanup fails, or the database query fails.
func (s *serviceImpl) Get(
	ctx context.Context,
	ownerType OwnerType,
	ownerKey string,
	namespace string,
	cacheKey string,
) (*Item, bool, error) {
	if err := s.validateIdentity(ownerType, ownerKey, namespace, cacheKey); err != nil {
		return nil, false, err
	}
	if err := s.CleanupExpired(ctx); err != nil {
		return nil, false, err
	}

	var row *entity.SysKvCache
	err := dao.SysKvCache.Ctx(ctx).Where(do.SysKvCache{
		OwnerType: ownerType.String(),
		OwnerKey:  strings.TrimSpace(ownerKey),
		Namespace: strings.TrimSpace(namespace),
		CacheKey:  strings.TrimSpace(cacheKey),
	}).Scan(&row)
	if err != nil {
		return nil, false, err
	}
	if row == nil {
		return nil, false, nil
	}
	return buildCacheItem(row), true, nil
}

// Set stores or replaces a string cache value for the specified owner, namespace,
// and cache key.
//
// Parameters:
//   - ctx: request-scoped context used for database access, tracing, and cancellation.
//   - ownerType: cache owner category, used to isolate entries across different business scopes.
//   - ownerKey: concrete owner identifier within ownerType, such as a module key or plugin key.
//   - namespace: logical group name used to organize related cache entries for the same owner.
//   - cacheKey: concrete key to write inside the namespace.
//   - value: string payload to persist in the cache entry.
//   - expireSeconds: entry lifetime in seconds; 0 means never expire, and positive values create an absolute expiration time.
//
// Returns:
//   - *Item: the latest cache entry snapshot after the value has been written successfully.
//   - error: returned when identity parameters are invalid, the value exceeds the allowed size,
//     expireSeconds is negative, expired-entry cleanup fails, or the upsert operation fails.
func (s *serviceImpl) Set(
	ctx context.Context,
	ownerType OwnerType,
	ownerKey string,
	namespace string,
	cacheKey string,
	value string,
	expireSeconds int64,
) (*Item, error) {
	if err := s.validateIdentity(ownerType, ownerKey, namespace, cacheKey); err != nil {
		return nil, err
	}
	if err := validateMaxByteLength("缓存值", value, maxValueBytes); err != nil {
		return nil, err
	}

	expireAt, err := normalizeExpireAt(expireSeconds)
	if err != nil {
		return nil, err
	}
	if err = s.CleanupExpired(ctx); err != nil {
		return nil, err
	}

	err = s.upsert(ctx, ownerType, ownerKey, namespace, cacheKey, do.SysKvCache{
		ValueKind:  ValueKindString,
		ValueBytes: []byte(value),
		ValueInt:   0,
		ExpireAt:   expireAt,
	})
	if err != nil {
		return nil, err
	}
	return &Item{
		Key:       strings.TrimSpace(cacheKey),
		ValueKind: ValueKindString,
		Value:     value,
		IntValue:  0,
		ExpireAt:  expireAt,
	}, nil
}

// Delete removes the cache entry identified by ownerType, ownerKey, namespace,
// and cacheKey.
//
// Parameters:
//   - ctx: request-scoped context used for database access, tracing, and cancellation.
//   - ownerType: cache owner category, used to isolate entries across different business scopes.
//   - ownerKey: concrete owner identifier within ownerType, such as a module key or plugin key.
//   - namespace: logical group name used to organize related cache entries for the same owner.
//   - cacheKey: concrete key to delete inside the namespace.
//
// Returns:
//   - error: returned when identity parameters are invalid or the delete statement fails.
//     Deleting a non-existent entry is treated as a successful no-op.
func (s *serviceImpl) Delete(
	ctx context.Context,
	ownerType OwnerType,
	ownerKey string,
	namespace string,
	cacheKey string,
) error {
	if err := s.validateIdentity(ownerType, ownerKey, namespace, cacheKey); err != nil {
		return err
	}
	_, err := dao.SysKvCache.Ctx(ctx).Where(do.SysKvCache{
		OwnerType: ownerType.String(),
		OwnerKey:  strings.TrimSpace(ownerKey),
		Namespace: strings.TrimSpace(namespace),
		CacheKey:  strings.TrimSpace(cacheKey),
	}).Delete()
	return err
}

// Incr increments an integer cache value by delta and returns the updated entry snapshot.
//
// Parameters:
//   - ctx: request-scoped context used for database access, tracing, and cancellation.
//   - ownerType: cache owner category, used to isolate entries across different business scopes.
//   - ownerKey: concrete owner identifier within ownerType, such as a module key or plugin key.
//   - namespace: logical group name used to organize related cache entries for the same owner.
//   - cacheKey: concrete key to increment inside the namespace.
//   - delta: increment amount added to the current integer value; when the entry does not exist, delta becomes the initial value.
//   - expireSeconds: new entry lifetime in seconds; 0 keeps the entry non-expiring when creating a new item and preserves the current expiration when updating an existing item.
//
// Returns:
//   - *Item: the latest cache entry snapshot after the increment succeeds.
//   - error: returned when identity parameters are invalid, expireSeconds is negative,
//     expired-entry cleanup fails, the existing entry is not stored as an integer, or any database operation fails.
func (s *serviceImpl) Incr(
	ctx context.Context,
	ownerType OwnerType,
	ownerKey string,
	namespace string,
	cacheKey string,
	delta int64,
	expireSeconds int64,
) (*Item, error) {
	if err := s.validateIdentity(ownerType, ownerKey, namespace, cacheKey); err != nil {
		return nil, err
	}

	expireAt, err := normalizeExpireAt(expireSeconds)
	if err != nil {
		return nil, err
	}
	if err = s.CleanupExpired(ctx); err != nil {
		return nil, err
	}

	var updatedItem *Item
	err = dao.SysKvCache.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		var row *entity.SysKvCache
		if scanErr := dao.SysKvCache.Ctx(ctx).Where(do.SysKvCache{
			OwnerType: ownerType.String(),
			OwnerKey:  strings.TrimSpace(ownerKey),
			Namespace: strings.TrimSpace(namespace),
			CacheKey:  strings.TrimSpace(cacheKey),
		}).Scan(&row); scanErr != nil {
			return scanErr
		}

		currentExpireAt := expireAt
		if row != nil {
			if row.ValueKind != ValueKindInt {
				return gerror.New("缓存值不是整数，无法执行自增")
			}
			if currentExpireAt == nil {
				currentExpireAt = row.ExpireAt
			}
		}

		nextValue := delta
		if row != nil {
			nextValue = row.ValueInt + delta
		}

		updateErr := s.upsert(ctx, ownerType, ownerKey, namespace, cacheKey, do.SysKvCache{
			ValueKind:  ValueKindInt,
			ValueBytes: []byte{},
			ValueInt:   nextValue,
			ExpireAt:   currentExpireAt,
		})
		if updateErr != nil {
			return updateErr
		}

		updatedItem = &Item{
			Key:       strings.TrimSpace(cacheKey),
			ValueKind: ValueKindInt,
			Value:     strconv.FormatInt(nextValue, 10),
			IntValue:  nextValue,
			ExpireAt:  currentExpireAt,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updatedItem, nil
}

// Expire updates the expiration policy of a cache entry without changing its value.
//
// Parameters:
//   - ctx: request-scoped context used for database access, tracing, and cancellation.
//   - ownerType: cache owner category, used to isolate entries across different business scopes.
//   - ownerKey: concrete owner identifier within ownerType, such as a module key or plugin key.
//   - namespace: logical group name used to organize related cache entries for the same owner.
//   - cacheKey: concrete key whose expiration policy should be updated.
//   - expireSeconds: new lifetime in seconds; 0 clears the expiration and makes the entry persistent.
//
// Returns:
//   - bool: whether an existing cache entry was found and updated.
//   - *gtime.Time: the normalized absolute expiration time; nil means the entry will not expire.
//   - error: returned when identity parameters are invalid, expireSeconds is negative,
//     expired-entry cleanup fails, or the database update fails.
func (s *serviceImpl) Expire(
	ctx context.Context,
	ownerType OwnerType,
	ownerKey string,
	namespace string,
	cacheKey string,
	expireSeconds int64,
) (bool, *gtime.Time, error) {
	if err := s.validateIdentity(ownerType, ownerKey, namespace, cacheKey); err != nil {
		return false, nil, err
	}

	expireAt, err := normalizeExpireAt(expireSeconds)
	if err != nil {
		return false, nil, err
	}
	if err = s.CleanupExpired(ctx); err != nil {
		return false, nil, err
	}

	affected, err := dao.SysKvCache.Ctx(ctx).Where(do.SysKvCache{
		OwnerType: ownerType.String(),
		OwnerKey:  strings.TrimSpace(ownerKey),
		Namespace: strings.TrimSpace(namespace),
		CacheKey:  strings.TrimSpace(cacheKey),
	}).Data(do.SysKvCache{
		ExpireAt: expireAt,
	}).UpdateAndGetAffected()
	if err != nil {
		return false, nil, err
	}
	return affected > 0, expireAt, nil
}

// CleanupExpired removes all cache entries whose expiration time is earlier than
// the current time.
//
// Parameters:
//   - ctx: request-scoped context used for database access, tracing, and cancellation.
//
// Returns:
//   - error: returned when the cleanup delete statement fails. When no expired entries
//     exist, the method returns nil.
func (s *serviceImpl) CleanupExpired(ctx context.Context) error {
	cols := dao.SysKvCache.Columns()
	_, err := dao.SysKvCache.Ctx(ctx).
		WhereNotNull(cols.ExpireAt).
		WhereLT(cols.ExpireAt, gtime.Now()).
		Delete()
	return err
}

func (s *serviceImpl) upsert(
	ctx context.Context,
	ownerType OwnerType,
	ownerKey string,
	namespace string,
	cacheKey string,
	data do.SysKvCache,
) error {
	var row *entity.SysKvCache
	err := dao.SysKvCache.Ctx(ctx).Where(do.SysKvCache{
		OwnerType: ownerType.String(),
		OwnerKey:  strings.TrimSpace(ownerKey),
		Namespace: strings.TrimSpace(namespace),
		CacheKey:  strings.TrimSpace(cacheKey),
	}).Scan(&row)
	if err != nil {
		return err
	}

	if row == nil {
		_, err = dao.SysKvCache.Ctx(ctx).Data(do.SysKvCache{
			OwnerType:  ownerType.String(),
			OwnerKey:   strings.TrimSpace(ownerKey),
			Namespace:  strings.TrimSpace(namespace),
			CacheKey:   strings.TrimSpace(cacheKey),
			ValueKind:  data.ValueKind,
			ValueBytes: data.ValueBytes,
			ValueInt:   data.ValueInt,
			ExpireAt:   data.ExpireAt,
		}).Insert()
		return err
	}

	_, err = dao.SysKvCache.Ctx(ctx).Where(do.SysKvCache{Id: row.Id}).Data(do.SysKvCache{
		ValueKind:  data.ValueKind,
		ValueBytes: data.ValueBytes,
		ValueInt:   data.ValueInt,
		ExpireAt:   data.ExpireAt,
	}).Update()
	return err
}

func (s *serviceImpl) validateIdentity(
	ownerType OwnerType,
	ownerKey string,
	namespace string,
	cacheKey string,
) error {
	if err := validateByteLength("所属类型", ownerType.String(), maxOwnerTypeBytes); err != nil {
		return err
	}
	if err := validateByteLength("所属标识", ownerKey, maxOwnerKeyBytes); err != nil {
		return err
	}
	if err := validateByteLength("缓存命名空间", namespace, maxNamespaceBytes); err != nil {
		return err
	}
	if err := validateByteLength("缓存键", cacheKey, maxCacheKeyBytes); err != nil {
		return err
	}
	return nil
}

func normalizeExpireAt(expireSeconds int64) (*gtime.Time, error) {
	if expireSeconds < 0 {
		return nil, gerror.New("缓存过期秒数不能为负数")
	}
	if expireSeconds == 0 {
		return nil, nil
	}
	return gtime.Now().Add(time.Duration(expireSeconds) * time.Second), nil
}

func validateByteLength(field string, value string, maxBytes int) error {
	if strings.TrimSpace(value) == "" {
		return gerror.Newf("%s不能为空", field)
	}
	if len([]byte(value)) > maxBytes {
		return gerror.Newf("%s长度超出限制，最大允许 %d 字节", field, maxBytes)
	}
	return nil
}

func validateMaxByteLength(field string, value string, maxBytes int) error {
	if len([]byte(value)) > maxBytes {
		return gerror.Newf("%s长度超出限制，最大允许 %d 字节", field, maxBytes)
	}
	return nil
}

func buildCacheItem(row *entity.SysKvCache) *Item {
	if row == nil {
		return nil
	}

	item := &Item{
		Key:       row.CacheKey,
		ValueKind: row.ValueKind,
		IntValue:  row.ValueInt,
		ExpireAt:  row.ExpireAt,
	}
	switch row.ValueKind {
	case ValueKindInt:
		item.Value = strconv.FormatInt(row.ValueInt, 10)
	default:
		item.Value = string(row.ValueBytes)
	}
	return item
}
