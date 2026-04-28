// This file defines upload-related configuration loading and runtime overrides.

package config

import (
	"context"
)

// Upload config defaults used when config.yaml omits the upload section.
const (
	defaultUploadPath    = "temp/upload"
	defaultUploadMaxSize = int64(20)
)

// UploadConfig holds file upload configuration.
type UploadConfig struct {
	Path    string `json:"path"`    // Upload directory
	MaxSize int64  `json:"maxSize"` // Max file size (MB)
}

// getStaticUploadConfig lazily loads the config-file-backed upload settings so
// static consumers can reuse one parsed object across the whole process.
func (s *serviceImpl) getStaticUploadConfig(ctx context.Context) *UploadConfig {
	return processStaticConfigCaches.upload.load(func() *UploadConfig {
		cfg := &UploadConfig{
			Path:    defaultUploadPath,
			MaxSize: defaultUploadMaxSize,
		}
		mustScanConfig(ctx, "upload", cfg)
		return cfg
	})
}

// GetUpload reads upload config from configuration file.
func (s *serviceImpl) GetUpload(ctx context.Context) (*UploadConfig, error) {
	cfg := cloneUploadConfig(s.getStaticUploadConfig(ctx))
	maxSize, err := s.resolveRuntimeInt64Override(ctx, RuntimeParamKeyUploadMaxSize, cfg.MaxSize)
	if err != nil {
		return nil, err
	}
	cfg.MaxSize = maxSize
	return cfg, nil
}

// GetUploadPath returns the runtime-resolved static upload directory loaded
// from config.yaml. Relative paths are anchored at the repository root when the
// host runs from a LinaPro checkout.
func (s *serviceImpl) GetUploadPath(ctx context.Context) string {
	cfg := s.getStaticUploadConfig(ctx)
	if cfg == nil {
		return resolveRuntimePathWithDefault("", defaultUploadPath)
	}
	return resolveRuntimePathWithDefault(cfg.Path, defaultUploadPath)
}

// GetUploadMaxSize returns the runtime-effective upload size ceiling in MB.
// Upload validation should call this directly so the hot path only reads the
// one field that can change at runtime.
func (s *serviceImpl) GetUploadMaxSize(ctx context.Context) (int64, error) {
	cfg := s.getStaticUploadConfig(ctx)
	if cfg == nil {
		return defaultUploadMaxSize, nil
	}
	return s.resolveRuntimeInt64Override(ctx, RuntimeParamKeyUploadMaxSize, cfg.MaxSize)
}
