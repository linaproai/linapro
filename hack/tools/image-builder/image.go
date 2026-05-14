// This file builds Docker image references and Docker CLI arguments.

package main

import (
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// deriveGitTag returns a git-derived image tag with latest as the final fallback.
func deriveGitTag(repoRoot string) (string, error) {
	cmd := exec.Command("git", "describe", "--tags", "--always", "--dirty", "--match", "v*")
	cmd.Dir = repoRoot
	output, err := cmd.Output()
	if err != nil {
		return "latest", nil
	}
	tag := strings.TrimSpace(string(output))
	if tag == "" {
		return "latest", nil
	}
	return tag, nil
}

// validateImageBuildRequest verifies Docker build settings before invoking Docker.
func validateImageBuildRequest(image imageConfig, build buildConfig) error {
	if image.Push {
		if strings.TrimSpace(image.Registry) == "" {
			return errors.New("push=true requires image.registry in hack/config.yaml, registry=<prefix>, or LINAPRO_IMAGE_REGISTRY")
		}
	}
	for _, target := range build.Targets {
		if target.OS != "linux" {
			return fmt.Errorf("Docker image builds require linux target platforms, got %s", target.String())
		}
	}
	if build.MultiPlatform() && !image.Push {
		return errors.New("multi-platform Docker image builds require push=1 so Docker buildx can publish a usable manifest")
	}
	return nil
}

// buildDockerImage runs docker build or buildx with the configured image platforms.
func buildDockerImage(repoRoot string, image imageConfig, build buildConfig, runner commandRunner, imageRef string) error {
	if build.MultiPlatform() {
		args := buildxDockerArgs(repoRoot, image, build, imageRef)
		fmt.Printf("Building multi-platform Docker image: %s (%s)\n", imageRef, build.Platform)
		dockerRunner := runner
		dockerRunner.Verbose = true
		return dockerRunner.Run(".", nil, "docker", args...)
	}
	target := build.Targets[0]
	args := dockerBuildArgs(repoRoot, image, target, imageRef)
	fmt.Printf("Building Docker image: %s\n", imageRef)
	dockerRunner := runner
	dockerRunner.Verbose = true
	return dockerRunner.Run(".", nil, "docker", args...)
}

// buildxDockerArgs returns Docker buildx arguments for multi-platform publishing.
func buildxDockerArgs(repoRoot string, image imageConfig, build buildConfig, imageRef string) []string {
	return []string{
		"buildx",
		"build",
		"--platform", build.Platform,
		"--build-arg", "BASE_IMAGE=" + image.BaseImage,
		"-f", filepath.Join(repoRoot, image.Dockerfile),
		"-t", imageRef,
		"--push",
		".",
	}
}

// dockerBuildArgs returns Docker build arguments for one local image platform.
func dockerBuildArgs(repoRoot string, image imageConfig, target targetPlatform, imageRef string) []string {
	return []string{
		"build",
		"--platform", target.String(),
		"--build-arg", "BASE_IMAGE=" + image.BaseImage,
		"--build-arg", "TARGETOS=" + target.OS,
		"--build-arg", "TARGETARCH=" + target.Arch,
		"-f", filepath.Join(repoRoot, image.Dockerfile),
		"-t", imageRef,
		".",
	}
}

// buildImageRef composes the final Docker image reference.
func buildImageRef(cfg imageConfig) string {
	name := cfg.Name + ":" + cfg.Tag
	if strings.TrimSpace(cfg.Registry) == "" {
		return name
	}
	return strings.Trim(cfg.Registry, "/") + "/" + name
}
