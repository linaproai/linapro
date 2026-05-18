// This file implements tenant resolver-chain dispatch for provider adapters.

package tenantcap

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"

	pkgtenantcap "lina-core/pkg/tenantcap"
)

// ResolverChain dispatches tenant resolution through configured resolvers.
type ResolverChain struct {
	resolvers map[pkgtenantcap.ResolverName]pkgtenantcap.Resolver
	order     []pkgtenantcap.ResolverName
}

// NewResolverChain creates a resolver chain with the supplied order.
func NewResolverChain(order []pkgtenantcap.ResolverName, resolvers ...pkgtenantcap.Resolver) *ResolverChain {
	chain := &ResolverChain{
		resolvers: make(map[pkgtenantcap.ResolverName]pkgtenantcap.Resolver, len(resolvers)),
		order:     normalizeResolverOrder(order),
	}
	for _, resolver := range resolvers {
		chain.Register(resolver)
	}
	return chain
}

// Register adds or replaces one resolver by name.
func (c *ResolverChain) Register(resolver pkgtenantcap.Resolver) {
	if c == nil || resolver == nil {
		return
	}
	name := resolver.Name()
	if strings.TrimSpace(string(name)) == "" {
		return
	}
	c.resolvers[name] = resolver
}

// Resolve returns the first resolver result that explicitly matches a tenant.
func (c *ResolverChain) Resolve(ctx context.Context, r *ghttp.Request) (*pkgtenantcap.ResolverResult, error) {
	if c == nil {
		return nil, nil
	}
	for _, name := range c.order {
		resolver := c.resolvers[name]
		if resolver == nil {
			continue
		}
		result, err := resolver.Resolve(ctx, r)
		if err != nil {
			return nil, err
		}
		if result != nil && result.Matched {
			return result, nil
		}
	}
	return nil, nil
}

// normalizeResolverOrder removes blanks and duplicate resolver names while preserving order.
func normalizeResolverOrder(order []pkgtenantcap.ResolverName) []pkgtenantcap.ResolverName {
	if len(order) == 0 {
		return nil
	}
	result := make([]pkgtenantcap.ResolverName, 0, len(order))
	seen := make(map[pkgtenantcap.ResolverName]struct{}, len(order))
	for _, name := range order {
		normalized := pkgtenantcap.ResolverName(strings.TrimSpace(string(name)))
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	return result
}
