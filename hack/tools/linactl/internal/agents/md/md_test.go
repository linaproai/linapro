// This file contains unit tests for the md subpackage. Tests cover
// agent registry integrity (link + native split, no rootCollision),
// single-file symlink creation/idempotency/mismatch/conflict semantics
// (which differ subtly from directory-kind resources only in conflict
// wording), and unlink guard rails preserving real files.

package md

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"linactl/internal/agents/common"
)

// newRepoFixture creates an isolated repository root with an AGENTS.md
// source file populated. It returns the absolute repo root.
func newRepoFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	source := filepath.Join(root, SourceFile)
	if err := os.WriteFile(source, []byte("# AGENTS.md test fixture\n"), 0o644); err != nil {
		t.Fatalf("create source file: %v", err)
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
	hasLink := false
	hasNative := false
	for _, spec := range specs {
		if spec.Name == "" {
			t.Fatalf("agent missing name: %+v", spec)
		}
		if _, dup := seen[spec.Name]; dup {
			t.Fatalf("duplicate agent name: %s", spec.Name)
		}
		seen[spec.Name] = struct{}{}
		switch spec.Category {
		case common.CategoryLink:
			hasLink = true
			if spec.ProjectPath == "" {
				t.Fatalf("link agent %s missing ProjectPath", spec.Name)
			}
		case common.CategoryNative:
			hasNative = true
		case common.CategoryRootCollision:
			t.Fatalf("md registry must not contain rootCollision agents (got %s)", spec.Name)
		default:
			t.Fatalf("agent %s has unexpected category %s", spec.Name, spec.Category)
		}
	}
	if !hasLink || !hasNative {
		t.Fatalf("expected both link-class and native-class agents in registry; hasLink=%v hasNative=%v",
			hasLink, hasNative)
	}
	// Required link agents.
	for _, required := range []string{"claude-code", "gemini-cli"} {
		if _, ok := seen[required]; !ok {
			t.Fatalf("expected required link agent %s in registry", required)
		}
	}
	// Required native agents.
	for _, required := range []string{"codex", "cursor"} {
		if _, ok := seen[required]; !ok {
			t.Fatalf("expected required native agent %s in registry", required)
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
	link := filepath.Join(root, "CLAUDE.md")
	info, err := os.Lstat(link)
	if err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("expected symlink at %s, err=%v info=%+v", link, err, info)
	}
	target, err := os.Readlink(link)
	if err != nil {
		t.Fatalf("readlink: %v", err)
	}
	// CLAUDE.md and AGENTS.md live in the same directory, so the
	// relative target is just the filename.
	if filepath.ToSlash(target) != "AGENTS.md" {
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

func TestApplyLinkNativeIsSkipped(t *testing.T) {
	root := newRepoFixture(t)
	results, err := ApplyLink(root, LinkRequest{Selectors: []string{"codex"}})
	if err != nil {
		t.Fatalf("apply native: %v", err)
	}
	if results[0].Status != common.StatusNative {
		t.Fatalf("expected native, got %s", results[0].Status)
	}
}

func TestApplyLinkConflictNeverDeletesRealFile(t *testing.T) {
	root := newRepoFixture(t)
	conflict := filepath.Join(root, "CLAUDE.md")
	if err := os.WriteFile(conflict, []byte("user-authored guide content"), 0o644); err != nil {
		t.Fatalf("write conflict: %v", err)
	}

	results, err := ApplyLink(root, LinkRequest{Selectors: []string{"claude-code"}, Force: true})
	if err != nil {
		t.Fatalf("apply with force: %v", err)
	}
	if results[0].Status != common.StatusConflict {
		t.Fatalf("expected conflict even with force, got %s", results[0].Status)
	}
	// The detail should mention "real file" rather than "real path" so
	// users can immediately spot they have an authored .md file.
	if results[0].Detail == "" || !contains(results[0].Detail, "real file") {
		t.Fatalf("expected detail to mention 'real file', got %q", results[0].Detail)
	}
	data, err := os.ReadFile(conflict)
	if err != nil {
		t.Fatalf("conflict file must survive: %v", err)
	}
	if string(data) != "user-authored guide content" {
		t.Fatalf("conflict file contents must survive untouched, got %q", string(data))
	}
}

func TestApplyLinkMismatchRequiresForce(t *testing.T) {
	root := newRepoFixture(t)
	link := filepath.Join(root, "CLAUDE.md")
	otherTarget := filepath.Join(root, "elsewhere.md")
	if err := os.WriteFile(otherTarget, []byte("foreign"), 0o644); err != nil {
		t.Fatalf("write foreign target: %v", err)
	}
	trySymlink(t, otherTarget, link)

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
	target, err := os.Readlink(link)
	if err != nil {
		t.Fatalf("readlink rebuilt: %v", err)
	}
	if filepath.ToSlash(target) != "AGENTS.md" {
		t.Fatalf("expected target AGENTS.md, got %s", target)
	}
}

func TestApplyUnlinkOnlyManagedSymlinks(t *testing.T) {
	root := newRepoFixture(t)
	if _, err := ApplyLink(root, LinkRequest{Selectors: []string{"claude-code"}}); err != nil {
		t.Fatalf("seed link: %v", err)
	}
	// Foreign symlink for gemini-cli pointing elsewhere.
	otherFile := filepath.Join(root, "elsewhere.md")
	if err := os.WriteFile(otherFile, []byte("foreign"), 0o644); err != nil {
		t.Fatalf("write elsewhere: %v", err)
	}
	geminiLink := filepath.Join(root, "GEMINI.md")
	trySymlink(t, otherFile, geminiLink)
	// Real authored guide file for qwen-code must be preserved.
	qwenReal := filepath.Join(root, "QWEN.md")
	if err := os.WriteFile(qwenReal, []byte("authored"), 0o644); err != nil {
		t.Fatalf("write qwen real: %v", err)
	}

	results, err := ApplyUnlink(root, UnlinkRequest{Selectors: []string{"claude-code", "gemini-cli", "qwen-code"}})
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
	if statusByName["gemini-cli"] != common.StatusSkippedForeignTarget {
		t.Fatalf("gemini-cli: expected skipped-foreign, got %s", statusByName["gemini-cli"])
	}
	if _, err := os.Lstat(geminiLink); err != nil {
		t.Fatalf("foreign gemini link must survive: %v", err)
	}
	if statusByName["qwen-code"] != common.StatusSkippedNotManaged {
		t.Fatalf("qwen-code: expected skipped-not-managed, got %s", statusByName["qwen-code"])
	}
	if data, err := os.ReadFile(qwenReal); err != nil || string(data) != "authored" {
		t.Fatalf("real qwen file must survive unlink unmodified: data=%q err=%v", string(data), err)
	}
}

func TestPlanListCoversAllAgents(t *testing.T) {
	root := newRepoFixture(t)
	results := PlanList(root)
	if len(results) != len(Agents()) {
		t.Fatalf("PlanList should cover every agent: got=%d want=%d", len(results), len(Agents()))
	}
}

// TestRegistryCoversCuratedAgents asserts that agents added during the
// FB-2 expansion are present with the correct category and (for link
// agents) the correct private guide file path. Each entry corresponds
// to an agent whose AGENTS.md behaviour is documented by an upstream
// source — see the inline comments next to the registry definitions in
// md_agents.go for citations. This guards against accidental removal
// or category flips during future edits and serves as a quick lookup
// reference for reviewers checking the curation rationale.
func TestRegistryCoversCuratedAgents(t *testing.T) {
	type expectation struct {
		category    common.Category
		projectPath string // empty for native agents
	}
	expected := map[string]expectation{
		// link-class additions from FB-2
		"aider-desk":  {common.CategoryLink, "CONVENTIONS.md"},
		"crush":       {common.CategoryLink, "CRUSH.md"},
		"iflow-cli":   {common.CategoryLink, "IFLOW.md"},
		"tabnine-cli": {common.CategoryLink, "TABNINE.md"},

		// native-class additions from FB-2 / FB-3
		"codebuddy":    {common.CategoryNative, ""},
		"devin":        {common.CategoryNative, ""},
		"droid":        {common.CategoryNative, ""},
		"forgecode":    {common.CategoryNative, ""},
		"goose":        {common.CategoryNative, ""},
		"hermes-agent": {common.CategoryNative, ""},
		"kode":         {common.CategoryNative, ""},
		"mistral-vibe": {common.CategoryNative, ""},
		"mux":          {common.CategoryNative, ""},
		"neovate":      {common.CategoryNative, ""},
		"openclaw":     {common.CategoryNative, ""},
		"openhands":    {common.CategoryNative, ""},
		"pi":           {common.CategoryNative, ""},
		"trae":         {common.CategoryNative, ""},
		"trae-cn":      {common.CategoryNative, ""},
	}
	for name, want := range expected {
		spec, ok := FindAgent(name)
		if !ok {
			t.Errorf("expected agent %q in registry", name)
			continue
		}
		if spec.Category != want.category {
			t.Errorf("agent %q: category=%s want=%s", name, spec.Category, want.category)
		}
		if spec.ProjectPath != want.projectPath {
			t.Errorf("agent %q: ProjectPath=%q want=%q", name, spec.ProjectPath, want.projectPath)
		}
	}
}

func TestLinkCandidatesExcludesNative(t *testing.T) {
	root := newRepoFixture(t)
	candidates := LinkCandidates(root)
	for _, entry := range candidates {
		if entry.Spec.SpecCategory() != common.CategoryLink {
			t.Fatalf("LinkCandidates leaked %s (category=%s)",
				entry.Spec.SpecName(), entry.Spec.SpecCategory())
		}
	}
}

// contains is a small, allocation-free substring helper used by the
// conflict-detail assertion above. We avoid importing strings just for
// strings.Contains in this test file to keep the import list minimal.
func contains(haystack string, needle string) bool {
	if len(needle) == 0 {
		return true
	}
	for index := 0; index+len(needle) <= len(haystack); index++ {
		if haystack[index:index+len(needle)] == needle {
			return true
		}
	}
	return false
}
