// This file hydrates plugin backend hook and resource declarations into
// manifest projections during source and dynamic manifest loading.

package catalog

import (
	"path/filepath"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gfile"
	"gopkg.in/yaml.v3"
)

// loadPluginBackendConfig loads plugin-owned hook and resource declarations into the manifest.
func loadPluginBackendConfig(manifest *Manifest) error {
	manifest.Hooks = make([]*HookSpec, 0)
	manifest.BackendResources = make(map[string]*ResourceSpec)

	if manifest.SourcePlugin != nil {
		return nil
	}

	if manifest.RuntimeArtifact != nil {
		manifest.Hooks = CloneHookSpecs(manifest.RuntimeArtifact.HookSpecs)
		manifest.BackendResources = CloneResourceSpecsToMap(manifest.RuntimeArtifact.ResourceSpecs)
		return nil
	}

	hookFiles, err := gfile.ScanDirFile(filepath.Join(manifest.RootDir, "backend", "hooks"), "*.yaml", false)
	if err != nil && !gfile.Exists(filepath.Join(manifest.RootDir, "backend", "hooks")) {
		err = nil
	}
	if err != nil {
		return err
	}
	for _, hookFile := range hookFiles {
		spec := &HookSpec{}
		if err = loadPluginYAMLFile(hookFile, spec); err != nil {
			return err
		}
		if err = ValidateHookSpec(manifest.ID, spec, hookFile); err != nil {
			return err
		}
		manifest.Hooks = append(manifest.Hooks, spec)
	}

	resourceFiles, err := gfile.ScanDirFile(filepath.Join(manifest.RootDir, "backend", "resources"), "*.yaml", false)
	if err != nil && !gfile.Exists(filepath.Join(manifest.RootDir, "backend", "resources")) {
		err = nil
	}
	if err != nil {
		return err
	}
	for _, resourceFile := range resourceFiles {
		spec := &ResourceSpec{}
		if err = loadPluginYAMLFile(resourceFile, spec); err != nil {
			return err
		}
		if err = ValidateResourceSpec(manifest.ID, spec, resourceFile); err != nil {
			return err
		}
		manifest.BackendResources[spec.Key] = spec
	}
	return nil
}

// loadPluginYAMLFile reads a YAML file at filePath and unmarshals it into target.
func loadPluginYAMLFile(filePath string, target interface{}) error {
	content := gfile.GetBytes(filePath)
	if len(content) == 0 {
		return gerror.Newf("plugin configuration file is empty: %s", filePath)
	}
	if err := yaml.Unmarshal(content, target); err != nil {
		return gerror.Wrapf(err, "parse plugin configuration file failed: %s", filePath)
	}
	return nil
}
