// This file provides host-side cleanup helpers for plugin-governed storage
// paths so lifecycle uninstall flows can purge plugin-owned files.

package wasm

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gfile"

	bridgehostservice "lina-core/pkg/pluginbridge/hostservice"
)

// PurgeAuthorizedStoragePaths removes all files under the given plugin's
// authorized storage paths and prunes the plugin storage root when it becomes empty.
func PurgeAuthorizedStoragePaths(
	ctx context.Context,
	pluginID string,
	hostServices []*bridgehostservice.HostServiceSpec,
) error {
	resourceConfig, err := buildStorageResourceConfigForPlugin(ctx, pluginID)
	if err != nil {
		return err
	}

	paths := collectAuthorizedStoragePaths(hostServices)
	if len(paths) == 0 {
		if gfile.Exists(resourceConfig.rootDir) {
			return gfile.Remove(resourceConfig.rootDir)
		}
		return nil
	}

	for _, authorizedPath := range paths {
		if err = purgeAuthorizedStoragePath(resourceConfig, authorizedPath); err != nil {
			return err
		}
	}
	return pruneEmptyStorageRoot(resourceConfig.rootDir)
}

// buildStorageResourceConfigForPlugin resolves the storage root for the given plugin ID.
func buildStorageResourceConfigForPlugin(
	ctx context.Context,
	pluginID string,
) (*storageResourceConfig, error) {
	normalizedPluginID := strings.TrimSpace(pluginID)
	if normalizedPluginID == "" {
		return nil, gerror.New("plugin id cannot be empty")
	}

	rootDir := filepath.Join(
		storageConfigSvc.GetPluginDynamicStoragePath(ctx),
		storageHostServiceRootDirName,
		storageHostServiceDirName,
		normalizedPluginID,
	)
	absoluteRootDir, err := filepath.Abs(rootDir)
	if err != nil {
		return nil, gerror.Wrap(err, "resolve storage resource root directory failed")
	}
	return &storageResourceConfig{
		rootDir:    filepath.Clean(absoluteRootDir),
		visibility: bridgehostservice.HostServiceStorageVisibilityPrivate,
	}, nil
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
	resourceConfig *storageResourceConfig,
	authorizedPath string,
) error {
	if resourceConfig == nil {
		return nil
	}

	normalizedPath, err := normalizeStorageAuthorizedPath(authorizedPath)
	if err != nil {
		return err
	}
	targetPath := strings.TrimSuffix(normalizedPath, "/")
	absolutePath, err := resourceConfig.resolveObjectPath(targetPath)
	if err != nil {
		return err
	}
	if !gfile.Exists(absolutePath) {
		return nil
	}
	return gfile.Remove(absolutePath)
}

// pruneEmptyStorageRoot deletes the plugin storage root once it becomes empty.
func pruneEmptyStorageRoot(rootDir string) error {
	normalizedRoot := strings.TrimSpace(rootDir)
	if normalizedRoot == "" || !gfile.Exists(normalizedRoot) {
		return nil
	}

	entries, err := os.ReadDir(normalizedRoot)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return gfile.Remove(normalizedRoot)
	}
	return nil
}
