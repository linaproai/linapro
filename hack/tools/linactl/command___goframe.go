// This file implements the hidden __goframe command for embedded GoFrame CLI dispatch.

package main

import (
	"context"
	"fmt"

	"linactl/internal/goframecli"
)

// runEmbeddedGoFrame executes a whitelisted embedded GoFrame code generation
// command. It is intentionally hidden from help output and only supports the
// code generation paths used by linactl ctrl and linactl dao.
func runEmbeddedGoFrame(ctx context.Context, a *app, input commandInput) error {
	configDir := input.Get("config-dir")
	for key := range input.Params {
		if key == "config_dir" {
			continue
		}
		return fmt.Errorf("embedded GoFrame only supports config-dir plus positional commands: gen ctrl or gen dao")
	}
	if configDir == "" {
		return fmt.Errorf("embedded GoFrame requires config-dir")
	}
	return goframecli.RunEmbedded(ctx, configDir, input.Args...)
}
