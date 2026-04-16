// This file defines typed order-direction helpers for applying governed sort
// clauses to GoFrame `gdb.Model` queries.

package gdbutil

import (
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
)

// OrderDirection represents one normalized sort direction accepted by backend
// query helpers.
type OrderDirection string

const (
	// OrderDirectionASC sorts records in ascending order.
	OrderDirectionASC OrderDirection = "asc"
	// OrderDirectionDESC sorts records in descending order.
	OrderDirectionDESC OrderDirection = "desc"
)

// String returns the canonical serialized value for the order direction.
func (direction OrderDirection) String() string {
	return string(direction)
}

// IsValid reports whether the direction is supported by the shared helper.
func (direction OrderDirection) IsValid() bool {
	switch direction {
	case OrderDirectionASC, OrderDirectionDESC:
		return true
	default:
		return false
	}
}

// NormalizeOrderDirection converts one raw input string into the canonical
// direction constant used by backend business code.
func NormalizeOrderDirection(value string) OrderDirection {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case OrderDirectionASC.String():
		return OrderDirectionASC
	case OrderDirectionDESC.String():
		return OrderDirectionDESC
	default:
		return ""
	}
}

// NormalizeOrderDirectionOrDefault converts one raw input string into a
// canonical direction constant and falls back to the supplied default when the
// input is empty or unsupported.
func NormalizeOrderDirectionOrDefault(value string, fallback OrderDirection) OrderDirection {
	normalized := NormalizeOrderDirection(value)
	if normalized.IsValid() {
		return normalized
	}
	if fallback.IsValid() {
		return fallback
	}
	return OrderDirectionDESC
}

// ApplyModelOrder applies one single-column order clause with the shared typed
// direction model so callers do not hand-build `ORDER BY` strings.
func ApplyModelOrder(model *gdb.Model, field string, direction OrderDirection) *gdb.Model {
	if model == nil || strings.TrimSpace(field) == "" {
		return model
	}
	if direction == OrderDirectionASC {
		return model.OrderAsc(field)
	}
	return model.OrderDesc(field)
}
