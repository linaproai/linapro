//go:build !wasip1

// This file verifies host-build guest host-service stubs return the shared
// unavailable sentinel instead of requiring each dynamic plugin to define its
// own unsupported service implementation.

package pluginbridge

import (
	"bytes"
	"testing"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/lockcap"
	"lina-core/pkg/plugin/capability/notifycap"
	"lina-core/pkg/plugin/capability/storagecap"
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
			_, err := Storage().Put(t.Context(), storagecap.PutInput{
				Path:        "demo.txt",
				Body:        bytes.NewReader([]byte("demo")),
				Size:        4,
				ContentType: "text/plain",
				Overwrite:   true,
			})
			return err
		}},
		{name: "storage provider statuses", run: func() error {
			_, err := Storage().ProviderStatuses(t.Context())
			return err
		}},
		{name: "network", run: func() error {
			_, err := Network().Request("https://example.com", &protocol.HostServiceNetworkRequest{})
			return err
		}},
		{name: "cache", run: func() error {
			_, _, err := Cache().Get(t.Context(), "demo", "key")
			return err
		}},
		{name: "lock", run: func() error {
			_, err := Lock().Acquire(t.Context(), lockcap.AcquireInput{Name: "demo", Lease: time.Second})
			return err
		}},
		{name: "config", run: func() error {
			_, err := New().Plugins().Config().String(t.Context(), "demo.greeting", "")
			return err
		}},
		{name: "jobs register", run: func() error {
			return NewDeclarations().Jobs().Register(&protocol.JobContract{Name: "demo", Pattern: "@every 1m", RequestType: "DemoJobReq"})
		}},
		{name: "notify", run: func() error {
			_, err := New().Notifications().Send(t.Context(), capmodel.CapabilityContext{}, notifycap.SendInput{ChannelKey: "demo"})
			return err
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
