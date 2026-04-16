// This file exposes root-facade list and manifest synchronization methods.

package plugin

import (
	"context"
	"strings"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/internal/service/plugin/internal/runtime"
)

// SyncSourcePlugins scans source plugin manifests and synchronizes default status.
func (s *serviceImpl) SyncSourcePlugins(ctx context.Context) error {
	_, err := s.SyncAndList(ctx)
	return err
}

// SyncAndList scans plugin manifests, synchronizes plugin registry rows, and
// returns the combined list of source and dynamic plugin items.
func (s *serviceImpl) SyncAndList(ctx context.Context) (*ListOutput, error) {
	manifests, err := s.catalogSvc.ScanManifests()
	if err != nil {
		return nil, err
	}

	covered := make(map[string]struct{}, len(manifests))
	items := make([]*PluginItem, 0, len(manifests))
	for _, manifest := range manifests {
		covered[manifest.ID] = struct{}{}
		registry, syncErr := s.catalogSvc.SyncManifest(ctx, manifest)
		if syncErr != nil {
			return nil, syncErr
		}
		items = append(items, s.runtimeSvc.BuildPluginItem(ctx, manifest, registry))
	}

	runtimeItems, err := s.runtimeSvc.BuildRuntimeItems(ctx, covered)
	if err != nil {
		return nil, err
	}
	items = append(items, runtimeItems...)
	runtime.SortPluginItems(items)
	return &ListOutput{List: items, Total: len(items)}, nil
}

// List returns the plugin list with optional in-memory filtering applied.
func (s *serviceImpl) List(ctx context.Context, in ListInput) (*ListOutput, error) {
	out, err := s.SyncAndList(ctx)
	if err != nil {
		return nil, err
	}
	filtered := make([]*PluginItem, 0, len(out.List))
	for _, item := range out.List {
		if in.ID != "" && !strings.Contains(item.Id, in.ID) {
			continue
		}
		if in.Name != "" && !strings.Contains(item.Name, in.Name) {
			continue
		}
		if in.Type != "" && !matchPluginType(item.Type, in.Type) {
			continue
		}
		if in.Status != nil && item.Enabled != *in.Status {
			continue
		}
		if in.Installed != nil && item.Installed != *in.Installed {
			continue
		}
		filtered = append(filtered, item)
	}
	return &ListOutput{List: filtered, Total: len(filtered)}, nil
}

func matchPluginType(actual string, expected string) bool {
	actualType := catalog.NormalizeType(actual)
	expectedType := catalog.NormalizeType(expected)
	if expectedType == "" {
		return true
	}
	return actualType == expectedType
}
