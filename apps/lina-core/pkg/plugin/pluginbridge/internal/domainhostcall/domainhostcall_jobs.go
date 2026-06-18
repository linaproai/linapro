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

// Search returns one bounded page of visible scheduled-job projections.
func (s jobsService) Search(_ context.Context, _ capmodel.CapabilityContext, input jobcap.SearchInput) (*capmodel.PageResult[*jobcap.Projection], error) {
	out := &capmodel.PageResult[*jobcap.Projection]{Items: []*jobcap.Projection{}}
	err := s.callJSONRequest(protocol.HostServiceJobs, protocol.HostServiceMethodJobsSearch, jobsSearchRequest{
		Keyword:  input.Keyword,
		Group:    input.Group,
		Status:   input.Status,
		PageNum:  input.Page.PageNum,
		PageSize: input.Page.PageSize,
	}, out)
	return out, err
}

// EnsureVisible rejects when any requested job is absent or invisible.
func (s jobsService) EnsureVisible(_ context.Context, _ capmodel.CapabilityContext, ids []jobcap.JobID) error {
	return s.callJSONRequest(protocol.HostServiceJobs, protocol.HostServiceMethodJobsEnsureVisible, idsRequest{IDs: jobIDsToStrings(ids)}, nil)
}

// jobIDsToStrings converts scheduled-job IDs to transport strings.
func jobIDsToStrings(ids []jobcap.JobID) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, string(id))
	}
	return out
}

// jobsSearchRequest carries governed scheduled-job search parameters.
type jobsSearchRequest struct {
	Keyword  string `json:"keyword,omitempty"`
	Group    string `json:"group,omitempty"`
	Status   string `json:"status,omitempty"`
	PageNum  int    `json:"pageNum,omitempty"`
	PageSize int    `json:"pageSize,omitempty"`
}

var _ jobcap.Service = (*jobsService)(nil)
