// This file implements link planning, apply, unlink and target inspection
// against the local filesystem. It uses os.Symlink, os.Readlink, os.Lstat,
// os.Remove and os.MkdirAll only, and never invokes platform commands such
// as ln or mklink. Real directories and files are never auto-removed.

package skilllink

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// Status is the outcome of one agent action emitted by the command output.
type Status string

const (
	// StatusNative indicates the agent project path is already SourceDir.
	StatusNative Status = "native"
	// StatusOK indicates the link already exists and points at SourceDir.
	StatusOK Status = "ok"
	// StatusCreated indicates a new link was just created.
	StatusCreated Status = "created"
	// StatusRebuilt indicates a mismatched link was removed and recreated.
	StatusRebuilt Status = "rebuilt"
	// StatusMismatch indicates a link exists but points at a different
	// target than SourceDir, and force was not provided.
	StatusMismatch Status = "mismatch"
	// StatusConflict indicates the project path exists as a real directory
	// or file. Auto-resolution is never attempted.
	StatusConflict Status = "conflict"
	// StatusSkippedRootCollision indicates a rootCollision agent was
	// skipped because force was not provided.
	StatusSkippedRootCollision Status = "skipped-root-collision"
	// StatusRemoved indicates an unlink call removed a managed link.
	StatusRemoved Status = "removed"
	// StatusSkippedForeignTarget indicates an unlink target is a symlink
	// pointing at a non-managed location and was preserved.
	StatusSkippedForeignTarget Status = "skipped-foreign"
	// StatusSkippedNotManaged indicates an unlink target is a real
	// directory or file and was preserved.
	StatusSkippedNotManaged Status = "skipped-not-managed"
	// StatusAbsent indicates the unlink target does not exist.
	StatusAbsent Status = "absent"
	// StatusError indicates an unrecoverable error processing the agent.
	// The caller renders Detail to show the underlying error message.
	StatusError Status = "error"
)

// Result describes the outcome of one agent action.
type Result struct {
	// Spec is the agent under inspection.
	Spec AgentSpec
	// Status is the action outcome.
	Status Status
	// Detail provides additional context for non-trivial statuses such as
	// the actual link target during a mismatch or the underlying error
	// message when Status is StatusError.
	Detail string
}

// IsError reports whether the result represents an unrecoverable error.
func (r Result) IsError() bool {
	return r.Status == StatusError
}

// Inspect returns the current Status and Detail for an agent without any
// filesystem mutation. It is used by the default no-selector listing flow.
func Inspect(repoRoot string, spec AgentSpec) Result {
	if spec.Category == CategoryNative {
		return Result{Spec: spec, Status: StatusNative}
	}
	target := filepath.Join(repoRoot, filepath.FromSlash(spec.ProjectPath))
	info, lstatErr := os.Lstat(target)
	if errors.Is(lstatErr, os.ErrNotExist) {
		if spec.Category == CategoryRootCollision {
			return Result{Spec: spec, Status: StatusSkippedRootCollision, Detail: "use FORCE=1 to create"}
		}
		return Result{Spec: spec, Status: StatusAbsent}
	}
	if lstatErr != nil {
		return Result{Spec: spec, Status: StatusError, Detail: lstatErr.Error()}
	}
	if info.Mode()&os.ModeSymlink == 0 {
		return Result{Spec: spec, Status: StatusConflict, Detail: "real path exists; resolve manually"}
	}
	matches, currentTarget, err := linkMatchesSource(repoRoot, target)
	if err != nil {
		return Result{Spec: spec, Status: StatusError, Detail: err.Error()}
	}
	if matches {
		return Result{Spec: spec, Status: StatusOK}
	}
	return Result{Spec: spec, Status: StatusMismatch, Detail: "-> " + currentTarget}
}

// PlanList returns inspection results for every agent in the registry.
func PlanList(repoRoot string) []Result {
	out := make([]Result, 0, len(agents))
	for _, spec := range agents {
		out = append(out, Inspect(repoRoot, spec))
	}
	return out
}

// ApplyLink executes the link request and returns one Result per resolved
// target. native agents are reported via StatusNative and skipped from any
// filesystem mutation.
func ApplyLink(repoRoot string, request LinkRequest) ([]Result, error) {
	if len(request.Selectors) == 0 {
		return nil, errors.New("no agent selected; pass agent=<name|all|csv>")
	}
	policy := targetPolicy{
		includeNative:        true,
		includeRootCollision: request.Force,
	}
	targets, err := resolveTargets(request.Selectors, policy)
	if err != nil {
		return nil, err
	}
	if len(targets) == 0 {
		return nil, errors.New("no agent selected")
	}
	results := make([]Result, 0, len(targets))
	for _, spec := range targets {
		results = append(results, applyOneLink(repoRoot, spec, request.Force))
	}
	return results, nil
}

// ApplyUnlink executes the unlink request and returns one Result per
// resolved target.
func ApplyUnlink(repoRoot string, request UnlinkRequest) ([]Result, error) {
	if len(request.Selectors) == 0 {
		return nil, errors.New("no agent selected; pass agent=<name|all|csv>")
	}
	// Unlink should not implicitly touch native or rootCollision paths.
	policy := targetPolicy{
		includeNative:        false,
		includeRootCollision: false,
	}
	targets, err := resolveTargets(request.Selectors, policy)
	if err != nil {
		return nil, err
	}
	if len(targets) == 0 {
		return nil, errors.New("no agent selected")
	}
	results := make([]Result, 0, len(targets))
	for _, spec := range targets {
		results = append(results, applyOneUnlink(repoRoot, spec))
	}
	return results, nil
}

// applyOneLink performs the link action for a single agent.
func applyOneLink(repoRoot string, spec AgentSpec, force bool) Result {
	if spec.Category == CategoryNative {
		return Result{Spec: spec, Status: StatusNative}
	}
	if spec.Category == CategoryRootCollision && !force {
		return Result{Spec: spec, Status: StatusSkippedRootCollision, Detail: "use FORCE=1 to create"}
	}
	target := filepath.Join(repoRoot, filepath.FromSlash(spec.ProjectPath))
	info, lstatErr := os.Lstat(target)
	switch {
	case errors.Is(lstatErr, os.ErrNotExist):
		return createLink(repoRoot, spec, target)
	case lstatErr != nil:
		return Result{Spec: spec, Status: StatusError, Detail: lstatErr.Error()}
	}
	if info.Mode()&os.ModeSymlink == 0 {
		return Result{Spec: spec, Status: StatusConflict, Detail: "real path exists; resolve manually"}
	}
	matches, currentTarget, err := linkMatchesSource(repoRoot, target)
	if err != nil {
		return Result{Spec: spec, Status: StatusError, Detail: err.Error()}
	}
	if matches {
		return Result{Spec: spec, Status: StatusOK}
	}
	if !force {
		return Result{Spec: spec, Status: StatusMismatch, Detail: "-> " + currentTarget + "; use FORCE=1 to rebuild"}
	}
	if removeErr := os.Remove(target); removeErr != nil {
		return Result{Spec: spec, Status: StatusError, Detail: "remove existing link: " + removeErr.Error()}
	}
	result := createLink(repoRoot, spec, target)
	if result.Status == StatusCreated {
		result.Status = StatusRebuilt
		result.Detail = "previous: -> " + currentTarget
	}
	return result
}

// createLink resolves the relative source path and creates a symlink at
// the project path. Parent directories are created on demand.
func createLink(repoRoot string, spec AgentSpec, target string) Result {
	if mkErr := os.MkdirAll(filepath.Dir(target), 0o755); mkErr != nil {
		return Result{Spec: spec, Status: StatusError, Detail: "create parent directory: " + mkErr.Error()}
	}
	source := filepath.Join(repoRoot, filepath.FromSlash(SourceDir))
	relativeSource, relErr := filepath.Rel(filepath.Dir(target), source)
	if relErr != nil {
		return Result{Spec: spec, Status: StatusError, Detail: "compute relative source: " + relErr.Error()}
	}
	if symErr := os.Symlink(relativeSource, target); symErr != nil {
		return Result{Spec: spec, Status: StatusError, Detail: symlinkErrorDetail(symErr)}
	}
	return Result{Spec: spec, Status: StatusCreated, Detail: "-> " + filepath.ToSlash(relativeSource)}
}

// applyOneUnlink performs the unlink action for a single agent.
func applyOneUnlink(repoRoot string, spec AgentSpec) Result {
	target := filepath.Join(repoRoot, filepath.FromSlash(spec.ProjectPath))
	info, lstatErr := os.Lstat(target)
	if errors.Is(lstatErr, os.ErrNotExist) {
		return Result{Spec: spec, Status: StatusAbsent}
	}
	if lstatErr != nil {
		return Result{Spec: spec, Status: StatusError, Detail: lstatErr.Error()}
	}
	if info.Mode()&os.ModeSymlink == 0 {
		return Result{Spec: spec, Status: StatusSkippedNotManaged, Detail: "real path; not removed"}
	}
	matches, currentTarget, err := linkMatchesSource(repoRoot, target)
	if err != nil {
		return Result{Spec: spec, Status: StatusError, Detail: err.Error()}
	}
	if !matches {
		return Result{Spec: spec, Status: StatusSkippedForeignTarget, Detail: "-> " + currentTarget}
	}
	if removeErr := os.Remove(target); removeErr != nil {
		return Result{Spec: spec, Status: StatusError, Detail: removeErr.Error()}
	}
	return Result{Spec: spec, Status: StatusRemoved}
}

// linkMatchesSource checks whether the symlink at target points at the
// canonical source directory. It accepts both absolute and relative target
// values and returns the resolved absolute target alongside the result so
// callers can render diagnostic detail.
func linkMatchesSource(repoRoot string, link string) (bool, string, error) {
	currentTarget, err := os.Readlink(link)
	if err != nil {
		return false, "", fmt.Errorf("readlink %s: %w", link, err)
	}
	resolved := currentTarget
	if !filepath.IsAbs(resolved) {
		resolved = filepath.Join(filepath.Dir(link), resolved)
	}
	resolvedClean, err := filepath.Abs(filepath.Clean(resolved))
	if err != nil {
		return false, currentTarget, fmt.Errorf("resolve link target: %w", err)
	}
	source := filepath.Join(repoRoot, filepath.FromSlash(SourceDir))
	sourceClean, err := filepath.Abs(filepath.Clean(source))
	if err != nil {
		return false, currentTarget, fmt.Errorf("resolve source path: %w", err)
	}
	return pathsEqual(resolvedClean, sourceClean), filepath.ToSlash(currentTarget), nil
}

// pathsEqual compares two cleaned paths using the platform's case-folding
// rules. On Windows path comparisons must be case-insensitive.
func pathsEqual(left string, right string) bool {
	if runtime.GOOS == "windows" {
		return equalFoldPath(left, right)
	}
	return left == right
}

// equalFoldPath performs an ASCII case-insensitive path comparison suitable
// for Windows file systems where path components compare case-insensitively.
func equalFoldPath(left string, right string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := 0; index < len(left); index++ {
		leftByte := left[index]
		rightByte := right[index]
		if leftByte >= 'A' && leftByte <= 'Z' {
			leftByte += 'a' - 'A'
		}
		if rightByte >= 'A' && rightByte <= 'Z' {
			rightByte += 'a' - 'A'
		}
		if leftByte != rightByte {
			return false
		}
	}
	return true
}

// symlinkErrorDetail formats an os.Symlink error with platform-specific
// guidance when permission is denied on Windows.
func symlinkErrorDetail(err error) string {
	if err == nil {
		return ""
	}
	message := err.Error()
	if runtime.GOOS == "windows" {
		return message + "; Windows requires Developer Mode or Administrator to create symlinks"
	}
	return message
}
