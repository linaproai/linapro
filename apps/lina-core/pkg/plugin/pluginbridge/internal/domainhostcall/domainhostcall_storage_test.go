// This file verifies storage host-service client helper behavior that remains
// internal to the domainhostcall implementation.

package domainhostcall

import (
	"testing"

	"lina-core/pkg/plugin/capability/storagecap"
)

// TestStorageListEffectiveLimit verifies guest storage list responses expose
// the same bounded limit semantics as storagecap.Service implementations.
func TestStorageListEffectiveLimit(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   int
		want int
	}{
		{name: "default", in: 0, want: storagecap.DefaultListLimit},
		{name: "negative default", in: -1, want: storagecap.DefaultListLimit},
		{name: "bounded", in: 10, want: 10},
		{name: "max", in: storagecap.MaxListLimit + 1, want: storagecap.MaxListLimit},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			if got := storageListEffectiveLimit(c.in); got != c.want {
				t.Fatalf("storageListEffectiveLimit(%d) = %d, want %d", c.in, got, c.want)
			}
		})
	}
}
