// This file defines upload-related configuration loading and runtime overrides.

package config

import (
	"context"
	"time"

	"lina-core/pkg/bizerr"
)

// Upload config defaults used when config.yaml omits the upload section.
const (
	defaultUploadPath    = "temp/upload"
	defaultUploadMaxSize = int64(100)
	// defaultUploadDirectUrlTTL is the default for sys.upload.directUrlTTL.
	defaultUploadDirectUrlTTL = time.Hour
	// defaultUploadDirectUrlTTLText is the sys_config seed / runtime-param text.
	defaultUploadDirectUrlTTLText = "1h"
	// maxUploadDirectUrlTTL is the maximum allowed direct access lifetime.
	maxUploadDirectUrlTTL = time.Hour
)

// UploadConfig holds file upload configuration.
type UploadConfig struct {
	Path         string        `json:"path"`         // Upload directory
	MaxSize      int64         `json:"maxSize"`      // Max file size (MB)
	DirectUrlTTL time.Duration `json:"directUrlTTL"` // Client direct access lifetime
}

// getStaticUploadConfig lazily loads the config-file-backed upload settings so
// static consumers can reuse one parsed object across the whole process.
func (s *serviceImpl) getStaticUploadConfig(ctx context.Context) *UploadConfig {
	return processStaticConfigCaches.upload.load(func() *UploadConfig {
		cfg := &UploadConfig{
			Path:         defaultUploadPath,
			MaxSize:      defaultUploadMaxSize,
			DirectUrlTTL: defaultUploadDirectUrlTTL,
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
	ttl, err := s.GetUploadDirectUrlTTL(ctx)
	if err != nil {
		return nil, err
	}
	cfg.DirectUrlTTL = ttl
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

// GetUploadDirectUrlTTL returns the runtime-effective lifetime for client
// direct object-storage access (presigned put/get and file-center direct-upload
// sessions). Values use Go duration text (for example 30m, 1h). The default is
// 1h; values must be positive and at most maxUploadDirectUrlTTL.
func (s *serviceImpl) GetUploadDirectUrlTTL(ctx context.Context) (time.Duration, error) {
	ttl, err := s.resolveRuntimeDurationOverride(ctx, RuntimeParamKeyUploadDirectUrlTTL, defaultUploadDirectUrlTTL)
	if err != nil {
		return 0, err
	}
	if ttl <= 0 || ttl > maxUploadDirectUrlTTL {
		return defaultUploadDirectUrlTTL, nil
	}
	return ttl, nil
}

// validateUploadDirectUrlTTLConfigValue validates sys.upload.directUrlTTL.
func validateUploadDirectUrlTTLConfigValue(key string, value string) error {
	duration, err := validatePositiveDurationValue(key, value)
	if err != nil {
		return err
	}
	if duration < time.Second {
		return bizerr.NewCode(CodeConfigParamDurationInvalid, bizerr.P("key", key))
	}
	if duration > maxUploadDirectUrlTTL {
		return bizerr.NewCode(
			CodeConfigParamDurationInvalid,
			bizerr.P("key", key),
		)
	}
	// Reject fractional-second values so signed URL TTL stays second-aligned.
	if duration%time.Second != 0 {
		return bizerr.NewCode(CodeConfigParamDurationInvalid, bizerr.P("key", key))
	}
	return nil
}
