// Package main implements LinaPro's cross-platform development command entrypoint.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
)

// main runs one linactl command invocation.
func main() {
	application := newApp(os.Stdout, os.Stderr, os.Stdin)
	if err := application.run(context.Background(), os.Args[1:]); err != nil {
		if errors.Is(err, errHelpRequested) {
			return
		}
		fmt.Fprintf(application.stderr, "linactl: %v\n", err)
		os.Exit(1)
	}
}
