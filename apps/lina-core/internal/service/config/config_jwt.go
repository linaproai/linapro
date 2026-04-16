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

// GetJwt reads JWT config from configuration file.
func (s *serviceImpl) GetJwt(ctx context.Context) *JwtConfig {
	cfg := &JwtConfig{
		Expire: 24 * time.Hour,
	}
	if secretVar := g.Cfg().MustGet(ctx, "jwt.secret"); secretVar != nil {
		cfg.Secret = secretVar.String()
	}
	cfg.Expire = mustLoadDurationConfig(ctx, "jwt.expire", cfg.Expire)
	return cfg
}
