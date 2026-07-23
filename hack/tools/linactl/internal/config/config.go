// Package config loads repository-level linactl defaults from hack/config.yaml.
// It centralizes the small YAML schema shared by build, image, and plugin
// commands so internal components do not depend on the main package.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Origin type identifiers for named plugin origins.
const (
	OriginTypeGit         = "git"
	OriginTypeMarketplace = "marketplace"
)

// Root stores repository-level tool configuration from hack/config.yaml.
type Root struct {
	Build   Build         `yaml:"build"`
	Image   Image         `yaml:"image"`
	Plugins PluginOrigins `yaml:"plugins"`
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

// PluginOrigins maps user-defined origin names to Git or marketplace origins.
// Names such as "official" or "public-market" are labels only.
type PluginOrigins map[string]PluginOrigin

// PluginOrigin stores one named plugin download origin.
//
// Git origins are repository-level: repo + root + ref apply to the whole
// checkout; items only list which plugin directories to copy.
// Marketplace origins use url; each item carries its own release version.
// Type is required (git | marketplace).
type PluginOrigin struct {
	Type string `yaml:"type"`
	// Repo is the Git repository URL for type=git.
	Repo string `yaml:"repo"`
	// Root is the path inside the Git repo that contains plugin directories.
	Root string `yaml:"root"`
	// Ref is the Git branch, tag, or commit for type=git (repository-level).
	Ref string `yaml:"ref"`
	// URL is the marketplace service base URL for type=marketplace.
	URL string `yaml:"url"`
	// Items lists plugins to install from this origin (mapping form only).
	Items []PluginItem `yaml:"items"`
}

// PluginItem stores one plugin id and optional marketplace version.
// Version is required for marketplace items and forbidden for git items.
type PluginItem struct {
	ID      string `yaml:"id"`
	Version string `yaml:"version"`
}

// UnmarshalYAML accepts only a mapping with id (and optional version).
func (item *PluginItem) UnmarshalYAML(value *yaml.Node) error {
	if value == nil {
		return fmt.Errorf("plugin item is empty")
	}
	if value.Kind != yaml.MappingNode {
		return fmt.Errorf("plugin item must be a mapping {id, version?}, not a scalar string")
	}
	var raw struct {
		ID      string `yaml:"id"`
		Version string `yaml:"version"`
	}
	if err := value.Decode(&raw); err != nil {
		return err
	}
	item.ID = strings.TrimSpace(raw.ID)
	item.Version = strings.TrimSpace(raw.Version)
	if item.ID == "" {
		return fmt.Errorf("plugin item id is required")
	}
	return nil
}

// ResolvedType validates and returns the required origin type.
func (o PluginOrigin) ResolvedType() (string, error) {
	explicit := strings.ToLower(strings.TrimSpace(o.Type))
	repo := strings.TrimSpace(o.Repo)
	url := strings.TrimSpace(o.URL)
	switch explicit {
	case OriginTypeGit:
		if repo == "" {
			return "", fmt.Errorf("git origin requires repo")
		}
		if url != "" {
			return "", fmt.Errorf("git origin must not set url")
		}
		return OriginTypeGit, nil
	case OriginTypeMarketplace:
		if url == "" {
			return "", fmt.Errorf("marketplace origin requires url")
		}
		if repo != "" {
			return "", fmt.Errorf("marketplace origin must not set repo")
		}
		return OriginTypeMarketplace, nil
	case "":
		return "", fmt.Errorf("origin type is required (git or marketplace)")
	default:
		return "", fmt.Errorf("unsupported origin type %q (want git or marketplace)", o.Type)
	}
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
		Plugins: PluginOrigins{},
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
	if cfg.Plugins == nil {
		cfg.Plugins = PluginOrigins{}
	}
	return cfg, nil
}
