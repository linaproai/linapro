package config

import (
	"context"
)

// UploadConfig holds file upload configuration.
type UploadConfig struct {
	Path    string `json:"path"`    // Upload directory
	MaxSize int64  `json:"maxSize"` // Max file size (MB)
}

// GetUpload reads upload config from configuration file.
func (s *serviceImpl) GetUpload(ctx context.Context) *UploadConfig {
	cfg := &UploadConfig{
		Path:    "temp/upload",
		MaxSize: 10,
	}
	mustScanConfig(ctx, "upload", cfg)
	return cfg
}
