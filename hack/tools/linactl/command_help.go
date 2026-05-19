// This file implements the help command and command usage rendering.

package main

import (
	"context"
	"fmt"
	"io"
	"sort"
)

// runHelp prints the top-level help output.
func runHelp(_ context.Context, a *app, input commandInput) error {
	return a.printHelp(input.HasBool("all"))
}

// printHelp writes the command overview and platform-specific entry examples.
func (a *app) printHelp(includeInternal bool) error {
	specs := commandRegistry()
	names := make([]string, 0, len(specs))
	for name, spec := range specs {
		if spec.Internal && !includeInternal {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)

	maxName := 0
	for _, name := range names {
		if len(name) > maxName {
			maxName = len(name)
		}
	}

	fmt.Fprintln(a.stdout, "Usage: linactl <command> [key=value] [--flag=value]")
	fmt.Fprintln(a.stdout)
	fmt.Fprintln(a.stdout, "Windows:")
	fmt.Fprintln(a.stdout, "  cmd.exe:     make help")
	fmt.Fprintln(a.stdout, "  PowerShell:  .\\make help")
	fmt.Fprintln(a.stdout)
	fmt.Fprintln(a.stdout, "Linux/macOS:")
	fmt.Fprintln(a.stdout, "  make help")
	fmt.Fprintln(a.stdout)
	fmt.Fprintln(a.stdout, "Available commands:")
	for _, name := range names {
		spec := specs[name]
		fmt.Fprintf(a.stdout, "  %-*s  %s\n", maxName, spec.Name, spec.Description)
	}
	return nil
}

// printCommandHelp writes usage for one command.
func printCommandHelp(out io.Writer, spec commandSpec) {
	fmt.Fprintf(out, "Usage: %s\n\n%s\n", spec.Usage, spec.Description)
}
