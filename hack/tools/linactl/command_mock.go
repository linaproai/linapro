// This file implements the mock command for optional demo data loading.

package main

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
)

// runMock loads optional mock data after explicit confirmation.
func runMock(ctx context.Context, a *app, input commandInput) error {
	if input.Get("confirm") != "mock" {
		return errors.New("mock requires explicit confirmation: linactl mock confirm=mock")
	}
	err := a.runCommand(ctx, commandOptions{Dir: filepath.Join(a.root, "apps", "lina-core")}, "go", "run", "main.go", "mock", "--confirm=mock", "--sql-source=local")
	if err != nil {
		return err
	}
	fmt.Fprintln(a.stdout, "Mock data load complete")
	return nil
}
