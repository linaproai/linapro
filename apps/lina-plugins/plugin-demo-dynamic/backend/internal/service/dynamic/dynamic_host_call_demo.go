// This file implements the host service demo business logic for the dynamic
// sample plugin.

package dynamicservice

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/pluginbridge"
)

// Host-call demo constants define the governed keys, paths, and sample values
// used by the dynamic plugin host-service showcase.
const (
	hostCallDemoStateKey           = "host_call_demo_visit_count"
	hostCallDemoStoragePath        = "host-call-demo/"
	hostCallDemoStoragePrefix      = "host-call-demo"
	hostCallDemoStorageContentType = "application/json"
	hostCallDemoNetworkURL         = "https://example.com"
	hostCallDemoNetworkMethodGet   = "GET"
	hostCallDemoDataTable          = "sys_plugin_node_state"
	hostCallDemoDesiredState       = "running"
	hostCallDemoCurrentStateNew    = "pending"
	hostCallDemoCurrentStateReady  = "running"
	hostCallDemoAnonymousUser      = "anonymous"
	hostCallDemoSummaryMessage     = "Host service demo executed through runtime, storage, network, and data services."
	hostCallDemoNetworkPreview     = 120
)

// BuildHostCallDemoPayload executes the host service demo and returns the
// response payload.
func (s *serviceImpl) BuildHostCallDemoPayload(input *HostCallDemoInput) (*hostCallDemoPayload, error) {
	username := hostCallDemoAnonymousUser
	if input != nil && strings.TrimSpace(input.Username) != "" {
		username = strings.TrimSpace(input.Username)
	}

	nowValue, err := s.runtimeSvc.Now()
	if err != nil {
		return nil, err
	}
	uuidValue, err := s.runtimeSvc.UUID()
	if err != nil {
		return nil, err
	}
	nodeValue, err := s.runtimeSvc.Node()
	if err != nil {
		return nil, err
	}
	if err = s.runtimeSvc.Log(
		int(pluginbridge.LogLevelInfo),
		"host service demo invoked",
		map[string]string{
			"username":  username,
			"requestId": hostCallDemoRequestID(input),
			"route":     hostCallDemoRoutePath(input),
			"demoKey":   uuidValue,
		},
	); err != nil {
		return nil, err
	}

	visitCount, found, err := s.runtimeSvc.StateGetInt(hostCallDemoStateKey)
	if err != nil || !found {
		visitCount = 0
	}
	visitCount++
	_ = s.runtimeSvc.StateSetInt(hostCallDemoStateKey, visitCount)

	storageSummary, err := s.runHostCallDemoStorage(hostCallDemoPluginID(input), uuidValue)
	if err != nil {
		return nil, err
	}
	dataSummary, err := s.runHostCallDemoData(hostCallDemoPluginID(input), uuidValue)
	if err != nil {
		return nil, err
	}
	networkSummary := s.runHostCallDemoNetwork(input, uuidValue)

	return &hostCallDemoPayload{
		VisitCount: visitCount,
		PluginID:   hostCallDemoPluginID(input),
		Runtime: hostCallDemoRuntimePayload{
			Now:  nowValue,
			UUID: uuidValue,
			Node: nodeValue,
		},
		Storage: *storageSummary,
		Network: *networkSummary,
		Data:    *dataSummary,
		Message: hostCallDemoSummaryMessage,
	}, nil
}

// runHostCallDemoStorage exercises governed storage APIs and summarizes the
// round-trip result.
func (s *serviceImpl) runHostCallDemoStorage(pluginID string, demoKey string) (*hostCallDemoStoragePayload, error) {
	objectPath := fmt.Sprintf("%s/%s.json", hostCallDemoStoragePrefix, demoKey)
	body, err := json.Marshal(&hostCallDemoStorageRecord{
		PluginID: pluginID,
		DemoKey:  demoKey,
	})
	if err != nil {
		return nil, gerror.Wrap(err, "marshal storage demo request body failed")
	}
	if _, err = s.storageSvc.Put(objectPath, body, hostCallDemoStorageContentType, true); err != nil {
		return nil, err
	}
	deleted := false
	defer func() {
		if !deleted {
			_ = s.storageSvc.Delete(objectPath)
		}
	}()

	readBody, _, found, err := s.storageSvc.Get(objectPath)
	if err != nil {
		return nil, err
	}
	if !found || string(readBody) != string(body) {
		return nil, gerror.New("storage demo object verification failed")
	}

	objects, err := s.storageSvc.List(hostCallDemoStoragePrefix, 10)
	if err != nil {
		return nil, err
	}
	if err = s.storageSvc.Delete(objectPath); err != nil {
		return nil, err
	}
	deleted = true

	_, statFound, err := s.storageSvc.Stat(objectPath)
	if err != nil {
		return nil, err
	}
	return &hostCallDemoStoragePayload{
		PathPrefix:  hostCallDemoStoragePath,
		ObjectPath:  objectPath,
		Stored:      true,
		ListedCount: len(objects),
		Deleted:     !statFound,
	}, nil
}

// runHostCallDemoData exercises governed structured-data APIs and summarizes
// the create/list/update/delete flow.
func (s *serviceImpl) runHostCallDemoData(pluginID string, demoKey string) (*hostCallDemoDataPayload, error) {
	createRecord, err := buildRecordMap(&hostCallDemoDataCreateRecord{
		PluginID:     pluginID,
		ReleaseID:    0,
		NodeKey:      "host-call-demo-" + demoKey,
		DesiredState: hostCallDemoDesiredState,
		CurrentState: hostCallDemoCurrentStateNew,
		Generation:   1,
		ErrorMessage: "",
	})
	if err != nil {
		return nil, err
	}
	createResult, err := s.dataSvc.Table(hostCallDemoDataTable).Insert(createRecord)
	if err != nil {
		return nil, err
	}
	if createResult == nil || createResult.Key == nil {
		return nil, gerror.New("data demo create did not return a record key")
	}

	recordKey := createResult.Key
	deleted := false
	defer func() {
		if !deleted {
			_, _ = s.dataSvc.Table(hostCallDemoDataTable).WhereKey(recordKey).Delete()
		}
	}()

	listRecords, listTotal, err := s.dataSvc.Table(hostCallDemoDataTable).
		Fields("id", "nodeKey", "currentState").
		WhereEq("pluginId", pluginID).
		WhereLike("nodeKey", demoKey).
		WhereIn("currentState", []string{hostCallDemoCurrentStateNew, hostCallDemoCurrentStateReady}).
		OrderDesc("id").
		Page(1, 10).
		All()
	if err != nil {
		return nil, err
	}
	if listTotal < 1 || len(listRecords) == 0 {
		return nil, gerror.New("data demo list did not find the created record")
	}
	countTotal, err := s.dataSvc.Table(hostCallDemoDataTable).
		WhereEq("pluginId", pluginID).
		WhereLike("nodeKey", demoKey).
		Count()
	if err != nil {
		return nil, err
	}
	recordKey = listRecords[0]["id"]

	updateRecord, err := buildRecordMap(&hostCallDemoDataUpdateRecord{
		CurrentState: hostCallDemoCurrentStateReady,
		ErrorMessage: "",
	})
	if err != nil {
		return nil, err
	}
	if _, err = s.dataSvc.Table(hostCallDemoDataTable).WhereKey(recordKey).Update(updateRecord); err != nil {
		return nil, err
	}

	record, found, err := s.dataSvc.Table(hostCallDemoDataTable).Fields("currentState").WhereKey(recordKey).One()
	if err != nil {
		return nil, err
	}
	if !found || record == nil || fmt.Sprint(record["currentState"]) != hostCallDemoCurrentStateReady {
		return nil, gerror.New("data demo get did not return the updated record")
	}

	if _, err = s.dataSvc.Table(hostCallDemoDataTable).WhereKey(recordKey).Delete(); err != nil {
		return nil, err
	}
	deleted = true

	return &hostCallDemoDataPayload{
		Table:      hostCallDemoDataTable,
		RecordKey:  fmt.Sprint(recordKey),
		ListTotal:  int(listTotal),
		CountTotal: int(countTotal),
		Updated:    true,
		Deleted:    true,
	}, nil
}

// runHostCallDemoNetwork exercises the governed outbound HTTP host service and
// captures a bounded preview of the response.
func (s *serviceImpl) runHostCallDemoNetwork(input *HostCallDemoInput, demoKey string) *hostCallDemoNetworkPayload {
	result := &hostCallDemoNetworkPayload{
		URL:         hostCallDemoNetworkURL,
		Skipped:     false,
		StatusCode:  0,
		ContentType: "",
		BodyPreview: "",
		Error:       "",
	}
	if input != nil && input.SkipNetwork {
		result.Skipped = true
		return result
	}

	response, err := s.httpSvc.Request(hostCallDemoNetworkURL, &pluginbridge.HostServiceNetworkRequest{
		Method: hostCallDemoNetworkMethodGet,
		Headers: map[string]string{
			"x-request-id": hostCallDemoRequestID(input) + "-" + demoKey,
		},
	})
	if err != nil {
		result.Error = err.Error()
		return result
	}
	result.StatusCode = int(response.StatusCode)
	result.ContentType = response.ContentType
	result.BodyPreview = buildHostCallDemoBodyPreview(response.Body)
	return result
}

// hostCallDemoPluginID returns the normalized plugin identifier from the input.
func hostCallDemoPluginID(input *HostCallDemoInput) string {
	if input == nil {
		return ""
	}
	return strings.TrimSpace(input.PluginID)
}

// hostCallDemoRequestID returns the normalized request identifier from the
// input.
func hostCallDemoRequestID(input *HostCallDemoInput) string {
	if input == nil {
		return ""
	}
	return strings.TrimSpace(input.RequestID)
}

// hostCallDemoRoutePath returns the normalized route path from the input.
func hostCallDemoRoutePath(input *HostCallDemoInput) string {
	if input == nil {
		return ""
	}
	return strings.TrimSpace(input.RoutePath)
}

// buildHostCallDemoBodyPreview truncates one response body to the configured
// preview length.
func buildHostCallDemoBodyPreview(body []byte) string {
	preview := strings.TrimSpace(string(body))
	if preview == "" {
		return ""
	}
	if len(preview) <= hostCallDemoNetworkPreview {
		return preview
	}
	return preview[:hostCallDemoNetworkPreview]
}
