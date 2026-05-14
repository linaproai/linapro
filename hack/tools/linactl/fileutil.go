// This file contains repository filesystem helpers used by linactl commands.

package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// discoverRepoRoot searches upward for the LinaPro repository root.
func discoverRepoRoot() (string, error) {
	current, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if fileExists(filepath.Join(current, "go.work")) && dirExists(filepath.Join(current, "apps", "lina-core")) {
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

// copyFile copies one regular file and creates the destination parent directory.
func copyFile(src string, dst string) error {
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

// copyDirContents recursively copies the contents of one directory.
func copyDirContents(src string, dst string) error {
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
			if err = copyDirContents(srcPath, dstPath); err != nil {
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
		if err = copyFile(srcPath, dstPath); err != nil {
			return err
		}
	}
	return nil
}
