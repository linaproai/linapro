// Package dictcap defines dictionary-domain capability contracts for plugins
// without exposing sys_dict_type or sys_dict_data storage.
package dictcap

import (
	"context"
	"lina-core/pkg/plugin/capability/capmodel"
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

// Service defines read-oriented dictionary capability methods.
type Service interface {
	// ResolveLabels resolves visible dictionary labels with opaque missing values.
	ResolveLabels(ctx context.Context, capCtx capmodel.CapabilityContext, input ResolveInput) (*capmodel.BatchResult[*LabelProjection, Value], error)
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
