// This file implements the guest-side jobs capability hostcall client.

package domainhostcall

import (
	"context"

	jobv1 "lina-core/api/job/v1"
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
func (s jobsService) BatchGet(_ context.Context, ids []jobcap.JobID) (*capmodel.BatchResult[*jobcap.JobInfo, jobcap.JobID], error) {
	out := &capmodel.BatchResult[*jobcap.JobInfo, jobcap.JobID]{Items: map[jobcap.JobID]*jobcap.JobInfo{}}
	err := s.callJSONRequest(protocol.HostServiceJobs, protocol.HostServiceMethodJobsBatchGet, idsRequest{IDs: jobIDsToStrings(ids)}, out)
	return out, err
}

// Get returns one visible job projection through the registered batch-read method.
func (s jobsService) Get(ctx context.Context, id jobcap.JobID) (*jobcap.JobInfo, error) {
	result, err := s.BatchGet(ctx, []jobcap.JobID{id})
	if err != nil || result == nil {
		return nil, err
	}
	if item, ok := result.Items[id]; ok {
		return item, nil
	}
	return nil, nil
}

// List returns one bounded page of visible scheduled-job projections.
func (s jobsService) List(_ context.Context, input jobcap.ListInput) (*capmodel.PageResult[*jobcap.JobInfo], error) {
	out := &capmodel.PageResult[*jobcap.JobInfo]{Items: []*jobcap.JobInfo{}}
	err := s.callJSONRequest(protocol.HostServiceJobs, protocol.HostServiceMethodJobsList, jobsListRequest{
		Keyword:  input.Keyword,
		Group:    input.Group,
		Status:   string(input.Status),
		PageNum:  input.Page.PageNum,
		PageSize: input.Page.PageSize,
	}, out)
	return out, err
}

// EnsureVisible rejects when any requested job is absent or invisible.
func (s jobsService) EnsureVisible(_ context.Context, ids []jobcap.JobID) error {
	return s.callJSONRequest(protocol.HostServiceJobs, protocol.HostServiceMethodJobsEnsureVisible, idsRequest{IDs: jobIDsToStrings(ids)}, nil)
}

// Create creates one governed scheduled job.
func (s jobsService) Create(_ context.Context, input jobcap.SaveInput) (jobcap.JobID, error) {
	var out jobcap.JobID
	err := s.callJSONRequest(protocol.HostServiceJobs, protocol.HostServiceMethodJobsCreate, input, &out)
	return out, err
}

// Update mutates one visible scheduled job.
func (s jobsService) Update(_ context.Context, input jobcap.UpdateInput) error {
	return s.callJSONRequest(protocol.HostServiceJobs, protocol.HostServiceMethodJobsUpdate, input, nil)
}

// Delete deletes one visible scheduled job.
func (s jobsService) Delete(_ context.Context, id jobcap.JobID) error {
	return s.callJSONRequest(protocol.HostServiceJobs, protocol.HostServiceMethodJobsDelete, jobIDRequest{JobID: string(id)}, nil)
}

// Run triggers one visible scheduled job.
func (s jobsService) Run(_ context.Context, id jobcap.JobID) error {
	return s.callJSONRequest(protocol.HostServiceJobs, protocol.HostServiceMethodJobsRun, jobIDRequest{JobID: string(id)}, nil)
}

// SetStatus changes one visible scheduled-job lifecycle status.
func (s jobsService) SetStatus(_ context.Context, id jobcap.JobID, status jobv1.Status) error {
	return s.callJSONRequest(protocol.HostServiceJobs, protocol.HostServiceMethodJobsSetStatus, jobsSetStatusRequest{
		JobID:  string(id),
		Status: string(status),
	}, nil)
}

// jobIDsToStrings converts scheduled-job IDs to transport strings.
func jobIDsToStrings(ids []jobcap.JobID) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, string(id))
	}
	return out
}

// jobsListRequest carries governed scheduled-job list parameters.
type jobsListRequest struct {
	Keyword  string `json:"keyword,omitempty"`
	Group    string `json:"group,omitempty"`
	Status   string `json:"status,omitempty"`
	PageNum  int    `json:"pageNum,omitempty"`
	PageSize int    `json:"pageSize,omitempty"`
}

type jobIDRequest struct {
	JobID string `json:"jobId"`
}

type jobsSetStatusRequest struct {
	JobID  string `json:"jobId"`
	Status string `json:"status"`
}

var _ jobcap.Service = (*jobsService)(nil)
