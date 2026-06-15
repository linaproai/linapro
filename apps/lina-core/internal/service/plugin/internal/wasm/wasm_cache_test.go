// This file covers compiled Wasm module cache invalidation and execution lease
// behavior so runtime refreshes cannot close an entry while a bridge call is
// still instantiating or executing.

package wasm

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tetratelabs/wazero"
)

// TestInvalidateCacheWaitsForActiveLease verifies cache invalidation removes
// stale entries for new callers while deferring runtime close until in-flight
// executions release their lease.
func TestInvalidateCacheWaitsForActiveLease(t *testing.T) {
	ctx := context.Background()
	runtime := newTestRuntimeImpl()
	artifactPath := writeMinimalWasmArtifact(t)
	defer runtime.InvalidateAllCache(ctx)

	lease, err := runtime.getOrCompileWasmModule(ctx, artifactPath)
	if err != nil {
		t.Fatalf("compile wasm module failed: %v", err)
	}
	firstRuntime := lease.runtime

	invalidated := make(chan struct{})
	go func() {
		runtime.InvalidateCache(ctx, artifactPath)
		close(invalidated)
	}()

	select {
	case <-invalidated:
		t.Fatal("expected cache invalidation to wait for the active lease")
	default:
	}
	waitForCacheEntryRemoval(t, runtime, artifactPath)

	freshLease, err := runtime.getOrCompileWasmModule(ctx, artifactPath)
	if err != nil {
		t.Fatalf("compile fresh wasm module after invalidation failed: %v", err)
	}
	if freshLease.runtime == firstRuntime {
		t.Fatal("expected new callers to receive a freshly compiled runtime")
	}
	freshLease.Release()

	lease.Release()
	<-invalidated

	if _, err := firstRuntime.CompileModule(ctx, minimalWasmBinary()); err == nil {
		t.Fatal("expected invalidated runtime to be closed after lease release")
	}
}

// TestInvalidateAllCacheWaitsForActiveLease verifies full cache invalidation
// follows the same deferred close contract as path-scoped invalidation.
func TestInvalidateAllCacheWaitsForActiveLease(t *testing.T) {
	ctx := context.Background()
	runtime := newTestRuntimeImpl()
	artifactPath := writeMinimalWasmArtifact(t)
	defer runtime.InvalidateAllCache(ctx)

	lease, err := runtime.getOrCompileWasmModule(ctx, artifactPath)
	if err != nil {
		t.Fatalf("compile wasm module failed: %v", err)
	}
	firstRuntime := lease.runtime

	invalidated := make(chan struct{})
	go func() {
		runtime.InvalidateAllCache(ctx)
		close(invalidated)
	}()

	select {
	case <-invalidated:
		t.Fatal("expected full cache invalidation to wait for the active lease")
	default:
	}
	waitForCacheEntryRemoval(t, runtime, artifactPath)

	freshLease, err := runtime.getOrCompileWasmModule(ctx, artifactPath)
	if err != nil {
		t.Fatalf("compile fresh wasm module after full invalidation failed: %v", err)
	}
	if freshLease.runtime == firstRuntime {
		t.Fatal("expected new callers to receive a freshly compiled runtime")
	}
	freshLease.Release()

	lease.Release()
	<-invalidated

	if _, err := firstRuntime.CompileModule(ctx, minimalWasmBinary()); err == nil {
		t.Fatal("expected invalidated runtime to be closed after lease release")
	}
}

// TestConcurrentSameArtifactCompilesOnce verifies the per-artifact single-flight
// path shares one compile among concurrent callers.
func TestConcurrentSameArtifactCompilesOnce(t *testing.T) {
	ctx := context.Background()
	runtime := newTestRuntimeImpl()
	artifactPath := writeMinimalWasmArtifact(t)
	defer runtime.InvalidateAllCache(ctx)

	var compileCount atomic.Int32
	runtime.compileHook = func(string) {
		compileCount.Add(1)
		time.Sleep(20 * time.Millisecond)
	}

	const callers = 8
	var wg sync.WaitGroup
	errs := make(chan error, callers)
	for i := 0; i < callers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			lease, err := runtime.getOrCompileWasmModule(ctx, artifactPath)
			if err != nil {
				errs <- err
				return
			}
			lease.Release()
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent compile failed: %v", err)
		}
	}
	if got := compileCount.Load(); got != 1 {
		t.Fatalf("expected one compile for concurrent same artifact, got %d", got)
	}
}

// TestCompileFailureRetries verifies failed compilation is not cached and a
// later valid artifact at the same path can compile successfully.
func TestCompileFailureRetries(t *testing.T) {
	ctx := context.Background()
	runtime := newTestRuntimeImpl()
	artifactPath := filepath.Join(t.TempDir(), "retry.wasm")
	if err := os.WriteFile(artifactPath, []byte("not-wasm"), 0o600); err != nil {
		t.Fatalf("write invalid wasm failed: %v", err)
	}
	defer runtime.InvalidateAllCache(ctx)

	var compileCount atomic.Int32
	runtime.compileHook = func(string) {
		compileCount.Add(1)
	}
	if _, err := runtime.getOrCompileWasmModule(ctx, artifactPath); err == nil {
		t.Fatal("expected invalid wasm compile to fail")
	}
	if err := os.WriteFile(artifactPath, minimalWasmBinary(), 0o600); err != nil {
		t.Fatalf("write valid wasm failed: %v", err)
	}
	lease, err := runtime.getOrCompileWasmModule(ctx, artifactPath)
	if err != nil {
		t.Fatalf("expected compile retry to succeed: %v", err)
	}
	lease.Release()
	if got := compileCount.Load(); got != 2 {
		t.Fatalf("expected failed compile plus retry, got %d", got)
	}
}

// TestDifferentArtifactsCompileIndependently verifies one artifact's in-flight
// compile does not make another artifact wait on the same single-flight entry.
func TestDifferentArtifactsCompileIndependently(t *testing.T) {
	ctx := context.Background()
	runtime := newTestRuntimeImpl()
	firstPath := writeMinimalWasmArtifactNamed(t, "first.wasm")
	secondPath := writeMinimalWasmArtifactNamed(t, "second.wasm")
	defer runtime.InvalidateAllCache(ctx)

	firstStarted := make(chan struct{})
	releaseFirst := make(chan struct{})
	var firstHookOnce sync.Once
	runtime.compileHook = func(artifactPath string) {
		if artifactPath != firstPath {
			return
		}
		firstHookOnce.Do(func() {
			close(firstStarted)
			<-releaseFirst
		})
	}

	firstErr := make(chan error, 1)
	go func() {
		lease, err := runtime.getOrCompileWasmModule(ctx, firstPath)
		if err == nil {
			lease.Release()
		}
		firstErr <- err
	}()
	<-firstStarted

	secondDone := make(chan error, 1)
	go func() {
		lease, err := runtime.getOrCompileWasmModule(ctx, secondPath)
		if err == nil {
			lease.Release()
		}
		secondDone <- err
	}()
	select {
	case err := <-secondDone:
		if err != nil {
			t.Fatalf("second artifact compile failed: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("different artifact compile was blocked by first artifact")
	}
	close(releaseFirst)
	if err := <-firstErr; err != nil {
		t.Fatalf("first artifact compile failed: %v", err)
	}
}

// waitForCacheEntryRemoval waits until invalidation has removed a stale entry
// from the global map, allowing the caller to compile a replacement entry.
func waitForCacheEntryRemoval(t *testing.T, runtime *runtimeImpl, artifactPath string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		runtime.cacheMu.RLock()
		_, ok := runtime.cache[artifactPath]
		runtime.cacheMu.RUnlock()
		if !ok {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("timed out waiting for cache entry removal")
}

func newTestRuntimeImpl() *runtimeImpl {
	return &runtimeImpl{
		hostServices: newTestHostServiceRuntime(),
		cache:        make(map[string]*wasmCacheEntry),
		inflight:     make(map[string]*wasmCompileInflight),
	}
}

// writeMinimalWasmArtifact writes a tiny module that imports the host bridge
// function, matching the dynamic plugin compile-time import contract.
func writeMinimalWasmArtifact(t *testing.T) string {
	t.Helper()
	return writeMinimalWasmArtifactNamed(t, "minimal.wasm")
}

// writeMinimalWasmArtifactNamed writes a minimal module with the given file name.
func writeMinimalWasmArtifactNamed(t *testing.T, name string) string {
	t.Helper()
	artifactPath := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(artifactPath, minimalWasmBinary(), 0o600); err != nil {
		t.Fatalf("write wasm artifact failed: %v", err)
	}
	return artifactPath
}

// minimalWasmBinary returns a valid module with the lina_env.host_call import.
func minimalWasmBinary() []byte {
	return []byte{
		0x00, 0x61, 0x73, 0x6d,
		0x01, 0x00, 0x00, 0x00,
		0x01, 0x0b, 0x02, 0x60, 0x03, 0x7f, 0x7f, 0x7f, 0x01, 0x7e, 0x60, 0x00, 0x00,
		0x02, 0x16, 0x01, 0x08, 0x6c, 0x69, 0x6e, 0x61, 0x5f, 0x65, 0x6e, 0x76, 0x09, 0x68, 0x6f, 0x73, 0x74, 0x5f, 0x63, 0x61, 0x6c, 0x6c, 0x00, 0x00,
		0x03, 0x02, 0x01, 0x01,
		0x07, 0x08, 0x01, 0x04, 0x6e, 0x6f, 0x6f, 0x70, 0x00, 0x01,
		0x0a, 0x04, 0x01, 0x02, 0x00, 0x0b,
	}
}

// TestMinimalWasmBinaryIsValid keeps the hand-authored binary readable by
// ensuring wazero can compile it with the same host imports as production.
func TestMinimalWasmBinaryIsValid(t *testing.T) {
	ctx := context.Background()
	rt := wazero.NewRuntime(ctx)
	defer func() {
		if err := rt.Close(ctx); err != nil {
			t.Fatalf("close wasm runtime failed: %v", err)
		}
	}()
	if err := newTestRuntimeImpl().registerHostCallModule(ctx, rt); err != nil {
		t.Fatalf("register host call module failed: %v", err)
	}
	compiled, err := rt.CompileModule(ctx, minimalWasmBinary())
	if err != nil {
		t.Fatalf("compile minimal wasm failed: %v", err)
	}
	if compiled == nil {
		t.Fatal("expected compiled module")
	}
}

// TestBridgeExecutionContextAddsDefaultDeadline verifies bridge calls without a
// caller deadline receive the host fallback timeout.
func TestBridgeExecutionContextAddsDefaultDeadline(t *testing.T) {
	ctx, cancel := bridgeExecutionContext(context.Background())
	if cancel == nil {
		t.Fatal("expected bridge context to return a cancel function")
	}
	defer cancel()

	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("expected bridge context to carry a deadline")
	}
	remaining := time.Until(deadline)
	if remaining <= 0 || remaining > defaultBridgeExecutionTimeout {
		t.Fatalf("expected deadline within default timeout, got remaining=%s", remaining)
	}
}

// TestBridgeExecutionContextKeepsCallerDeadline verifies the bridge guard does
// not extend a tighter caller deadline.
func TestBridgeExecutionContextKeepsCallerDeadline(t *testing.T) {
	parent, parentCancel := context.WithTimeout(context.Background(), time.Second)
	defer parentCancel()
	parentDeadline, ok := parent.Deadline()
	if !ok {
		t.Fatal("expected parent context to carry a deadline")
	}

	ctx, cancel := bridgeExecutionContext(parent)
	if cancel != nil {
		t.Fatal("expected no replacement cancel function for caller deadline")
	}
	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("expected returned context to keep a deadline")
	}
	if !deadline.Equal(parentDeadline) {
		t.Fatalf("expected caller deadline %s, got %s", parentDeadline, deadline)
	}
}

// TestNewWasmRuntimeCancelsInfiniteLoop verifies the runtime configuration
// terminates non-returning guest code when the call context expires.
func TestNewWasmRuntimeCancelsInfiniteLoop(t *testing.T) {
	ctx := context.Background()
	rt := newWasmRuntime(ctx)
	defer func() {
		if err := rt.Close(context.Background()); err != nil {
			t.Fatalf("close wasm runtime failed: %v", err)
		}
	}()

	compiled, err := rt.CompileModule(ctx, infiniteLoopWasmBinary())
	if err != nil {
		t.Fatalf("compile infinite loop wasm failed: %v", err)
	}
	module, err := rt.InstantiateModule(ctx, compiled, wazero.NewModuleConfig().WithName("infinite-loop"))
	if err != nil {
		t.Fatalf("instantiate infinite loop wasm failed: %v", err)
	}

	callCtx, cancel := context.WithTimeout(ctx, 20*time.Millisecond)
	defer cancel()
	_, err = module.ExportedFunction("infinite_loop").Call(callCtx)
	if err == nil {
		t.Fatal("expected infinite loop to stop when context expires")
	}
	if !strings.Contains(err.Error(), context.DeadlineExceeded.Error()) {
		t.Fatalf("expected deadline exceeded error, got: %v", err)
	}
}

// TestNewWasmRuntimeRejectsMemoryAboveLimit verifies the host-side memory
// ceiling rejects modules whose memory minimum exceeds the configured limit.
func TestNewWasmRuntimeRejectsMemoryAboveLimit(t *testing.T) {
	ctx := context.Background()
	rt := newWasmRuntime(ctx)
	defer func() {
		if err := rt.Close(ctx); err != nil {
			t.Fatalf("close wasm runtime failed: %v", err)
		}
	}()

	_, err := rt.CompileModule(ctx, wasmBinaryWithMemoryMinimum(defaultWasmMemoryLimitPages+1))
	if err == nil {
		t.Fatal("expected wasm compile to reject memory above host limit")
	}
	if !strings.Contains(err.Error(), "memory") {
		t.Fatalf("expected memory limit error, got: %v", err)
	}
}

// infiniteLoopWasmBinary returns a tiny module exporting infinite_loop, matching
// the wazero context-cancellation example without importing testdata.
func infiniteLoopWasmBinary() []byte {
	return []byte{
		0x00, 0x61, 0x73, 0x6d,
		0x01, 0x00, 0x00, 0x00,
		0x01, 0x04, 0x01, 0x60, 0x00, 0x00,
		0x03, 0x02, 0x01, 0x00,
		0x07, 0x11, 0x01, 0x0d, 0x69, 0x6e, 0x66, 0x69, 0x6e, 0x69, 0x74, 0x65, 0x5f, 0x6c, 0x6f, 0x6f, 0x70, 0x00, 0x00,
		0x0a, 0x09, 0x01, 0x07, 0x00, 0x03, 0x40, 0x0c, 0x00, 0x0b, 0x0b,
	}
}

// wasmBinaryWithMemoryMinimum returns a minimal module with one memory section.
func wasmBinaryWithMemoryMinimum(minPages uint32) []byte {
	content := append([]byte{0x01, 0x00}, encodeU32LEB128(minPages)...)
	out := []byte{
		0x00, 0x61, 0x73, 0x6d,
		0x01, 0x00, 0x00, 0x00,
		0x05,
	}
	out = append(out, encodeU32LEB128(uint32(len(content)))...)
	out = append(out, content...)
	return out
}

// encodeU32LEB128 encodes one uint32 for compact WebAssembly section payloads.
func encodeU32LEB128(value uint32) []byte {
	var encoded []byte
	for {
		next := byte(value & 0x7f)
		value >>= 7
		if value != 0 {
			next |= 0x80
		}
		encoded = append(encoded, next)
		if value == 0 {
			return encoded
		}
	}
}
