// Package dictcap defines dictionary-domain capability contracts for plugins
// without exposing sys_dict_type or sys_dict_data storage.
package dictcap

import (
	"context"

	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/statusflag"
)

// Service defines governed dictionary capability methods. Reads use bounded
// batch/page sizes and tenant fallback metadata; refresh validates the target
// type and records cache revision impact through the domain owner.
type Service interface {
	// Type returns governed dictionary type subresource methods.
	Type() TypeService
	// Value returns governed dictionary value subresource methods.
	Value() ValueService
	// Refresh invalidates or reloads dictionary data for one type after
	// validating scope, audit source, idempotent cache revision, and cross-node
	// refresh semantics.
	Refresh(ctx context.Context, dictType Type) error
}

// TypeService defines governed dictionary type methods.
type TypeService interface {
	// Get returns one visible dictionary type. Risk: read; resource: type ID;
	// context: actor and tenant; data permission: target visibility;
	// performance: single read; audit/cache: read-only.
	Get(ctx context.Context, id int) (*TypeInfo, error)
	// BatchGet returns visible dictionary types and opaque missing IDs.
	BatchGet(ctx context.Context, ids []int) (*capmodel.BatchResult[*TypeInfo, int], error)
	// List returns bounded visible dictionary type candidates.
	List(ctx context.Context, input ListTypesInput) (*capmodel.PageResult[*TypeInfo], error)
	// EnsureVisible rejects when any requested type ID is absent or invisible.
	EnsureVisible(ctx context.Context, ids []int) error
	// EnsureKeysVisible rejects when any requested type key is absent or invisible.
	EnsureKeysVisible(ctx context.Context, keys []Type) error
	// Create creates one dictionary type through the owner.
	Create(ctx context.Context, input CreateTypeInput) (int, error)
	// Update updates one visible dictionary type through the owner.
	Update(ctx context.Context, input UpdateTypeInput) error
	// Delete deletes one visible dictionary type through the owner.
	Delete(ctx context.Context, id int) error
}

// ValueService defines governed dictionary value methods.
type ValueService interface {
	// Get returns one visible dictionary value by row ID.
	Get(ctx context.Context, id int) (*ValueInfo, error)
	// BatchGet returns visible dictionary values by type and value.
	BatchGet(ctx context.Context, input BatchGetValuesInput) (*capmodel.BatchResult[*ValueInfo, Value], error)
	// List returns one bounded page of visible dictionary values.
	List(ctx context.Context, input ListValuesInput) (*capmodel.PageResult[*ValueInfo], error)
	// ResolveLabels resolves visible dictionary labels with opaque missing values.
	ResolveLabels(ctx context.Context, input ResolveInput) (*capmodel.BatchResult[*LabelInfo, Value], error)
	// EnsureVisible rejects when any requested dictionary value row is absent or invisible.
	EnsureVisible(ctx context.Context, ids []int) error
	// EnsureValuesVisible rejects when any requested dictionary value is absent or invisible.
	EnsureValuesVisible(ctx context.Context, input ResolveInput) error
	// Create creates one dictionary value through the owner.
	Create(ctx context.Context, input CreateValueInput) (int, error)
	// Update updates one visible dictionary value through the owner.
	Update(ctx context.Context, input UpdateValueInput) error
	// Delete deletes one visible dictionary value through the owner.
	Delete(ctx context.Context, id int) error
	// DeleteByType deletes values under one visible dictionary type.
	DeleteByType(ctx context.Context, dictType Type) error
}

const (
	// MaxBatchGetTypes limits one dictionary type batch read.
	MaxBatchGetTypes = 100
	// MaxEnsureTypeKeysVisible limits one dictionary type key visibility check.
	MaxEnsureTypeKeysVisible = 200
	// MaxListTypesPageSize limits one dictionary type page.
	MaxListTypesPageSize = 200
	// MaxBatchGetValues limits one dictionary value batch read.
	MaxBatchGetValues = 200
	// MaxEnsureValuesVisible limits one dictionary value visibility check.
	MaxEnsureValuesVisible = 200
	// MaxListValuesPageSize limits one dictionary candidate page.
	MaxListValuesPageSize = 200
)

// Type identifies one dictionary type.
type Type string

// Value identifies one dictionary value.
type Value string

// LabelInfo describes one dictionary value label.
type LabelInfo struct {
	// Type is the dictionary type.
	Type Type
	// Value is the dictionary value.
	Value Value
	// LabelKey is the stable runtime i18n key.
	LabelKey string
	// Label is the optional locale-resolved label.
	Label string
}

// TypeInfo describes one dictionary type visible to plugins.
type TypeInfo struct {
	// ID is the host dictionary type identifier.
	ID int
	// Type is the dictionary type key.
	Type Type
	// Name is the display name.
	Name string
	// Status is the lifecycle status.
	Status statusflag.Enabled
	// LabelKey is the optional runtime i18n key.
	LabelKey string
	// Label is the optional locale-resolved label.
	Label string
}

// ValueInfo describes one dictionary value visible to plugins.
type ValueInfo struct {
	// ID is the host dictionary data identifier.
	ID int
	// Type is the dictionary type key.
	Type Type
	// Value is the dictionary value.
	Value Value
	// LabelKey is the stable runtime i18n key.
	LabelKey string
	// Label is the optional locale-resolved label.
	Label string
	// Sort controls dictionary display ordering.
	Sort int
	// Status is the lifecycle status.
	Status statusflag.Enabled
}

// ListTypesInput constrains dictionary type listing.
type ListTypesInput struct {
	// Keyword filters by type key or display name.
	Keyword string
	// Type filters by dictionary type key.
	Type Type
	// Status optionally filters by lifecycle status.
	Status *statusflag.Enabled
	// IncludeLabel asks the owner to resolve current-locale labels.
	IncludeLabel bool
	// Page constrains result size and stable sorting.
	Page capmodel.PageRequest
}

// CreateTypeInput describes one dictionary type create request.
type CreateTypeInput struct {
	// Type is the dictionary type key.
	Type Type
	// Name is the display name.
	Name string
	// Status is the lifecycle status.
	Status statusflag.Enabled
	// Remark stores optional operator notes.
	Remark string
}

// UpdateTypeInput describes one dictionary type update request.
type UpdateTypeInput struct {
	// ID identifies the target dictionary type row.
	ID int
	// Type optionally updates the dictionary type key.
	Type *Type
	// Name optionally updates the display name.
	Name *string
	// Status optionally updates the lifecycle status.
	Status *statusflag.Enabled
	// Remark optionally updates operator notes.
	Remark *string
}

// ResolveInput constrains dictionary label resolution.
type ResolveInput struct {
	// Type is the dictionary type to resolve.
	Type Type
	// Values contains requested values.
	Values []Value
	// IncludeLabel asks the domain to resolve labels in the current locale.
	IncludeLabel bool
}

// ListValuesInput constrains dictionary value candidate listing.
type ListValuesInput struct {
	// Type is the dictionary type to list.
	Type Type
	// Status optionally filters dictionary data by lifecycle status.
	Status *statusflag.Enabled
	// IncludeLabel asks the domain to resolve labels in the current locale.
	IncludeLabel bool
	// Page constrains candidate size and sorting.
	Page capmodel.PageRequest
}

// BatchGetValuesInput describes a bounded dictionary value batch read.
type BatchGetValuesInput struct {
	// Type is the dictionary type containing the values.
	Type Type
	// Values are dictionary values to read.
	Values []Value
	// IncludeLabel asks the owner to resolve current-locale labels.
	IncludeLabel bool
}

// CreateValueInput describes one dictionary value create request.
type CreateValueInput struct {
	// Type is the dictionary type key.
	Type Type
	// Value is the stable dictionary value.
	Value Value
	// Label is the source display label.
	Label string
	// Sort controls display ordering.
	Sort int
	// TagStyle stores optional UI tag style metadata.
	TagStyle string
	// CssClass stores optional UI CSS class metadata.
	CssClass string
	// Status is the lifecycle status.
	Status statusflag.Enabled
	// Remark stores optional operator notes.
	Remark string
}

// UpdateValueInput describes one dictionary value update request.
type UpdateValueInput struct {
	// ID identifies the target dictionary value row.
	ID int
	// Type optionally updates the dictionary type key.
	Type *Type
	// Value optionally updates the stable dictionary value.
	Value *Value
	// Label optionally updates the source display label.
	Label *string
	// Sort optionally updates display ordering.
	Sort *int
	// TagStyle optionally updates UI tag style metadata.
	TagStyle *string
	// CssClass optionally updates UI CSS class metadata.
	CssClass *string
	// Status optionally updates lifecycle status.
	Status *statusflag.Enabled
	// Remark optionally updates operator notes.
	Remark *string
}
