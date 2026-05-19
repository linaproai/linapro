// This file implements the test.host command for host-owned E2E tests.

package main

import "context"

// runTestHost starts the host-owned Playwright E2E test suite.
func runTestHost(ctx context.Context, a *app, _ commandInput) error {
	return runTest(ctx, a, commandInput{Params: map[string]string{"scope": "host"}})
}
