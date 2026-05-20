// This file implements the i18n.check command for runtime i18n governance.

package main

import (
	"context"

	"linactl/internal/runtimei18n"
)

// runI18nCheck invokes all runtime i18n governance checks.
func runI18nCheck(_ context.Context, a *app, _ commandInput) error {
	return runtimei18n.RunCheck(a.root, a.stdout)
}
