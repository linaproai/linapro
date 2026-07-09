// This file keeps PageRequest normalization and ID parsing coupled to the
// capability model instead of routing them through a separate helper package.

package capmodel

import (
	"strconv"
	"strings"

	"lina-core/pkg/bizerr"
)

// Normalize applies conservative defaults and hard limits to one page request.
func (page PageRequest) Normalize() (int, int) {
	pageNum := page.PageNum
	if pageNum <= 0 {
		pageNum = 1
	}
	pageSize := page.PageSize
	if pageSize <= 0 {
		pageSize = page.Limit
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 200 {
		pageSize = 200
	}
	return pageNum, pageSize
}

// ParseInt64IDs decodes string domain IDs into host-owned int64 keys.
func ParseInt64IDs[ID ~string](ids []ID, invalid func(ID)) ([]int64, map[int64]ID) {
	parsedIDs := make([]int64, 0, len(ids))
	requested := make(map[int64]ID, len(ids))
	for _, id := range ids {
		parsedID, err := strconv.ParseInt(strings.TrimSpace(string(id)), 10, 64)
		if err != nil || parsedID <= 0 {
			if invalid != nil {
				invalid(id)
			}
			continue
		}
		if _, exists := requested[parsedID]; exists {
			continue
		}
		requested[parsedID] = id
		parsedIDs = append(parsedIDs, parsedID)
	}
	return parsedIDs, requested
}

// ParsePositiveInt64Strings decodes positive host-owned integer IDs.
func ParsePositiveInt64Strings(values []string) ([]int64, error) {
	ids := make([]int64, 0, len(values))
	seen := make(map[int64]struct{}, len(values))
	for _, value := range values {
		parsedID, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
		if err != nil || parsedID <= 0 {
			if err != nil {
				return nil, err
			}
			return nil, bizerr.NewCode(CodeCapabilityDenied)
		}
		if _, exists := seen[parsedID]; exists {
			continue
		}
		seen[parsedID] = struct{}{}
		ids = append(ids, parsedID)
	}
	return ids, nil
}
