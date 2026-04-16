// This file manages requested-versus-authorized host service snapshots for
// dynamic plugin releases.

package catalog

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"gopkg.in/yaml.v3"

	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/pkg/pluginbridge"
)

// HostServiceAuthorizationInput describes the host-confirmed authorization
// result submitted during install or enable flows.
type HostServiceAuthorizationInput struct {
	// Services narrows one or more resource-scoped host service declarations.
	Services []*HostServiceAuthorizationDecision
}

// HostServiceAuthorizationDecision describes the confirmed methods and
// resource refs for one logical host service.
type HostServiceAuthorizationDecision struct {
	// Service is the logical host service identifier.
	Service string
	// Methods optionally narrows the allowed service methods.
	Methods []string
	// Paths lists the confirmed logical storage paths for this service.
	Paths []string
	// ResourceRefs lists the confirmed resource refs for this service.
	ResourceRefs []string
	// Tables lists the confirmed data tables for this service.
	Tables []string
}

// HasResourceScopedHostServices reports whether any host service declaration
// requires host confirmation because it contains governed paths, resource refs or tables.
func HasResourceScopedHostServices(specs []*pluginbridge.HostServiceSpec) bool {
	for _, spec := range specs {
		if spec == nil {
			continue
		}
		if len(spec.Paths) > 0 || len(spec.Resources) > 0 || len(spec.Tables) > 0 {
			return true
		}
	}
	return false
}

// ParseManifestSnapshot unmarshals one persisted release manifest snapshot.
func (s *serviceImpl) ParseManifestSnapshot(content string) (*ManifestSnapshot, error) {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return nil, nil
	}
	snapshot := &ManifestSnapshot{}
	if err := yaml.Unmarshal([]byte(trimmed), snapshot); err != nil {
		return nil, gerror.Wrap(err, "解析插件 release manifest_snapshot 失败")
	}
	snapshot.RequestedHostServices = pluginbridge.NormalizeHostServiceSpecs(snapshot.RequestedHostServices)
	snapshot.AuthorizedHostServices = pluginbridge.NormalizeHostServiceSpecs(snapshot.AuthorizedHostServices)
	return snapshot, nil
}

// PersistReleaseHostServiceAuthorization writes the current requested and
// authorized host service snapshot into the matching release row.
func (s *serviceImpl) PersistReleaseHostServiceAuthorization(
	ctx context.Context,
	manifest *Manifest,
	input *HostServiceAuthorizationInput,
) (*ManifestSnapshot, error) {
	if manifest == nil {
		return nil, gerror.New("插件清单不能为空")
	}

	release, err := s.GetRelease(ctx, manifest.ID, manifest.Version)
	if err != nil {
		return nil, err
	}
	if release == nil {
		return nil, gerror.Newf("插件 release 不存在: %s@%s", manifest.ID, manifest.Version)
	}

	existingSnapshot, err := s.ParseManifestSnapshot(release.ManifestSnapshot)
	if err != nil {
		return nil, err
	}

	snapshot := s.buildManifestSnapshotModel(manifest)
	if existingSnapshot != nil {
		snapshot.HostServiceAuthConfirmed = existingSnapshot.HostServiceAuthConfirmed
		snapshot.AuthorizedHostServices = pluginbridge.NormalizeHostServiceSpecs(existingSnapshot.AuthorizedHostServices)
	}

	if !snapshot.HostServiceAuthRequired {
		snapshot.AuthorizedHostServices = pluginbridge.NormalizeHostServiceSpecs(snapshot.RequestedHostServices)
		snapshot.HostServiceAuthConfirmed = false
	} else if input != nil {
		snapshot.AuthorizedHostServices, err = BuildAuthorizedHostServiceSpecs(snapshot.RequestedHostServices, input)
		if err != nil {
			return nil, err
		}
		snapshot.HostServiceAuthConfirmed = true
	}

	content, err := yaml.Marshal(snapshot)
	if err != nil {
		return nil, gerror.Wrap(err, "生成插件 release 授权快照失败")
	}

	if _, err = dao.SysPluginRelease.Ctx(ctx).
		Where(do.SysPluginRelease{Id: release.Id}).
		Data(do.SysPluginRelease{ManifestSnapshot: string(content)}).
		Update(); err != nil {
		return nil, err
	}
	return snapshot, nil
}

// BuildAuthorizedHostServiceSpecs applies one host confirmation input onto the
// requested host service declarations and returns the final authorization
// snapshot used by runtime enforcement.
func BuildAuthorizedHostServiceSpecs(
	requested []*pluginbridge.HostServiceSpec,
	input *HostServiceAuthorizationInput,
) ([]*pluginbridge.HostServiceSpec, error) {
	requestedSpecs := pluginbridge.NormalizeHostServiceSpecs(requested)
	if len(requestedSpecs) == 0 {
		return []*pluginbridge.HostServiceSpec{}, nil
	}
	if input == nil {
		return requestedSpecs, nil
	}

	type decisionState struct {
		methods      map[string]struct{}
		paths        map[string]struct{}
		resourceRefs map[string]struct{}
		tables       map[string]struct{}
	}

	serviceMap := make(map[string]*pluginbridge.HostServiceSpec, len(requestedSpecs))
	for _, spec := range requestedSpecs {
		if spec == nil {
			continue
		}
		serviceMap[spec.Service] = spec
	}

	decisionMap := make(map[string]*decisionState, len(input.Services))
	for _, item := range input.Services {
		if item == nil {
			return nil, gerror.New("宿主服务授权项不能为空")
		}
		service := strings.TrimSpace(strings.ToLower(item.Service))
		spec, ok := serviceMap[service]
		if !ok {
			return nil, gerror.Newf("宿主服务授权包含未声明的 service: %s", item.Service)
		}

		state := &decisionState{
			methods:      make(map[string]struct{}),
			paths:        make(map[string]struct{}),
			resourceRefs: make(map[string]struct{}),
			tables:       make(map[string]struct{}),
		}
		for _, method := range item.Methods {
			normalizedMethod := strings.TrimSpace(strings.ToLower(method))
			if normalizedMethod == "" {
				continue
			}
			if !containsString(spec.Methods, normalizedMethod) {
				return nil, gerror.Newf("宿主服务 %s 授权包含未声明的 method: %s", service, method)
			}
			state.methods[normalizedMethod] = struct{}{}
		}

		pathSet := buildHostServicePathSet(spec.Paths)
		for _, declaredPath := range item.Paths {
			normalizedPath := strings.TrimSpace(declaredPath)
			if normalizedPath == "" {
				continue
			}
			if _, ok = pathSet[normalizedPath]; !ok {
				return nil, gerror.Newf("宿主服务 %s 授权包含未声明的 path: %s", service, declaredPath)
			}
			state.paths[normalizedPath] = struct{}{}
		}

		resourceSet := buildHostServiceResourceSet(spec.Resources)
		for _, ref := range item.ResourceRefs {
			normalizedRef := strings.TrimSpace(ref)
			if normalizedRef == "" {
				continue
			}
			if _, ok = resourceSet[normalizedRef]; !ok {
				return nil, gerror.Newf("宿主服务 %s 授权包含未声明的 resourceRef: %s", service, ref)
			}
			state.resourceRefs[normalizedRef] = struct{}{}
		}
		tableSet := buildHostServiceTableSet(spec.Tables)
		for _, table := range item.Tables {
			normalizedTable := strings.TrimSpace(table)
			if normalizedTable == "" {
				continue
			}
			if _, ok = tableSet[normalizedTable]; !ok {
				return nil, gerror.Newf("宿主服务 %s 授权包含未声明的 table: %s", service, table)
			}
			state.tables[normalizedTable] = struct{}{}
		}
		decisionMap[service] = state
	}

	authorized := make([]*pluginbridge.HostServiceSpec, 0, len(requestedSpecs))
	for _, spec := range requestedSpecs {
		if spec == nil {
			continue
		}
		// Services without governed targets are effectively capability-only and
		// can be copied through directly. Path/resource/table-scoped services are
		// included only when the host explicitly keeps some confirmed targets.
		if len(spec.Paths) == 0 && len(spec.Resources) == 0 && len(spec.Tables) == 0 {
			authorized = append(authorized, spec)
			continue
		}

		decision, ok := decisionMap[spec.Service]
		if !ok {
			continue
		}

		methods := spec.Methods
		if len(decision.methods) > 0 {
			methods = filterMethodsBySet(spec.Methods, decision.methods)
		}
		if len(methods) == 0 {
			continue
		}

		if len(spec.Paths) > 0 {
			paths := filterPathsBySet(spec.Paths, decision.paths)
			if len(paths) == 0 {
				continue
			}
			authorized = append(authorized, &pluginbridge.HostServiceSpec{
				Service: spec.Service,
				Methods: methods,
				Paths:   paths,
			})
			continue
		}

		if len(spec.Tables) > 0 {
			tables := filterTablesBySet(spec.Tables, decision.tables)
			if len(tables) == 0 {
				continue
			}
			authorized = append(authorized, &pluginbridge.HostServiceSpec{
				Service: spec.Service,
				Methods: methods,
				Tables:  tables,
			})
			continue
		}

		resources := filterResourcesBySet(spec.Resources, decision.resourceRefs)
		if len(resources) == 0 {
			continue
		}

		authorized = append(authorized, &pluginbridge.HostServiceSpec{
			Service:   spec.Service,
			Methods:   methods,
			Resources: resources,
		})
	}
	return pluginbridge.NormalizeHostServiceSpecs(authorized), nil
}

func buildHostServicePathSet(paths []string) map[string]struct{} {
	set := make(map[string]struct{}, len(paths))
	for _, item := range paths {
		normalizedPath := strings.TrimSpace(item)
		if normalizedPath != "" {
			set[normalizedPath] = struct{}{}
		}
	}
	return set
}

func buildHostServiceResourceSet(resources []*pluginbridge.HostServiceResourceSpec) map[string]struct{} {
	set := make(map[string]struct{}, len(resources))
	for _, resource := range resources {
		if resource == nil {
			continue
		}
		ref := strings.TrimSpace(resource.Ref)
		if ref != "" {
			set[ref] = struct{}{}
		}
	}
	return set
}

func buildHostServiceTableSet(tables []string) map[string]struct{} {
	set := make(map[string]struct{}, len(tables))
	for _, table := range tables {
		normalizedTable := strings.TrimSpace(table)
		if normalizedTable != "" {
			set[normalizedTable] = struct{}{}
		}
	}
	return set
}

func filterMethodsBySet(methods []string, allowed map[string]struct{}) []string {
	if len(allowed) == 0 {
		return []string{}
	}
	filtered := make([]string, 0, len(methods))
	for _, method := range methods {
		if _, ok := allowed[method]; ok {
			filtered = append(filtered, method)
		}
	}
	return filtered
}

func filterResourcesBySet(
	resources []*pluginbridge.HostServiceResourceSpec,
	allowed map[string]struct{},
) []*pluginbridge.HostServiceResourceSpec {
	if len(allowed) == 0 {
		return []*pluginbridge.HostServiceResourceSpec{}
	}
	filtered := make([]*pluginbridge.HostServiceResourceSpec, 0, len(resources))
	for _, resource := range resources {
		if resource == nil {
			continue
		}
		if _, ok := allowed[strings.TrimSpace(resource.Ref)]; !ok {
			continue
		}
		filtered = append(filtered, &pluginbridge.HostServiceResourceSpec{
			Ref:             resource.Ref,
			AllowMethods:    append([]string(nil), resource.AllowMethods...),
			HeaderAllowList: append([]string(nil), resource.HeaderAllowList...),
			TimeoutMs:       resource.TimeoutMs,
			MaxBodyBytes:    resource.MaxBodyBytes,
			Attributes:      cloneStringMap(resource.Attributes),
		})
	}
	return filtered
}

func filterPathsBySet(paths []string, allowed map[string]struct{}) []string {
	if len(allowed) == 0 {
		return []string{}
	}
	filtered := make([]string, 0, len(paths))
	for _, item := range paths {
		normalizedPath := strings.TrimSpace(item)
		if normalizedPath == "" {
			continue
		}
		if _, ok := allowed[normalizedPath]; ok {
			filtered = append(filtered, normalizedPath)
		}
	}
	return filtered
}

func filterTablesBySet(tables []string, allowed map[string]struct{}) []string {
	if len(allowed) == 0 {
		return []string{}
	}
	filtered := make([]string, 0, len(tables))
	for _, table := range tables {
		normalizedTable := strings.TrimSpace(table)
		if normalizedTable == "" {
			continue
		}
		if _, ok := allowed[normalizedTable]; ok {
			filtered = append(filtered, normalizedTable)
		}
	}
	return filtered
}

func cloneStringMap(source map[string]string) map[string]string {
	if len(source) == 0 {
		return nil
	}
	target := make(map[string]string, len(source))
	for key, value := range source {
		target[key] = value
	}
	return target
}

func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
