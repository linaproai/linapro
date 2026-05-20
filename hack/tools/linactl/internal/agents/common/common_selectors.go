// This file implements selector parsing and target resolution against a
// generic agent registry. ResolveTargets is parameterized over the
// concrete AgentSpec type so callers receive their own type back without
// runtime assertions; a SpecLike constraint lets the resolver consult
// Name and Category fields uniformly.

package common

import (
	"fmt"
	"slices"
	"sort"
	"strings"
)

// SelectorAll is the special selector value that targets every link-class
// agent. native and rootCollision agents are skipped by default with this
// selector; rootCollision agents only execute when force=true is also set.
const SelectorAll = "all"

// ParseSelectors parses a comma-separated agent selector value. Empty
// values and whitespace-only tokens are dropped. An empty input yields a
// nil slice.
func ParseSelectors(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		token := strings.TrimSpace(part)
		if token == "" {
			continue
		}
		out = append(out, token)
	}
	return out
}

// TargetPolicy controls which agent categories an "all" selector expands
// to. Specific agent names always match regardless of policy; the policy
// only affects the special "all" expansion.
type TargetPolicy struct {
	// IncludeNative includes native-class agents in expansion. They are
	// always reported in status output regardless of this flag.
	IncludeNative bool
	// IncludeRootCollision includes rootCollision-class agents. Should be
	// set to true only when force is also true.
	IncludeRootCollision bool
}

// ResolveTargets returns the agents in the registry matched by selectors.
// When selectors contains SelectorAll the policy filters apply; otherwise
// specific agent names are looked up and missing names are returned as a
// single error listing every unknown name in input order.
//
// The result is sorted by SpecName for stable rendering.
func ResolveTargets[S SpecLike](selectors []string, registry []S, policy TargetPolicy) ([]S, error) {
	if len(selectors) == 0 {
		return nil, nil
	}
	if hasAll(selectors) {
		out := make([]S, 0, len(registry))
		for _, spec := range registry {
			switch spec.SpecCategory() {
			case CategoryNative:
				if policy.IncludeNative {
					out = append(out, spec)
				}
			case CategoryLink:
				out = append(out, spec)
			case CategoryRootCollision:
				if policy.IncludeRootCollision {
					out = append(out, spec)
				}
			}
		}
		return out, nil
	}
	byName := make(map[string]S, len(registry))
	for _, spec := range registry {
		byName[spec.SpecName()] = spec
	}
	seen := make(map[string]struct{}, len(selectors))
	out := make([]S, 0, len(selectors))
	var unknown []string
	for _, name := range selectors {
		if _, exists := seen[name]; exists {
			continue
		}
		seen[name] = struct{}{}
		spec, ok := byName[name]
		if !ok {
			unknown = append(unknown, name)
			continue
		}
		out = append(out, spec)
	}
	if len(unknown) > 0 {
		return nil, fmt.Errorf("unknown agent(s): %s", strings.Join(unknown, ", "))
	}
	sort.Slice(out, func(left, right int) bool {
		return out[left].SpecName() < out[right].SpecName()
	})
	return out, nil
}

// hasAll reports whether a selector list contains SelectorAll.
func hasAll(selectors []string) bool {
	return slices.Contains(selectors, SelectorAll)
}
