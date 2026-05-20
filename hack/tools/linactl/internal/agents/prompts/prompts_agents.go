// This file defines the supported agent registry for the prompts resource.
// Each entry declares a per-agent SourcePath because different agents may
// link different prompt catalogs in the future. The current registry only
// links the OpenSpec /opsx slash command directory, but the structure
// supports adding more entries (or per-agent multiple bindings) later.

package prompts

import (
	"path/filepath"
	"sort"

	"linactl/internal/agents/common"
)

// AgentSpec describes one supported agent's project-level prompts/commands
// binding. It implements common.SpecLike so the resource-agnostic engine
// in the common subpackage can operate on it uniformly.
type AgentSpec struct {
	// Name is the CLI identifier used on the command line (e.g. claude-code).
	Name string
	// DisplayName is the human-readable label rendered in status output.
	DisplayName string
	// SourcePath is the canonical repo-relative source directory the
	// managed symlink must point at. Different agents may use different
	// source directories under .agents/prompts/.
	SourcePath string
	// ProjectPath is the project-relative target directory where the
	// symlink should live (e.g. .claude/commands/opsx).
	ProjectPath string
	// Category indicates how the agent's project path should be handled.
	// Only common.CategoryLink and common.CategoryNative are meaningful
	// for prompts agents; the rootCollision case does not apply because
	// no known prompts agent uses a bare repo-root path.
	Category common.Category
}

// SpecName implements common.SpecLike.
func (s AgentSpec) SpecName() string { return s.Name }

// SpecDisplayName implements common.SpecLike.
func (s AgentSpec) SpecDisplayName() string { return s.DisplayName }

// SpecCategory implements common.SpecLike.
func (s AgentSpec) SpecCategory() common.Category { return s.Category }

// SpecSourcePath implements common.SpecLike. Each prompts agent declares
// its own source path explicitly.
func (s AgentSpec) SpecSourcePath() string { return s.SourcePath }

// SpecProjectPath implements common.SpecLike.
func (s AgentSpec) SpecProjectPath() string { return s.ProjectPath }

// SpecKind implements common.SpecLike. Prompts bindings always manage
// directory-level symlinks.
func (s AgentSpec) SpecKind() common.Kind { return common.KindDir }

// agents is the canonical agent registry. The list is sorted alphabetically
// by Name in init() so callers can rely on stable iteration order.
//
// Initial coverage focuses on the four mainstream agents that have a
// clearly-defined commands/prompts surface:
//   - claude-code: Claude Code reads .claude/commands/<name>/ as slash
//     commands.
//   - cursor: Cursor reads .cursor/commands/<name>/ as slash commands.
//   - codex: OpenAI Codex CLI reads .codex/prompts/<name>/ as named
//     prompts.
//   - gemini-cli: Gemini CLI reads .gemini/commands/<name>/ as slash
//     commands.
//
// All four currently link the OpenSpec /opsx slash command source at
// .agents/prompts/opsx into the agent-specific layout. Additional sources
// can be added later without changing the engine.
var agents = []AgentSpec{
	{
		Name:        "claude-code",
		DisplayName: "Claude Code",
		SourcePath:  ".agents/prompts/opsx",
		ProjectPath: ".claude/commands/opsx",
		Category:    common.CategoryLink,
	},
	{
		Name:        "codex",
		DisplayName: "Codex",
		SourcePath:  ".agents/prompts/opsx",
		ProjectPath: ".codex/prompts/opsx",
		Category:    common.CategoryLink,
	},
	{
		Name:        "cursor",
		DisplayName: "Cursor",
		SourcePath:  ".agents/prompts/opsx",
		ProjectPath: ".cursor/commands/opsx",
		Category:    common.CategoryLink,
	},
	{
		Name:        "gemini-cli",
		DisplayName: "Gemini CLI",
		SourcePath:  ".agents/prompts/opsx",
		ProjectPath: ".gemini/commands/opsx",
		Category:    common.CategoryLink,
	},
}

// init normalizes registry data once at package load: every entry's
// SourcePath and ProjectPath are forced to forward-slash form and the
// list is sorted by Name for stable iteration.
func init() {
	for index := range agents {
		agents[index].SourcePath = filepath.ToSlash(agents[index].SourcePath)
		agents[index].ProjectPath = filepath.ToSlash(agents[index].ProjectPath)
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
