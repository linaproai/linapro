// This file defines stable JSON payload models used by the dynamic sample
// plugin backend responses and demo helper records.

package dynamicservice

import (
	"encoding/json"

	"github.com/gogf/gf/v2/errors/gerror"
)

// backendSummaryPayload defines the summary response returned by the dynamic
// sample backend endpoint.
type backendSummaryPayload struct {
	Message       string  `json:"message"`
	PluginID      string  `json:"pluginId"`
	PublicPath    string  `json:"publicPath"`
	Access        string  `json:"access"`
	Permission    string  `json:"permission"`
	Authenticated bool    `json:"authenticated"`
	Username      *string `json:"username,omitempty"`
	IsSuperAdmin  *bool   `json:"isSuperAdmin,omitempty"`
}

// demoRecordListPayload defines the paged list response returned by sample
// record queries.
type demoRecordListPayload struct {
	List  []*demoRecordPayload `json:"list"`
	Total int                  `json:"total"`
}

// demoRecordPayload defines one demo-record JSON payload returned by the
// dynamic sample API.
type demoRecordPayload struct {
	Id             string `json:"id"`
	Title          string `json:"title"`
	Content        string `json:"content"`
	AttachmentName string `json:"attachmentName"`
	HasAttachment  bool   `json:"hasAttachment"`
	CreatedAt      string `json:"createdAt"`
	UpdatedAt      string `json:"updatedAt"`
}

// demoRecordDeletePayload defines the delete response returned by sample CRUD
// endpoints.
type demoRecordDeletePayload struct {
	Id      string `json:"id"`
	Deleted bool   `json:"deleted"`
}

// demoRecordAttachmentDownloadPayload describes one attachment download
// response before controller serialization.
type demoRecordAttachmentDownloadPayload struct {
	OriginalName string
	ContentType  string
	Body         []byte
}

// demoRecordEntity is the internal sample record representation loaded from
// structured data storage.
type demoRecordEntity struct {
	Id             string `json:"id"`
	Title          string `json:"title"`
	Content        string `json:"content"`
	AttachmentName string `json:"attachmentName"`
	AttachmentPath string `json:"attachmentPath"`
	CreatedAt      string `json:"createdAt"`
	UpdatedAt      string `json:"updatedAt"`
}

// hostCallDemoPayload defines the top-level host-service demo response.
type hostCallDemoPayload struct {
	VisitCount int                        `json:"visitCount"`
	PluginID   string                     `json:"pluginId"`
	Runtime    hostCallDemoRuntimePayload `json:"runtime"`
	Storage    hostCallDemoStoragePayload `json:"storage"`
	Network    hostCallDemoNetworkPayload `json:"network"`
	Data       hostCallDemoDataPayload    `json:"data"`
	Message    string                     `json:"message"`
}

// hostCallDemoRuntimePayload summarizes runtime host-call results.
type hostCallDemoRuntimePayload struct {
	Now  string `json:"now"`
	UUID string `json:"uuid"`
	Node string `json:"node"`
}

// hostCallDemoStoragePayload summarizes storage host-call results.
type hostCallDemoStoragePayload struct {
	PathPrefix  string `json:"pathPrefix"`
	ObjectPath  string `json:"objectPath"`
	Stored      bool   `json:"stored"`
	ListedCount int    `json:"listedCount"`
	Deleted     bool   `json:"deleted"`
}

// hostCallDemoStorageRecord is the JSON object persisted into governed storage
// during the demo.
type hostCallDemoStorageRecord struct {
	PluginID string `json:"pluginId"`
	DemoKey  string `json:"demoKey"`
}

// hostCallDemoDataPayload summarizes structured-data host-call results.
type hostCallDemoDataPayload struct {
	Table      string `json:"table"`
	RecordKey  string `json:"recordKey"`
	ListTotal  int    `json:"listTotal"`
	CountTotal int    `json:"countTotal"`
	Updated    bool   `json:"updated"`
	Deleted    bool   `json:"deleted"`
}

// hostCallDemoDataCreateRecord defines the sample record inserted during the
// structured-data demo.
type hostCallDemoDataCreateRecord struct {
	PluginID     string `json:"pluginId"`
	ReleaseID    int    `json:"releaseId"`
	NodeKey      string `json:"nodeKey"`
	DesiredState string `json:"desiredState"`
	CurrentState string `json:"currentState"`
	Generation   int    `json:"generation"`
	ErrorMessage string `json:"errorMessage"`
}

// hostCallDemoDataUpdateRecord defines the sample fields updated during the
// structured-data demo.
type hostCallDemoDataUpdateRecord struct {
	CurrentState string `json:"currentState"`
	ErrorMessage string `json:"errorMessage"`
}

// hostCallDemoNetworkPayload summarizes outbound network host-call results.
type hostCallDemoNetworkPayload struct {
	URL         string `json:"url"`
	Skipped     bool   `json:"skipped"`
	StatusCode  int    `json:"statusCode"`
	ContentType string `json:"contentType"`
	BodyPreview string `json:"bodyPreview"`
	Error       string `json:"error"`
}

// boolPointer allocates one boolean pointer for optional JSON response fields.
func boolPointer(value bool) *bool {
	return &value
}

// stringPointer allocates one string pointer for optional JSON response fields.
func stringPointer(value string) *string {
	return &value
}

// buildRecordMap marshals one typed record into a generic map used by the
// structured-data client.
func buildRecordMap(record any) (map[string]any, error) {
	content, err := json.Marshal(record)
	if err != nil {
		return nil, gerror.Wrap(err, "marshal demo record failed")
	}

	payload := make(map[string]any)
	if err = json.Unmarshal(content, &payload); err != nil {
		return nil, gerror.Wrap(err, "unmarshal demo record failed")
	}
	return payload, nil
}

// parseDemoRecordEntity converts one generic structured-data row into the
// internal demoRecordEntity shape.
func parseDemoRecordEntity(record map[string]any) (*demoRecordEntity, error) {
	content, err := json.Marshal(record)
	if err != nil {
		return nil, gerror.Wrap(err, "marshal dynamic demo record failed")
	}

	entity := &demoRecordEntity{}
	if err = json.Unmarshal(content, entity); err != nil {
		return nil, gerror.Wrap(err, "unmarshal dynamic demo record failed")
	}
	return entity, nil
}
