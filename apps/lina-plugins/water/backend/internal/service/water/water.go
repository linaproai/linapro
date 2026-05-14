// Package water implements watermark snapshot processing for the water source plugin.
package water

import (
	"context"
	"sync"
)

// Service defines the water plugin service contract.
type Service interface {
	// SubmitSnap submits one asynchronous watermark snapshot task.
	SubmitSnap(ctx context.Context, in SubmitSnapInput) (*SubmitSnapOutput, error)
	// Preview synchronously renders one watermark preview.
	Preview(ctx context.Context, in PreviewInput) (*ProcessOutput, error)
	// GetTask returns one recent watermark task snapshot.
	GetTask(ctx context.Context, taskID string) (*TaskSnapshot, error)
}

// Interface compliance assertion for the default water service implementation.
var _ Service = (*serviceImpl)(nil)

// Package-level singleton keeps the in-memory queue and task store shared by controllers.
var (
	defaultService Service
	serviceOnce    sync.Once
)

// serviceImpl implements Service.
type serviceImpl struct {
	queue *taskQueue // queue executes asynchronous watermark tasks.
	store *taskStore // store keeps recent task status snapshots.
}

// New creates and returns the shared water service instance.
func New() Service {
	serviceOnce.Do(func() {
		store := newTaskStore(defaultTaskStoreCapacity)
		defaultService = &serviceImpl{
			queue: newTaskQueue(store),
			store: store,
		}
	})
	return defaultService
}
