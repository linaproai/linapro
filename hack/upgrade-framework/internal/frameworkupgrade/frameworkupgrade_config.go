// This file loads development-only upgrade metadata from hack/config.yaml and
// configures GoFrame to use the same file for database access.

package frameworkupgrade

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"
	"gopkg.in/yaml.v3"
)

// Relative config paths used by the development-only upgrade tool.
const (
	hackConfigRelativePath = "apps/lina-core/hack/config.yaml"
)

// projectConfig stores only the hack-config sections needed by the upgrade tool.
type projectConfig struct {
	FrameworkUpgrade UpgradeConfig `yaml:"frameworkUpgrade"`
}

// UpgradeConfig stores the framework-upgrade metadata declared in hack/config.yaml.
type UpgradeConfig struct {
	Version       string `yaml:"version"`       // Version stores the current framework version for upgrade comparison.
	RepositoryURL string `yaml:"repositoryUrl"` // RepositoryURL stores the default upstream framework repository URL.
}

// ConfigureGoFrameConfig binds GoFrame global config loading to hack/config.yaml
// so development-only upgrade SQL execution never reads runtime config files.
func ConfigureGoFrameConfig(repoRoot string) error {
	configPath := filepath.Join(repoRoot, hackConfigRelativePath)
	adapter, err := gcfg.NewAdapterFile(configPath)
	if err != nil {
		return gerror.Wrapf(err, "初始化 hack 配置适配器失败: %s", configPath)
	}
	g.Cfg().SetAdapter(adapter)
	return nil
}

// readCurrentUpgradeMetadata reads upgrade metadata from the current project hack config.
func readCurrentUpgradeMetadata(repoRoot string) (UpgradeConfig, error) {
	configPath := filepath.Join(repoRoot, hackConfigRelativePath)
	config, err := readUpgradeMetadataFile(configPath)
	if err != nil {
		return UpgradeConfig{}, err
	}
	return *config, nil
}

// readTargetUpgradeMetadata reads upgrade metadata from the target release hack config.
func readTargetUpgradeMetadata(targetCloneDir string) (*UpgradeConfig, error) {
	configPath := filepath.Join(targetCloneDir, hackConfigRelativePath)
	return readUpgradeMetadataFile(configPath)
}

// readUpgradeMetadataFile reads one hack/config.yaml file and returns the frameworkUpgrade section.
func readUpgradeMetadataFile(configPath string) (*UpgradeConfig, error) {
	content, err := os.ReadFile(configPath)
	if err != nil {
		return nil, gerror.Wrapf(err, "读取升级配置失败: %s", configPath)
	}

	cfg := &projectConfig{}
	if err = yaml.Unmarshal(content, cfg); err != nil {
		return nil, gerror.Wrap(err, "解析 hack/config.yaml 失败")
	}
	cfg.FrameworkUpgrade.Version = strings.TrimSpace(cfg.FrameworkUpgrade.Version)
	cfg.FrameworkUpgrade.RepositoryURL = strings.TrimSpace(cfg.FrameworkUpgrade.RepositoryURL)
	if cfg.FrameworkUpgrade.Version == "" {
		return nil, gerror.New("hack/config.yaml 缺少 frameworkUpgrade.version")
	}
	if _, err = parseSemanticVersion(cfg.FrameworkUpgrade.Version); err != nil {
		return nil, err
	}
	return &cfg.FrameworkUpgrade, nil
}
