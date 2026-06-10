// This file implements the guest-side infra capability hostcall client.

package domainhostcall

import (
	"context"

	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/infracap"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// infraService adapts infrastructure status reads to host services.
type infraService struct{ baseService }

// Infra creates the infrastructure-domain guest client.
func Infra(invoker Invoker) infracap.Service {
	return infraService{baseService: newBaseService(invoker)}
}

// BatchGetStatus returns visible component status projections.
func (s infraService) BatchGetStatus(_ context.Context, _ capmodel.CapabilityContext, ids []infracap.ComponentID) (*capmodel.BatchResult[*infracap.StatusProjection, infracap.ComponentID], error) {
	out := &capmodel.BatchResult[*infracap.StatusProjection, infracap.ComponentID]{Items: map[infracap.ComponentID]*infracap.StatusProjection{}}
	err := s.callJSONRequest(protocol.HostServiceInfra, protocol.HostServiceMethodInfraBatchGetStatus, idsRequest{IDs: componentIDsToStrings(ids)}, out)
	return out, err
}

// componentIDsToStrings converts infrastructure component IDs to strings.
func componentIDsToStrings(ids []infracap.ComponentID) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, string(id))
	}
	return out
}

var _ infracap.Service = (*infraService)(nil)
