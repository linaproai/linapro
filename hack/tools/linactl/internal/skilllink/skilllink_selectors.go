// This file implements selector parsing and the public Plan/Apply/Unlink
// entry points used by linactl skills.link and skills.unlink commands.

package skilllink

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

// LinkRequest captures one skills.link invocation parameters.
type LinkRequest struct {
	// Selectors is the list of agent names provided by the caller. An empty
	// list means "no selection" and the command should print status only.
	// A list containing "all" expands to every link-class agent.
	Selectors []string
	// Force enables rebuilding mismatched symlinks and creating
	// rootCollision agents. It never affects real directories or files.
	Force bool
}

// UnlinkRequest captures one skills.unlink invocation parameters.
type UnlinkRequest struct {
	// Selectors mirrors LinkRequest.Selectors but applies to unlink flow.
	// Empty selectors means "no selection" and the command should refuse
	// to perform any change.
	Selectors []string
}

// ParseSelectors parses a comma-separated agent selector value.
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

// resolveTargets expands selectors against the agent registry and returns the
// intended target list for an action. The caller policy parameter decides
// which agent categories are eligible when selector contains SelectorAll.
type targetPolicy struct {
	// includeNative includes native-class agents in expansion. They are
	// always reported in status output regardless of this flag.
	includeNative bool
	// includeRootCollision includes rootCollision-class agents. Should be
	// set to true only when force is also true.
	includeRootCollision bool
}

// resolveTargets returns the agents matched by selectors. When selectors
// contains SelectorAll the policy filters apply; otherwise specific agent
// names are looked up directly and missing names are returned as errors.
func resolveTargets(selectors []string, policy targetPolicy) ([]AgentSpec, error) {
	if len(selectors) == 0 {
		return nil, nil
	}
	if hasAll(selectors) {
		out := make([]AgentSpec, 0, len(agents))
		for _, spec := range agents {
			switch spec.Category {
			case CategoryNative:
				if policy.includeNative {
					out = append(out, spec)
				}
			case CategoryLink:
				out = append(out, spec)
			case CategoryRootCollision:
				if policy.includeRootCollision {
					out = append(out, spec)
				}
			}
		}
		return out, nil
	}
	seen := make(map[string]struct{}, len(selectors))
	out := make([]AgentSpec, 0, len(selectors))
	var unknown []string
	for _, name := range selectors {
		if _, exists := seen[name]; exists {
			continue
		}
		seen[name] = struct{}{}
		spec, ok := FindAgent(name)
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
		return out[left].Name < out[right].Name
	})
	return out, nil
}

// hasAll reports whether a selector list contains SelectorAll.
func hasAll(selectors []string) bool {
	return slices.Contains(selectors, SelectorAll)
}
