// This file implements the in-memory watermark task status store.

package water

import (
	"container/list"
	"context"
	"sync"

	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/pkg/bizerr"
)

// taskRecord stores mutable task status in memory.
type taskRecord struct {
	TaskSnapshot
}

// taskStore keeps recent task snapshots with FIFO eviction.
type taskStore struct {
	mu       sync.RWMutex
	capacity int
	items    map[string]*list.Element
	order    *list.List
}

// newTaskStore creates one bounded task status store.
func newTaskStore(capacity int) *taskStore {
	if capacity <= 0 {
		capacity = defaultTaskStoreCapacity
	}
	return &taskStore{
		capacity: capacity,
		items:    make(map[string]*list.Element, capacity),
		order:    list.New(),
	}
}

// create inserts one queued task snapshot.
func (s *taskStore) create(taskID string, in SubmitSnapInput) {
	now := gtime.Now().String()
	record := &taskRecord{
		TaskSnapshot: TaskSnapshot{
			TaskId:      taskID,
			Status:      TaskStatusQueued,
			Message:     "等待处理",
			Tenant:      in.Tenant,
			DeviceId:    in.DeviceId,
			Source:      StrategySourceNone,
			SourceLabel: strategySourceLabel(StrategySourceNone),
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if existing, ok := s.items[taskID]; ok {
		existing.Value = record
		s.order.MoveToBack(existing)
		return
	}
	element := s.order.PushBack(record)
	s.items[taskID] = element
	for len(s.items) > s.capacity {
		front := s.order.Front()
		if front == nil {
			break
		}
		oldRecord, ok := front.Value.(*taskRecord)
		if ok {
			delete(s.items, oldRecord.TaskId)
		}
		s.order.Remove(front)
	}
}

// update mutates one task snapshot if it still exists.
func (s *taskStore) update(taskID string, mutate func(record *taskRecord)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	element, ok := s.items[taskID]
	if !ok {
		return
	}
	record, ok := element.Value.(*taskRecord)
	if !ok || record == nil {
		return
	}
	mutate(record)
	record.UpdatedAt = gtime.Now().String()
	s.order.MoveToBack(element)
}

// get returns one task snapshot by ID.
func (s *taskStore) get(_ context.Context, taskID string) (*TaskSnapshot, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	element, ok := s.items[taskID]
	if !ok {
		return nil, bizerr.NewCode(CodeWaterTaskNotFound)
	}
	record, ok := element.Value.(*taskRecord)
	if !ok || record == nil {
		return nil, bizerr.NewCode(CodeWaterTaskNotFound)
	}
	snapshot := record.TaskSnapshot
	return &snapshot, nil
}
