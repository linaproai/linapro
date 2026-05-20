// This file contains unit tests for the skills subpackage. Tests cover
// agent registry integrity, selector parsing/expansion, link/unlink state
// transitions and conflict guarding.

package skills

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// newRepoFixture creates an isolated repository root with the canonical
// .agents/skills directory populated. It returns the absolute repo root.
func newRepoFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".agents", "skills"), 0o755); err != nil {
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
		switch spec.Category {
		case CategoryNative:
			if spec.ProjectPath != SourceDir {
				t.Fatalf("native agent %s must use ProjectPath=%s, got %s", spec.Name, SourceDir, spec.ProjectPath)
			}
		case CategoryLink:
			if spec.ProjectPath == "" || spec.ProjectPath == SourceDir {
				t.Fatalf("link agent %s has invalid ProjectPath %q", spec.Name, spec.ProjectPath)
			}
			if !strings.HasPrefix(spec.ProjectPath, ".") {
				t.Fatalf("link agent %s ProjectPath must live under a dotted directory, got %s", spec.Name, spec.ProjectPath)
			}
		case CategoryRootCollision:
			if spec.ProjectPath != "skills" {
				t.Fatalf("rootCollision agent %s must use ProjectPath=skills, got %s", spec.Name, spec.ProjectPath)
			}
		default:
			t.Fatalf("agent %s has unknown category %s", spec.Name, spec.Category)
		}
	}
}

func TestParseSelectors(t *testing.T) {
	cases := []struct {
		input string
		want  []string
	}{
		{input: "", want: nil},
		{input: "  ", want: nil},
		{input: "claude-code", want: []string{"claude-code"}},
		{input: "claude-code,codebuddy", want: []string{"claude-code", "codebuddy"}},
		{input: " claude-code , , qoder ", want: []string{"claude-code", "qoder"}},
		{input: "all", want: []string{"all"}},
	}
	for _, testCase := range cases {
		got := ParseSelectors(testCase.input)
		if len(got) != len(testCase.want) {
			t.Fatalf("ParseSelectors(%q) length got=%v want=%v", testCase.input, got, testCase.want)
		}
		for index := range got {
			if got[index] != testCase.want[index] {
				t.Fatalf("ParseSelectors(%q)[%d] got=%s want=%s", testCase.input, index, got[index], testCase.want[index])
			}
		}
	}
}

func TestApplyLinkCreatesAndIsIdempotent(t *testing.T) {
	root := newRepoFixture(t)
	results, err := ApplyLink(root, LinkRequest{Selectors: []string{"claude-code"}})
	if err != nil {
		t.Fatalf("first apply: %v", err)
	}
	if len(results) != 1 || results[0].Status != StatusCreated {
		t.Fatalf("expected created, got %+v", results)
	}
	link := filepath.Join(root, ".claude", "skills")
	info, err := os.Lstat(link)
	if err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("expected symlink at %s, err=%v info=%+v", link, err, info)
	}

	again, err := ApplyLink(root, LinkRequest{Selectors: []string{"claude-code"}})
	if err != nil {
		t.Fatalf("second apply: %v", err)
	}
	if again[0].Status != StatusOK {
		t.Fatalf("expected ok on second apply, got %s", again[0].Status)
	}
}

func TestApplyLinkNativeIsSkipped(t *testing.T) {
	root := newRepoFixture(t)
	results, err := ApplyLink(root, LinkRequest{Selectors: []string{"cursor"}})
	if err != nil {
		t.Fatalf("apply native: %v", err)
	}
	if results[0].Status != StatusNative {
		t.Fatalf("expected native, got %s", results[0].Status)
	}
	if _, statErr := os.Lstat(filepath.Join(root, ".cursor", "skills")); !os.IsNotExist(statErr) {
		t.Fatalf("native agent must not create any path; stat err=%v", statErr)
	}
}

func TestApplyLinkMismatchRequiresForce(t *testing.T) {
	root := newRepoFixture(t)
	link := filepath.Join(root, ".claude", "skills")
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
	if results[0].Status != StatusMismatch {
		t.Fatalf("expected mismatch, got %s detail=%s", results[0].Status, results[0].Detail)
	}

	rebuilt, err := ApplyLink(root, LinkRequest{Selectors: []string{"claude-code"}, Force: true})
	if err != nil {
		t.Fatalf("apply rebuild: %v", err)
	}
	if rebuilt[0].Status != StatusRebuilt {
		t.Fatalf("expected rebuilt, got %s detail=%s", rebuilt[0].Status, rebuilt[0].Detail)
	}
	target, err := os.Readlink(link)
	if err != nil {
		t.Fatalf("readlink rebuilt: %v", err)
	}
	if filepath.ToSlash(target) != "../.agents/skills" {
		t.Fatalf("expected relative target ../.agents/skills, got %s", target)
	}
}

func TestApplyLinkConflictNeverDeletes(t *testing.T) {
	root := newRepoFixture(t)
	conflict := filepath.Join(root, ".claude", "skills")
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
	if results[0].Status != StatusConflict {
		t.Fatalf("expected conflict even with force, got %s", results[0].Status)
	}
	if _, err := os.Stat(sentinel); err != nil {
		t.Fatalf("sentinel must survive conflict; err=%v", err)
	}
}

func TestApplyLinkRootCollisionRequiresForce(t *testing.T) {
	root := newRepoFixture(t)
	defaultRun, err := ApplyLink(root, LinkRequest{Selectors: []string{"openclaw"}})
	if err != nil {
		t.Fatalf("apply openclaw default: %v", err)
	}
	if defaultRun[0].Status != StatusSkippedRootCollision {
		t.Fatalf("expected skipped-root-collision when allRoot is implicit; got %s", defaultRun[0].Status)
	}

	forced, err := ApplyLink(root, LinkRequest{Selectors: []string{"openclaw"}, Force: true})
	if err != nil {
		t.Fatalf("apply openclaw force: %v", err)
	}
	if forced[0].Status != StatusCreated {
		t.Fatalf("expected created with force, got %s", forced[0].Status)
	}
	if _, err := os.Lstat(filepath.Join(root, "skills")); err != nil {
		t.Fatalf("expected skills/ symlink: %v", err)
	}
}

func TestApplyLinkRootCollisionWithAllPolicy(t *testing.T) {
	root := newRepoFixture(t)
	results, err := ApplyLink(root, LinkRequest{Selectors: []string{"all"}})
	if err != nil {
		t.Fatalf("apply all: %v", err)
	}
	for _, result := range results {
		if result.Spec.SpecCategory() == CategoryRootCollision {
			t.Fatalf("rootCollision agent %s should not appear in default all expansion", result.Spec.SpecName())
		}
	}
}

func TestApplyUnlinkOnlyManagedLinks(t *testing.T) {
	root := newRepoFixture(t)
	// Native: nothing to unlink, skipped via policy filter.
	if _, err := ApplyUnlink(root, UnlinkRequest{Selectors: []string{"cursor"}}); err == nil {
		// resolveTargets keeps the spec because we matched by name; expect StatusAbsent.
	}
	// Set up a managed link.
	if _, err := ApplyLink(root, LinkRequest{Selectors: []string{"claude-code"}}); err != nil {
		t.Fatalf("seed link: %v", err)
	}
	// Set up a foreign symlink for codebuddy.
	otherTarget := filepath.Join(root, "elsewhere")
	if err := os.MkdirAll(otherTarget, 0o755); err != nil {
		t.Fatalf("mkdir elsewhere: %v", err)
	}
	codebuddyLink := filepath.Join(root, ".codebuddy", "skills")
	if err := os.MkdirAll(filepath.Dir(codebuddyLink), 0o755); err != nil {
		t.Fatalf("mkdir codebuddy parent: %v", err)
	}
	trySymlink(t, otherTarget, codebuddyLink)
	// Set up a real directory for windsurf (must be preserved).
	windsurfReal := filepath.Join(root, ".windsurf", "skills")
	if err := os.MkdirAll(windsurfReal, 0o755); err != nil {
		t.Fatalf("mkdir windsurf real: %v", err)
	}
	sentinel := filepath.Join(windsurfReal, "keep.md")
	if err := os.WriteFile(sentinel, []byte("keep"), 0o644); err != nil {
		t.Fatalf("write sentinel: %v", err)
	}

	results, err := ApplyUnlink(root, UnlinkRequest{Selectors: []string{"claude-code", "codebuddy", "windsurf"}})
	if err != nil {
		t.Fatalf("apply unlink: %v", err)
	}
	statusByName := make(map[string]Status, len(results))
	for _, result := range results {
		statusByName[result.Spec.SpecName()] = result.Status
	}
	if statusByName["claude-code"] != StatusRemoved {
		t.Fatalf("claude-code: expected removed, got %s", statusByName["claude-code"])
	}
	if _, statErr := os.Lstat(filepath.Join(root, ".claude", "skills")); !os.IsNotExist(statErr) {
		t.Fatalf("claude-code link must be gone; stat err=%v", statErr)
	}
	if statusByName["codebuddy"] != StatusSkippedForeignTarget {
		t.Fatalf("codebuddy: expected skipped-foreign, got %s", statusByName["codebuddy"])
	}
	if _, statErr := os.Lstat(codebuddyLink); statErr != nil {
		t.Fatalf("foreign codebuddy link must survive; stat err=%v", statErr)
	}
	if statusByName["windsurf"] != StatusSkippedNotManaged {
		t.Fatalf("windsurf: expected skipped-not-managed, got %s", statusByName["windsurf"])
	}
	if _, statErr := os.Stat(sentinel); statErr != nil {
		t.Fatalf("real windsurf directory must survive unlink: %v", statErr)
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
