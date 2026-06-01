// This file implements the db.upgrade command for replaying host upgrade SQL.

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

// runUpgrade replays host SQL assets against the configured database after
// explicit confirmation.
func runUpgrade(ctx context.Context, a *app, input commandInput) error {
	if input.Get("confirm") != "upgrade" {
		return errors.New("db.upgrade requires explicit confirmation: linactl db.upgrade confirm=upgrade")
	}

	var output bytes.Buffer
	err := a.runCommand(ctx, commandOptions{
		Dir:    filepath.Join(a.root, "apps", "lina-core"),
		Stdout: io.MultiWriter(a.stdout, &output),
		Stderr: io.MultiWriter(a.stderr, &output),
	}, "go", "run", "main.go", "upgrade", "--confirm=upgrade", "--sql-source=local")
	if err != nil {
		text := strings.ToLower(output.String())
		if toolutil.IsConnectionFailure(text) {
			fmt.Fprintln(a.stderr, "PostgreSQL is not ready. Start PostgreSQL first.")
			fmt.Fprintln(a.stderr, "Local example: docker run --name linapro-postgres -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=linapro -p 5432:5432 -d postgres:14-alpine")
		}
		return err
	}
	fmt.Fprintln(a.stdout, "Database upgrade complete")
	return nil
}
