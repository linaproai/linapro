// This file verifies plugin management read-model helpers and cache behavior.

package management

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
)

// TestListCacheLoadOrBuildSharesConcurrentBuild verifies cold concurrent
// summary-list requests share one in-flight build for the same cache key.
func TestListCacheLoadOrBuildSharesConcurrentBuild(t *testing.T) {
	cache := NewListCache()
	key := ListCacheKey{
		Locale:               "zh-CN",
		RuntimeBundleVersion: 1,
		RuntimeRevision:      7,
	}

	var builds int32
	started := make(chan struct{})
	release := make(chan struct{})
	const callers = 8
	errs := make(chan error, callers)
	var wg sync.WaitGroup
	for range callers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			out, err := cache.LoadOrBuild(key, func() (*ListOutput, error) {
				if atomic.AddInt32(&builds, 1) == 1 {
					close(started)
				}
				<-release
				return &ListOutput{List: []*PluginItem{{}}, Total: 1}, nil
			})
			if err != nil {
				errs <- err
				return
			}
			if out == nil || out.Total != 1 || len(out.List) != 1 || out.List[0] == nil {
				errs <- fmt.Errorf("unexpected cache output: %#v", out)
			}
		}()
	}

	<-started
	close(release)
	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	if got := atomic.LoadInt32(&builds); got != 1 {
		t.Fatalf("expected one shared cache build, got %d", got)
	}
}
