// This file defines upload-related configuration loading and runtime overrides.

package config

import (
	"context"
)

const (
	defaultUploadPath    = "temp/upload"
	defaultUploadMaxSize = int64(10)
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
func (s *serviceImpl) GetUpload(ctx context.Context) *UploadConfig {
	cfg := cloneUploadConfig(s.getStaticUploadConfig(ctx))
	cfg.MaxSize = s.applyRuntimeInt64Override(ctx, RuntimeParamKeyUploadMaxSize, cfg.MaxSize)
	return cfg
}

// GetUploadPath returns the static upload directory loaded from config.yaml.
// File storage initialization and static-file routes can reuse this method to
// avoid allocating a full UploadConfig on every request.
func (s *serviceImpl) GetUploadPath(ctx context.Context) string {
	cfg := s.getStaticUploadConfig(ctx)
	if cfg == nil {
		return defaultUploadPath
	}
	return cfg.Path
}

// GetUploadMaxSize returns the runtime-effective upload size ceiling in MB.
// Upload validation should call this directly so the hot path only reads the
// one field that can change at runtime.
func (s *serviceImpl) GetUploadMaxSize(ctx context.Context) int64 {
	cfg := s.getStaticUploadConfig(ctx)
	if cfg == nil {
		return defaultUploadMaxSize
	}
	return s.applyRuntimeInt64Override(ctx, RuntimeParamKeyUploadMaxSize, cfg.MaxSize)
}
