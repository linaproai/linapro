// Package servermon exposes a narrowed host server-monitor service contract to
// source plugins.
package servermon

import (
	"context"
	"time"

	internalservermon "lina-core/internal/service/servermon"
)

// NodeMonitorData aliases the host server-monitor node snapshot type.
type NodeMonitorData = internalservermon.NodeMonitorData

// DBInfo aliases the host database-monitor snapshot type.
type DBInfo = internalservermon.DBInfo

// Service defines the server-monitor operations published to source plugins.
type Service interface {
	// CollectAndStore collects metrics and stores them in the database.
	CollectAndStore(ctx context.Context)
	// GetLatest returns the latest monitor records for each node.
	GetLatest(ctx context.Context, nodeName string) ([]*NodeMonitorData, error)
	// GetDBInfo collects database metrics on demand.
	GetDBInfo(ctx context.Context) *DBInfo
	// CleanupStale deletes monitor records older than the provided threshold.
	CleanupStale(ctx context.Context, threshold time.Duration) (int64, error)
}

// serviceAdapter bridges the internal server-monitor service into the published plugin contract.
type serviceAdapter struct {
	service internalservermon.Service
}

// New creates and returns the published server-monitor service adapter.
func New() Service {
	return &serviceAdapter{service: internalservermon.New()}
}

// CollectAndStore collects metrics and stores them in the database.
func (s *serviceAdapter) CollectAndStore(ctx context.Context) {
	if s == nil || s.service == nil {
		return
	}
	s.service.CollectAndStore(ctx)
}

// GetLatest returns the latest monitor records for each node.
func (s *serviceAdapter) GetLatest(ctx context.Context, nodeName string) ([]*NodeMonitorData, error) {
	if s == nil || s.service == nil {
		return []*NodeMonitorData{}, nil
	}
	return s.service.GetLatest(ctx, nodeName)
}

// GetDBInfo collects database metrics on demand.
func (s *serviceAdapter) GetDBInfo(ctx context.Context) *DBInfo {
	if s == nil || s.service == nil {
		return &DBInfo{}
	}
	return s.service.GetDBInfo(ctx)
}

// CleanupStale deletes monitor records older than the provided threshold.
func (s *serviceAdapter) CleanupStale(ctx context.Context, threshold time.Duration) (int64, error) {
	if s == nil || s.service == nil {
		return 0, nil
	}
	return s.service.CleanupStale(ctx, threshold)
}
