// This file wires runtime reconciliation and topology dependencies into the
// dynamic plugin lifecycle service.

package lifecycle

// SetReconciler wires the runtime package's reconcile provider.
func (s *serviceImpl) SetReconciler(r ReconcileProvider) {
	s.reconciler = r
}

// SetTopology wires the cluster topology provider.
func (s *serviceImpl) SetTopology(t TopologyProvider) {
	s.topology = t
}
