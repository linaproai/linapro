// This file contains collector and snapshot methods for startup metrics,
// keeping the main file focused on metric contracts and context helpers.

package startupstats

import (
	"sort"
	"time"
)

// Add increments one startup counter by delta.
func (c *Collector) Add(counter Counter, delta int) {
	if c == nil || delta == 0 {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.counters[counter] += delta
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
