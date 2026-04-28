// This file defines JWT-related configuration loading and duration migration.

package config

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// JwtConfig holds JWT authentication configuration.
type JwtConfig struct {
	Secret string        `json:"secret"` // JWT secret key
	Expire time.Duration `json:"expire"` // Expire is the JWT token validity duration.
}

// getStaticJwtConfig lazily loads the config-file-backed JWT settings for reuse
// across the process.
func (s *serviceImpl) getStaticJwtConfig(ctx context.Context) *JwtConfig {
	return processStaticConfigCaches.jwt.load(func() *JwtConfig {
		cfg := &JwtConfig{
			Expire: 24 * time.Hour,
		}
		if secretVar := g.Cfg().MustGet(ctx, "jwt.secret"); secretVar != nil {
			cfg.Secret = secretVar.String()
		}
		cfg.Expire = mustLoadDurationConfig(ctx, "jwt.expire", cfg.Expire)
		return cfg
	})
}

// GetJwt reads JWT config from configuration file.
func (s *serviceImpl) GetJwt(ctx context.Context) (*JwtConfig, error) {
	cfg := cloneJwtConfig(s.getStaticJwtConfig(ctx))
	expire, err := s.resolveRuntimeDurationOverride(ctx, RuntimeParamKeyJWTExpire, cfg.Expire)
	if err != nil {
		return nil, err
	}
	cfg.Expire = expire
	return cfg, nil
}

// GetJwtSecret returns the static JWT signing secret loaded from config.yaml.
func (s *serviceImpl) GetJwtSecret(ctx context.Context) string {
	cfg := s.getStaticJwtConfig(ctx)
	if cfg == nil {
		return ""
	}
	return cfg.Secret
}

// GetJwtExpire returns the runtime-effective JWT expiration duration.
func (s *serviceImpl) GetJwtExpire(ctx context.Context) (time.Duration, error) {
	cfg := s.getStaticJwtConfig(ctx)
	if cfg == nil {
		return 24 * time.Hour, nil
	}
	return s.resolveRuntimeDurationOverride(ctx, RuntimeParamKeyJWTExpire, cfg.Expire)
}
