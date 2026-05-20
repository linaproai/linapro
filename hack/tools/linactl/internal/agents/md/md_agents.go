// This file defines the supported agent registry for the md resource.
// Every entry's SourcePath is fixed to AGENTS.md at the repo root; the
// project-side ProjectPath is the agent-specific guide file name.

package md

import (
	"path/filepath"
	"sort"

	"linactl/internal/agents/common"
)

// SourceFile is the canonical project-spec source file all md bindings
// point at. It lives at the repo root and is shared by every link-class
// agent in the registry.
const SourceFile = "AGENTS.md"

// AgentSpec describes one supported agent's project-level guide file
// binding. It implements common.SpecLike so the resource-agnostic engine
// in the common subpackage can operate on it uniformly.
type AgentSpec struct {
	// Name is the CLI identifier used on the command line (e.g. claude-code).
	Name string
	// DisplayName is the human-readable label rendered in status output.
	DisplayName string
	// ProjectPath is the project-relative target file path where the
	// symlink should live (e.g. CLAUDE.md, GEMINI.md, .junie/guidelines.md).
	// For native agents this field is left empty because they read
	// AGENTS.md directly.
	ProjectPath string
	// Category indicates how the agent should be handled. Only
	// common.CategoryLink and common.CategoryNative are meaningful for
	// md agents.
	Category common.Category
}

// SpecName implements common.SpecLike.
func (s AgentSpec) SpecName() string { return s.Name }

// SpecDisplayName implements common.SpecLike.
func (s AgentSpec) SpecDisplayName() string { return s.DisplayName }

// SpecCategory implements common.SpecLike.
func (s AgentSpec) SpecCategory() common.Category { return s.Category }

// SpecSourcePath implements common.SpecLike. All md bindings link to the
// canonical AGENTS.md file at the repo root.
func (s AgentSpec) SpecSourcePath() string { return SourceFile }

// SpecProjectPath implements common.SpecLike. Native agents return an
// empty string here; the engine never reads it for native agents.
func (s AgentSpec) SpecProjectPath() string { return s.ProjectPath }

// SpecKind implements common.SpecLike. Md bindings always manage
// single-file symlinks.
func (s AgentSpec) SpecKind() common.Kind { return common.KindFile }

// agents is the canonical agent registry covering both link-class agents
// (which need a private guide file symlinked to AGENTS.md) and native-class
// agents (which read AGENTS.md natively and are listed for visibility
// only). The registry is sorted alphabetically by Name in init() so
// callers can rely on stable iteration order.
var agents = []AgentSpec{
	// link (private guide file -> AGENTS.md)
	{Name: "augment", DisplayName: "Augment", ProjectPath: ".augment-guidelines", Category: common.CategoryLink},
	{Name: "claude-code", DisplayName: "Claude Code", ProjectPath: "CLAUDE.md", Category: common.CategoryLink},
	{Name: "continue", DisplayName: "Continue", ProjectPath: ".continuerules", Category: common.CategoryLink},
	{Name: "gemini-cli", DisplayName: "Gemini CLI", ProjectPath: "GEMINI.md", Category: common.CategoryLink},
	{Name: "junie", DisplayName: "Junie", ProjectPath: ".junie/guidelines.md", Category: common.CategoryLink},
	{Name: "qwen-code", DisplayName: "Qwen Code", ProjectPath: "QWEN.md", Category: common.CategoryLink},
	{Name: "roo", DisplayName: "Roo Code", ProjectPath: ".roo/rules/AGENTS.md", Category: common.CategoryLink},
	{Name: "windsurf", DisplayName: "Windsurf", ProjectPath: ".windsurfrules", Category: common.CategoryLink},

	// native (agent reads AGENTS.md natively at repo root; nothing to
	// link). Listed for visibility in status output only.
	{Name: "amp", DisplayName: "Amp", Category: common.CategoryNative},
	{Name: "antigravity", DisplayName: "Antigravity", Category: common.CategoryNative},
	{Name: "cline", DisplayName: "Cline", Category: common.CategoryNative},
	{Name: "codex", DisplayName: "Codex", Category: common.CategoryNative},
	{Name: "cursor", DisplayName: "Cursor", Category: common.CategoryNative},
	{Name: "deepagents", DisplayName: "Deep Agents", Category: common.CategoryNative},
	{Name: "dexto", DisplayName: "Dexto", Category: common.CategoryNative},
	{Name: "firebender", DisplayName: "Firebender", Category: common.CategoryNative},
	{Name: "github-copilot", DisplayName: "GitHub Copilot", Category: common.CategoryNative},
	{Name: "kimi-cli", DisplayName: "Kimi Code CLI", Category: common.CategoryNative},
	{Name: "opencode", DisplayName: "OpenCode", Category: common.CategoryNative},
	{Name: "replit", DisplayName: "Replit", Category: common.CategoryNative},
	{Name: "universal", DisplayName: "Universal", Category: common.CategoryNative},
	{Name: "warp", DisplayName: "Warp", Category: common.CategoryNative},
}

// init normalizes registry data once at package load: ProjectPath is
// forced to forward-slash form (only meaningful for link-class entries
// that span subdirectories) and the list is sorted by Name.
func init() {
	for index := range agents {
		if agents[index].ProjectPath != "" {
			agents[index].ProjectPath = filepath.ToSlash(agents[index].ProjectPath)
		}
	}
	sort.Slice(agents, func(left, right int) bool {
		return agents[left].Name < agents[right].Name
	})
}

// Agents returns a defensive copy of the supported agent registry sorted
// by agent name. Callers must not mutate the returned slice.
func Agents() []AgentSpec {
	out := make([]AgentSpec, len(agents))
	copy(out, agents)
	return out
}

// FindAgent returns the AgentSpec for the given name, or false if not found.
func FindAgent(name string) (AgentSpec, bool) {
	for _, spec := range agents {
		if spec.Name == name {
			return spec, true
		}
	}
	return AgentSpec{}, false
}
