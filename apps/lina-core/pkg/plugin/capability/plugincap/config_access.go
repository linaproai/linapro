// This file contains the concrete GoFrame configuration adapter operations for
// source plugins. It keeps lookup, scan, and typed conversion behavior outside
// the package entrypoint while preserving the published config contract.

package plugincap

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gcfg"
	"github.com/gogf/gf/v2/os/gfile"
)

// Get returns the raw configuration value for the given key.
func (s *serviceAdapter) Get(ctx context.Context, key string) (*gvar.Var, error) {
	normalizedKey, err := normalizeConfigKey(key)
	if err != nil {
		return nil, err
	}
	cfg, source, err := s.resolveConfig(ctx)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		return nil, nil
	}
	value, err := cfg.Get(ctx, normalizedKey)
	if err != nil {
		return nil, gerror.Wrapf(err, "read plugin config key failed plugin=%s source=%s key=%s", s.pluginID, source, normalizedKey)
	}
	return value, nil
}

// Exists reports whether the given configuration key exists.
func (s *serviceAdapter) Exists(ctx context.Context, key string) (bool, error) {
	value, err := s.Get(ctx, key)
	if err != nil {
		return false, err
	}
	return !isMissing(value), nil
}

// Scan scans the configuration section into target.
func (s *serviceAdapter) Scan(ctx context.Context, key string, target any) error {
	if target == nil {
		return gerror.New("plugin config scan target cannot be nil")
	}

	value, err := s.Get(ctx, key)
	if err != nil {
		return err
	}
	if isMissing(value) {
		return nil
	}
	if err := value.Scan(target); err != nil {
		return gerror.Wrapf(err, "scan plugin config %s failed", key)
	}
	return nil
}

// String reads a string value or returns defaultValue when the key is absent or blank.
func (s *serviceAdapter) String(ctx context.Context, key string, defaultValue string) (string, error) {
	value, err := s.Get(ctx, key)
	if err != nil {
		return "", err
	}
	if isMissing(value) {
		return defaultValue, nil
	}

	raw := value.String()
	if strings.TrimSpace(raw) == "" {
		return defaultValue, nil
	}
	return raw, nil
}

// Bool reads a bool value or returns defaultValue when the key is absent.
func (s *serviceAdapter) Bool(ctx context.Context, key string, defaultValue bool) (bool, error) {
	value, err := s.Get(ctx, key)
	if err != nil {
		return false, err
	}
	if isMissing(value) {
		return defaultValue, nil
	}
	return value.Bool(), nil
}

// Int reads an int value or returns defaultValue when the key is absent.
func (s *serviceAdapter) Int(ctx context.Context, key string, defaultValue int) (int, error) {
	value, err := s.Get(ctx, key)
	if err != nil {
		return 0, err
	}
	if isMissing(value) {
		return defaultValue, nil
	}
	return value.Int(), nil
}

// Duration reads a time.Duration value from a duration string or returns defaultValue when the key is absent or blank.
func (s *serviceAdapter) Duration(ctx context.Context, key string, defaultValue time.Duration) (time.Duration, error) {
	value, err := s.Get(ctx, key)
	if err != nil {
		return 0, err
	}
	if isMissing(value) {
		return defaultValue, nil
	}

	raw := strings.TrimSpace(value.String())
	if raw == "" {
		return defaultValue, nil
	}
	duration, err := time.ParseDuration(raw)
	if err != nil {
		return 0, gerror.Wrapf(err, "parse plugin config %s duration failed", key)
	}
	return duration, nil
}

// isMissing reports whether a GoFrame config lookup returned no concrete value.
func isMissing(value *gvar.Var) bool {
	return value == nil || value.IsNil()
}

// normalizeConfigKey validates and normalizes plugin-local config keys.
func normalizeConfigKey(key string) (string, error) {
	normalized := strings.TrimSpace(key)
	if normalized == "" || normalized == "." {
		return "", gerror.New("plugin config key cannot be empty or root")
	}
	if strings.HasPrefix(normalized, ".") || strings.HasSuffix(normalized, ".") {
		return "", gerror.Newf("plugin config key is invalid: %s", key)
	}
	return normalized, nil
}

// resolveConfig returns the first configured plugin-scoped config source.
func (s *serviceAdapter) resolveConfig(ctx context.Context) (*gcfg.Config, string, error) {
	if s == nil || strings.TrimSpace(s.pluginID) == "" {
		return nil, "", gerror.New("plugin config service requires plugin scope")
	}
	for _, candidate := range s.fileCandidates() {
		if strings.TrimSpace(candidate.path) == "" || !gfile.Exists(candidate.path) {
			continue
		}
		cfg, err := buildConfigFromFile(candidate.path)
		if err != nil {
			return nil, "", gerror.Wrapf(err, "create plugin config reader failed plugin=%s source=%s", s.pluginID, candidate.kind)
		}
		return cfg, candidate.kind, nil
	}
	if content := s.artifactConfigContent(); len(content) > 0 {
		cfg, err := buildConfigFromContent(content)
		if err != nil {
			return nil, "", gerror.Wrapf(err, "create artifact plugin config reader failed plugin=%s", s.pluginID)
		}
		return cfg, "artifact", nil
	}
	return nil, "", nil
}

// configFileCandidate records one scoped runtime config file candidate.
type configFileCandidate struct {
	kind string
	path string
}

// fileCandidates returns production before development candidates.
func (s *serviceAdapter) fileCandidates() []configFileCandidate {
	candidates := make([]configFileCandidate, 0, 2)
	if root := resolveProductionConfigRoot(s.productionRoot); root != "" {
		candidates = append(candidates, configFileCandidate{
			kind: "production",
			path: filepath.Join(root, "plugins", s.pluginID, RuntimeConfigFileName),
		})
	}
	if root := resolveDevelopmentRoot(s.developmentRoot); root != "" {
		candidates = append(candidates, configFileCandidate{
			kind: "development",
			path: filepath.Join(root, "apps", "lina-plugins", s.pluginID, "manifest", "config", RuntimeConfigFileName),
		})
	}
	return candidates
}

// artifactConfigContent returns the configured release-bound default config.
func (s *serviceAdapter) artifactConfigContent() []byte {
	if s == nil || len(s.artifactConfigs) == 0 {
		return nil
	}
	content := s.artifactConfigs[strings.TrimSpace(s.pluginID)]
	if len(content) == 0 {
		return nil
	}
	return append([]byte(nil), content...)
}

// resolveProductionConfigRoot resolves the GoFrame config root used for
// production plugin config discovery.
func resolveProductionConfigRoot(override string) string {
	if cleaned := cleanFilesystemPath(override); cleaned != "" {
		return cleaned
	}
	if envPath := cleanFilesystemPath(os.Getenv("GF_GCFG_PATH")); envPath != "" {
		return envPath
	}
	workingDir, err := os.Getwd()
	if err == nil {
		if repoRoot, rootErr := findRepositoryRoot(workingDir); rootErr == nil {
			return filepath.Join(repoRoot, "apps", "lina-core", "manifest", "config")
		}
	}
	if selfDir := cleanFilesystemPath(gfile.SelfDir()); selfDir != "" {
		return filepath.Join(selfDir, "config")
	}
	return ""
}

// resolveDevelopmentRoot resolves the repository root used for source plugin
// development config discovery.
func resolveDevelopmentRoot(override string) string {
	if cleaned := cleanFilesystemPath(override); cleaned != "" {
		return cleaned
	}
	workingDir, err := os.Getwd()
	if err != nil {
		return ""
	}
	repoRoot, err := findRepositoryRoot(workingDir)
	if err != nil {
		return ""
	}
	return repoRoot
}

// cleanFilesystemPath trims and cleans a filesystem path without making it
// absolute.
func cleanFilesystemPath(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	return filepath.Clean(trimmed)
}

// findRepositoryRoot locates the LinaPro repository root from startDir.
func findRepositoryRoot(startDir string) (string, error) {
	current, err := filepath.Abs(startDir)
	if err != nil {
		return "", err
	}
	for depth := 0; depth < 12; depth++ {
		if isRepositoryRoot(current) {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return "", gerror.Newf("repository root not found from %s", startDir)
}

// isRepositoryRoot reports whether dir contains LinaPro repository markers.
func isRepositoryRoot(dir string) bool {
	if strings.TrimSpace(dir) == "" {
		return false
	}
	if gfile.Exists(filepath.Join(dir, "go.work")) && gfile.Exists(filepath.Join(dir, "apps", "lina-core")) {
		return true
	}
	return gfile.Exists(filepath.Join(dir, "apps", "lina-core", "go.mod")) &&
		gfile.Exists(filepath.Join(dir, "apps", "lina-plugins"))
}
