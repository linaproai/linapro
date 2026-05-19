// This file implements the image.build command for staging image build artifacts.

package main

import (
	"context"

	"linactl/internal/imagebuilder"
)

// runImageBuild stages image build artifacts without running docker build.
func runImageBuild(ctx context.Context, a *app, input commandInput) error {
	if err := runBuild(ctx, a, input); err != nil {
		return err
	}
	return imagebuilder.RunWithOutput(ctx, a.root, input, func(runCtx context.Context, dir string, name string, args ...string) error {
		return a.runCommand(runCtx, commandOptions{Dir: dir}, name, args...)
	}, a.stdout, a.stderr, "--build-only")
}
