// Package jobhandler implements the in-memory scheduled-job handler registry
// shared by job management, the scheduler, and plugin lifecycle hooks.
package jobhandler

import (
	"context"
	"encoding/json"
	"sync"

	"lina-core/internal/service/jobmeta"
)

// InvokeFunc defines one registered scheduled-job callback.
type InvokeFunc func(ctx context.Context, params json.RawMessage) (result any, err error)

// HandlerDef defines one registry entry with execution callback and metadata.
type HandlerDef struct {
	Ref          string                // Ref is the stable handler reference.
	DisplayName  string                // DisplayName is shown in the UI.
	Description  string                // Description explains the handler purpose.
	ParamsSchema string                // ParamsSchema stores the accepted JSON Schema subset.
	Source       jobmeta.HandlerSource // Source identifies whether the handler comes from host or plugin code.
	PluginID     string                // PluginID stores the owning plugin when Source=plugin.
	Invoke       InvokeFunc            // Invoke executes the handler.
}

// HandlerInfo defines one display-safe handler snapshot exposed to callers.
type HandlerInfo struct {
	Ref          string                // Ref is the stable handler reference.
	DisplayName  string                // DisplayName is shown in the UI.
	Description  string                // Description explains the handler purpose.
	ParamsSchema string                // ParamsSchema stores the accepted JSON Schema subset.
	Source       jobmeta.HandlerSource // Source identifies whether the handler comes from host or plugin code.
	PluginID     string                // PluginID stores the owning plugin when Source=plugin.
}

// ChangeCallback receives registry change notifications after a handler is
// registered or unregistered.
type ChangeCallback func(ref string, exists bool)

// Registry defines the scheduled-job handler registry contract.
type Registry interface {
	// Register stores one handler definition and rejects duplicate refs.
	// The definition is normalized before storage; missing refs, names,
	// callbacks, unsupported sources, invalid plugin ownership, duplicate refs,
	// or invalid parameter schemas return business errors. Successful calls
	// notify subscribers outside the registry lock.
	Register(def HandlerDef) error
	// Unregister removes one handler definition and notifies change observers.
	// Blank or unknown refs are ignored, making plugin lifecycle cleanup
	// idempotent.
	Unregister(ref string)
	// Lookup returns one registered handler definition by ref. The returned
	// definition includes the invocation callback and must only be used inside
	// trusted scheduler execution paths.
	Lookup(ref string) (HandlerDef, bool)
	// List returns all registered handlers sorted by ref. The result excludes
	// callback functions and is safe to expose to management views.
	List() []HandlerInfo
	// SubscribeChanges registers one change callback and returns its unsubscribe
	// function. Nil callbacks receive a no-op unsubscribe function; callbacks are
	// invoked after lock release and should avoid blocking scheduler refresh.
	SubscribeChanges(callback ChangeCallback) func()
}

// Ensure serviceImpl implements Registry.
var _ Registry = (*serviceImpl)(nil)

// serviceImpl implements the in-memory handler registry.
type serviceImpl struct {
	mu         sync.RWMutex           // mu protects handler and callback state.
	handlers   map[string]HandlerDef  // handlers stores definitions by stable ref.
	callbackID int                    // callbackID allocates deterministic observer keys.
	callbacks  map[int]ChangeCallback // callbacks stores registry observers.
}

// New creates and returns one empty handler registry.
func New() Registry {
	return &serviceImpl{
		handlers:  make(map[string]HandlerDef),
		callbacks: make(map[int]ChangeCallback),
	}
}
