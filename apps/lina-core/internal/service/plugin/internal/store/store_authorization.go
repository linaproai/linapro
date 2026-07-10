// This file manages requested-versus-authorized host service snapshots for
// dynamic plugin releases.

package store

import (
	"context"
	"reflect"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"gopkg.in/yaml.v3"

	pluginv1 "lina-core/api/plugin/v1"
	"lina-core/internal/dao"
	"lina-core/internal/model/do"
	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/pkg/plugin/pluginbridge/protocol"
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
	// Owner is the owner plugin ID for plugin-owned host services.
	Owner string
	// Service is the logical host service identifier.
	Service string
	// Version is the owner capability protocol version for plugin-owned host services.
	Version string
	// Methods optionally narrows the allowed service methods.
	Methods []string
	// Paths lists the confirmed logical storage paths for this service.
	Paths []string
	// ResourceRefs lists the confirmed resource refs for this service.
	ResourceRefs []string
	// Tables lists the confirmed data tables for this service.
	Tables []string
	// Keys lists the confirmed host config keys for this service.
	Keys []string
}

// HasResourceScopedHostServices reports whether any host service declaration
// requires host confirmation because it contains governed paths, resource refs or tables.
func HasResourceScopedHostServices(specs []*protocol.HostServiceSpec) bool {
	for _, spec := range specs {
		if spec == nil {
			continue
		}
		if len(spec.Paths) > 0 || len(spec.Resources) > 0 || len(spec.Tables) > 0 || len(spec.Keys) > 0 {
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
	cacheKey := manifestSnapshotCacheKey(trimmed)
	if cached := s.getCachedManifestSnapshot(cacheKey); cached != nil {
		return cached, nil
	}
	snapshot := &ManifestSnapshot{}
	root := &yaml.Node{}
	if err := yaml.Unmarshal([]byte(trimmed), root); err != nil {
		return nil, gerror.Wrap(err, "parse plugin release manifest_snapshot failed")
	}
	if err := root.Decode(snapshot); err != nil {
		return nil, gerror.Wrap(err, "decode plugin release manifest_snapshot failed")
	}
	if strings.TrimSpace(snapshot.Distribution) == "" {
		return nil, gerror.Newf("plugin release manifest_snapshot distribution is required: %s", snapshot.ID)
	}
	if snapshot.Distribution != pluginv1.PluginDistributionManaged.String() &&
		snapshot.Distribution != pluginv1.PluginDistributionBuiltin.String() {
		return nil, gerror.Newf("plugin release manifest_snapshot distribution only supports managed/builtin: %s", snapshot.ID)
	}
	requestedHostServices, err := protocol.NormalizeHostServiceSpecsForPlugin(snapshot.ID, snapshot.RequestedHostServices)
	if err != nil {
		return nil, gerror.Wrap(err, "parse requested plugin host service snapshot failed")
	}
	authorizedHostServices, err := protocol.NormalizeHostServiceSpecsForPlugin(snapshot.ID, snapshot.AuthorizedHostServices)
	if err != nil {
		return nil, gerror.Wrap(err, "parse authorized plugin host service snapshot failed")
	}
	snapshot.RequestedHostServices = requestedHostServices
	snapshot.AuthorizedHostServices = authorizedHostServices
	s.storeCachedManifestSnapshot(cacheKey, snapshot)
	return cloneManifestSnapshot(snapshot), nil
}

// applyExistingHostServiceAuthorization carries forward an existing release
// authorization only when it was confirmed against the same requested
// hostServices declaration. Same-version dynamic artifacts can be rebuilt with
// new hostServices, so a confirmed snapshot for an older declaration must not
// shadow the current release request.
func applyExistingHostServiceAuthorization(
	snapshot *ManifestSnapshot,
	existingSnapshot *ManifestSnapshot,
) error {
	if snapshot == nil || existingSnapshot == nil {
		return nil
	}
	snapshot.UninstallPurgeStorageData = existingSnapshot.UninstallPurgeStorageData

	matches, err := hostServiceSpecsEqual(snapshot.ID, snapshot.RequestedHostServices, existingSnapshot.RequestedHostServices)
	if err != nil {
		return err
	}
	if !matches {
		return nil
	}

	authorizedHostServices, err := protocol.NormalizeHostServiceSpecsForPlugin(snapshot.ID, existingSnapshot.AuthorizedHostServices)
	if err != nil {
		return err
	}
	snapshot.AuthorizedHostServices = authorizedHostServices
	snapshot.HostServiceAuthConfirmed = existingSnapshot.HostServiceAuthConfirmed
	return nil
}

// hostServiceSpecsEqual compares two host service snapshots after protocol
// normalization so ordering and declaration casing do not affect drift checks.
func hostServiceSpecsEqual(
	pluginID string,
	left []*protocol.HostServiceSpec,
	right []*protocol.HostServiceSpec,
) (bool, error) {
	leftSpecs, err := protocol.NormalizeHostServiceSpecsForPlugin(pluginID, left)
	if err != nil {
		return false, err
	}
	rightSpecs, err := protocol.NormalizeHostServiceSpecsForPlugin(pluginID, right)
	if err != nil {
		return false, err
	}
	return reflect.DeepEqual(leftSpecs, rightSpecs), nil
}

// PersistReleaseHostServiceAuthorization writes the current requested and
// authorized host service snapshot into the matching release row.
func (s *serviceImpl) PersistReleaseHostServiceAuthorization(
	ctx context.Context,
	manifest *catalog.Manifest,
	input *HostServiceAuthorizationInput,
) (*ManifestSnapshot, error) {
	if manifest == nil {
		return nil, gerror.New("plugin manifest cannot be nil")
	}

	release, err := s.GetRelease(ctx, manifest.ID, manifest.Version)
	if err != nil {
		return nil, err
	}
	if release == nil {
		return nil, gerror.Newf("plugin release does not exist: %s@%s", manifest.ID, manifest.Version)
	}

	existingSnapshot, err := s.ParseManifestSnapshot(release.ManifestSnapshot)
	if err != nil {
		return nil, err
	}

	snapshot, err := s.buildManifestSnapshotModel(manifest)
	if err != nil {
		return nil, err
	}
	if existingSnapshot != nil {
		if err = applyExistingHostServiceAuthorization(snapshot, existingSnapshot); err != nil {
			return nil, err
		}
	}

	if !snapshot.HostServiceAuthRequired {
		authorizedHostServices, normalizeErr := protocol.NormalizeHostServiceSpecsForPlugin(manifest.ID, snapshot.RequestedHostServices)
		if normalizeErr != nil {
			return nil, normalizeErr
		}
		snapshot.AuthorizedHostServices = authorizedHostServices
		snapshot.HostServiceAuthConfirmed = false
	} else if input != nil {
		snapshot.AuthorizedHostServices, err = BuildAuthorizedHostServiceSpecsForPlugin(manifest.ID, snapshot.RequestedHostServices, input)
		if err != nil {
			return nil, err
		}
		snapshot.HostServiceAuthConfirmed = true
	} else if !snapshot.HostServiceAuthConfirmed {
		snapshot.AuthorizedHostServices, err = BuildAuthorizedHostServiceSpecsForPlugin(manifest.ID, snapshot.RequestedHostServices, nil)
		if err != nil {
			return nil, err
		}
		snapshot.HostServiceAuthConfirmed = true
	}

	content, err := yaml.Marshal(snapshot)
	if err != nil {
		return nil, gerror.Wrap(err, "build plugin release authorization snapshot failed")
	}

	if _, err = dao.SysPluginRelease.Ctx(ctx).
		Where(do.SysPluginRelease{Id: release.Id}).
		Data(do.SysPluginRelease{ManifestSnapshot: string(content)}).
		Update(); err != nil {
		return nil, err
	}
	s.invalidateReleaseManifestCacheForPlugin(release.PluginId)
	if _, err = s.RefreshStartupReleaseByID(ctx, release.Id); err != nil {
		return nil, err
	}
	return snapshot, nil
}

// BuildAuthorizedHostServiceSpecs applies one host confirmation input onto the
// requested host service declarations and returns the final authorization
// snapshot used by runtime enforcement.
func BuildAuthorizedHostServiceSpecs(
	requested []*protocol.HostServiceSpec,
	input *HostServiceAuthorizationInput,
) ([]*protocol.HostServiceSpec, error) {
	return BuildAuthorizedHostServiceSpecsForPlugin("", requested, input)
}

// BuildAuthorizedHostServiceSpecsForPlugin applies one host confirmation input
// and enforces plugin-owned data tables in the final authorization snapshot.
//
//nolint:cyclop // Authorization merging keeps requested, confirmed, and denied host-service branches explicit.
func BuildAuthorizedHostServiceSpecsForPlugin(
	pluginID string,
	requested []*protocol.HostServiceSpec,
	input *HostServiceAuthorizationInput,
) ([]*protocol.HostServiceSpec, error) {
	requestedSpecs, err := protocol.NormalizeHostServiceSpecsForPlugin(pluginID, requested)
	if err != nil {
		return nil, err
	}
	if len(requestedSpecs) == 0 {
		return []*protocol.HostServiceSpec{}, nil
	}
	if input == nil {
		return requestedSpecs, nil
	}

	type decisionState struct {
		methods      map[string]struct{}
		paths        map[string]struct{}
		resourceRefs map[string]struct{}
		tables       map[string]struct{}
		keys         map[string]struct{}
	}

	serviceMap := make(map[string]*protocol.HostServiceSpec, len(requestedSpecs))
	for _, spec := range requestedSpecs {
		if spec == nil {
			continue
		}
		serviceMap[protocol.HostServiceSpecIdentity(spec)] = spec
	}

	decisionMap := make(map[string]*decisionState, len(input.Services))
	for _, item := range input.Services {
		if item == nil {
			return nil, gerror.New("host service authorization item cannot be nil")
		}
		owner := strings.TrimSpace(strings.ToLower(item.Owner))
		service := strings.TrimSpace(strings.ToLower(item.Service))
		version := strings.TrimSpace(strings.ToLower(item.Version))
		if owner == "" && version != "" {
			return nil, gerror.Newf("host service authorization %s version requires owner", service)
		}
		if owner != "" && version == "" {
			return nil, gerror.Newf("host service authorization %s owner %s requires version", service, owner)
		}
		identity := protocol.HostServiceIdentity(owner, service, version)
		spec, ok := serviceMap[identity]
		if !ok {
			return nil, gerror.Newf(
				"host service authorization contains undeclared service: %s",
				protocol.HostServiceIdentityLabel(owner, service, version),
			)
		}
		if _, exists := decisionMap[identity]; exists {
			return nil, gerror.Newf(
				"host service authorization cannot be declared more than once: %s",
				protocol.HostServiceSpecLabel(spec),
			)
		}
		label := protocol.HostServiceSpecLabel(spec)

		state := &decisionState{
			methods:      make(map[string]struct{}),
			paths:        make(map[string]struct{}),
			resourceRefs: make(map[string]struct{}),
			tables:       make(map[string]struct{}),
			keys:         make(map[string]struct{}),
		}
		for _, method := range item.Methods {
			normalizedMethod := strings.TrimSpace(strings.ToLower(method))
			if normalizedMethod == "" {
				continue
			}
			if !containsString(spec.Methods, normalizedMethod) {
				return nil, gerror.Newf("host service %s authorization contains undeclared method: %s", label, method)
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
				return nil, gerror.Newf("host service %s authorization contains undeclared path: %s", label, declaredPath)
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
				return nil, gerror.Newf("host service %s authorization contains undeclared resourceRef: %s", label, ref)
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
				return nil, gerror.Newf("host service %s authorization contains undeclared table: %s", label, table)
			}
			state.tables[normalizedTable] = struct{}{}
		}
		keySet := buildHostServiceKeySet(spec.Keys)
		for _, key := range item.Keys {
			normalizedKey := strings.TrimSpace(key)
			if normalizedKey == "" {
				continue
			}
			if _, ok = keySet[normalizedKey]; !ok {
				return nil, gerror.Newf("host service %s authorization contains undeclared key: %s", label, key)
			}
			state.keys[normalizedKey] = struct{}{}
		}
		decisionMap[identity] = state
	}

	authorized := make([]*protocol.HostServiceSpec, 0, len(requestedSpecs))
	for _, spec := range requestedSpecs {
		if spec == nil {
			continue
		}
		// Services without governed targets are effectively capability-only and
		// can be copied through directly. Path/resource/table/key-scoped services are
		// included only when the host explicitly keeps some confirmed targets.
		if len(spec.Paths) == 0 && len(spec.Resources) == 0 && len(spec.Tables) == 0 && len(spec.Keys) == 0 {
			authorized = append(authorized, spec)
			continue
		}

		decision, ok := decisionMap[protocol.HostServiceSpecIdentity(spec)]
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
			authorized = append(authorized, &protocol.HostServiceSpec{
				Owner:   spec.Owner,
				Service: spec.Service,
				Version: spec.Version,
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
			authorized = append(authorized, &protocol.HostServiceSpec{
				Owner:   spec.Owner,
				Service: spec.Service,
				Version: spec.Version,
				Methods: methods,
				Tables:  tables,
			})
			continue
		}

		if len(spec.Keys) > 0 {
			keys := filterKeysBySet(spec.Keys, decision.keys)
			if len(keys) == 0 {
				continue
			}
			authorized = append(authorized, &protocol.HostServiceSpec{
				Owner:   spec.Owner,
				Service: spec.Service,
				Version: spec.Version,
				Methods: methods,
				Keys:    keys,
			})
			continue
		}

		resources := filterResourcesBySet(spec.Resources, decision.resourceRefs)
		if len(resources) == 0 {
			continue
		}

		authorized = append(authorized, &protocol.HostServiceSpec{
			Owner:     spec.Owner,
			Service:   spec.Service,
			Version:   spec.Version,
			Methods:   methods,
			Resources: resources,
		})
	}
	return protocol.NormalizeHostServiceSpecsForPlugin(pluginID, authorized)
}

// buildHostServiceKeySet normalizes declared key-scoped authorizations into one lookup set.
func buildHostServiceKeySet(keys []string) map[string]struct{} {
	set := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		normalizedKey := strings.TrimSpace(key)
		if normalizedKey != "" {
			set[normalizedKey] = struct{}{}
		}
	}
	return set
}

// buildHostServicePathSet normalizes declared path-scoped authorizations into
// one lookup set for decision validation.
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

// buildHostServiceResourceSet normalizes declared resource refs into one
// lookup set for authorization validation.
func buildHostServiceResourceSet(resources []*protocol.HostServiceResourceSpec) map[string]struct{} {
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

// buildHostServiceTableSet normalizes declared table-scoped authorizations into
// one lookup set for decision validation.
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

// filterMethodsBySet narrows one ordered method slice to the confirmed set.
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

// filterResourcesBySet narrows resource refs to the confirmed set while
// cloning attribute data for the persisted authorization snapshot.
func filterResourcesBySet(
	resources []*protocol.HostServiceResourceSpec,
	allowed map[string]struct{},
) []*protocol.HostServiceResourceSpec {
	if len(allowed) == 0 {
		return []*protocol.HostServiceResourceSpec{}
	}
	filtered := make([]*protocol.HostServiceResourceSpec, 0, len(resources))
	for _, resource := range resources {
		if resource == nil {
			continue
		}
		if _, ok := allowed[strings.TrimSpace(resource.Ref)]; !ok {
			continue
		}
		filtered = append(filtered, &protocol.HostServiceResourceSpec{
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

// filterPathsBySet narrows declared paths to the confirmed authorization set.
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

// filterTablesBySet narrows declared tables to the confirmed authorization set.
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

// filterKeysBySet narrows one ordered key slice to the confirmed set.
func filterKeysBySet(keys []string, allowed map[string]struct{}) []string {
	if len(allowed) == 0 {
		return []string{}
	}
	filtered := make([]string, 0, len(keys))
	for _, key := range keys {
		if _, ok := allowed[key]; ok {
			filtered = append(filtered, key)
		}
	}
	return filtered
}

// cloneStringMap copies resource attribute maps for safe snapshot reuse.
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

// containsString reports whether target appears in items without normalizing case.
func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
