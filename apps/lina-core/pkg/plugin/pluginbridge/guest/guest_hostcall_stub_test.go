//go:build !wasip1

// This file verifies host-build guest host-service stubs return the shared
// unavailable sentinel instead of requiring each dynamic plugin to define its
// own unsupported service implementation.

package guest

import (
	"testing"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// TestHostCallStubsReturnUnavailable verifies representative host service
// clients fail with the shared non-WASI sentinel during ordinary Go tests.
func TestHostCallStubsReturnUnavailable(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		run  func() error
	}{
		{name: "runtime", run: func() error {
			_, err := Runtime().Now()
			return err
		}},
		{name: "storage", run: func() error {
			_, err := Storage().PutText("demo.txt", "demo", "text/plain", true)
			return err
		}},
		{name: "network", run: func() error {
			_, err := Network().Request("https://example.com", &protocol.HostServiceNetworkRequest{})
			return err
		}},
		{name: "cache", run: func() error {
			_, _, err := Cache().Get("demo", "key")
			return err
		}},
		{name: "lock", run: func() error {
			_, err := Lock().Acquire("demo", 1000)
			return err
		}},
		{name: "config", run: func() error {
			_, _, err := New().Plugins().Config().String("demo.greeting")
			return err
		}},
		{name: "notify", run: func() error {
			_, err := Notify().Send("demo", &protocol.HostServiceNotifySendRequest{})
			return err
		}},
		{name: "cron", run: func() error {
			return Cron().Register(&protocol.CronContract{})
		}},
		{name: "host runtime", run: func() error {
			_, _, err := HostConfig().Bool("i18n.enabled")
			return err
		}},
		{name: "manifest", run: func() error {
			_, _, err := Manifest().GetText("metadata.yaml")
			return err
		}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			if err := c.run(); !gerror.Is(err, ErrHostCallsUnavailable) {
				t.Fatalf("expected ErrHostCallsUnavailable, got %v", err)
			}
		})
	}
}
