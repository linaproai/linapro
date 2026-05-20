// This file contains unit tests for the prompts subpackage. Tests cover
// agent registry integrity, link/unlink state transitions, source-path
// matching (since each prompts agent has its own SourcePath), and
// candidate filtering for the interactive flow.

package prompts

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"linactl/internal/agents/common"
)

// newRepoFixture creates an isolated repository root with the canonical
// .agents/prompts/opsx source directory populated. It returns the
// absolute repo root.
func newRepoFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".agents", "prompts", "opsx"), 0o755); err != nil {
		t.Fatalf("create source dir: %v", err)
	}
	return root
}

// trySymlink attempts a symlink and skips the test on Windows when the
// platform refuses to create one (no Developer Mode / Administrator).
func trySymlink(t *testing.T, oldname string, newname string) {
	t.Helper()
	if err := os.Symlink(oldname, newname); err != nil {
		if runtime.GOOS == "windows" {
			t.Skipf("skip: symlink unsupported on this Windows host: %v", err)
		}
		t.Fatalf("symlink: %v", err)
	}
}

func TestAgentsRegistryIntegrity(t *testing.T) {
	specs := Agents()
	if len(specs) == 0 {
		t.Fatalf("expected non-empty agent registry")
	}
	seen := make(map[string]struct{}, len(specs))
	for _, spec := range specs {
		if spec.Name == "" {
			t.Fatalf("agent missing name: %+v", spec)
		}
		if _, dup := seen[spec.Name]; dup {
			t.Fatalf("duplicate agent name: %s", spec.Name)
		}
		seen[spec.Name] = struct{}{}
		if spec.SourcePath == "" {
			t.Fatalf("agent %s missing SourcePath", spec.Name)
		}
		if spec.ProjectPath == "" {
			t.Fatalf("agent %s missing ProjectPath", spec.Name)
		}
		if spec.Category != common.CategoryLink && spec.Category != common.CategoryNative {
			t.Fatalf("agent %s has unexpected category %s", spec.Name, spec.Category)
		}
	}
	// Initial coverage requires the four mainstream agents.
	for _, required := range []string{"claude-code", "codex", "cursor", "gemini-cli"} {
		if _, ok := seen[required]; !ok {
			t.Fatalf("expected required agent %s in initial registry", required)
		}
	}
}

func TestApplyLinkCreatesAndIsIdempotent(t *testing.T) {
	root := newRepoFixture(t)
	results, err := ApplyLink(root, LinkRequest{Selectors: []string{"claude-code"}})
	if err != nil {
		t.Fatalf("first apply: %v", err)
	}
	if len(results) != 1 || results[0].Status != common.StatusCreated {
		t.Fatalf("expected created, got %+v", results)
	}
	link := filepath.Join(root, ".claude", "commands", "opsx")
	info, err := os.Lstat(link)
	if err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("expected symlink at %s, err=%v info=%+v", link, err, info)
	}
	target, err := os.Readlink(link)
	if err != nil {
		t.Fatalf("readlink: %v", err)
	}
	// Relative target should walk up from .claude/commands/ to .agents/prompts/opsx.
	if filepath.ToSlash(target) != "../../.agents/prompts/opsx" {
		t.Fatalf("unexpected relative target %s", target)
	}

	again, err := ApplyLink(root, LinkRequest{Selectors: []string{"claude-code"}})
	if err != nil {
		t.Fatalf("second apply: %v", err)
	}
	if again[0].Status != common.StatusOK {
		t.Fatalf("expected ok on second apply, got %s", again[0].Status)
	}
}

func TestApplyLinkAllExpandsToLinkClass(t *testing.T) {
	root := newRepoFixture(t)
	results, err := ApplyLink(root, LinkRequest{Selectors: []string{"all"}})
	if err != nil {
		t.Fatalf("apply all: %v", err)
	}
	if len(results) != len(Agents()) {
		t.Fatalf("expected %d results, got %d", len(Agents()), len(results))
	}
	for _, result := range results {
		if result.Status != common.StatusCreated {
			t.Fatalf("agent %s expected created, got %s", result.Spec.SpecName(), result.Status)
		}
	}
}

func TestApplyLinkMismatchRequiresForce(t *testing.T) {
	root := newRepoFixture(t)
	link := filepath.Join(root, ".claude", "commands", "opsx")
	if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
		t.Fatalf("mkdir parent: %v", err)
	}
	other := filepath.Join(root, "other-target")
	if err := os.MkdirAll(other, 0o755); err != nil {
		t.Fatalf("mkdir other: %v", err)
	}
	trySymlink(t, other, link)

	results, err := ApplyLink(root, LinkRequest{Selectors: []string{"claude-code"}})
	if err != nil {
		t.Fatalf("apply mismatch: %v", err)
	}
	if results[0].Status != common.StatusMismatch {
		t.Fatalf("expected mismatch, got %s", results[0].Status)
	}

	rebuilt, err := ApplyLink(root, LinkRequest{Selectors: []string{"claude-code"}, Force: true})
	if err != nil {
		t.Fatalf("apply rebuild: %v", err)
	}
	if rebuilt[0].Status != common.StatusRebuilt {
		t.Fatalf("expected rebuilt, got %s", rebuilt[0].Status)
	}
}

func TestApplyLinkConflictNeverDeletes(t *testing.T) {
	root := newRepoFixture(t)
	conflict := filepath.Join(root, ".claude", "commands", "opsx")
	if err := os.MkdirAll(conflict, 0o755); err != nil {
		t.Fatalf("mkdir conflict: %v", err)
	}
	sentinel := filepath.Join(conflict, "sentinel.txt")
	if err := os.WriteFile(sentinel, []byte("keep"), 0o644); err != nil {
		t.Fatalf("write sentinel: %v", err)
	}

	results, err := ApplyLink(root, LinkRequest{Selectors: []string{"claude-code"}, Force: true})
	if err != nil {
		t.Fatalf("apply with force: %v", err)
	}
	if results[0].Status != common.StatusConflict {
		t.Fatalf("expected conflict even with force, got %s", results[0].Status)
	}
	if _, err := os.Stat(sentinel); err != nil {
		t.Fatalf("sentinel must survive conflict; err=%v", err)
	}
}

func TestApplyUnlinkOnlyManagedLinks(t *testing.T) {
	root := newRepoFixture(t)
	if _, err := ApplyLink(root, LinkRequest{Selectors: []string{"claude-code"}}); err != nil {
		t.Fatalf("seed link: %v", err)
	}
	// Foreign symlink for cursor pointing at an unrelated location.
	otherTarget := filepath.Join(root, "elsewhere")
	if err := os.MkdirAll(otherTarget, 0o755); err != nil {
		t.Fatalf("mkdir elsewhere: %v", err)
	}
	cursorLink := filepath.Join(root, ".cursor", "commands", "opsx")
	if err := os.MkdirAll(filepath.Dir(cursorLink), 0o755); err != nil {
		t.Fatalf("mkdir cursor parent: %v", err)
	}
	trySymlink(t, otherTarget, cursorLink)
	// Real directory for codex must be preserved.
	codexReal := filepath.Join(root, ".codex", "prompts", "opsx")
	if err := os.MkdirAll(codexReal, 0o755); err != nil {
		t.Fatalf("mkdir codex real: %v", err)
	}

	results, err := ApplyUnlink(root, UnlinkRequest{Selectors: []string{"claude-code", "cursor", "codex"}})
	if err != nil {
		t.Fatalf("apply unlink: %v", err)
	}
	statusByName := make(map[string]common.Status, len(results))
	for _, result := range results {
		statusByName[result.Spec.SpecName()] = result.Status
	}
	if statusByName["claude-code"] != common.StatusRemoved {
		t.Fatalf("claude-code: expected removed, got %s", statusByName["claude-code"])
	}
	if statusByName["cursor"] != common.StatusSkippedForeignTarget {
		t.Fatalf("cursor: expected skipped-foreign, got %s", statusByName["cursor"])
	}
	if _, err := os.Lstat(cursorLink); err != nil {
		t.Fatalf("foreign cursor link must survive: %v", err)
	}
	if statusByName["codex"] != common.StatusSkippedNotManaged {
		t.Fatalf("codex: expected skipped-not-managed, got %s", statusByName["codex"])
	}
	if _, err := os.Stat(codexReal); err != nil {
		t.Fatalf("real codex directory must survive unlink: %v", err)
	}
}

func TestApplyLinkRequiresSelector(t *testing.T) {
	root := newRepoFixture(t)
	if _, err := ApplyLink(root, LinkRequest{}); err == nil {
		t.Fatalf("expected error when no selector provided")
	}
	if _, err := ApplyUnlink(root, UnlinkRequest{}); err == nil {
		t.Fatalf("expected error when no selector provided")
	}
}

func TestPlanListCoversAllAgents(t *testing.T) {
	root := newRepoFixture(t)
	results := PlanList(root)
	if len(results) != len(Agents()) {
		t.Fatalf("PlanList should cover every agent: got=%d want=%d", len(results), len(Agents()))
	}
}

func TestLinkCandidatesIncludesAllLinkClass(t *testing.T) {
	root := newRepoFixture(t)
	candidates := LinkCandidates(root)
	if len(candidates) != len(Agents()) {
		// All four initial agents are link-class.
		t.Fatalf("expected all link-class agents in candidates, got %d", len(candidates))
	}
}

func TestUnlinkCandidatesOnlyManagedLinks(t *testing.T) {
	root := newRepoFixture(t)
	if _, err := ApplyLink(root, LinkRequest{Selectors: []string{"claude-code"}}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	candidates := UnlinkCandidates(root)
	if len(candidates) != 1 {
		t.Fatalf("expected exactly 1 managed candidate, got %d", len(candidates))
	}
	if candidates[0].Spec.SpecName() != "claude-code" {
		t.Fatalf("expected claude-code candidate, got %s", candidates[0].Spec.SpecName())
	}
}
