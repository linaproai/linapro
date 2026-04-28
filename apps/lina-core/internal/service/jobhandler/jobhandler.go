// Package jobhandler implements the in-memory scheduled-job handler registry
// shared by job management, the scheduler, and plugin lifecycle hooks.
package jobhandler

import (
	"context"
	"encoding/json"
	"sort"
	"strings"
	"sync"

	"lina-core/internal/service/jobmeta"
	"lina-core/pkg/bizerr"
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
	Register(def HandlerDef) error
	// Unregister removes one handler definition and notifies change observers.
	Unregister(ref string)
	// Lookup returns one registered handler definition by ref.
	Lookup(ref string) (HandlerDef, bool)
	// List returns all registered handlers sorted by ref.
	List() []HandlerInfo
	// SubscribeChanges registers one change callback and returns its unsubscribe function.
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

// Register stores one handler definition and rejects duplicate refs.
func (s *serviceImpl) Register(def HandlerDef) error {
	def.Ref = strings.TrimSpace(def.Ref)
	def.DisplayName = strings.TrimSpace(def.DisplayName)
	def.Description = strings.TrimSpace(def.Description)
	def.PluginID = strings.TrimSpace(def.PluginID)
	if def.Ref == "" {
		return bizerr.NewCode(CodeJobHandlerRefRequired)
	}
	if def.DisplayName == "" {
		return bizerr.NewCode(CodeJobHandlerDisplayNameRequired)
	}
	if def.Invoke == nil {
		return bizerr.NewCode(CodeJobHandlerCallbackRequired)
	}
	if !def.Source.IsValid() {
		return bizerr.NewCode(CodeJobHandlerSourceUnsupported)
	}
	if def.Source == jobmeta.HandlerSourcePlugin && def.PluginID == "" {
		return bizerr.NewCode(CodeJobHandlerPluginIDRequired)
	}
	if def.Source == jobmeta.HandlerSourceHost {
		def.PluginID = ""
	}

	schemaText, err := normalizeSchema(def.ParamsSchema)
	if err != nil {
		return err
	}
	def.ParamsSchema = schemaText

	s.mu.Lock()
	if _, exists := s.handlers[def.Ref]; exists {
		s.mu.Unlock()
		return bizerr.NewCode(CodeJobHandlerExists, bizerr.P("ref", def.Ref))
	}
	s.handlers[def.Ref] = def
	callbacks := s.snapshotCallbacksLocked()
	s.mu.Unlock()

	notifyCallbacks(callbacks, def.Ref, true)
	return nil
}

// Unregister removes one handler definition and notifies change observers.
func (s *serviceImpl) Unregister(ref string) {
	trimmedRef := strings.TrimSpace(ref)
	if trimmedRef == "" {
		return
	}

	s.mu.Lock()
	if _, exists := s.handlers[trimmedRef]; !exists {
		s.mu.Unlock()
		return
	}
	delete(s.handlers, trimmedRef)
	callbacks := s.snapshotCallbacksLocked()
	s.mu.Unlock()

	notifyCallbacks(callbacks, trimmedRef, false)
}

// Lookup returns one registered handler definition by ref.
func (s *serviceImpl) Lookup(ref string) (HandlerDef, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	def, ok := s.handlers[strings.TrimSpace(ref)]
	return def, ok
}

// List returns all registered handlers sorted by ref.
func (s *serviceImpl) List() []HandlerInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]HandlerInfo, 0, len(s.handlers))
	for _, def := range s.handlers {
		items = append(items, HandlerInfo{
			Ref:          def.Ref,
			DisplayName:  def.DisplayName,
			Description:  def.Description,
			ParamsSchema: def.ParamsSchema,
			Source:       def.Source,
			PluginID:     def.PluginID,
		})
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Ref < items[j].Ref
	})
	return items
}

// SubscribeChanges registers one change callback and returns its unsubscribe function.
func (s *serviceImpl) SubscribeChanges(callback ChangeCallback) func() {
	if callback == nil {
		return func() {}
	}

	s.mu.Lock()
	s.callbackID++
	currentID := s.callbackID
	s.callbacks[currentID] = callback
	s.mu.Unlock()

	return func() {
		s.mu.Lock()
		delete(s.callbacks, currentID)
		s.mu.Unlock()
	}
}

// snapshotCallbacksLocked clones all callbacks while the write lock is held.
func (s *serviceImpl) snapshotCallbacksLocked() []ChangeCallback {
	callbacks := make([]ChangeCallback, 0, len(s.callbacks))
	for _, callback := range s.callbacks {
		callbacks = append(callbacks, callback)
	}
	return callbacks
}

// notifyCallbacks executes all registry change observers outside the registry lock.
func notifyCallbacks(callbacks []ChangeCallback, ref string, exists bool) {
	for _, callback := range callbacks {
		callback(ref, exists)
	}
}
