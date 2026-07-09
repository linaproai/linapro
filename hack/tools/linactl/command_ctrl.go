// This file implements the ctrl command for GoFrame controller generation.

package main

import (
	"context"
	"fmt"

	"linactl/internal/goframecli"
)

// runCtrl runs the embedded GoFrame gen ctrl command in the selected backend
// directory without requiring an external gf binary.
func runCtrl(ctx context.Context, a *app, input commandInput) error {
	target, err := resolveGoFrameTargetInput(a.root, input, false)
	if err != nil {
		return err
	}
	return runGoFrameTarget(ctx, a, target, "gen", "ctrl")
}

func resolveGoFrameTargetInput(root string, input commandInput, requireConfig bool) (goframecli.Target, error) {
	if len(input.Args) > 0 {
		return goframecli.Target{}, fmt.Errorf("GoFrame code generation accepts target parameters only; use dir=<path>")
	}
	for key := range input.Params {
		if key == "dir" {
			continue
		}
		return goframecli.Target{}, fmt.Errorf("GoFrame code generation parameter %s is not supported; use dir=<path>", key)
	}
	return goframecli.ResolveTarget(root, goframecli.TargetOptions{
		Dir:           input.Get("dir"),
		DirSet:        input.Has("dir"),
		RequireConfig: requireConfig,
	})
}

func runGoFrameTarget(ctx context.Context, a *app, target goframecli.Target, args ...string) error {
	return goframecli.Run(ctx, target, a.executable, a.runCommand, args...)
}
