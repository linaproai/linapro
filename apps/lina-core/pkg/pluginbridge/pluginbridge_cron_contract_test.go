// This file verifies dynamic-plugin cron contract normalization and protocol
// helper construction.

package pluginbridge

import "testing"

// TestNormalizeCronScope verifies raw scope values normalize to canonical
// plugin bridge enums.
func TestNormalizeCronScope(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    string
		expected CronScope
	}{
		{
			name:     "default all-node",
			input:    "",
			expected: CronScopeAllNode,
		},
		{
			name:     "master only",
			input:    " master_only ",
			expected: CronScopeMasterOnly,
		},
		{
			name:     "case insensitive all-node",
			input:    "ALL_NODE",
			expected: CronScopeAllNode,
		},
		{
			name:     "invalid",
			input:    "edge_only",
			expected: "",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if actual := NormalizeCronScope(testCase.input); actual != testCase.expected {
				t.Fatalf("NormalizeCronScope(%q): got %q want %q", testCase.input, actual, testCase.expected)
			}
		})
	}
}

// TestNormalizeCronConcurrency verifies raw concurrency values normalize to
// canonical plugin bridge enums.
func TestNormalizeCronConcurrency(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    string
		expected CronConcurrency
	}{
		{
			name:     "default singleton",
			input:    "",
			expected: CronConcurrencySingleton,
		},
		{
			name:     "parallel",
			input:    " parallel ",
			expected: CronConcurrencyParallel,
		},
		{
			name:     "case insensitive singleton",
			input:    "SINGLETON",
			expected: CronConcurrencySingleton,
		},
		{
			name:     "invalid",
			input:    "serial",
			expected: "",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if actual := NormalizeCronConcurrency(testCase.input); actual != testCase.expected {
				t.Fatalf(
					"NormalizeCronConcurrency(%q): got %q want %q",
					testCase.input,
					actual,
					testCase.expected,
				)
			}
		})
	}
}

// TestBuildPluginHandlerRefs verifies shared handler reference helpers return
// stable protocol strings.
func TestBuildPluginHandlerRefs(t *testing.T) {
	t.Parallel()

	handlerRef, err := BuildPluginHandlerRef("plugin-demo", "sync")
	if err != nil {
		t.Fatalf("expected generic handler ref to build, got error: %v", err)
	}
	if handlerRef != "plugin:plugin-demo/sync" {
		t.Fatalf("expected generic handler ref, got %s", handlerRef)
	}

	cronRef, err := BuildPluginCronHandlerRef("plugin-demo", "heartbeat")
	if err != nil {
		t.Fatalf("expected cron handler ref to build, got error: %v", err)
	}
	if cronRef != "plugin:plugin-demo/cron:heartbeat" {
		t.Fatalf("expected cron handler ref, got %s", cronRef)
	}
}

// TestBuildDeclaredCronRoutePath verifies declared cron jobs derive one stable
// synthetic runtime route path.
func TestBuildDeclaredCronRoutePath(t *testing.T) {
	t.Parallel()

	if routePath := BuildDeclaredCronRoutePath(nil); routePath != DeclaredCronRouteBasePath {
		t.Fatalf("expected nil contract to use base path, got %s", routePath)
	}

	if routePath := BuildDeclaredCronRoutePath(&CronContract{Name: "heartbeat"}); routePath != "/@cron/heartbeat" {
		t.Fatalf("expected named contract route path, got %s", routePath)
	}

	if routePath := BuildDeclaredCronRoutePath(&CronContract{InternalPath: "cron-heartbeat"}); routePath != "/cron-heartbeat" {
		t.Fatalf("expected internal path normalization, got %s", routePath)
	}
}
