// This file implements host-service manifest declaration validation and normalization.

package hostservice

import (
	"net/url"
	"path"
	"sort"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
)

// ValidateHostServiceSpecs validates and normalizes host service declarations
// in-place. Data host-service declarations require ValidateHostServiceSpecsForPlugin
// so table ownership is always checked against a concrete plugin ID.
func ValidateHostServiceSpecs(specs []*HostServiceSpec) error {
	return validateHostServiceSpecs(specs, "")
}

// ValidateHostServiceSpecsForPlugin validates and normalizes host service
// declarations in-place, additionally enforcing plugin-owned data tables.
func ValidateHostServiceSpecsForPlugin(pluginID string, specs []*HostServiceSpec) error {
	return validateHostServiceSpecs(specs, pluginID)
}

// validateHostServiceSpecs applies structural host-service validation and, when
// pluginID is present, data-service table ownership validation.
func validateHostServiceSpecs(specs []*HostServiceSpec, pluginID string) error {
	if len(specs) == 0 {
		return nil
	}

	seenServices := make(map[string]struct{}, len(specs))
	for _, spec := range specs {
		if spec == nil {
			return gerror.New("host service declaration cannot be nil")
		}
		spec.Service = normalizeHostServiceName(spec.Service)
		if spec.Service == "" {
			return gerror.New("host service name cannot be empty")
		}
		if _, ok := hostServiceMethodCapabilityMap[spec.Service]; !ok {
			return gerror.Newf("unknown host service declaration: %s", spec.Service)
		}
		if _, exists := seenServices[spec.Service]; exists {
			return gerror.Newf("host service cannot be declared more than once: %s", spec.Service)
		}
		seenServices[spec.Service] = struct{}{}

		methodSeen := make(map[string]struct{}, len(spec.Methods))
		methods := make([]string, 0, len(spec.Methods))
		for _, rawMethod := range spec.Methods {
			method := normalizeHostServiceMethod(rawMethod)
			if method == "" {
				return gerror.Newf("host service %s method cannot be empty", spec.Service)
			}
			if RequiredCapabilityForHostServiceMethod(spec.Service, method) == "" {
				return gerror.Newf("host service %s does not support method: %s", spec.Service, method)
			}
			if _, exists := methodSeen[method]; exists {
				return gerror.Newf("host service %s method cannot be duplicated: %s", spec.Service, method)
			}
			methodSeen[method] = struct{}{}
			methods = append(methods, method)
		}
		if len(methods) == 0 {
			return gerror.Newf("host service %s must declare at least one method", spec.Service)
		}
		sort.Strings(methods)
		spec.Methods = methods

		tableSeen := make(map[string]struct{}, len(spec.Tables))
		tables := make([]string, 0, len(spec.Tables))
		for _, rawTable := range spec.Tables {
			table := strings.TrimSpace(rawTable)
			if table == "" {
				return gerror.Newf("host service %s table cannot be empty", spec.Service)
			}
			if _, exists := tableSeen[table]; exists {
				return gerror.Newf("host service %s table cannot be duplicated: %s", spec.Service, table)
			}
			tableSeen[table] = struct{}{}
			tables = append(tables, table)
		}
		sort.Strings(tables)
		spec.Tables = tables

		keySeen := make(map[string]struct{}, len(spec.Keys))
		keys := make([]string, 0, len(spec.Keys))
		for _, rawKey := range spec.Keys {
			key := strings.TrimSpace(rawKey)
			if key == "" || key == "." {
				return gerror.Newf("host service %s key cannot be empty or root", spec.Service)
			}
			if _, exists := keySeen[key]; exists {
				return gerror.Newf("host service %s key cannot be duplicated: %s", spec.Service, key)
			}
			keySeen[key] = struct{}{}
			keys = append(keys, key)
		}
		sort.Strings(keys)
		spec.Keys = keys

		pathSeen := make(map[string]struct{}, len(spec.Paths))
		paths := make([]string, 0, len(spec.Paths))
		for _, rawPath := range spec.Paths {
			normalizedPath, err := normalizeDeclaredPathForService(spec.Service, rawPath)
			if err != nil {
				return gerror.Wrapf(err, "host service %s has invalid path", spec.Service)
			}
			if _, exists := pathSeen[normalizedPath]; exists {
				return gerror.Newf("host service %s path cannot be duplicated: %s", spec.Service, normalizedPath)
			}
			pathSeen[normalizedPath] = struct{}{}
			paths = append(paths, normalizedPath)
		}
		sort.Strings(paths)
		spec.Paths = paths

		resourceSeen := make(map[string]struct{}, len(spec.Resources))
		resources := make([]*HostServiceResourceSpec, 0, len(spec.Resources))
		for _, resource := range spec.Resources {
			if resource == nil {
				return gerror.Newf("host service %s resource declaration cannot be nil", spec.Service)
			}
			resource.Ref = strings.TrimSpace(resource.Ref)
			if resource.Ref == "" {
				return gerror.Newf("host service %s resource ref cannot be empty", spec.Service)
			}
			if _, exists := resourceSeen[resource.Ref]; exists {
				return gerror.Newf("host service %s resource ref cannot be duplicated: %s", spec.Service, resource.Ref)
			}
			resourceSeen[resource.Ref] = struct{}{}
			resource.AllowMethods = normalizeUpperStringSlice(resource.AllowMethods)
			resource.HeaderAllowList = normalizeLowerStringSlice(resource.HeaderAllowList)
			resource.Attributes = normalizeStringMap(resource.Attributes)
			resources = append(resources, resource)
		}
		sort.Slice(resources, func(i, j int) bool {
			return resources[i].Ref < resources[j].Ref
		})
		spec.Resources = resources

		if _, ok := hostServicesWithPaths[spec.Service]; ok {
			if len(spec.Tables) > 0 {
				return gerror.Newf("host service %s cannot declare tables", spec.Service)
			}
			if len(spec.Keys) > 0 {
				return gerror.Newf("host service %s cannot declare keys", spec.Service)
			}
			if len(spec.Resources) > 0 {
				return gerror.Newf("host service %s cannot declare resource refs", spec.Service)
			}
			if len(spec.Paths) == 0 {
				return gerror.Newf("host service %s must declare at least one path", spec.Service)
			}
			continue
		}

		if _, ok := hostServicesWithTables[spec.Service]; ok {
			if len(spec.Paths) > 0 {
				return gerror.Newf("host service %s cannot declare paths", spec.Service)
			}
			if len(spec.Keys) > 0 {
				return gerror.Newf("host service %s cannot declare keys", spec.Service)
			}
			if len(spec.Resources) > 0 {
				return gerror.Newf("host service %s cannot declare resources", spec.Service)
			}
			if len(spec.Tables) == 0 {
				return gerror.Newf("host service %s must declare at least one table", spec.Service)
			}
			if spec.Service == HostServiceData {
				if strings.TrimSpace(pluginID) == "" {
					return gerror.New("host service data requires plugin-aware validation")
				}
				if err := validateDataServiceTablesForPlugin(pluginID, spec.Tables); err != nil {
					return err
				}
			}
			continue
		}
		if _, ok := hostServicesWithKeys[spec.Service]; ok {
			if len(spec.Paths) > 0 {
				return gerror.Newf("host service %s cannot declare paths", spec.Service)
			}
			if len(spec.Tables) > 0 {
				return gerror.Newf("host service %s cannot declare tables", spec.Service)
			}
			if len(spec.Resources) > 0 {
				return gerror.Newf("host service %s cannot declare resources", spec.Service)
			}
			if len(spec.Keys) == 0 {
				return gerror.Newf("host service %s must declare at least one key", spec.Service)
			}
			continue
		}
		if len(spec.Tables) > 0 {
			return gerror.Newf("host service %s cannot declare tables", spec.Service)
		}
		if len(spec.Paths) > 0 {
			return gerror.Newf("host service %s cannot declare paths", spec.Service)
		}
		if len(spec.Keys) > 0 {
			return gerror.Newf("host service %s cannot declare keys", spec.Service)
		}

		methodResourceKind := hostServiceResourceKindForMethods(spec.Service, spec.Methods)
		if methodResourceKind == HostServiceResourceNone {
			if len(spec.Resources) > 0 {
				return gerror.Newf("host service %s cannot declare resources", spec.Service)
			}
			continue
		}
		if methodResourceKind != "" && methodResourceKind != HostServiceResourceRef {
			return gerror.Newf("host service %s uses unsupported resource declaration kind: %s", spec.Service, methodResourceKind)
		}
		if len(spec.Resources) == 0 {
			return gerror.Newf("host service %s must declare at least one resource", spec.Service)
		}
		if spec.Service == HostServiceNetwork {
			for _, resource := range spec.Resources {
				if resource == nil {
					continue
				}
				if len(resource.AllowMethods) > 0 || len(resource.HeaderAllowList) > 0 || resource.TimeoutMs > 0 || resource.MaxBodyBytes > 0 || len(resource.Attributes) > 0 {
					return gerror.Newf("host service %s only allows url declarations and cannot include extra governance fields: %s", spec.Service, resource.Ref)
				}
				if err := validateNetworkURLPattern(resource.Ref); err != nil {
					return gerror.Wrapf(err, "host service %s has invalid url", spec.Service)
				}
			}
		}
	}

	sort.Slice(specs, func(i, j int) bool {
		return specs[i].Service < specs[j].Service
	})
	return nil
}

// validateDataServiceTablesForPlugin restricts dynamic data-service access to
// the current plugin namespace.
func validateDataServiceTablesForPlugin(pluginID string, tables []string) error {
	normalizedPluginID := normalizePluginIDForTableNamespace(pluginID)
	if normalizedPluginID == "" {
		return gerror.New("host service data requires plugin ID")
	}
	ownedTable := "plugin_" + normalizedPluginID
	ownedPrefix := ownedTable + "_"
	for _, table := range tables {
		normalizedTable := strings.ToLower(strings.TrimSpace(table))
		if normalizedTable == "" {
			continue
		}
		if strings.HasPrefix(normalizedTable, "sys_") {
			return gerror.Newf("host service data cannot declare host core table: %s", table)
		}
		if normalizedTable != ownedTable && !strings.HasPrefix(normalizedTable, ownedPrefix) {
			return gerror.Newf(
				"host service data table must belong to plugin %s namespace %s*: %s",
				pluginID,
				ownedPrefix,
				table,
			)
		}
	}
	return nil
}

// normalizePluginIDForTableNamespace converts a plugin ID into the database
// namespace used by plugin-owned tables.
func normalizePluginIDForTableNamespace(pluginID string) string {
	trimmed := strings.ToLower(strings.TrimSpace(pluginID))
	if trimmed == "" {
		return ""
	}
	replacer := strings.NewReplacer("-", "_", ".", "_")
	return replacer.Replace(trimmed)
}

// hostServiceResourceKindForMethods returns the resource shape required by the
// declared methods. Mixed none+resource methods require resource declarations
// only when at least one resource-bound method is present.
func hostServiceResourceKindForMethods(service string, methods []string) HostServiceResourceKind {
	methodResources := hostServiceMethodResourceMap[service]
	if len(methodResources) == 0 {
		if _, ok := hostServicesWithoutResources[service]; ok {
			return HostServiceResourceNone
		}
		return ""
	}
	requiresResource := false
	for _, rawMethod := range methods {
		method := normalizeHostServiceMethod(rawMethod)
		switch methodResources[method] {
		case HostServiceResourceRef, HostServiceResourceReserved:
			requiresResource = true
		case HostServiceResourceNone, "":
		default:
			return methodResources[method]
		}
	}
	if requiresResource {
		return HostServiceResourceRef
	}
	return HostServiceResourceNone
}

// NormalizeHostServiceSpecs returns deep-cloned and normalized host service
// declarations. Data declarations must use NormalizeHostServiceSpecsForPlugin.
func NormalizeHostServiceSpecs(specs []*HostServiceSpec) ([]*HostServiceSpec, error) {
	return normalizeHostServiceSpecs(specs, "")
}

// NormalizeHostServiceSpecsForPlugin returns deep-cloned and normalized host
// service declarations while enforcing plugin-owned data table declarations.
func NormalizeHostServiceSpecsForPlugin(pluginID string, specs []*HostServiceSpec) ([]*HostServiceSpec, error) {
	return normalizeHostServiceSpecs(specs, pluginID)
}

// normalizeHostServiceSpecs clones declarations before validation mutates them.
func normalizeHostServiceSpecs(specs []*HostServiceSpec, pluginID string) ([]*HostServiceSpec, error) {
	if len(specs) == 0 {
		return []*HostServiceSpec{}, nil
	}
	cloned := make([]*HostServiceSpec, 0, len(specs))
	for _, item := range specs {
		if item == nil {
			continue
		}
		next := &HostServiceSpec{
			Service: normalizeHostServiceName(item.Service),
			Methods: normalizeLowerStringSlice(item.Methods),
			Paths:   normalizePathSliceForService(normalizeHostServiceName(item.Service), item.Paths),
			Tables:  normalizeTableSlice(item.Tables),
			Keys:    normalizeKeySlice(item.Keys),
		}
		if len(item.Resources) > 0 {
			next.Resources = make([]*HostServiceResourceSpec, 0, len(item.Resources))
			for _, resource := range item.Resources {
				if resource == nil {
					continue
				}
				next.Resources = append(next.Resources, &HostServiceResourceSpec{
					Ref:             strings.TrimSpace(resource.Ref),
					AllowMethods:    normalizeUpperStringSlice(resource.AllowMethods),
					HeaderAllowList: normalizeLowerStringSlice(resource.HeaderAllowList),
					TimeoutMs:       resource.TimeoutMs,
					MaxBodyBytes:    resource.MaxBodyBytes,
					Attributes:      normalizeStringMap(resource.Attributes),
				})
			}
		}
		cloned = append(cloned, next)
	}
	if err := validateHostServiceSpecs(cloned, pluginID); err != nil {
		return nil, err
	}
	return cloned, nil
}

// normalizeDeclaredPathForService validates service-specific path resources.
func normalizeDeclaredPathForService(service string, value string) (string, error) {
	if service == HostServiceManifest {
		return normalizeManifestDeclaredPath(value)
	}
	return normalizeStorageDeclaredPath(value)
}

// MustNormalizeHostServiceSpecs returns normalized declarations or panics for
// compile-time constants whose invalid form must fail fast.
func MustNormalizeHostServiceSpecs(specs []*HostServiceSpec) []*HostServiceSpec {
	normalized, err := NormalizeHostServiceSpecs(specs)
	if err != nil {
		panic(err)
	}
	return normalized
}

// MustNormalizeHostServiceSpecsForPlugin returns plugin-aware normalized
// declarations or panics for compile-time constants whose invalid form must fail fast.
func MustNormalizeHostServiceSpecsForPlugin(pluginID string, specs []*HostServiceSpec) []*HostServiceSpec {
	normalized, err := NormalizeHostServiceSpecsForPlugin(pluginID, specs)
	if err != nil {
		panic(err)
	}
	return normalized
}

// validateNetworkURLPattern validates one authorized network URL pattern before
// it is accepted into manifest state.
func validateNetworkURLPattern(rawValue string) error {
	trimmed := strings.TrimSpace(rawValue)
	if trimmed == "" {
		return gerror.New("url cannot be empty")
	}
	if !strings.Contains(trimmed, "://") {
		return gerror.New("url must include a scheme")
	}
	if strings.Contains(trimmed, "?") || strings.Contains(trimmed, "#") {
		return gerror.New("url pattern cannot include query or fragment")
	}
	parsed, err := url.Parse(trimmed)
	if err != nil {
		return gerror.Wrap(err, "failed to parse url pattern")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return gerror.New("url scheme only supports http/https")
	}
	if strings.TrimSpace(parsed.Host) == "" {
		return gerror.New("url is missing host")
	}
	return nil
}

// normalizeStorageDeclaredPath validates one logical storage path and preserves
// trailing-slash semantics for prefix grants.
func normalizeStorageDeclaredPath(value string) (string, error) {
	raw := strings.ReplaceAll(strings.TrimSpace(value), "\\", "/")
	if raw == "" {
		return "", gerror.New("path cannot be empty")
	}
	if strings.HasPrefix(raw, "/") {
		return "", gerror.Newf("path cannot be absolute: %s", value)
	}
	if len(raw) >= 2 && ((raw[0] >= 'A' && raw[0] <= 'Z') || (raw[0] >= 'a' && raw[0] <= 'z')) && raw[1] == ':' {
		return "", gerror.Newf("path cannot contain a host drive prefix: %s", value)
	}

	isPrefix := strings.HasSuffix(raw, "/")
	trimmed := strings.TrimSuffix(raw, "/")
	if trimmed == "" {
		return "", gerror.New("path cannot be empty")
	}

	normalized := path.Clean(trimmed)
	if normalized == "." || normalized == ".." || strings.HasPrefix(normalized, "../") {
		return "", gerror.Newf("path is invalid: %s", value)
	}
	if isPrefix {
		return normalized + "/", nil
	}
	return normalized, nil
}

// normalizeManifestDeclaredPath validates one manifest resource path or glob.
func normalizeManifestDeclaredPath(value string) (string, error) {
	if strings.Contains(strings.TrimSpace(value), "://") {
		return "", gerror.Newf("path cannot be a url: %s", value)
	}
	normalizedPath, err := normalizeStorageDeclaredPath(value)
	if err != nil {
		return "", err
	}
	if normalizedPath == "manifest" || strings.HasPrefix(normalizedPath, "manifest/") {
		return "", gerror.Newf("path must be relative to manifest root: %s", value)
	}
	return normalizedPath, nil
}
