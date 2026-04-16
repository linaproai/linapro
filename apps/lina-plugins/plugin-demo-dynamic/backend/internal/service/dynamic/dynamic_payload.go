// This file defines stable JSON payload models used by the dynamic sample
// plugin backend responses and demo helper records.

package dynamicservice

import (
	"encoding/json"

	"github.com/gogf/gf/v2/errors/gerror"
)

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

type hostCallDemoPayload struct {
	VisitCount int                        `json:"visitCount"`
	PluginID   string                     `json:"pluginId"`
	Runtime    hostCallDemoRuntimePayload `json:"runtime"`
	Storage    hostCallDemoStoragePayload `json:"storage"`
	Network    hostCallDemoNetworkPayload `json:"network"`
	Data       hostCallDemoDataPayload    `json:"data"`
	Message    string                     `json:"message"`
}

type hostCallDemoRuntimePayload struct {
	Now  string `json:"now"`
	UUID string `json:"uuid"`
	Node string `json:"node"`
}

type hostCallDemoStoragePayload struct {
	PathPrefix  string `json:"pathPrefix"`
	ObjectPath  string `json:"objectPath"`
	Stored      bool   `json:"stored"`
	ListedCount int    `json:"listedCount"`
	Deleted     bool   `json:"deleted"`
}

type hostCallDemoStorageRecord struct {
	PluginID string `json:"pluginId"`
	DemoKey  string `json:"demoKey"`
}

type hostCallDemoDataPayload struct {
	Table      string `json:"table"`
	RecordKey  string `json:"recordKey"`
	ListTotal  int    `json:"listTotal"`
	CountTotal int    `json:"countTotal"`
	Updated    bool   `json:"updated"`
	Deleted    bool   `json:"deleted"`
}

type hostCallDemoDataCreateRecord struct {
	PluginID     string `json:"pluginId"`
	ReleaseID    int    `json:"releaseId"`
	NodeKey      string `json:"nodeKey"`
	DesiredState string `json:"desiredState"`
	CurrentState string `json:"currentState"`
	Generation   int    `json:"generation"`
	ErrorMessage string `json:"errorMessage"`
}

type hostCallDemoDataUpdateRecord struct {
	CurrentState string `json:"currentState"`
	ErrorMessage string `json:"errorMessage"`
}

type hostCallDemoNetworkPayload struct {
	URL         string `json:"url"`
	Skipped     bool   `json:"skipped"`
	StatusCode  int    `json:"statusCode"`
	ContentType string `json:"contentType"`
	BodyPreview string `json:"bodyPreview"`
	Error       string `json:"error"`
}

func boolPointer(value bool) *bool {
	return &value
}

func stringPointer(value string) *string {
	return &value
}

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
