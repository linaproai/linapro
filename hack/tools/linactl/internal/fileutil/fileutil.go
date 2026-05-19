// Package fileutil provides repository filesystem helpers for linactl commands
// and internal components. It keeps copy, hashing, discovery, and guarded
// removal logic out of command entry files while preserving cross-platform
// standard-library implementations.
package fileutil

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
)

// DiscoverRepoRoot searches upward for the LinaPro repository root.
func DiscoverRepoRoot() (string, error) {
	current, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if FileExists(filepath.Join(current, "go.work")) && DirExists(filepath.Join(current, "apps", "lina-core")) {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return "", errors.New("cannot find LinaPro repository root")
}

// CopyFile copies one regular file and creates the destination parent directory.
func CopyFile(src string, dst string) error {
	input, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open %s: %w", src, err)
	}
	defer func() {
		if closeErr := input.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "warning: close %s: %v\n", src, closeErr)
		}
	}()

	if err = os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("create parent directory for %s: %w", dst, err)
	}
	output, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("create %s: %w", dst, err)
	}
	defer func() {
		if closeErr := output.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "warning: close %s: %v\n", dst, closeErr)
		}
	}()

	if _, err = io.Copy(output, input); err != nil {
		return fmt.Errorf("copy %s to %s: %w", src, dst, err)
	}
	return nil
}

// CopyDirContents recursively copies the contents of one directory.
func CopyDirContents(src string, dst string) error {
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return fmt.Errorf("create %s: %w", dst, err)
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("read %s: %w", src, err)
	}
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			if err = CopyDirContents(srcPath, dstPath); err != nil {
				return err
			}
			continue
		}
		if entry.Type()&os.ModeSymlink != 0 {
			target, readErr := os.Readlink(srcPath)
			if readErr != nil {
				return fmt.Errorf("read symlink %s: %w", srcPath, readErr)
			}
			if err = os.Symlink(target, dstPath); err != nil {
				return fmt.Errorf("create symlink %s: %w", dstPath, err)
			}
			continue
		}
		if err = CopyFile(srcPath, dstPath); err != nil {
			return err
		}
	}
	return nil
}

// CopyPluginDir recursively copies a plugin directory while excluding nested
// Git metadata from the source checkout.
func CopyPluginDir(src string, dst string) error {
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return fmt.Errorf("create %s: %w", dst, err)
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("read %s: %w", src, err)
	}
	for _, entry := range entries {
		if entry.Name() == ".git" {
			continue
		}
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			if err = CopyPluginDir(srcPath, dstPath); err != nil {
				return err
			}
			continue
		}
		if entry.Type()&os.ModeSymlink != 0 {
			target, readErr := os.Readlink(srcPath)
			if readErr != nil {
				return fmt.Errorf("read symlink %s: %w", srcPath, readErr)
			}
			if err = os.Symlink(target, dstPath); err != nil {
				return fmt.Errorf("create symlink %s: %w", dstPath, err)
			}
			continue
		}
		if err = CopyFile(srcPath, dstPath); err != nil {
			return err
		}
	}
	return nil
}

// HashDirectory calculates a stable content hash for regular files and
// symlinks below root while ignoring Git metadata.
func HashDirectory(root string) (string, error) {
	var entries []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == root {
			return nil
		}
		if entry.IsDir() && entry.Name() == ".git" {
			return filepath.SkipDir
		}
		if entry.IsDir() {
			return nil
		}
		relative, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		entries = append(entries, filepath.ToSlash(relative))
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("walk %s: %w", root, err)
	}
	sort.Strings(entries)

	hasher := sha256.New()
	for _, relative := range entries {
		path := filepath.Join(root, filepath.FromSlash(relative))
		info, statErr := os.Lstat(path)
		if statErr != nil {
			return "", fmt.Errorf("stat %s: %w", path, statErr)
		}
		if _, writeErr := io.WriteString(hasher, relative+"\n"); writeErr != nil {
			return "", fmt.Errorf("hash path %s: %w", path, writeErr)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			target, readErr := os.Readlink(path)
			if readErr != nil {
				return "", fmt.Errorf("read symlink %s: %w", path, readErr)
			}
			if _, writeErr := io.WriteString(hasher, "symlink:"+target+"\n"); writeErr != nil {
				return "", fmt.Errorf("hash symlink %s: %w", path, writeErr)
			}
			continue
		}
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return "", fmt.Errorf("read %s: %w", path, readErr)
		}
		if _, writeErr := hasher.Write(content); writeErr != nil {
			return "", fmt.Errorf("hash %s: %w", path, writeErr)
		}
		if _, writeErr := io.WriteString(hasher, "\n"); writeErr != nil {
			return "", fmt.Errorf("hash separator %s: %w", path, writeErr)
		}
	}
	return "sha256:" + hex.EncodeToString(hasher.Sum(nil)), nil
}

// RemoveDirectoryIfExists removes a directory tree when present and reports
// non-directory paths as an error.
func RemoveDirectoryIfExists(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("stat %s: %w", path, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s exists but is not a directory", path)
	}
	if err = os.RemoveAll(path); err != nil {
		return fmt.Errorf("remove %s: %w", path, err)
	}
	return nil
}

// FileExists reports whether a path exists and is a regular file.
func FileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// DirExists reports whether a path exists and is a directory.
func DirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
