// This file implements distributed KV cache CRUD, increment, expire, and cleanup behaviors.

package kvcache

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/model/entity"
	"lina-core/pkg/bizerr"
)

// Get returns the current cache entry snapshot identified by ownerType and one
// scoped cache key.
//
// Parameters:
//   - ctx: request-scoped context used for database access, tracing, and cancellation.
//   - ownerType: cache owner category, used to isolate entries across different business scopes.
//   - cacheKey: scoped cache key that encodes ownerKey, namespace, and the logical key.
//
// Returns:
//   - *Item: the cache entry snapshot when the entry exists, including value kind, value, and expiration time.
//   - bool: whether the cache entry exists after expired data has been cleaned up.
//   - error: returned when the scoped cache key is invalid, expired-entry cleanup fails, or the database query fails.
func (s *serviceImpl) Get(
	ctx context.Context,
	ownerType OwnerType,
	cacheKey string,
) (*Item, bool, error) {
	identity, err := s.resolveIdentity(ownerType, cacheKey)
	if err != nil {
		return nil, false, err
	}
	if err := s.CleanupExpired(ctx); err != nil {
		return nil, false, err
	}

	var row *entity.SysKvCache
	err = dao.SysKvCache.Ctx(ctx).Where(do.SysKvCache{
		OwnerType: ownerType.String(),
		OwnerKey:  identity.ownerKey,
		Namespace: identity.namespace,
		CacheKey:  identity.cacheKey,
	}).Scan(&row)
	if err != nil {
		return nil, false, err
	}
	if row == nil {
		return nil, false, nil
	}
	return buildCacheItem(row), true, nil
}

// GetInt returns the current integer cache value identified by ownerType and
// one scoped cache key.
//
// Parameters:
//   - ctx: request-scoped context used for database access, tracing, and cancellation.
//   - ownerType: cache owner category, used to isolate entries across different business scopes.
//   - cacheKey: scoped cache key that encodes ownerKey, namespace, and the logical key.
//
// Returns:
//   - int64: the integer cache value when the entry exists and is stored as an integer.
//   - bool: whether the cache entry exists after optional single-row expiration cleanup.
//   - error: returned when the scoped cache key is invalid, the existing entry is not stored
//     as an integer, or the database query/delete fails.
func (s *serviceImpl) GetInt(
	ctx context.Context,
	ownerType OwnerType,
	cacheKey string,
) (int64, bool, error) {
	identity, err := s.resolveIdentity(ownerType, cacheKey)
	if err != nil {
		return 0, false, err
	}

	// Keep revision reads lightweight: unlike Get/Incr, this path intentionally
	// skips global CleanupExpired so watcher polling stays read-dominant.
	var row *entity.SysKvCache
	err = dao.SysKvCache.Ctx(ctx).Where(do.SysKvCache{
		OwnerType: ownerType.String(),
		OwnerKey:  identity.ownerKey,
		Namespace: identity.namespace,
		CacheKey:  identity.cacheKey,
	}).Scan(&row)
	if err != nil {
		return 0, false, err
	}
	if row == nil {
		return 0, false, nil
	}
	if row.ExpireAt != nil && row.ExpireAt.Before(gtime.Now()) {
		// Lazily remove only the expired row we just touched so hot keys do not
		// pay the cost of scanning unrelated cache entries.
		_, err = dao.SysKvCache.Ctx(ctx).Where(do.SysKvCache{Id: row.Id}).Delete()
		return 0, false, err
	}
	if row.ValueKind != ValueKindInt {
		return 0, false, bizerr.NewCode(CodeKVCacheValueNotInteger)
	}
	return row.ValueInt, true, nil
}

// Set stores or replaces a string cache value for the specified scoped cache key.
//
// Parameters:
//   - ctx: request-scoped context used for database access, tracing, and cancellation.
//   - ownerType: cache owner category, used to isolate entries across different business scopes.
//   - cacheKey: scoped cache key that encodes ownerKey, namespace, and the logical key.
//   - value: string payload to persist in the cache entry.
//   - expireSeconds: entry lifetime in seconds; 0 means never expire, and positive values create an absolute expiration time.
//
// Returns:
//   - *Item: the latest cache entry snapshot after the value has been written successfully.
//   - error: returned when the scoped cache key is invalid, the value exceeds the allowed size,
//     expireSeconds is negative, expired-entry cleanup fails, or the upsert operation fails.
func (s *serviceImpl) Set(
	ctx context.Context,
	ownerType OwnerType,
	cacheKey string,
	value string,
	expireSeconds int64,
) (*Item, error) {
	identity, err := s.resolveIdentity(ownerType, cacheKey)
	if err != nil {
		return nil, err
	}
	if err := validateMaxByteLength("value", value, maxValueBytes); err != nil {
		return nil, err
	}

	expireAt, err := normalizeExpireAt(expireSeconds)
	if err != nil {
		return nil, err
	}
	if err = s.CleanupExpired(ctx); err != nil {
		return nil, err
	}

	err = s.upsert(ctx, ownerType, identity, do.SysKvCache{
		ValueKind:  ValueKindString,
		ValueBytes: []byte(value),
		ValueInt:   0,
		ExpireAt:   expireAt,
	})
	if err != nil {
		return nil, err
	}
	return &Item{
		Key:       identity.cacheKey,
		ValueKind: ValueKindString,
		Value:     value,
		IntValue:  0,
		ExpireAt:  expireAt,
	}, nil
}

// Delete removes the cache entry identified by ownerType and one scoped cache key.
//
// Parameters:
//   - ctx: request-scoped context used for database access, tracing, and cancellation.
//   - ownerType: cache owner category, used to isolate entries across different business scopes.
//   - cacheKey: scoped cache key that encodes ownerKey, namespace, and the logical key.
//
// Returns:
//   - error: returned when the scoped cache key is invalid or the delete statement fails.
//     Deleting a non-existent entry is treated as a successful no-op.
func (s *serviceImpl) Delete(
	ctx context.Context,
	ownerType OwnerType,
	cacheKey string,
) error {
	identity, err := s.resolveIdentity(ownerType, cacheKey)
	if err != nil {
		return err
	}
	_, err = dao.SysKvCache.Ctx(ctx).Where(do.SysKvCache{
		OwnerType: ownerType.String(),
		OwnerKey:  identity.ownerKey,
		Namespace: identity.namespace,
		CacheKey:  identity.cacheKey,
	}).Delete()
	return err
}

// Incr increments an integer cache value by delta and returns the updated entry snapshot.
//
// Parameters:
//   - ctx: request-scoped context used for database access, tracing, and cancellation.
//   - ownerType: cache owner category, used to isolate entries across different business scopes.
//   - cacheKey: scoped cache key that encodes ownerKey, namespace, and the logical key.
//   - delta: increment amount added to the current integer value; when the entry does not exist, delta becomes the initial value.
//   - expireSeconds: new entry lifetime in seconds; 0 keeps the entry non-expiring when creating a new item and preserves the current expiration when updating an existing item.
//
// Returns:
//   - *Item: the latest cache entry snapshot after the increment succeeds.
//   - error: returned when the scoped cache key is invalid, expireSeconds is negative,
//     expired-entry cleanup fails, the existing entry is not stored as an integer, or any database operation fails.
func (s *serviceImpl) Incr(
	ctx context.Context,
	ownerType OwnerType,
	cacheKey string,
	delta int64,
	expireSeconds int64,
) (*Item, error) {
	identity, err := s.resolveIdentity(ownerType, cacheKey)
	if err != nil {
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
			OwnerKey:  identity.ownerKey,
			Namespace: identity.namespace,
			CacheKey:  identity.cacheKey,
		}).Scan(&row); scanErr != nil {
			return scanErr
		}

		currentExpireAt := expireAt
		if row != nil {
			if row.ValueKind != ValueKindInt {
				return bizerr.NewCode(CodeKVCacheIncrementValueNotInteger)
			}
			if currentExpireAt == nil {
				currentExpireAt = row.ExpireAt
			}
		}

		nextValue := delta
		if row != nil {
			nextValue = row.ValueInt + delta
		}

		updateErr := s.upsert(ctx, ownerType, identity, do.SysKvCache{
			ValueKind:  ValueKindInt,
			ValueBytes: []byte{},
			ValueInt:   nextValue,
			ExpireAt:   currentExpireAt,
		})
		if updateErr != nil {
			return updateErr
		}

		updatedItem = &Item{
			Key:       identity.cacheKey,
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
//   - cacheKey: scoped cache key that encodes ownerKey, namespace, and the logical key.
//   - expireSeconds: new lifetime in seconds; 0 clears the expiration and makes the entry persistent.
//
// Returns:
//   - bool: whether an existing cache entry was found and updated.
//   - *gtime.Time: the normalized absolute expiration time; nil means the entry will not expire.
//   - error: returned when the scoped cache key is invalid, expireSeconds is negative,
//     expired-entry cleanup fails, or the database update fails.
func (s *serviceImpl) Expire(
	ctx context.Context,
	ownerType OwnerType,
	cacheKey string,
	expireSeconds int64,
) (bool, *gtime.Time, error) {
	identity, err := s.resolveIdentity(ownerType, cacheKey)
	if err != nil {
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
		OwnerKey:  identity.ownerKey,
		Namespace: identity.namespace,
		CacheKey:  identity.cacheKey,
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

// upsert inserts one cache entry when absent or updates the existing entry in
// place.
func (s *serviceImpl) upsert(
	ctx context.Context,
	ownerType OwnerType,
	identity *cacheIdentity,
	data do.SysKvCache,
) error {
	var row *entity.SysKvCache
	err := dao.SysKvCache.Ctx(ctx).Where(do.SysKvCache{
		OwnerType: ownerType.String(),
		OwnerKey:  identity.ownerKey,
		Namespace: identity.namespace,
		CacheKey:  identity.cacheKey,
	}).Scan(&row)
	if err != nil {
		return err
	}

	if row == nil {
		_, err = dao.SysKvCache.Ctx(ctx).Data(do.SysKvCache{
			OwnerType:  ownerType.String(),
			OwnerKey:   identity.ownerKey,
			Namespace:  identity.namespace,
			CacheKey:   identity.cacheKey,
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

// resolveIdentity parses and validates one public cache key under the provided
// owner type.
func (s *serviceImpl) resolveIdentity(
	ownerType OwnerType,
	cacheKey string,
) (*cacheIdentity, error) {
	identity, err := parseCacheKey(cacheKey)
	if err != nil {
		return nil, err
	}
	if err = s.validateIdentity(ownerType, identity.ownerKey, identity.namespace, identity.cacheKey); err != nil {
		return nil, err
	}
	return identity, nil
}

// validateIdentity validates the byte-length constraints for one decoded cache
// identity.
func (s *serviceImpl) validateIdentity(
	ownerType OwnerType,
	ownerKey string,
	namespace string,
	cacheKey string,
) error {
	if err := validateByteLength("ownerType", ownerType.String(), maxOwnerTypeBytes); err != nil {
		return err
	}
	if err := validateByteLength("ownerKey", ownerKey, maxOwnerKeyBytes); err != nil {
		return err
	}
	if err := validateByteLength("namespace", namespace, maxNamespaceBytes); err != nil {
		return err
	}
	if err := validateByteLength("cacheKey", cacheKey, maxCacheKeyBytes); err != nil {
		return err
	}
	return nil
}

// normalizeExpireAt converts an expiration duration in seconds into an
// absolute expiration time, or nil for persistent entries.
func normalizeExpireAt(expireSeconds int64) (*gtime.Time, error) {
	if expireSeconds < 0 {
		return nil, bizerr.NewCode(CodeKVCacheExpireSecondsNegative)
	}
	if expireSeconds == 0 {
		return nil, nil
	}
	return gtime.Now().Add(time.Duration(expireSeconds) * time.Second), nil
}

// validateByteLength enforces a non-empty string field with a maximum byte
// length.
func validateByteLength(field string, value string, maxBytes int) error {
	if strings.TrimSpace(value) == "" {
		return bizerr.NewCode(CodeKVCacheFieldRequired, bizerr.P("field", field))
	}
	if len([]byte(value)) > maxBytes {
		return bizerr.NewCode(
			CodeKVCacheFieldTooLong,
			bizerr.P("field", field),
			bizerr.P("maxBytes", maxBytes),
		)
	}
	return nil
}

// validateMaxByteLength enforces only the maximum byte length for an optional
// string field.
func validateMaxByteLength(field string, value string, maxBytes int) error {
	if len([]byte(value)) > maxBytes {
		return bizerr.NewCode(
			CodeKVCacheValueTooLong,
			bizerr.P("field", field),
			bizerr.P("maxBytes", maxBytes),
		)
	}
	return nil
}

// buildCacheItem converts one persisted cache row into the public cache item
// snapshot returned by the service.
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
