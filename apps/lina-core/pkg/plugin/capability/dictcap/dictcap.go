// Package dictcap defines dictionary-domain capability contracts for plugins
// without exposing sys_dict_type or sys_dict_data storage.
package dictcap

import (
	"context"
	"lina-core/pkg/plugin/capability/capmodel"
)

const (
	// MaxEnsureValuesVisible limits one dictionary value visibility check.
	MaxEnsureValuesVisible = 200
	// MaxListValuesPageSize limits one dictionary candidate page.
	MaxListValuesPageSize = 200
)

// Type identifies one dictionary type.
type Type string

// Value identifies one dictionary value.
type Value string

// LabelProjection describes one dictionary value label.
type LabelProjection struct {
	// Type is the dictionary type.
	Type Type
	// Value is the dictionary value.
	Value Value
	// LabelKey is the stable runtime i18n key.
	LabelKey string
	// Label is the optional locale-resolved label.
	Label string
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
	Status *int
	// IncludeLabel asks the domain to resolve labels in the current locale.
	IncludeLabel bool
	// Page constrains candidate size and sorting.
	Page capmodel.PageRequest
}

// Service defines read-oriented dictionary capability methods.
type Service interface {
	// ResolveLabels resolves visible dictionary labels with opaque missing values.
	ResolveLabels(ctx context.Context, capCtx capmodel.CapabilityContext, input ResolveInput) (*capmodel.BatchResult[*LabelProjection, Value], error)
	// ListValues returns one bounded page of visible dictionary value candidates.
	ListValues(ctx context.Context, capCtx capmodel.CapabilityContext, input ListValuesInput) (*capmodel.PageResult[*LabelProjection], error)
	// EnsureValuesVisible rejects when any requested dictionary value is absent or invisible.
	EnsureValuesVisible(ctx context.Context, capCtx capmodel.CapabilityContext, input ResolveInput) error
}

// AdminService defines dictionary management commands.
type AdminService interface {
	// Refresh invalidates or reloads dictionary projections for one type.
	Refresh(ctx context.Context, capCtx capmodel.CapabilityContext, dictType Type) error
}

// ScopeService defines host-internal dictionary governance helpers.
type ScopeService interface {
	// EnsureValuesVisible rejects when any dictionary value is outside caller scope.
	EnsureValuesVisible(ctx context.Context, capCtx capmodel.CapabilityContext, input ResolveInput) error
}
