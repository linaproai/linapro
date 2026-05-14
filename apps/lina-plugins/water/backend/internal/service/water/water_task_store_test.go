// This file tests the watermark task status store.

package water

import (
	"context"
	"testing"
)

// TestTaskStoreEvictsOldest verifies bounded FIFO task status retention.
func TestTaskStoreEvictsOldest(t *testing.T) {
	store := newTaskStore(1)
	store.create("task-1", SubmitSnapInput{Tenant: "tenant-a", DeviceId: "device-a"})
	store.create("task-2", SubmitSnapInput{Tenant: "tenant-b", DeviceId: "device-b"})
	if _, err := store.get(context.Background(), "task-1"); err == nil {
		t.Fatal("expected task-1 to be evicted")
	}
	task, err := store.get(context.Background(), "task-2")
	if err != nil {
		t.Fatalf("expected task-2 to exist: %v", err)
	}
	if task.Tenant != "tenant-b" {
		t.Fatalf("expected tenant-b, got %q", task.Tenant)
	}
}
