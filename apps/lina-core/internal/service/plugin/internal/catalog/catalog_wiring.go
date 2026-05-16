// This file wires sibling plugin subcomponents into the catalog service without
// introducing package import cycles.

package catalog

// SetBackendLoader wires the integration package's backend config loader.
func (s *serviceImpl) SetBackendLoader(loader BackendConfigLoader) {
	s.backendLoader = loader
}

// SetArtifactParser wires the runtime package's WASM artifact parser.
func (s *serviceImpl) SetArtifactParser(parser ArtifactParser) {
	s.artifactParser = parser
}

// SetDynamicManifestLoader wires the runtime package's active manifest loader.
func (s *serviceImpl) SetDynamicManifestLoader(loader DynamicManifestLoader) {
	s.dynamicManifestLoader = loader
}

// SetNodeStateSyncer wires the runtime package's node state syncer.
func (s *serviceImpl) SetNodeStateSyncer(syncer NodeStateSyncer) {
	s.nodeStateSyncer = syncer
}

// SetMenuSyncer wires the integration package's menu syncer.
func (s *serviceImpl) SetMenuSyncer(syncer MenuSyncer) {
	s.menuSyncer = syncer
}

// SetResourceRefSyncer wires the integration package's resource reference syncer.
func (s *serviceImpl) SetResourceRefSyncer(syncer ResourceRefSyncer) {
	s.resourceRefSyncer = syncer
}

// SetReleaseStateSyncer wires the runtime package's release state syncer.
func (s *serviceImpl) SetReleaseStateSyncer(syncer ReleaseStateSyncer) {
	s.releaseStateSyncer = syncer
}

// SetHookDispatcher wires the integration package's hook event dispatcher.
func (s *serviceImpl) SetHookDispatcher(dispatcher HookDispatcher) {
	s.hookDispatcher = dispatcher
}
