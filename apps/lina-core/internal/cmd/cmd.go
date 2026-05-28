// Package cmd assembles the Lina core command tree together with shared
// helpers used by host bootstrap and maintenance operations.
package cmd

import (
	"fmt"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// Main is the root command object for the Lina core CLI.
type Main struct {
	g.Meta `name:"main" root:"http"`
}

// Sensitive command names that require explicit confirmation.
const (
	initCommandName = "init"
	mockCommandName = "mock"
)

// requireCommandConfirmation validates the explicit confirmation value for a
// sensitive command before any destructive step is executed.
func requireCommandConfirmation(commandName string, confirmValue string) error {
	expectedValue := expectedCommandConfirmation(commandName)
	if confirmValue == expectedValue {
		return nil
	}
	return gerror.Newf(
		"command %s performs sensitive upgrade or database operations and requires explicit confirmation; use %s or %s",
		commandName,
		makeConfirmationExample(commandName),
		goRunConfirmationExample(commandName),
	)
}

// expectedCommandConfirmation returns the required confirmation token for the
// given command.
func expectedCommandConfirmation(commandName string) string {
	return commandName
}

// makeConfirmationExample returns the safe make invocation for the command.
func makeConfirmationExample(commandName string) string {
	return fmt.Sprintf("make %s confirm=%s", commandName, expectedCommandConfirmation(commandName))
}

// goRunConfirmationExample returns the safe go run invocation for the command.
func goRunConfirmationExample(commandName string) string {
	return fmt.Sprintf("go run main.go %s --confirm=%s", commandName, expectedCommandConfirmation(commandName))
}
