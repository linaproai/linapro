// Package config loads repository-level linactl defaults from hack/config.yaml.
// It centralizes the small YAML schema shared by build, image, and plugin
// commands so internal components do not depend on the main package.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Root stores repository-level tool configuration from hack/config.yaml.
type Root struct {
	Build   Build   `yaml:"build"`
	Image   Image   `yaml:"image"`
	Plugins Plugins `yaml:"plugins"`
}

// Build stores user-facing build defaults.
type Build struct {
	Platforms  []string `yaml:"platforms"`
	CGOEnabled bool     `yaml:"cgoEnabled"`
	OutputDir  string   `yaml:"outputDir"`
	BinaryName string   `yaml:"binaryName"`
}

// Image stores user-facing image metadata defaults.
type Image struct {
	Name       string `yaml:"name"`
	Tag        string `yaml:"tag"`
	Registry   string `yaml:"registry"`
	Push       bool   `yaml:"push"`
	BaseImage  string `yaml:"baseImage"`
	Dockerfile string `yaml:"dockerfile"`
}

// Plugins stores source-plugin workspace management configuration.
type Plugins struct {
	Sources map[string]PluginSource `yaml:"sources"`
}

// PluginSource stores one configured plugin source repository.
type PluginSource struct {
	Repo  string   `yaml:"repo"`
	Root  string   `yaml:"root"`
	Ref   string   `yaml:"ref"`
	Items []string `yaml:"items"`
}

// Load reads repository tool defaults from hack/config.yaml or an override.
func Load(root string, configPath string) (Root, error) {
	cfg := Root{
		Build: Build{
			Platforms:  []string{"auto"},
			CGOEnabled: false,
			OutputDir:  filepath.Join("temp", "output"),
			BinaryName: "lina",
		},
		Image: Image{
			Name:       "linapro",
			BaseImage:  "alpine:3.22",
			Dockerfile: filepath.Join("hack", "docker", "Dockerfile"),
		},
	}

	if configPath == "" {
		configPath = filepath.Join("hack", "config.yaml")
	}
	if !filepath.IsAbs(configPath) {
		configPath = filepath.Join(root, configPath)
	}
	content, err := os.ReadFile(configPath)
	if err != nil {
		return cfg, fmt.Errorf("read config %s: %w", configPath, err)
	}
	if err = yaml.Unmarshal(content, &cfg); err != nil {
		return cfg, fmt.Errorf("parse config %s: %w", configPath, err)
	}
	return cfg, nil
}
