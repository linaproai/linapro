// This file implements shared confirmation guardrails for sensitive database
// bootstrap commands.

package cmd

import (
	"fmt"

	"github.com/gogf/gf/v2/errors/gerror"
)

// Sensitive database command names that require explicit confirmation.
const (
	initCommandName = "init"
	mockCommandName = "mock"
)

// requireCommandConfirmation validates the explicit confirmation value for a
// sensitive database command before any SQL is executed.
func requireCommandConfirmation(commandName string, confirmValue string) error {
	expectedValue := expectedCommandConfirmation(commandName)
	if confirmValue == expectedValue {
		return nil
	}
	return gerror.Newf(
		"命令 %s 涉及敏感数据库操作，必须显式确认后才能执行。请使用 %s 或 %s",
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
