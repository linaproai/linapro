// This file hydrates plugin backend hook and resource declarations from
// plugin-root tool configuration or runtime artifacts.

package catalog

import (
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gfile"
	"gopkg.in/yaml.v3"

	"lina-core/pkg/plugin/pluginhost"
)

// LoadPluginBackendConfig loads plugin-owned hook and resource declarations into the manifest.
func LoadPluginBackendConfig(manifest *Manifest) error {
	return loadPluginBackendConfig(manifest)
}

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

	cfg, configPath, err := loadPluginHackConfig(manifest.RootDir)
	if err != nil {
		return err
	}
	for index, rawSpec := range cfg.Wasm.Hooks {
		specLabel := pluginHackConfigItemLabel(configPath, "wasm.hooks", index)
		spec, convertErr := buildHookSpecFromPluginHackConfig(manifest.ID, rawSpec, specLabel)
		if convertErr != nil {
			return convertErr
		}
		if err = ValidateHookSpec(manifest.ID, spec, specLabel); err != nil {
			return err
		}
		manifest.Hooks = append(manifest.Hooks, spec)
	}
	for index, spec := range cfg.Wasm.Resources {
		specLabel := pluginHackConfigItemLabel(configPath, "wasm.resources", index)
		if err = ValidateResourceSpec(manifest.ID, spec, specLabel); err != nil {
			return err
		}
		manifest.BackendResources[spec.Key] = spec
	}
	return nil
}

// pluginHackConfig stores builder-owned plugin tool configuration from
// plugin-root hack/config.yaml.
type pluginHackConfig struct {
	Wasm pluginHackWasmConfig `yaml:"wasm"`
}

// pluginHackWasmConfig stores dynamic WASM builder configuration.
type pluginHackWasmConfig struct {
	Hooks     []*pluginHackWasmHookSpec `yaml:"hooks"`
	Resources []*ResourceSpec           `yaml:"resources"`
}

// pluginHackWasmHookSpec stores hook metadata in human-readable config form.
type pluginHackWasmHookSpec struct {
	Event        pluginhost.ExtensionPoint        `yaml:"event"`
	Action       pluginhost.HookAction            `yaml:"action,omitempty"`
	Mode         pluginhost.CallbackExecutionMode `yaml:"mode,omitempty"`
	Table        string                           `yaml:"table,omitempty"`
	Fields       map[string]string                `yaml:"fields,omitempty"`
	Timeout      string                           `yaml:"timeout,omitempty"`
	Sleep        string                           `yaml:"sleep,omitempty"`
	ErrorMessage string                           `yaml:"errorMessage,omitempty"`
}

// UnmarshalYAML keeps the hook config schema explicit so removed millisecond
// fields cannot be accepted or silently ignored.
func (spec *pluginHackWasmHookSpec) UnmarshalYAML(value *yaml.Node) error {
	if value == nil || value.Kind != yaml.MappingNode {
		return gerror.New("plugin hook config entry must be a mapping")
	}
	allowedFields := map[string]struct{}{
		"event":        {},
		"action":       {},
		"mode":         {},
		"table":        {},
		"fields":       {},
		"timeout":      {},
		"sleep":        {},
		"errorMessage": {},
	}
	for index := 0; index+1 < len(value.Content); index += 2 {
		key := strings.TrimSpace(value.Content[index].Value)
		if _, ok := allowedFields[key]; !ok {
			return gerror.Newf("plugin hook config field is not supported: %s", key)
		}
	}

	type rawPluginHackWasmHookSpec pluginHackWasmHookSpec
	var raw rawPluginHackWasmHookSpec
	if err := value.Decode(&raw); err != nil {
		return err
	}
	*spec = pluginHackWasmHookSpec(raw)
	return nil
}

// loadPluginHackConfig reads the plugin-root development tool configuration.
func loadPluginHackConfig(pluginRoot string) (*pluginHackConfig, string, error) {
	if strings.TrimSpace(pluginRoot) == "" {
		return &pluginHackConfig{}, "", nil
	}
	configPath := filepath.Join(pluginRoot, "hack", "config.yaml")
	if !gfile.Exists(configPath) {
		return &pluginHackConfig{}, configPath, nil
	}

	content := gfile.GetBytes(configPath)
	if len(content) == 0 {
		return nil, configPath, gerror.Newf("plugin configuration file is empty: %s", configPath)
	}
	cfg := &pluginHackConfig{}
	if err := yaml.Unmarshal(content, cfg); err != nil {
		return nil, configPath, gerror.Wrapf(err, "parse plugin configuration file failed: %s", configPath)
	}
	return cfg, configPath, nil
}

// buildHookSpecFromPluginHackConfig converts human-readable hook config into
// the millisecond-based contract used by runtime hook execution.
func buildHookSpecFromPluginHackConfig(
	pluginID string,
	rawSpec *pluginHackWasmHookSpec,
	specLabel string,
) (*HookSpec, error) {
	if rawSpec == nil {
		return nil, gerror.Newf("plugin hook cannot be nil: %s", specLabel)
	}

	spec := &HookSpec{
		Event:        rawSpec.Event,
		Action:       rawSpec.Action,
		Mode:         rawSpec.Mode,
		Table:        rawSpec.Table,
		Fields:       rawSpec.Fields,
		ErrorMessage: rawSpec.ErrorMessage,
	}
	var err error
	spec.TimeoutMs, err = parseOptionalPluginHackDuration(pluginID, specLabel, "timeout", rawSpec.Timeout)
	if err != nil {
		return nil, err
	}
	spec.SleepMs, err = parseOptionalPluginHackDuration(pluginID, specLabel, "sleep", rawSpec.Sleep)
	if err != nil {
		return nil, err
	}
	return spec, nil
}

// parseOptionalPluginHackDuration parses an optional duration string from
// hack/config.yaml into millisecond precision.
func parseOptionalPluginHackDuration(pluginID string, specLabel string, fieldName string, value string) (int, error) {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		return 0, nil
	}
	duration, err := time.ParseDuration(trimmedValue)
	if err != nil {
		return 0, gerror.Wrapf(err, "plugin %s duration must include a valid unit for %s %s", fieldName, pluginID, specLabel)
	}
	if duration <= 0 {
		return 0, gerror.Newf("plugin %s duration must be greater than 0 for %s %s", fieldName, pluginID, specLabel)
	}
	if duration%time.Millisecond != 0 {
		return 0, gerror.Newf("plugin %s duration must use millisecond precision for %s %s", fieldName, pluginID, specLabel)
	}

	durationMs := duration.Milliseconds()
	maxInt := int64(int(^uint(0) >> 1))
	if durationMs > maxInt {
		return 0, gerror.Newf("plugin %s duration is too large for %s %s", fieldName, pluginID, specLabel)
	}
	return int(durationMs), nil
}

func pluginHackConfigItemLabel(configPath string, fieldPath string, index int) string {
	return configPath + " " + fieldPath + "[" + strconv.Itoa(index) + "]"
}
