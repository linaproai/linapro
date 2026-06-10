// This file provides host-side cleanup helpers for plugin-governed storage
// paths so lifecycle uninstall flows can purge plugin-owned files.

package wasm

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/capability/storagecap"
	bridgehostservice "lina-core/pkg/plugin/pluginbridge/protocol"
)

// PurgeAuthorizedStoragePaths removes all objects under the given plugin's
// authorized storage paths through the plugin-scoped storage domain service.
func PurgeAuthorizedStoragePaths(
	ctx context.Context,
	storageSvc storagecap.Service,
	hostServices []*bridgehostservice.HostServiceSpec,
) error {
	if storageSvc == nil {
		return gerror.New("storage cleanup service is not configured")
	}

	paths := collectAuthorizedStoragePaths(hostServices)
	for _, authorizedPath := range paths {
		if err := purgeAuthorizedStoragePath(ctx, storageSvc, authorizedPath); err != nil {
			return err
		}
	}
	return nil
}

// collectAuthorizedStoragePaths collects unique authorized storage paths from host services.
func collectAuthorizedStoragePaths(hostServices []*bridgehostservice.HostServiceSpec) []string {
	seen := make(map[string]struct{})
	paths := make([]string, 0)
	for _, spec := range hostServices {
		if spec == nil || spec.Service != bridgehostservice.HostServiceStorage {
			continue
		}
		for _, item := range spec.Paths {
			normalizedPath, err := normalizeStorageAuthorizedPath(item)
			if err != nil || normalizedPath == "" {
				continue
			}
			if _, ok := seen[normalizedPath]; ok {
				continue
			}
			seen[normalizedPath] = struct{}{}
			paths = append(paths, normalizedPath)
		}
	}
	return paths
}

// purgeAuthorizedStoragePath removes the authorized path or prefix contents from storage.
func purgeAuthorizedStoragePath(
	ctx context.Context,
	storageSvc storagecap.Service,
	authorizedPath string,
) error {
	if storageSvc == nil {
		return nil
	}

	normalizedPath, err := normalizeStorageAuthorizedPath(authorizedPath)
	if err != nil {
		return err
	}
	if !strings.HasSuffix(normalizedPath, "/") {
		return storageSvc.Delete(ctx, storagecap.DeleteInput{Path: normalizedPath})
	}
	for {
		output, err := storageSvc.List(ctx, storagecap.ListInput{
			Prefix: normalizedPath,
			Limit:  storagecap.MaxListLimit,
		})
		if err != nil {
			return err
		}
		if output == nil || len(output.Objects) == 0 {
			return nil
		}
		deleted := 0
		for _, object := range output.Objects {
			if object == nil || strings.TrimSpace(object.Path) == "" {
				continue
			}
			if err = storageSvc.Delete(ctx, storagecap.DeleteInput{Path: object.Path}); err != nil {
				return err
			}
			deleted++
		}
		effectiveLimit := output.Limit
		if effectiveLimit <= 0 {
			effectiveLimit = storagecap.MaxListLimit
		}
		if deleted == 0 || len(output.Objects) < effectiveLimit {
			return nil
		}
	}
}
