// Package operlogtype defines semantic operation-log types owned by the
// monitor-operlog plugin.
package operlogtype

import "strings"

// OperType identifies one semantic operation-log category.
type OperType string

// Canonical operation-log types.
const (
	// OperTypeCreate marks create/write requests.
	OperTypeCreate OperType = "create"
	// OperTypeUpdate marks update/write requests.
	OperTypeUpdate OperType = "update"
	// OperTypeDelete marks delete/write requests.
	OperTypeDelete OperType = "delete"
	// OperTypeExport marks read-side export requests.
	OperTypeExport OperType = "export"
	// OperTypeImport marks import requests.
	OperTypeImport OperType = "import"
	// OperTypeOther marks audited requests that do not map to the core verbs.
	OperTypeOther OperType = "other"
)

// publishedOperTypes preserves the stable ordering used by validators and
// error-message construction.
var publishedOperTypes = []OperType{
	OperTypeCreate,
	OperTypeUpdate,
	OperTypeDelete,
	OperTypeExport,
	OperTypeImport,
	OperTypeOther,
}

// String returns the canonical persisted value.
func (operType OperType) String() string {
	return string(operType)
}

// Normalize converts a raw value into one canonical operation-log type.
func Normalize(value string) OperType {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case OperTypeCreate.String():
		return OperTypeCreate
	case OperTypeUpdate.String():
		return OperTypeUpdate
	case OperTypeDelete.String():
		return OperTypeDelete
	case OperTypeExport.String():
		return OperTypeExport
	case OperTypeImport.String():
		return OperTypeImport
	case OperTypeOther.String():
		return OperTypeOther
	default:
		return ""
	}
}

// IsSupported reports whether the operation type is one of the published values.
func IsSupported(operType OperType) bool {
	for _, published := range publishedOperTypes {
		if operType == published {
			return true
		}
	}
	return false
}

// PublishedValues returns the canonical published values as plain strings.
func PublishedValues() []string {
	values := make([]string, 0, len(publishedOperTypes))
	for _, operType := range publishedOperTypes {
		values = append(values, operType.String())
	}
	return values
}
