// This file implements the guest-side jobs capability hostcall client.

package domainhostcall

import (
	"context"

	"lina-core/pkg/plugin/capability/capmodel"
	"lina-core/pkg/plugin/capability/jobcap"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// jobsService adapts scheduled-job projection reads to host services.
type jobsService struct{ baseService }

// Jobs creates the scheduled-job domain guest client.
func Jobs(invoker Invoker) jobcap.Service {
	return jobsService{baseService: newBaseService(invoker)}
}

// BatchGet returns visible job projections and opaque missing IDs.
func (s jobsService) BatchGet(_ context.Context, _ capmodel.CapabilityContext, ids []jobcap.JobID) (*capmodel.BatchResult[*jobcap.Projection, jobcap.JobID], error) {
	out := &capmodel.BatchResult[*jobcap.Projection, jobcap.JobID]{Items: map[jobcap.JobID]*jobcap.Projection{}}
	err := s.callJSONRequest(protocol.HostServiceJobs, protocol.HostServiceMethodJobsBatchGet, idsRequest{IDs: jobIDsToStrings(ids)}, out)
	return out, err
}

// jobIDsToStrings converts scheduled-job IDs to transport strings.
func jobIDsToStrings(ids []jobcap.JobID) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, string(id))
	}
	return out
}

var _ jobcap.Service = (*jobsService)(nil)
