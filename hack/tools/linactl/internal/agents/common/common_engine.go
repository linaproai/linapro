// This file implements the resource-agnostic link state machine. Inspect,
// ApplyOneLink and ApplyOneUnlink operate on any SpecLike value and produce
// the same Status outcomes regardless of whether the binding manages a
// directory or a single-file symlink. Callers (resource subpackages)
// supply the concrete AgentSpec values via the SpecLike interface and
// receive Result values they can pass directly to Render.
//
// Real directories and files are never automatically removed, even when
// force is true. force only acts on "target is already a symlink but
// points elsewhere" mismatches.

package common

import (
	"errors"
	"os"
	"path/filepath"
)

// Inspect returns the current Status and Detail for an agent without any
// filesystem mutation. It is used by the default no-selector listing flow
// shared across every resource subpackage.
func Inspect(repoRoot string, spec SpecLike) Result {
	if spec.SpecCategory() == CategoryNative {
		return Result{Spec: spec, Status: StatusNative}
	}
	target := filepath.Join(repoRoot, filepath.FromSlash(spec.SpecProjectPath()))
	info, lstatErr := os.Lstat(target)
	if errors.Is(lstatErr, os.ErrNotExist) {
		if spec.SpecCategory() == CategoryRootCollision {
			return Result{Spec: spec, Status: StatusSkippedRootCollision, Detail: "use FORCE=1 to create"}
		}
		return Result{Spec: spec, Status: StatusAbsent}
	}
	if lstatErr != nil {
		return Result{Spec: spec, Status: StatusError, Detail: lstatErr.Error()}
	}
	if info.Mode()&os.ModeSymlink == 0 {
		return Result{Spec: spec, Status: StatusConflict, Detail: conflictDetail(spec.SpecKind())}
	}
	source := absoluteSource(repoRoot, spec)
	matches, currentTarget, err := LinkMatchesSource(target, source)
	if err != nil {
		return Result{Spec: spec, Status: StatusError, Detail: err.Error()}
	}
	if matches {
		return Result{Spec: spec, Status: StatusOK}
	}
	return Result{Spec: spec, Status: StatusMismatch, Detail: "-> " + currentTarget}
}

// ApplyOneLink performs the link action for a single agent.
func ApplyOneLink(repoRoot string, spec SpecLike, force bool) Result {
	if spec.SpecCategory() == CategoryNative {
		return Result{Spec: spec, Status: StatusNative}
	}
	if spec.SpecCategory() == CategoryRootCollision && !force {
		return Result{Spec: spec, Status: StatusSkippedRootCollision, Detail: "use FORCE=1 to create"}
	}
	target := filepath.Join(repoRoot, filepath.FromSlash(spec.SpecProjectPath()))
	info, lstatErr := os.Lstat(target)
	switch {
	case errors.Is(lstatErr, os.ErrNotExist):
		return createLink(repoRoot, spec, target)
	case lstatErr != nil:
		return Result{Spec: spec, Status: StatusError, Detail: lstatErr.Error()}
	}
	if info.Mode()&os.ModeSymlink == 0 {
		return Result{Spec: spec, Status: StatusConflict, Detail: conflictDetail(spec.SpecKind())}
	}
	source := absoluteSource(repoRoot, spec)
	matches, currentTarget, err := LinkMatchesSource(target, source)
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

// ApplyOneUnlink performs the unlink action for a single agent. It only
// removes symlinks pointing at the canonical source path; real files and
// directories, and symlinks pointing elsewhere, are preserved.
func ApplyOneUnlink(repoRoot string, spec SpecLike) Result {
	target := filepath.Join(repoRoot, filepath.FromSlash(spec.SpecProjectPath()))
	info, lstatErr := os.Lstat(target)
	if errors.Is(lstatErr, os.ErrNotExist) {
		return Result{Spec: spec, Status: StatusAbsent}
	}
	if lstatErr != nil {
		return Result{Spec: spec, Status: StatusError, Detail: lstatErr.Error()}
	}
	if info.Mode()&os.ModeSymlink == 0 {
		return Result{Spec: spec, Status: StatusSkippedNotManaged, Detail: notManagedDetail(spec.SpecKind())}
	}
	source := absoluteSource(repoRoot, spec)
	matches, currentTarget, err := LinkMatchesSource(target, source)
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

// createLink resolves the relative source path and creates a symlink at
// the project path. Parent directories are created on demand for both
// directory-kind and file-kind bindings.
func createLink(repoRoot string, spec SpecLike, target string) Result {
	if mkErr := os.MkdirAll(filepath.Dir(target), 0o755); mkErr != nil {
		return Result{Spec: spec, Status: StatusError, Detail: "create parent directory: " + mkErr.Error()}
	}
	source := absoluteSource(repoRoot, spec)
	relativeSource, relErr := filepath.Rel(filepath.Dir(target), source)
	if relErr != nil {
		return Result{Spec: spec, Status: StatusError, Detail: "compute relative source: " + relErr.Error()}
	}
	if symErr := os.Symlink(relativeSource, target); symErr != nil {
		return Result{Spec: spec, Status: StatusError, Detail: SymlinkErrorDetail(symErr)}
	}
	return Result{Spec: spec, Status: StatusCreated, Detail: "-> " + filepath.ToSlash(relativeSource)}
}

// absoluteSource returns the OS-native absolute path of the canonical
// source for a spec.
func absoluteSource(repoRoot string, spec SpecLike) string {
	return filepath.Join(repoRoot, filepath.FromSlash(spec.SpecSourcePath()))
}

// conflictDetail returns the detail message used when a non-symlink path
// blocks linking. Directory- and file-kind bindings phrase the conflict
// slightly differently so users can spot which kind of object survived.
func conflictDetail(kind Kind) string {
	if kind == KindFile {
		return "real file exists; resolve manually"
	}
	return "real path exists; resolve manually"
}

// notManagedDetail returns the detail message used by ApplyOneUnlink when
// the target is a real directory or file. The wording mirrors
// conflictDetail so the table column reads consistently across resources.
func notManagedDetail(kind Kind) string {
	if kind == KindFile {
		return "real file; not removed"
	}
	return "real path; not removed"
}
