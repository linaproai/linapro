// Package main boots the runtime i18n verification tool and keeps the
// repository-root package limited to process startup.
package main

import (
	"fmt"
	"os"
)

// main delegates all command parsing and execution to the tool implementation.
func main() {
	exitCode, err := run(os.Args[1:], os.Stdout)
	if err != nil {
		if _, writeErr := fmt.Fprintln(os.Stderr, err); writeErr != nil {
			os.Exit(1)
		}
		os.Exit(1)
	}
	os.Exit(exitCode)
}
