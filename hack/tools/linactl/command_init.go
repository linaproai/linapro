// This file implements the init command for database initialization.

package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"linactl/internal/toolutil"
)

// runInit initializes the configured database after explicit confirmation.
func runInit(ctx context.Context, a *app, input commandInput) error {
	if input.Get("confirm") != "init" {
		return errors.New("init requires explicit confirmation: linactl init confirm=init")
	}

	args := []string{"run", "main.go", "init", "--confirm=init", "--sql-source=local"}
	if rebuild := input.Get("rebuild"); rebuild != "" {
		args = append(args, "--rebuild="+rebuild)
	}

	var output bytes.Buffer
	err := a.runCommand(ctx, commandOptions{
		Dir:    filepath.Join(a.root, "apps", "lina-core"),
		Stdout: io.MultiWriter(a.stdout, &output),
		Stderr: io.MultiWriter(a.stderr, &output),
	}, "go", args...)
	if err != nil {
		text := strings.ToLower(output.String())
		if toolutil.IsConnectionFailure(text) {
			fmt.Fprintln(a.stderr, "PostgreSQL is not ready. Start PostgreSQL first.")
			fmt.Fprintln(a.stderr, "Local example: docker run --name linapro-postgres -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=linapro -p 5432:5432 -d postgres:14-alpine")
		}
		return err
	}
	fmt.Fprintln(a.stdout, "Database initialization complete")
	return nil
}
