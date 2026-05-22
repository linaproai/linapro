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
// callers can rely on stable iteration order. Coverage is intentionally
// curated: an agent only appears here when its public documentation,
// official source repository or vendor support page provides reliable
// evidence about how it consumes AGENTS.md. Agents whose AGENTS.md
// behaviour cannot be confirmed from public sources are deliberately
// omitted, even when they appear in the skills registry — guessing
// would mislead users into linking the wrong file. Add a new entry
// only when the source of truth is recorded in the inline comment
// next to it.
var agents = []AgentSpec{
	// link (private guide file -> AGENTS.md)
	{Name: "aider-desk", DisplayName: "AiderDesk", ProjectPath: "CONVENTIONS.md", Category: common.CategoryLink},      // Aider-AI/conventions repo: aider reads CONVENTIONS.md.
	{Name: "augment", DisplayName: "Augment", ProjectPath: ".augment-guidelines", Category: common.CategoryLink},      // Augment Code docs: project guidelines live in .augment-guidelines.
	{Name: "claude-code", DisplayName: "Claude Code", ProjectPath: "CLAUDE.md", Category: common.CategoryLink},        // Anthropic docs: Claude Code reads CLAUDE.md (single-file).
	{Name: "continue", DisplayName: "Continue", ProjectPath: ".continuerules", Category: common.CategoryLink},         // Continue docs: rules live in .continuerules at repo root.
	{Name: "crush", DisplayName: "Crush", ProjectPath: "CRUSH.md", Category: common.CategoryLink},                     // charmbracelet/crush README: default context file is CRUSH.md.
	{Name: "gemini-cli", DisplayName: "Gemini CLI", ProjectPath: "GEMINI.md", Category: common.CategoryLink},          // Gemini CLI docs: project context file is GEMINI.md.
	{Name: "iflow-cli", DisplayName: "iFlow CLI", ProjectPath: "IFLOW.md", Category: common.CategoryLink},             // iflow-ai/iflow-cli docs: memory file is IFLOW.md.
	{Name: "junie", DisplayName: "Junie", ProjectPath: ".junie/guidelines.md", Category: common.CategoryLink},         // JetBrains Junie docs: guidelines live in .junie/guidelines.md.
	{Name: "qwen-code", DisplayName: "Qwen Code", ProjectPath: "QWEN.md", Category: common.CategoryLink},              // Qwen Code docs: project context file is QWEN.md.
	{Name: "roo", DisplayName: "Roo Code", ProjectPath: ".roo/rules/AGENTS.md", Category: common.CategoryLink},        // Roo Code docs: rules dir hosts AGENTS.md as a single file.
	{Name: "tabnine-cli", DisplayName: "Tabnine CLI", ProjectPath: "TABNINE.md", Category: common.CategoryLink},       // balcsida/tabnine-patch confirms Tabnine CLI default context file is TABNINE.md.
	{Name: "windsurf", DisplayName: "Windsurf", ProjectPath: ".windsurfrules", Category: common.CategoryLink},         // Windsurf docs: project rules live in .windsurfrules.

	// native (agent reads AGENTS.md natively at repo root; nothing to
	// link). Listed for visibility in status output only.
	{Name: "amp", DisplayName: "Amp", Category: common.CategoryNative},                              // Sourcegraph Amp docs: native AGENTS.md support.
	{Name: "antigravity", DisplayName: "Antigravity", Category: common.CategoryNative},              // Antigravity docs: native AGENTS.md support.
	{Name: "cline", DisplayName: "Cline", Category: common.CategoryNative},                          // Cline docs: native AGENTS.md support.
	{Name: "codebuddy", DisplayName: "CodeBuddy", Category: common.CategoryNative},                  // Tencent CodeBuddy docs: prefers CODEBUDDY.md but auto-falls back to AGENTS.md when CODEBUDDY.md is absent — registering as native lets `clone && go` projects work zero-config.
	{Name: "codex", DisplayName: "Codex", Category: common.CategoryNative},                          // OpenAI Codex CLI docs: native AGENTS.md (the format author).
	{Name: "cursor", DisplayName: "Cursor", Category: common.CategoryNative},                        // Cursor 1.0+ docs: native AGENTS.md support.
	{Name: "deepagents", DisplayName: "Deep Agents", Category: common.CategoryNative},               // Deep Agents docs: native AGENTS.md support.
	{Name: "devin", DisplayName: "Devin for Terminal", Category: common.CategoryNative},             // docs.devin.ai/onboard-devin/agents-md: native AGENTS.md support.
	{Name: "dexto", DisplayName: "Dexto", Category: common.CategoryNative},                          // Dexto docs: native AGENTS.md support.
	{Name: "droid", DisplayName: "Droid", Category: common.CategoryNative},                          // Factory Droid docs (docs.factory.ai/cli/configuration/agents-md): native AGENTS.md.
	{Name: "firebender", DisplayName: "Firebender", Category: common.CategoryNative},                // Firebender docs: native AGENTS.md support.
	{Name: "forgecode", DisplayName: "ForgeCode", Category: common.CategoryNative},                  // forge-agents/forge: ACP-based universal CLI, native AGENTS.md.
	{Name: "github-copilot", DisplayName: "GitHub Copilot", Category: common.CategoryNative},        // GitHub Copilot Coding Agent docs: native AGENTS.md.
	{Name: "goose", DisplayName: "Goose", Category: common.CategoryNative},                          // goose-docs.ai/context-engineering: native AGENTS.md (alongside .goosehints).
	{Name: "hermes-agent", DisplayName: "Hermes Agent", Category: common.CategoryNative},            // hermesagent.org.cn docs: AGENTS.md is part of the layered config (with SOUL.md).
	{Name: "kimi-cli", DisplayName: "Kimi Code CLI", Category: common.CategoryNative},               // Kimi CLI docs: native AGENTS.md support.
	{Name: "kode", DisplayName: "Kode", Category: common.CategoryNative},                            // shareAI-lab/Kode-Agent README: "Kode supports the AGENTS.md standard".
	{Name: "mistral-vibe", DisplayName: "Mistral Vibe", Category: common.CategoryNative},            // mistralai/mistral-vibe ships its own AGENTS.md upstream.
	{Name: "mux", DisplayName: "Mux", Category: common.CategoryNative},                              // mux.coder.com/AGENTS confirms native AGENTS.md format.
	{Name: "neovate", DisplayName: "Neovate", Category: common.CategoryNative},                      // neovateai.dev/docs/rules: project AGENTS.md and global ~/.neovate/AGENTS.md.
	{Name: "opencode", DisplayName: "OpenCode", Category: common.CategoryNative},                    // OpenCode docs: native AGENTS.md (rules system).
	{Name: "openclaw", DisplayName: "OpenClaw", Category: common.CategoryNative},                    // OpenClaw docs: AGENTS.md is one of the six core configuration files.
	{Name: "openhands", DisplayName: "OpenHands", Category: common.CategoryNative},                  // docs.openhands.dev: AGENTS.md auto-loaded at repo root.
	{Name: "pi", DisplayName: "Pi", Category: common.CategoryNative},                                // earendil-works/pi: "Pi loads AGENTS.md or CLAUDE.md at startup".
	{Name: "replit", DisplayName: "Replit", Category: common.CategoryNative},                        // Replit docs: native AGENTS.md support.
	{Name: "trae", DisplayName: "Trae", Category: common.CategoryNative},                            // Trae community docs: "trae 支持默认读取 AGENTS.md 文件".
	{Name: "trae-cn", DisplayName: "Trae CN", Category: common.CategoryNative},                      // Trae CN inherits Trae's native AGENTS.md support.
	{Name: "universal", DisplayName: "Universal", Category: common.CategoryNative},                  // Generic placeholder used by tools following the AGENTS.md open standard.
	{Name: "warp", DisplayName: "Warp", Category: common.CategoryNative},                            // Warp docs: native AGENTS.md support.
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
