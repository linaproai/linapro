// This file verifies the small helpers coupled to the capability model.

package capmodel

import (
	"testing"

	"lina-core/pkg/bizerr"
)

func TestPageRequestNormalize(t *testing.T) {
	tests := []struct {
		name     string
		request  PageRequest
		pageNum  int
		pageSize int
	}{
		{
			name:     "defaults",
			request:  PageRequest{},
			pageNum:  1,
			pageSize: 20,
		},
		{
			name:     "limit fallback",
			request:  PageRequest{Limit: 15},
			pageNum:  1,
			pageSize: 15,
		},
		{
			name:     "hard limit",
			request:  PageRequest{PageNum: 2, PageSize: 999},
			pageNum:  2,
			pageSize: 200,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pageNum, pageSize := tc.request.Normalize()
			if pageNum != tc.pageNum || pageSize != tc.pageSize {
				t.Fatalf("unexpected normalized page: got %d/%d want %d/%d", pageNum, pageSize, tc.pageNum, tc.pageSize)
			}
		})
	}
}

func TestParsePositiveInt64Strings(t *testing.T) {
	ids, err := ParsePositiveInt64Strings([]string{" 7 ", "7", "9"})
	if err != nil {
		t.Fatalf("parse ids: %v", err)
	}
	if len(ids) != 2 || ids[0] != 7 || ids[1] != 9 {
		t.Fatalf("unexpected ids: %v", ids)
	}

	if _, err = ParsePositiveInt64Strings([]string{"0"}); !bizerr.Is(err, CodeCapabilityDenied) {
		t.Fatalf("expected denied error for non-positive id, got %v", err)
	}

	if _, err = ParsePositiveInt64Strings([]string{"bad"}); err == nil {
		t.Fatalf("expected parse error for invalid integer input, got %v", err)
	}
}

func TestParseInt64IDs(t *testing.T) {
	var invalid []string
	parsed, requested := ParseInt64IDs([]string{"3", " 3 ", "5", "bad"}, func(id string) {
		invalid = append(invalid, id)
	})
	if len(parsed) != 2 || parsed[0] != 3 || parsed[1] != 5 {
		t.Fatalf("unexpected parsed ids: %v", parsed)
	}
	if len(requested) != 2 || requested[3] != "3" || requested[5] != "5" {
		t.Fatalf("unexpected requested map: %#v", requested)
	}
	if len(invalid) != 1 || invalid[0] != "bad" {
		t.Fatalf("unexpected invalid ids: %v", invalid)
	}
}
