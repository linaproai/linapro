// This file validates and stages host binaries for Docker image builds.

package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// validateExistingBinary checks that the image input binary exists.
func validateExistingBinary(binaryPath string) error {
	info, err := os.Stat(binaryPath)
	if err != nil {
		return fmt.Errorf("build binary is missing: %s; run make build first", binaryPath)
	}
	if info.IsDir() {
		return fmt.Errorf("build binary path is a directory: %s", binaryPath)
	}
	fmt.Printf("✓ build binary is ready: %s\n", binaryPath)
	return nil
}

// stageImageBinary copies the standard build output into the Dockerfile input path.
func stageImageBinary(source string, target string) error {
	if sameFile(source, target) {
		fmt.Printf("✓ image binary staged: %s\n", target)
		return nil
	}
	if err := copyFile(source, target); err != nil {
		return err
	}
	fmt.Printf("✓ image binary staged: %s\n", target)
	return nil
}

// buildOutputBinaryPath returns the expected make build binary path for one target.
func buildOutputBinaryPath(repoRoot string, build buildConfig, target targetPlatform) string {
	return filepath.Join(repoRoot, buildOutputBinaryRelPath(build, target))
}

// buildOutputBinaryRelPath returns the repository-relative binary path for one target.
func buildOutputBinaryRelPath(build buildConfig, target targetPlatform) string {
	if build.MultiPlatform() {
		return filepath.Join(build.OutputDir, target.DirName(), build.BinaryName)
	}
	return filepath.Join(build.OutputDir, build.BinaryName)
}

// defaultBuildBinaryPath returns the printable primary build artifact path.
func defaultBuildBinaryPath(build buildConfig) string {
	if build.MultiPlatform() && len(build.Targets) > 0 {
		return buildOutputBinaryRelPath(build, build.Targets[0])
	}
	return filepath.Join(build.OutputDir, build.BinaryName)
}

// imageStagedBinaryPath returns the Dockerfile input binary path for one target.
func imageStagedBinaryPath(repoRoot string, build buildConfig, target targetPlatform) string {
	return filepath.Join(repoRoot, conventionImageBinaryRoot, target.OS, target.Arch, conventionImageBinaryName)
}

// sameFile reports whether two paths point to the same filesystem object.
func sameFile(source string, target string) bool {
	sourceAbs, sourceErr := filepath.Abs(source)
	targetAbs, targetErr := filepath.Abs(target)
	if sourceErr == nil && targetErr == nil && sourceAbs == targetAbs {
		return true
	}
	sourceInfo, sourceErr := os.Stat(source)
	if sourceErr != nil {
		return false
	}
	targetInfo, targetErr := os.Stat(target)
	if targetErr != nil {
		return false
	}
	return os.SameFile(sourceInfo, targetInfo)
}

// copyFile copies one regular file preserving permission bits.
func copyFile(source string, target string) error {
	if sameFile(source, target) {
		return nil
	}
	info, err := os.Stat(source)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := in.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "warning: close source file %s: %v\n", source, closeErr)
		}
	}()
	out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return err
	}
	if _, err = io.Copy(out, in); err != nil {
		if closeErr := out.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "warning: close target file %s after copy failure: %v\n", target, closeErr)
		}
		return err
	}
	return out.Close()
}
