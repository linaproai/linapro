// This file implements host-service manifest declaration validation and normalization.

package hostservice

import (
	"net/url"
	"path"
	"sort"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
)

// ValidateHostServiceSpecs validates and normalizes host service declarations in-place.
func ValidateHostServiceSpecs(specs []*HostServiceSpec) error {
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

		if len(spec.Methods) == 0 {
			spec.Methods = defaultHostServiceMethods(spec.Service)
		}
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

		pathSeen := make(map[string]struct{}, len(spec.Paths))
		paths := make([]string, 0, len(spec.Paths))
		for _, rawPath := range spec.Paths {
			normalizedPath, err := normalizeStorageDeclaredPath(rawPath)
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
			if len(spec.Resources) > 0 {
				return gerror.Newf("host service %s cannot declare resources", spec.Service)
			}
			if len(spec.Tables) == 0 {
				return gerror.Newf("host service %s must declare at least one table", spec.Service)
			}
			continue
		}
		if len(spec.Tables) > 0 {
			return gerror.Newf("host service %s cannot declare tables", spec.Service)
		}
		if len(spec.Paths) > 0 {
			return gerror.Newf("host service %s cannot declare paths", spec.Service)
		}

		if _, ok := hostServicesWithoutResources[spec.Service]; ok {
			if len(spec.Resources) > 0 {
				return gerror.Newf("host service %s cannot declare resources", spec.Service)
			}
			continue
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

// defaultHostServiceMethods returns service-specific default method grants.
func defaultHostServiceMethods(service string) []string {
	if service == HostServiceConfig {
		return []string{
			HostServiceMethodConfigGet,
			HostServiceMethodConfigExists,
			HostServiceMethodConfigString,
			HostServiceMethodConfigBool,
			HostServiceMethodConfigInt,
			HostServiceMethodConfigDuration,
		}
	}
	return nil
}

// NormalizeHostServiceSpecs returns deep-cloned and normalized host service declarations.
func NormalizeHostServiceSpecs(specs []*HostServiceSpec) ([]*HostServiceSpec, error) {
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
			Paths:   normalizeStoragePathSlice(item.Paths),
			Tables:  normalizeTableSlice(item.Tables),
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
	if err := ValidateHostServiceSpecs(cloned); err != nil {
		return nil, err
	}
	return cloned, nil
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
