// This file contains the in-memory job handler registry implementation,
// including normalization, duplicate detection, and change notifications.

package jobhandler

import (
	"sort"
	"strings"

	"lina-core/internal/service/jobmeta"
	"lina-core/pkg/bizerr"
)

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
