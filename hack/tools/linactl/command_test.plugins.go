// This file implements the test.plugins command for official plugin E2E tests.

package main

import "context"

// runTestPlugins starts the official plugin Playwright E2E test suite.
func runTestPlugins(ctx context.Context, a *app, _ commandInput) error {
	return runTest(ctx, a, commandInput{Params: map[string]string{"scope": "plugins"}})
}
