// This file resolves embedded GoFrame code generation targets. It keeps the
// GoFrame working directory separate from the config directory so plugin-level
// tool configuration can live outside the plugin backend module.

package goframecli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Target describes one resolved GoFrame code generation target.
type Target struct {
	WorkDir   string
	ConfigDir string
}

// TargetOptions describes one requested GoFrame code generation target.
type TargetOptions struct {
	Dir           string
	DirSet        bool
	RequireConfig bool
}

// ResolveTarget resolves the backend working directory and GoFrame config
// directory for code generation.
func ResolveTarget(root string, options TargetOptions) (Target, error) {
	if options.DirSet && strings.TrimSpace(options.Dir) == "" {
		return Target{}, fmt.Errorf("GoFrame code generation parameter dir is empty")
	}

	target := filepath.Join("apps", "lina-core")
	if options.DirSet {
		target = options.Dir
	}

	workDir, err := normalizeTargetDir(root, target)
	if err != nil {
		return Target{}, err
	}
	if err := ValidateProjectDir(workDir); err != nil {
		return Target{}, err
	}

	configDir := filepath.Join(workDir, "hack")
	if isStandardPluginBackend(root, workDir) {
		configDir = filepath.Join(filepath.Dir(workDir), "hack")
	}
	targetInfo := Target{
		WorkDir:   workDir,
		ConfigDir: configDir,
	}
	if options.RequireConfig {
		if err := ValidateConfigDir(configDir); err != nil {
			return Target{}, err
		}
	}
	return targetInfo, nil
}

// ValidateProjectDir checks that a code generation target is a directory.
func ValidateProjectDir(targetDir string) error {
	if targetDir == "" {
		return fmt.Errorf("GoFrame code generation target directory is empty")
	}
	info, err := os.Stat(targetDir)
	if err != nil {
		return fmt.Errorf("check GoFrame code generation target %s: %w", targetDir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("GoFrame code generation target %s is not a directory", targetDir)
	}
	return nil
}

// ValidateConfigDir checks that a config directory has a GoFrame CLI config.
func ValidateConfigDir(configDir string) error {
	if configDir == "" {
		return fmt.Errorf("GoFrame code generation config directory is empty")
	}
	if info, err := os.Stat(configDir); err == nil {
		if !info.IsDir() {
			return fmt.Errorf("GoFrame code generation config directory %s is not a directory", configDir)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("check GoFrame code generation config directory %s: %w", configDir, err)
	}
	configPath := filepath.Join(configDir, "config.yaml")
	configInfo, err := os.Stat(configPath)
	if err != nil {
		return fmt.Errorf("GoFrame code generation config directory %s is missing config.yaml: %w", configDir, err)
	}
	if configInfo.IsDir() {
		return fmt.Errorf("GoFrame code generation config directory %s has config.yaml as a directory", configDir)
	}
	return nil
}

func normalizeTargetDir(root string, target string) (string, error) {
	target = strings.TrimSpace(target)
	if target == "" {
		return "", fmt.Errorf("GoFrame code generation target directory is empty")
	}
	if !filepath.IsAbs(target) {
		target = filepath.Join(root, filepath.FromSlash(target))
	}
	absolute, err := filepath.Abs(target)
	if err != nil {
		return "", fmt.Errorf("resolve GoFrame code generation target %s: %w", target, err)
	}
	return absolute, nil
}

func isStandardPluginBackend(root string, workDir string) bool {
	if filepath.Base(workDir) != "backend" {
		return false
	}
	pluginRoot := filepath.Dir(workDir)
	pluginsRoot := filepath.Join(root, "apps", "lina-plugins")
	relative, err := filepath.Rel(pluginsRoot, pluginRoot)
	if err != nil || relative == "." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) || relative == ".." || filepath.IsAbs(relative) {
		return false
	}
	if strings.Contains(relative, string(filepath.Separator)) {
		return false
	}
	manifestPath := filepath.Join(pluginRoot, "plugin.yaml")
	manifestID, err := readPluginID(manifestPath)
	if err != nil {
		return false
	}
	return manifestID == filepath.Base(pluginRoot)
}

func readPluginID(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	var root yaml.Node
	if err = yaml.Unmarshal(content, &root); err != nil {
		return "", err
	}
	if len(root.Content) == 0 {
		return "", nil
	}
	doc := root.Content[0]
	if doc.Kind != yaml.MappingNode {
		return "", nil
	}
	for index := 0; index+1 < len(doc.Content); index += 2 {
		if doc.Content[index].Value == "id" {
			return strings.TrimSpace(doc.Content[index+1].Value), nil
		}
	}
	return "", nil
}
