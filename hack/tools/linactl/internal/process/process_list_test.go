//go:build linux || darwin

package process

import (
	"os"
	"testing"
)

func TestListIncludesCurrentProcess(t *testing.T) {
	infos, err := List()
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	currentPID := os.Getpid()
	for _, info := range infos {
		if info.PID == currentPID && len(info.Args) > 0 {
			return
		}
	}
	t.Fatalf("List did not include current process %d", currentPID)
}
