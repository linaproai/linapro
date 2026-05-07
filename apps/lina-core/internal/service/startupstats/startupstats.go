// Package startupstats records short-lived host startup metrics.
package startupstats

import (
	"context"
	"sort"
	"sync"
	"time"
)

// contextKey stores the startup collector on one host startup context.
type contextKey struct{}

// Counter names one startup metric recorded by the collector.
type Counter string

// Phase names one observed startup phase recorded by the collector.
type Phase string

// Startup metric counter names.
const (
	CounterCatalogSnapshotBuilds      Counter = "catalog_snapshot_builds"
	CounterIntegrationSnapshotBuilds  Counter = "integration_snapshot_builds"
	CounterJobSnapshotBuilds          Counter = "job_snapshot_builds"
	CounterPluginScans                Counter = "plugin_scans"
	CounterPluginScanItems            Counter = "plugin_scan_items"
	CounterPluginSyncChanged          Counter = "plugin_sync_changed"
	CounterPluginSyncNoop             Counter = "plugin_sync_noop"
	CounterPluginMenuSyncChanged      Counter = "plugin_menu_sync_changed"
	CounterPluginMenuSyncNoop         Counter = "plugin_menu_sync_noop"
	CounterPluginResourceSyncChanged  Counter = "plugin_resource_sync_changed"
	CounterPluginResourceSyncNoop     Counter = "plugin_resource_sync_noop"
	CounterBuiltinJobProjections      Counter = "builtin_job_projections"
	CounterBuiltinJobProjectionNoop   Counter = "builtin_job_projection_noop"
	CounterPersistentJobStartupLoaded Counter = "persistent_job_startup_loaded"
)

// Startup phase names used by HTTP runtime startup orchestration.
const (
	// PhasePluginBootstrapAutoEnable measures startup plugin auto-enable work.
	PhasePluginBootstrapAutoEnable Phase = "plugin_bootstrap_auto_enable"
	// PhasePluginSourceUpgradeReadiness measures source-plugin upgrade readiness checks.
	PhasePluginSourceUpgradeReadiness Phase = "plugin_source_upgrade_readiness"
	// PhasePluginLifecycleAttach measures plugin lifecycle handler attachment.
	PhasePluginLifecycleAttach Phase = "plugin_lifecycle_attach"
	// PhaseCronStart measures cron and persistent scheduled-job startup.
	PhaseCronStart Phase = "cron_start"
)

// Collector stores counters and phase durations for one HTTP startup pass.
type Collector struct {
	mu       sync.Mutex
	started  time.Time
	counters map[Counter]int
	phases   map[Phase]time.Duration
}

// Snapshot is an immutable view of the collected startup statistics.
type Snapshot struct {
	StartedAt time.Time
	Elapsed   time.Duration
	Counters  map[Counter]int
	Phases    map[Phase]time.Duration
}

// New creates a startup statistics collector.
func New() *Collector {
	return &Collector{
		started:  time.Now(),
		counters: make(map[Counter]int),
		phases:   make(map[Phase]time.Duration),
	}
}

// WithCollector returns a child context carrying the startup collector.
func WithCollector(ctx context.Context, collector *Collector) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if collector == nil {
		return ctx
	}
	return context.WithValue(ctx, contextKey{}, collector)
}

// FromContext returns the startup collector stored on the context.
func FromContext(ctx context.Context) *Collector {
	if ctx == nil {
		return nil
	}
	collector, ok := ctx.Value(contextKey{}).(*Collector)
	if !ok {
		return nil
	}
	return collector
}

// Add increments one startup counter by delta.
func Add(ctx context.Context, counter Counter, delta int) {
	collector := FromContext(ctx)
	if collector == nil || delta == 0 {
		return
	}
	collector.Add(counter, delta)
}

// Add increments one startup counter by delta.
func (c *Collector) Add(counter Counter, delta int) {
	if c == nil || delta == 0 {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.counters[counter] += delta
}

// Observe records the duration of one startup phase.
func Observe(ctx context.Context, phase Phase, fn func() error) error {
	started := time.Now()
	err := fn()
	if collector := FromContext(ctx); collector != nil {
		collector.RecordPhase(phase, time.Since(started))
	}
	return err
}

// RecordPhase adds one observed phase duration to the collector.
func (c *Collector) RecordPhase(phase Phase, duration time.Duration) {
	if c == nil || phase == "" {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.phases[phase] += duration
}

// Snapshot returns an immutable copy of the current collector state.
func (c *Collector) Snapshot() Snapshot {
	if c == nil {
		return Snapshot{}
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	counters := make(map[Counter]int, len(c.counters))
	for key, value := range c.counters {
		counters[key] = value
	}
	phases := make(map[Phase]time.Duration, len(c.phases))
	for key, value := range c.phases {
		phases[key] = value
	}
	return Snapshot{
		StartedAt: c.started,
		Elapsed:   time.Since(c.started),
		Counters:  counters,
		Phases:    phases,
	}
}

// CounterValue returns the current value for one counter.
func (s Snapshot) CounterValue(counter Counter) int {
	if s.Counters == nil {
		return 0
	}
	return s.Counters[counter]
}

// PhaseNames returns deterministic phase names present in the snapshot.
func (s Snapshot) PhaseNames() []Phase {
	names := make([]Phase, 0, len(s.Phases))
	for name := range s.Phases {
		names = append(names, name)
	}
	sort.Slice(names, func(i, j int) bool {
		return names[i] < names[j]
	})
	return names
}
