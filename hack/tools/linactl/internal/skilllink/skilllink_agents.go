// This file defines the supported agent registry and shared types for
// skilllink. The registry is hand-maintained to mirror the project paths
// published in https://github.com/vercel-labs/skills#supported-agents and
// classifies each agent into native, link or rootCollision categories.

package skilllink

import (
	"path/filepath"
	"sort"
)

// SourceDir is the canonical skills source directory relative to the
// repository root. All managed symlinks point at this directory.
const SourceDir = ".agents/skills"

// Category classifies an agent's project skill path.
type Category string

const (
	// CategoryNative marks agents whose project path is already SourceDir.
	// No symlink is required for these agents.
	CategoryNative Category = "native"
	// CategoryLink marks agents whose project path differs from SourceDir
	// and lives under a tool-specific dotted directory; a relative symlink
	// is created at the project path pointing back to SourceDir.
	CategoryLink Category = "link"
	// CategoryRootCollision marks agents whose project path is exactly
	// "skills" at the repository root. Creating that link would shadow any
	// real skills/ directory in the repo, so it is skipped by default and
	// only enabled when the caller passes FORCE=1.
	CategoryRootCollision Category = "rootCollision"
)

// AgentSpec describes one supported agent's project-level skill location.
type AgentSpec struct {
	// Name is the CLI identifier used on the command line (e.g. claude-code).
	Name string
	// DisplayName is the human-readable label rendered in status output.
	DisplayName string
	// ProjectPath is the project-relative skills directory path.
	ProjectPath string
	// Category indicates how the agent's project path should be handled.
	Category Category
}

// agents is the canonical agent registry. The list is sorted alphabetically
// by Name in init() so callers can rely on stable iteration order.
var agents = []AgentSpec{
	// native (project path == .agents/skills)
	{Name: "amp", DisplayName: "Amp", ProjectPath: ".agents/skills", Category: CategoryNative},
	{Name: "antigravity", DisplayName: "Antigravity", ProjectPath: ".agents/skills", Category: CategoryNative},
	{Name: "cline", DisplayName: "Cline", ProjectPath: ".agents/skills", Category: CategoryNative},
	{Name: "codex", DisplayName: "Codex", ProjectPath: ".agents/skills", Category: CategoryNative},
	{Name: "cursor", DisplayName: "Cursor", ProjectPath: ".agents/skills", Category: CategoryNative},
	{Name: "deepagents", DisplayName: "Deep Agents", ProjectPath: ".agents/skills", Category: CategoryNative},
	{Name: "dexto", DisplayName: "Dexto", ProjectPath: ".agents/skills", Category: CategoryNative},
	{Name: "firebender", DisplayName: "Firebender", ProjectPath: ".agents/skills", Category: CategoryNative},
	{Name: "gemini-cli", DisplayName: "Gemini CLI", ProjectPath: ".agents/skills", Category: CategoryNative},
	{Name: "github-copilot", DisplayName: "GitHub Copilot", ProjectPath: ".agents/skills", Category: CategoryNative},
	{Name: "kimi-cli", DisplayName: "Kimi Code CLI", ProjectPath: ".agents/skills", Category: CategoryNative},
	{Name: "opencode", DisplayName: "OpenCode", ProjectPath: ".agents/skills", Category: CategoryNative},
	{Name: "replit", DisplayName: "Replit", ProjectPath: ".agents/skills", Category: CategoryNative},
	{Name: "universal", DisplayName: "Universal", ProjectPath: ".agents/skills", Category: CategoryNative},
	{Name: "warp", DisplayName: "Warp", ProjectPath: ".agents/skills", Category: CategoryNative},

	// link (project path differs from SourceDir, not at repo root)
	{Name: "adal", DisplayName: "AdaL", ProjectPath: ".adal/skills", Category: CategoryLink},
	{Name: "aider-desk", DisplayName: "AiderDesk", ProjectPath: ".aider-desk/skills", Category: CategoryLink},
	{Name: "augment", DisplayName: "Augment", ProjectPath: ".augment/skills", Category: CategoryLink},
	{Name: "bob", DisplayName: "IBM Bob", ProjectPath: ".bob/skills", Category: CategoryLink},
	{Name: "claude-code", DisplayName: "Claude Code", ProjectPath: ".claude/skills", Category: CategoryLink},
	{Name: "codearts-agent", DisplayName: "CodeArts Agent", ProjectPath: ".codeartsdoer/skills", Category: CategoryLink},
	{Name: "codebuddy", DisplayName: "CodeBuddy", ProjectPath: ".codebuddy/skills", Category: CategoryLink},
	{Name: "codemaker", DisplayName: "Codemaker", ProjectPath: ".codemaker/skills", Category: CategoryLink},
	{Name: "codestudio", DisplayName: "Code Studio", ProjectPath: ".codestudio/skills", Category: CategoryLink},
	{Name: "command-code", DisplayName: "Command Code", ProjectPath: ".commandcode/skills", Category: CategoryLink},
	{Name: "continue", DisplayName: "Continue", ProjectPath: ".continue/skills", Category: CategoryLink},
	{Name: "cortex", DisplayName: "Cortex Code", ProjectPath: ".cortex/skills", Category: CategoryLink},
	{Name: "crush", DisplayName: "Crush", ProjectPath: ".crush/skills", Category: CategoryLink},
	{Name: "devin", DisplayName: "Devin for Terminal", ProjectPath: ".devin/skills", Category: CategoryLink},
	{Name: "droid", DisplayName: "Droid", ProjectPath: ".factory/skills", Category: CategoryLink},
	{Name: "forgecode", DisplayName: "ForgeCode", ProjectPath: ".forge/skills", Category: CategoryLink},
	{Name: "goose", DisplayName: "Goose", ProjectPath: ".goose/skills", Category: CategoryLink},
	{Name: "hermes-agent", DisplayName: "Hermes Agent", ProjectPath: ".hermes/skills", Category: CategoryLink},
	{Name: "iflow-cli", DisplayName: "iFlow CLI", ProjectPath: ".iflow/skills", Category: CategoryLink},
	{Name: "junie", DisplayName: "Junie", ProjectPath: ".junie/skills", Category: CategoryLink},
	{Name: "kilo", DisplayName: "Kilo Code", ProjectPath: ".kilocode/skills", Category: CategoryLink},
	{Name: "kiro-cli", DisplayName: "Kiro CLI", ProjectPath: ".kiro/skills", Category: CategoryLink},
	{Name: "kode", DisplayName: "Kode", ProjectPath: ".kode/skills", Category: CategoryLink},
	{Name: "mcpjam", DisplayName: "MCPJam", ProjectPath: ".mcpjam/skills", Category: CategoryLink},
	{Name: "mistral-vibe", DisplayName: "Mistral Vibe", ProjectPath: ".vibe/skills", Category: CategoryLink},
	{Name: "mux", DisplayName: "Mux", ProjectPath: ".mux/skills", Category: CategoryLink},
	{Name: "neovate", DisplayName: "Neovate", ProjectPath: ".neovate/skills", Category: CategoryLink},
	{Name: "openhands", DisplayName: "OpenHands", ProjectPath: ".openhands/skills", Category: CategoryLink},
	{Name: "pi", DisplayName: "Pi", ProjectPath: ".pi/skills", Category: CategoryLink},
	{Name: "pochi", DisplayName: "Pochi", ProjectPath: ".pochi/skills", Category: CategoryLink},
	{Name: "qoder", DisplayName: "Qoder", ProjectPath: ".qoder/skills", Category: CategoryLink},
	{Name: "qwen-code", DisplayName: "Qwen Code", ProjectPath: ".qwen/skills", Category: CategoryLink},
	{Name: "roo", DisplayName: "Roo Code", ProjectPath: ".roo/skills", Category: CategoryLink},
	{Name: "rovodev", DisplayName: "Rovo Dev", ProjectPath: ".rovodev/skills", Category: CategoryLink},
	{Name: "tabnine-cli", DisplayName: "Tabnine CLI", ProjectPath: ".tabnine/agent/skills", Category: CategoryLink},
	{Name: "trae", DisplayName: "Trae", ProjectPath: ".trae/skills", Category: CategoryLink},
	{Name: "trae-cn", DisplayName: "Trae CN", ProjectPath: ".trae/skills", Category: CategoryLink},
	{Name: "windsurf", DisplayName: "Windsurf", ProjectPath: ".windsurf/skills", Category: CategoryLink},
	{Name: "zencoder", DisplayName: "Zencoder", ProjectPath: ".zencoder/skills", Category: CategoryLink},

	// rootCollision (project path is "skills" at the repo root)
	{Name: "openclaw", DisplayName: "OpenClaw", ProjectPath: "skills", Category: CategoryRootCollision},
}

// init normalizes registry data once at package load.
func init() {
	for index := range agents {
		agents[index].ProjectPath = filepath.ToSlash(agents[index].ProjectPath)
	}
	sort.Slice(agents, func(left, right int) bool {
		return agents[left].Name < agents[right].Name
	})
}

// Agents returns a defensive copy of the supported agent registry sorted by
// agent name. Callers must not mutate the returned slice.
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
