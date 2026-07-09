// This file implements the dao command for GoFrame DAO generation.

package main

import "context"

// runDao runs the embedded GoFrame gen dao command in the selected backend
// directory without requiring an external gf binary.
func runDao(ctx context.Context, a *app, input commandInput) error {
	target, err := resolveGoFrameTargetInput(a.root, input, true)
	if err != nil {
		return err
	}
	return runGoFrameTarget(ctx, a, target, "gen", "dao")
}
