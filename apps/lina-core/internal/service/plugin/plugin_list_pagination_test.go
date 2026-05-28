// This file covers plugin management list pagination windowing.

package plugin

import "testing"

// TestPaginatePluginItems verifies plugin list pagination windows the filtered
// result correctly, defaults a non-positive page number to the first page, and
// disables paging for a non-positive page size so internal callers keep the
// full projection.
func TestPaginatePluginItems(t *testing.T) {
	items := make([]*PluginItem, 5)
	for i := range items {
		items[i] = &PluginItem{}
	}
	indexOf := func(target *PluginItem) int {
		for i, item := range items {
			if item == target {
				return i
			}
		}
		return -1
	}

	cases := []struct {
		name        string
		pageNum     int
		pageSize    int
		wantIndexes []int
	}{
		{"first page", 1, 2, []int{0, 1}},
		{"second page", 2, 2, []int{2, 3}},
		{"last partial page", 3, 2, []int{4}},
		{"out of range page", 4, 2, []int{}},
		{"non-positive page defaults to first", 0, 2, []int{0, 1}},
		{"no pagination returns all", 1, 0, []int{0, 1, 2, 3, 4}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := paginatePluginItems(items, tc.pageNum, tc.pageSize)
			if len(got) != len(tc.wantIndexes) {
				t.Fatalf("expected %d items, got %d", len(tc.wantIndexes), len(got))
			}
			for i, want := range tc.wantIndexes {
				if indexOf(got[i]) != want {
					t.Fatalf("page item %d: expected original index %d, got %d", i, want, indexOf(got[i]))
				}
			}
		})
	}
}
