// This file exposes runtime login parameters managed through sys_config.

package config

import (
	"context"
	"net"
	"strings"
)

// LoginConfig holds runtime login controls.
type LoginConfig struct {
	BlackIPList []string `json:"blackIPList"` // BlackIPList contains blocked source IPs or CIDR ranges.
}

// GetLogin reads runtime login parameters from sys_config.
func (s *serviceImpl) GetLogin(ctx context.Context) *LoginConfig {
	cfg := &LoginConfig{}
	snapshot := s.getRuntimeParamSnapshot(ctx)
	if snapshot == nil {
		return cfg
	}
	cfg.BlackIPList = snapshot.loginBlacklist()
	return cfg
}

// IsLoginIPBlacklisted reports whether the input IP is denied by the
// runtime-effective blacklist snapshot. Auth hot paths should call this method
// directly so they avoid allocating a LoginConfig and reparsing blacklist rules.
func (s *serviceImpl) IsLoginIPBlacklisted(ctx context.Context, ip string) bool {
	snapshot := s.getRuntimeParamSnapshot(ctx)
	if snapshot == nil {
		return false
	}
	return snapshot.isLoginIPBlacklisted(ip)
}

// IsBlacklisted reports whether the input IP matches one configured deny rule.
func (c *LoginConfig) IsBlacklisted(ip string) bool {
	if c == nil {
		return false
	}
	return newLoginBlacklistMatcher(c.BlackIPList).matches(ip)
}

// loginBlacklistMatcher stores one parsed blacklist snapshot so request hot
// paths can reuse exact-IP and CIDR parsing results across many checks.
type loginBlacklistMatcher struct {
	exact []net.IP
	cidrs []*net.IPNet
}

// newLoginBlacklistMatcher parses configured blacklist rules once and drops
// invalid entries defensively because validation happens before values reach
// sys_config, while stale historical data should not break request handling.
func newLoginBlacklistMatcher(values []string) *loginBlacklistMatcher {
	if len(values) == 0 {
		return nil
	}

	matcher := &loginBlacklistMatcher{
		exact: make([]net.IP, 0, len(values)),
		cidrs: make([]*net.IPNet, 0, len(values)),
	}
	for _, item := range values {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		if direct := net.ParseIP(trimmed); direct != nil {
			matcher.exact = append(matcher.exact, direct)
			continue
		}
		_, ipNet, err := net.ParseCIDR(trimmed)
		if err != nil || ipNet == nil {
			continue
		}
		matcher.cidrs = append(matcher.cidrs, ipNet)
	}
	if len(matcher.exact) == 0 && len(matcher.cidrs) == 0 {
		return nil
	}
	return matcher
}

// matches checks one client IP against the pre-parsed direct IPs and CIDR
// ranges. The caller still parses the incoming IP per request, but configured
// rules are no longer reparsed on every login attempt.
func (m *loginBlacklistMatcher) matches(ip string) bool {
	if m == nil {
		return false
	}

	parsedIP := net.ParseIP(strings.TrimSpace(ip))
	if parsedIP == nil {
		return false
	}
	for _, direct := range m.exact {
		if direct.Equal(parsedIP) {
			return true
		}
	}
	for _, ipNet := range m.cidrs {
		if ipNet.Contains(parsedIP) {
			return true
		}
	}
	return false
}
