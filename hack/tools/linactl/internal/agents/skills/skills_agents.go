// This file defines the supported agent registry and AgentSpec type for
// the skills subpackage. The registry is hand-maintained to mirror the
// project paths published in
// https://github.com/vercel-labs/skills#supported-agents and classifies
// each agent into native, link or rootCollision categories.

package skills

import (
	"path/filepath"
	"sort"

	"linactl/internal/agents/common"
)

// SourceDir is the canonical skills source directory relative to the
// repository root. All managed symlinks point at this directory.
const SourceDir = ".agents/skills"

// Category aliases the resource-agnostic category enum defined in the
// common subpackage so existing call sites keep using skills.CategoryLink
// etc. without churn.
type Category = common.Category

// Re-exported category constants. New code should prefer common.Category*
// directly; these aliases exist so the skills subpackage's external surface
// keeps working during the multi-resource refactor.
const (
	CategoryNative        = common.CategoryNative
	CategoryLink          = common.CategoryLink
	CategoryRootCollision = common.CategoryRootCollision
)

// AgentSpec describes one supported agent's project-level skill location.
// It implements common.SpecLike so the resource-agnostic engine in the
// common subpackage can operate on it uniformly.
type AgentSpec struct {
	// Name is the CLI identifier used on the command line (e.g. claude-code).
	Name string
	// DisplayName is the human-readable label rendered in status output.
	DisplayName string
	// SourcePath is the canonical source path the managed symlink points
	// at. For skills agents this is always SourceDir; the field is kept
	// explicit so the engine reads it via SpecSourcePath() like every
	// other resource.
	SourcePath string
	// ProjectPath is the project-relative skills directory path.
	ProjectPath string
	// Category indicates how the agent's project path should be handled.
	Category common.Category
}

// SpecName implements common.SpecLike.
func (s AgentSpec) SpecName() string { return s.Name }

// SpecDisplayName implements common.SpecLike.
func (s AgentSpec) SpecDisplayName() string { return s.DisplayName }

// SpecCategory implements common.SpecLike.
func (s AgentSpec) SpecCategory() common.Category { return s.Category }

// SpecSourcePath implements common.SpecLike. Skills agents always link to
// SourceDir; the field defaults to SourceDir during registry init.
func (s AgentSpec) SpecSourcePath() string { return s.SourcePath }

// SpecProjectPath implements common.SpecLike.
func (s AgentSpec) SpecProjectPath() string { return s.ProjectPath }

// SpecKind implements common.SpecLike. Skills bindings always manage
// directory-level symlinks.
func (s AgentSpec) SpecKind() common.Kind { return common.KindDir }

// agents is the canonical agent registry. The list is sorted alphabetically
// by Name in init() so callers can rely on stable iteration order.
var agents = []AgentSpec{
	// native (project path == .agents/skills)
	{Name: "amp", DisplayName: "Amp", ProjectPath: ".agents/skills", Category: common.CategoryNative},
	{Name: "antigravity", DisplayName: "Antigravity", ProjectPath: ".agents/skills", Category: common.CategoryNative},
	{Name: "cline", DisplayName: "Cline", ProjectPath: ".agents/skills", Category: common.CategoryNative},
	{Name: "codex", DisplayName: "Codex", ProjectPath: ".agents/skills", Category: common.CategoryNative},
	{Name: "cursor", DisplayName: "Cursor", ProjectPath: ".agents/skills", Category: common.CategoryNative},
	{Name: "deepagents", DisplayName: "Deep Agents", ProjectPath: ".agents/skills", Category: common.CategoryNative},
	{Name: "dexto", DisplayName: "Dexto", ProjectPath: ".agents/skills", Category: common.CategoryNative},
	{Name: "firebender", DisplayName: "Firebender", ProjectPath: ".agents/skills", Category: common.CategoryNative},
	{Name: "gemini-cli", DisplayName: "Gemini CLI", ProjectPath: ".agents/skills", Category: common.CategoryNative},
	{Name: "github-copilot", DisplayName: "GitHub Copilot", ProjectPath: ".agents/skills", Category: common.CategoryNative},
	{Name: "kimi-cli", DisplayName: "Kimi Code CLI", ProjectPath: ".agents/skills", Category: common.CategoryNative},
	{Name: "opencode", DisplayName: "OpenCode", ProjectPath: ".agents/skills", Category: common.CategoryNative},
	{Name: "replit", DisplayName: "Replit", ProjectPath: ".agents/skills", Category: common.CategoryNative},
	{Name: "universal", DisplayName: "Universal", ProjectPath: ".agents/skills", Category: common.CategoryNative},
	{Name: "warp", DisplayName: "Warp", ProjectPath: ".agents/skills", Category: common.CategoryNative},

	// link (project path differs from SourceDir, not at repo root)
	{Name: "adal", DisplayName: "AdaL", ProjectPath: ".adal/skills", Category: common.CategoryLink},
	{Name: "aider-desk", DisplayName: "AiderDesk", ProjectPath: ".aider-desk/skills", Category: common.CategoryLink},
	{Name: "augment", DisplayName: "Augment", ProjectPath: ".augment/skills", Category: common.CategoryLink},
	{Name: "bob", DisplayName: "IBM Bob", ProjectPath: ".bob/skills", Category: common.CategoryLink},
	{Name: "claude-code", DisplayName: "Claude Code", ProjectPath: ".claude/skills", Category: common.CategoryLink},
	{Name: "codearts-agent", DisplayName: "CodeArts Agent", ProjectPath: ".codeartsdoer/skills", Category: common.CategoryLink},
	{Name: "codebuddy", DisplayName: "CodeBuddy", ProjectPath: ".codebuddy/skills", Category: common.CategoryLink},
	{Name: "codemaker", DisplayName: "Codemaker", ProjectPath: ".codemaker/skills", Category: common.CategoryLink},
	{Name: "codestudio", DisplayName: "Code Studio", ProjectPath: ".codestudio/skills", Category: common.CategoryLink},
	{Name: "command-code", DisplayName: "Command Code", ProjectPath: ".commandcode/skills", Category: common.CategoryLink},
	{Name: "continue", DisplayName: "Continue", ProjectPath: ".continue/skills", Category: common.CategoryLink},
	{Name: "cortex", DisplayName: "Cortex Code", ProjectPath: ".cortex/skills", Category: common.CategoryLink},
	{Name: "crush", DisplayName: "Crush", ProjectPath: ".crush/skills", Category: common.CategoryLink},
	{Name: "devin", DisplayName: "Devin for Terminal", ProjectPath: ".devin/skills", Category: common.CategoryLink},
	{Name: "droid", DisplayName: "Droid", ProjectPath: ".factory/skills", Category: common.CategoryLink},
	{Name: "forgecode", DisplayName: "ForgeCode", ProjectPath: ".forge/skills", Category: common.CategoryLink},
	{Name: "goose", DisplayName: "Goose", ProjectPath: ".goose/skills", Category: common.CategoryLink},
	{Name: "hermes-agent", DisplayName: "Hermes Agent", ProjectPath: ".hermes/skills", Category: common.CategoryLink},
	{Name: "iflow-cli", DisplayName: "iFlow CLI", ProjectPath: ".iflow/skills", Category: common.CategoryLink},
	{Name: "junie", DisplayName: "Junie", ProjectPath: ".junie/skills", Category: common.CategoryLink},
	{Name: "kilo", DisplayName: "Kilo Code", ProjectPath: ".kilocode/skills", Category: common.CategoryLink},
	{Name: "kiro-cli", DisplayName: "Kiro CLI", ProjectPath: ".kiro/skills", Category: common.CategoryLink},
	{Name: "kode", DisplayName: "Kode", ProjectPath: ".kode/skills", Category: common.CategoryLink},
	{Name: "mcpjam", DisplayName: "MCPJam", ProjectPath: ".mcpjam/skills", Category: common.CategoryLink},
	{Name: "mistral-vibe", DisplayName: "Mistral Vibe", ProjectPath: ".vibe/skills", Category: common.CategoryLink},
	{Name: "mux", DisplayName: "Mux", ProjectPath: ".mux/skills", Category: common.CategoryLink},
	{Name: "neovate", DisplayName: "Neovate", ProjectPath: ".neovate/skills", Category: common.CategoryLink},
	{Name: "openhands", DisplayName: "OpenHands", ProjectPath: ".openhands/skills", Category: common.CategoryLink},
	{Name: "pi", DisplayName: "Pi", ProjectPath: ".pi/skills", Category: common.CategoryLink},
	{Name: "pochi", DisplayName: "Pochi", ProjectPath: ".pochi/skills", Category: common.CategoryLink},
	{Name: "qoder", DisplayName: "Qoder", ProjectPath: ".qoder/skills", Category: common.CategoryLink},
	{Name: "qwen-code", DisplayName: "Qwen Code", ProjectPath: ".qwen/skills", Category: common.CategoryLink},
	{Name: "roo", DisplayName: "Roo Code", ProjectPath: ".roo/skills", Category: common.CategoryLink},
	{Name: "rovodev", DisplayName: "Rovo Dev", ProjectPath: ".rovodev/skills", Category: common.CategoryLink},
	{Name: "tabnine-cli", DisplayName: "Tabnine CLI", ProjectPath: ".tabnine/agent/skills", Category: common.CategoryLink},
	{Name: "trae", DisplayName: "Trae", ProjectPath: ".trae/skills", Category: common.CategoryLink},
	{Name: "trae-cn", DisplayName: "Trae CN", ProjectPath: ".trae/skills", Category: common.CategoryLink},
	{Name: "windsurf", DisplayName: "Windsurf", ProjectPath: ".windsurf/skills", Category: common.CategoryLink},
	{Name: "zencoder", DisplayName: "Zencoder", ProjectPath: ".zencoder/skills", Category: common.CategoryLink},

	// rootCollision (project path is "skills" at the repo root)
	{Name: "openclaw", DisplayName: "OpenClaw", ProjectPath: "skills", Category: common.CategoryRootCollision},
}

// init normalizes registry data once at package load: every entry receives
// SourceDir as its SourcePath (skills agents always link to the canonical
// skills source) and ProjectPath is forced to forward-slash form.
func init() {
	for index := range agents {
		agents[index].SourcePath = SourceDir
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
