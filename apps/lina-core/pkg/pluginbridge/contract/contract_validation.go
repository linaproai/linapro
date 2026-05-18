// This file implements bridge and route contract normalization and validation.

package contract

import (
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
)

// ValidateRouteContracts validates one plugin's route declarations in-place.
func ValidateRouteContracts(pluginID string, routes []*RouteContract) error {
	seen := make(map[string]struct{}, len(routes))
	for _, route := range routes {
		if route == nil {
			return gerror.New("dynamic route contract cannot be nil")
		}
		normalizeRouteContract(route)
		if route.Path == "" {
			return gerror.New("dynamic route path cannot be empty")
		}
		if !strings.HasPrefix(route.Path, "/") {
			return gerror.Newf("dynamic route path must start with /: %s", route.Path)
		}
		if route.Method == "" {
			return gerror.Newf("dynamic route method cannot be empty: %s", route.Path)
		}
		switch route.Access {
		case "", AccessLogin:
			route.Access = AccessLogin
		case AccessPublic:
		default:
			return gerror.Newf("dynamic route access only supports public/login: %s %s", route.Method, route.Path)
		}
		if route.Access == AccessPublic {
			if route.Permission != "" {
				return gerror.Newf("public dynamic route cannot declare permission: %s %s", route.Method, route.Path)
			}
		}
		if route.Permission != "" {
			parts := strings.Split(route.Permission, ":")
			if len(parts) != 3 {
				return gerror.Newf("dynamic route permission must use {pluginId}:{resource}:{action} format: %s", route.Permission)
			}
			if strings.TrimSpace(parts[0]) != strings.TrimSpace(pluginID) {
				return gerror.Newf("dynamic route permission must start with prefix %s: %s", pluginID, route.Permission)
			}
			if strings.TrimSpace(parts[1]) == "" || strings.TrimSpace(parts[2]) == "" {
				return gerror.Newf("dynamic route permission resource and action cannot be empty: %s", route.Permission)
			}
		}
		key := route.Method + " " + route.Path
		if _, ok := seen[key]; ok {
			return gerror.Newf("dynamic route method and path cannot be duplicated: %s", key)
		}
		seen[key] = struct{}{}
	}
	return nil
}

// NormalizeBridgeSpec normalizes bridge defaults in-place.
func NormalizeBridgeSpec(spec *BridgeSpec) {
	if spec == nil {
		return
	}
	spec.ABIVersion = normalizeLower(spec.ABIVersion, ABIVersionV1)
	spec.RuntimeKind = normalizeLower(spec.RuntimeKind, RuntimeKindWasm)
	spec.RequestCodec = normalizeLower(spec.RequestCodec, "")
	spec.ResponseCodec = normalizeLower(spec.ResponseCodec, "")
	spec.AllocExport = strings.TrimSpace(spec.AllocExport)
	spec.ExecuteExport = strings.TrimSpace(spec.ExecuteExport)
	if spec.AllocExport == "" {
		spec.AllocExport = DefaultGuestAllocExport
	}
	if spec.ExecuteExport == "" {
		spec.ExecuteExport = DefaultGuestExecuteExport
	}
}

// ValidateBridgeSpec validates bridge ABI compatibility in-place.
func ValidateBridgeSpec(spec *BridgeSpec) error {
	if spec == nil {
		return nil
	}
	NormalizeBridgeSpec(spec)
	if spec.ABIVersion != ABIVersionV1 {
		return gerror.Newf("dynamic route bridge ABI version is unsupported: %s", spec.ABIVersion)
	}
	if spec.RuntimeKind != RuntimeKindWasm {
		return gerror.Newf("dynamic route bridge runtimeKind only supports wasm: %s", spec.RuntimeKind)
	}
	if !spec.RouteExecution {
		return nil
	}
	if spec.RequestCodec != CodecProtobuf || spec.ResponseCodec != CodecProtobuf {
		return gerror.Newf(
			"dynamic route bridge executable mode only supports protobuf codecs: request=%s response=%s",
			spec.RequestCodec,
			spec.ResponseCodec,
		)
	}
	if spec.AllocExport == "" || spec.ExecuteExport == "" {
		return gerror.New("dynamic route bridge executable mode is missing guest export functions")
	}
	return nil
}

// normalizeRouteContract trims and normalizes one route contract in-place
// before validation or serialization.
func normalizeRouteContract(route *RouteContract) {
	route.Path = strings.TrimSpace(route.Path)
	route.Method = strings.ToUpper(strings.TrimSpace(route.Method))
	route.Summary = strings.TrimSpace(route.Summary)
	route.Description = strings.TrimSpace(route.Description)
	route.Access = strings.ToLower(strings.TrimSpace(route.Access))
	route.Permission = strings.TrimSpace(route.Permission)
	route.Meta = normalizeRouteMeta(route.Meta)
	route.RequestType = strings.TrimSpace(route.RequestType)
	if len(route.Tags) > 0 {
		tags := make([]string, 0, len(route.Tags))
		for _, item := range route.Tags {
			normalized := strings.TrimSpace(item)
			if normalized == "" {
				continue
			}
			tags = append(tags, normalized)
		}
		route.Tags = tags
	}
}

// normalizeRouteMeta trims plugin-defined route metadata and drops empty keys or values.
func normalizeRouteMeta(meta map[string]string) map[string]string {
	if len(meta) == 0 {
		return nil
	}
	normalized := make(map[string]string, len(meta))
	for key, value := range meta {
		normalizedKey := strings.TrimSpace(key)
		normalizedValue := strings.TrimSpace(value)
		if normalizedKey == "" || normalizedValue == "" {
			continue
		}
		normalized[normalizedKey] = normalizedValue
	}
	if len(normalized) == 0 {
		return nil
	}
	return normalized
}

// normalizeLower trims and lowercases one string, applying the default when
// the normalized result is empty.
func normalizeLower(value string, defaultValue string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" {
		return defaultValue
	}
	return normalized
}
