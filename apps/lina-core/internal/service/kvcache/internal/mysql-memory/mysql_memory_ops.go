// This file implements MySQL MEMORY KV cache CRUD, increment, expire, and
// cleanup behaviors.

package mysqlmemory

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

// expiredCleanupBatchSize bounds each global cleanup pass so repeated
// cluster-node invocations stay idempotent and avoid full-table pressure.
const expiredCleanupBatchSize = 500

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
//   - bool: whether the unexpired cache entry exists.
//   - error: returned when the scoped cache key is invalid or the database query fails.
func (b *MySQLMemoryBackend) Get(
	ctx context.Context,
	ownerType OwnerType,
	cacheKey string,
) (*Item, bool, error) {
	identity, err := b.resolveIdentity(ownerType, cacheKey)
	if err != nil {
		return nil, false, err
	}

	var row *entity.SysKvCache
	cols := dao.SysKvCache.Columns()
	err = dao.SysKvCache.Ctx(ctx).Where(do.SysKvCache{
		OwnerType: ownerType.String(),
		OwnerKey:  identity.ownerKey,
		Namespace: identity.namespace,
		CacheKey:  identity.cacheKey,
	}).Wheref("(%s IS NULL OR %s > ?)", cols.ExpireAt, cols.ExpireAt, gtime.Now()).Scan(&row)
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
//   - bool: whether the unexpired cache entry exists.
//   - error: returned when the scoped cache key is invalid, the existing entry is not stored
//     as an integer, or the database query fails.
func (b *MySQLMemoryBackend) GetInt(
	ctx context.Context,
	ownerType OwnerType,
	cacheKey string,
) (int64, bool, error) {
	identity, err := b.resolveIdentity(ownerType, cacheKey)
	if err != nil {
		return 0, false, err
	}

	var row *entity.SysKvCache
	cols := dao.SysKvCache.Columns()
	err = dao.SysKvCache.Ctx(ctx).Where(do.SysKvCache{
		OwnerType: ownerType.String(),
		OwnerKey:  identity.ownerKey,
		Namespace: identity.namespace,
		CacheKey:  identity.cacheKey,
	}).Wheref("(%s IS NULL OR %s > ?)", cols.ExpireAt, cols.ExpireAt, gtime.Now()).Scan(&row)
	if err != nil {
		return 0, false, err
	}
	if row == nil {
		return 0, false, nil
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
//   - ttl: entry lifetime; 0 means never expire, and positive values create an absolute expiration time.
//
// Returns:
//   - *Item: the latest cache entry snapshot after the value has been written successfully.
//   - error: returned when the scoped cache key is invalid, the value exceeds the allowed size,
//     ttl is negative or the upsert operation fails.
func (b *MySQLMemoryBackend) Set(
	ctx context.Context,
	ownerType OwnerType,
	cacheKey string,
	value string,
	ttl time.Duration,
) (*Item, error) {
	identity, err := b.resolveIdentity(ownerType, cacheKey)
	if err != nil {
		return nil, err
	}
	if err := validateMaxByteLength("value", value, maxValueBytes); err != nil {
		return nil, err
	}

	expireAt, err := normalizeExpireAt(ttl)
	if err != nil {
		return nil, err
	}
	err = b.upsert(ctx, ownerType, identity, do.SysKvCache{
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
func (b *MySQLMemoryBackend) Delete(
	ctx context.Context,
	ownerType OwnerType,
	cacheKey string,
) error {
	identity, err := b.resolveIdentity(ownerType, cacheKey)
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
//   - ttl: new entry lifetime; 0 keeps the entry non-expiring when creating a new item and preserves the current expiration when updating an existing item.
//
// Returns:
//   - *Item: the latest cache entry snapshot after the increment succeeds.
//   - error: returned when the scoped cache key is invalid, ttl is negative,
//     expired-entry cleanup fails, the existing entry is not stored as an integer, or any database operation fails.
func (b *MySQLMemoryBackend) Incr(
	ctx context.Context,
	ownerType OwnerType,
	cacheKey string,
	delta int64,
	ttl time.Duration,
) (*Item, error) {
	identity, err := b.resolveIdentity(ownerType, cacheKey)
	if err != nil {
		return nil, err
	}

	expireAt, err := normalizeExpireAt(ttl)
	if err != nil {
		return nil, err
	}
	if err = b.cleanupExpiredIdentity(ctx, ownerType, identity); err != nil {
		return nil, err
	}

	var updatedItem *Item
	err = dao.SysKvCache.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		var row *entity.SysKvCache
		_, insertErr := dao.SysKvCache.Ctx(ctx).Data(do.SysKvCache{
			OwnerType:  ownerType.String(),
			OwnerKey:   identity.ownerKey,
			Namespace:  identity.namespace,
			CacheKey:   identity.cacheKey,
			ValueKind:  ValueKindInt,
			ValueBytes: []byte{},
			ValueInt:   0,
			ExpireAt:   expireAt,
		}).InsertIgnore()
		if insertErr != nil {
			return insertErr
		}

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

		updateData := do.SysKvCache{
			ValueKind:  ValueKindInt,
			ValueBytes: []byte{},
			ValueInt:   gdb.Raw("LAST_INSERT_ID(value_int + " + strconv.FormatInt(delta, 10) + ")"),
		}
		if expireAt != nil {
			updateData.ExpireAt = expireAt
		}
		_, updateErr := dao.SysKvCache.Ctx(ctx).Where(do.SysKvCache{
			OwnerType: ownerType.String(),
			OwnerKey:  identity.ownerKey,
			Namespace: identity.namespace,
			CacheKey:  identity.cacheKey,
			ValueKind: ValueKindInt,
		}).Data(updateData).Update()
		if updateErr != nil {
			return updateErr
		}
		nextValueVar, valueErr := dao.SysKvCache.Ctx(ctx).Raw("SELECT LAST_INSERT_ID()").Value()
		if valueErr != nil {
			return valueErr
		}
		nextValue := nextValueVar.Int64()

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
//   - ttl: new lifetime; 0 clears the expiration and makes the entry persistent.
//
// Returns:
//   - bool: whether an existing cache entry was found and updated.
//   - *gtime.Time: the normalized absolute expiration time; nil means the entry will not expire.
//   - error: returned when the scoped cache key is invalid, ttl is negative,
//     expired-entry cleanup fails, or the database update fails.
func (b *MySQLMemoryBackend) Expire(
	ctx context.Context,
	ownerType OwnerType,
	cacheKey string,
	ttl time.Duration,
) (bool, *gtime.Time, error) {
	identity, err := b.resolveIdentity(ownerType, cacheKey)
	if err != nil {
		return false, nil, err
	}

	expireAt, err := normalizeExpireAt(ttl)
	if err != nil {
		return false, nil, err
	}
	if err = b.cleanupExpiredIdentity(ctx, ownerType, identity); err != nil {
		return false, nil, err
	}

	var affected int64
	if expireAt == nil {
		affected, err = b.clearIdentityExpireAt(ctx, ownerType, identity)
	} else {
		affected, err = dao.SysKvCache.Ctx(ctx).Where(do.SysKvCache{
			OwnerType: ownerType.String(),
			OwnerKey:  identity.ownerKey,
			Namespace: identity.namespace,
			CacheKey:  identity.cacheKey,
		}).Data(do.SysKvCache{
			ExpireAt: expireAt,
		}).UpdateAndGetAffected()
	}
	if err != nil {
		return false, nil, err
	}
	return affected > 0, expireAt, nil
}

// CleanupExpired removes one bounded batch of cache entries whose expiration
// time is earlier than the current time.
//
// Parameters:
//   - ctx: request-scoped context used for database access, tracing, and cancellation.
//
// Returns:
//   - error: returned when the cleanup delete statement fails. When no expired entries
//     exist, the method returns nil.
func (b *MySQLMemoryBackend) CleanupExpired(ctx context.Context) error {
	cols := dao.SysKvCache.Columns()
	_, err := dao.SysKvCache.Ctx(ctx).
		WhereNotNull(cols.ExpireAt).
		WhereLT(cols.ExpireAt, gtime.Now()).
		OrderAsc(cols.ExpireAt).
		Limit(expiredCleanupBatchSize).
		Delete()
	return err
}

// upsert inserts one cache entry when absent and always updates the matching
// unique key in place so concurrent writers follow last-write-wins semantics.
func (b *MySQLMemoryBackend) upsert(
	ctx context.Context,
	ownerType OwnerType,
	identity *cacheIdentity,
	data do.SysKvCache,
) error {
	insertData := do.SysKvCache{
		OwnerType:  ownerType.String(),
		OwnerKey:   identity.ownerKey,
		Namespace:  identity.namespace,
		CacheKey:   identity.cacheKey,
		ValueKind:  data.ValueKind,
		ValueBytes: data.ValueBytes,
		ValueInt:   data.ValueInt,
		ExpireAt:   data.ExpireAt,
	}
	_, err := dao.SysKvCache.Ctx(ctx).Data(insertData).InsertIgnore()
	if err != nil {
		return err
	}

	updateModel := dao.SysKvCache.Ctx(ctx).Where(do.SysKvCache{
		OwnerType: ownerType.String(),
		OwnerKey:  identity.ownerKey,
		Namespace: identity.namespace,
		CacheKey:  identity.cacheKey,
	}).Data(do.SysKvCache{
		ValueKind:  data.ValueKind,
		ValueBytes: data.ValueBytes,
		ValueInt:   data.ValueInt,
		ExpireAt:   data.ExpireAt,
	})
	if data.ExpireAt == nil {
		cols := dao.SysKvCache.Columns()
		updateModel = updateModel.Fields(cols.ValueKind, cols.ValueBytes, cols.ValueInt)
	}
	_, err = updateModel.Update()
	if err == nil && data.ExpireAt == nil {
		_, err = b.clearIdentityExpireAt(ctx, ownerType, identity)
	}
	return err
}

// clearIdentityExpireAt clears the expiration column for one cache identity.
func (b *MySQLMemoryBackend) clearIdentityExpireAt(
	ctx context.Context,
	ownerType OwnerType,
	identity *cacheIdentity,
) (int64, error) {
	cols := dao.SysKvCache.Columns()
	return dao.SysKvCache.Ctx(ctx).Where(do.SysKvCache{
		OwnerType: ownerType.String(),
		OwnerKey:  identity.ownerKey,
		Namespace: identity.namespace,
		CacheKey:  identity.cacheKey,
	}).Data(cols.ExpireAt, gdb.Raw("NULL")).UpdateAndGetAffected()
}

// cleanupExpiredIdentity deletes one expired cache row before a write path
// mutates the same identity.
func (b *MySQLMemoryBackend) cleanupExpiredIdentity(
	ctx context.Context,
	ownerType OwnerType,
	identity *cacheIdentity,
) error {
	cols := dao.SysKvCache.Columns()
	_, err := dao.SysKvCache.Ctx(ctx).
		Where(do.SysKvCache{
			OwnerType: ownerType.String(),
			OwnerKey:  identity.ownerKey,
			Namespace: identity.namespace,
			CacheKey:  identity.cacheKey,
		}).
		WhereNotNull(cols.ExpireAt).
		WhereLT(cols.ExpireAt, gtime.Now()).
		Delete()
	return err
}

// resolveIdentity parses and validates one public cache key under the provided
// owner type.
func (b *MySQLMemoryBackend) resolveIdentity(
	ownerType OwnerType,
	cacheKey string,
) (*cacheIdentity, error) {
	identity, err := parseCacheKey(cacheKey)
	if err != nil {
		return nil, err
	}
	if err = b.validateIdentity(ownerType, identity.ownerKey, identity.namespace, identity.cacheKey); err != nil {
		return nil, err
	}
	return identity, nil
}

// validateIdentity validates the byte-length constraints for one decoded cache
// identity.
func (b *MySQLMemoryBackend) validateIdentity(
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
func normalizeExpireAt(ttl time.Duration) (*gtime.Time, error) {
	if ttl < 0 {
		return nil, bizerr.NewCode(CodeKVCacheExpireSecondsNegative)
	}
	if ttl == 0 {
		return nil, nil
	}
	return gtime.Now().Add(ttl), nil
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
