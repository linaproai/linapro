// This file implements the image command for production Docker image builds.

package main

import (
	"context"

	"linactl/internal/imagebuilder"
)

// runImage builds and optionally pushes a production Docker image.
func runImage(ctx context.Context, a *app, input commandInput) error {
	run := func(runCtx context.Context, dir string, name string, args ...string) error {
		return a.runCommand(runCtx, commandOptions{Dir: dir}, name, args...)
	}
	if err := imagebuilder.RunWithOutput(ctx, a.root, input, run, a.stdout, a.stderr, "--preflight"); err != nil {
		return err
	}
	if err := runBuild(ctx, a, input); err != nil {
		return err
	}
	return imagebuilder.RunWithOutput(ctx, a.root, input, run, a.stdout, a.stderr)
}
