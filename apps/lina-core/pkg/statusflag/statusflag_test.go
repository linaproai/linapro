// This file verifies the small 0/1 flag contracts remain explicit per business
// meaning while preserving their serialized integer values.

package statusflag

import "testing"

// TestFlagIntValues verifies the published flags keep the stable 0/1 contract
// used by API DTOs and persisted rows.
func TestFlagIntValues(t *testing.T) {
	if Disabled.Int() != 0 || EnabledValue.Int() != 1 {
		t.Fatalf("unexpected enabled flag values: disabled=%d enabled=%d", Disabled.Int(), EnabledValue.Int())
	}
	if Hidden.Int() != 0 || Visible.Int() != 1 {
		t.Fatalf("unexpected visibility flag values: hidden=%d visible=%d", Hidden.Int(), Visible.Int())
	}
	if Uninstalled.Int() != 0 || Installed.Int() != 1 {
		t.Fatalf("unexpected installation flag values: uninstalled=%d installed=%d", Uninstalled.Int(), Installed.Int())
	}
	if Unread.Int() != 0 || Read.Int() != 1 {
		t.Fatalf("unexpected read-state flag values: unread=%d read=%d", Unread.Int(), Read.Int())
	}
	if No.Int() != 0 || Yes.Int() != 1 {
		t.Fatalf("unexpected yes/no flag values: no=%d yes=%d", No.Int(), Yes.Int())
	}
}

// TestFlagSupportChecks verifies unsupported integer values are not accepted as
// published flag constants.
func TestFlagSupportChecks(t *testing.T) {
	if !EnabledValue.IsSupported() || Enabled(2).IsSupported() {
		t.Fatal("enabled support check returned an unexpected result")
	}
	if !Visible.IsSupported() || Visibility(2).IsSupported() {
		t.Fatal("visibility support check returned an unexpected result")
	}
	if !Installed.IsSupported() || Installation(2).IsSupported() {
		t.Fatal("installation support check returned an unexpected result")
	}
	if !Read.IsSupported() || ReadState(2).IsSupported() {
		t.Fatal("read-state support check returned an unexpected result")
	}
	if !Yes.IsSupported() || YesNo(2).IsSupported() {
		t.Fatal("yes/no support check returned an unexpected result")
	}
}
