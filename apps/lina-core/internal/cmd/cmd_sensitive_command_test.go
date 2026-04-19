// This file verifies explicit confirmation is required before sensitive host
// database commands execute.

package cmd

import (
	"strings"
	"testing"
)

// TestRequireCommandConfirmation verifies sensitive command confirmation tokens
// are enforced for both init and mock operations.
func TestRequireCommandConfirmation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		commandName    string
		confirmValue   string
		wantErr        bool
		wantSubstrings []string
	}{
		{
			name:         "init accepts matching confirmation",
			commandName:  initCommandName,
			confirmValue: initCommandName,
		},
		{
			name:         "mock accepts matching confirmation",
			commandName:  mockCommandName,
			confirmValue: mockCommandName,
		},
		{
			name:         "init rejects missing confirmation",
			commandName:  initCommandName,
			confirmValue: "",
			wantErr:      true,
			wantSubstrings: []string{
				"命令 init 涉及敏感数据库操作",
				makeConfirmationExample(initCommandName),
				goRunConfirmationExample(initCommandName),
			},
		},
		{
			name:         "mock rejects wrong confirmation",
			commandName:  mockCommandName,
			confirmValue: initCommandName,
			wantErr:      true,
			wantSubstrings: []string{
				"命令 mock 涉及敏感数据库操作",
				makeConfirmationExample(mockCommandName),
				goRunConfirmationExample(mockCommandName),
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := requireCommandConfirmation(tt.commandName, tt.confirmValue)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for command %q", tt.commandName)
				}
				for _, substring := range tt.wantSubstrings {
					if !strings.Contains(err.Error(), substring) {
						t.Fatalf("expected error %q to contain %q", err.Error(), substring)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}
