// This file executes external commands from the repository root.

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// commandRunner executes external tools from the repository root.
type commandRunner struct {
	Root    string
	Verbose bool
}

// Run executes one external command in a repository-relative directory.
func (r commandRunner) Run(dir string, env []string, name string, args ...string) error {
	workingDir := filepath.Join(r.Root, dir)
	cmd := exec.Command(name, args...)
	cmd.Dir = workingDir
	if len(env) > 0 {
		cmd.Env = append(os.Environ(), env...)
	}
	display := name + " " + strings.Join(args, " ")
	if r.Verbose {
		fmt.Printf("+ %s\n", display)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("run %s in %s failed: %w\n%s", display, workingDir, err, string(output))
	}
	return nil
}
