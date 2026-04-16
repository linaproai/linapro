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
func (s *serviceImpl) BuildHostCallDemoPayload(request *pluginbridge.BridgeRequestEnvelopeV1) (*hostCallDemoPayload, error) {
	username := hostCallDemoAnonymousUser
	if request.Identity != nil && request.Identity.Username != "" {
		username = request.Identity.Username
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
	if err = s.runtimeSvc.Log(int(pluginbridge.LogLevelInfo), "host service demo invoked", map[string]string{
		"username":  username,
		"requestId": request.RequestID,
		"route":     request.Route.InternalPath,
		"demoKey":   uuidValue,
	}); err != nil {
		return nil, err
	}

	visitCount, found, err := s.runtimeSvc.StateGetInt(hostCallDemoStateKey)
	if err != nil || !found {
		visitCount = 0
	}
	visitCount++
	_ = s.runtimeSvc.StateSetInt(hostCallDemoStateKey, visitCount)

	storageSummary, err := s.runHostCallDemoStorage(request.PluginID, uuidValue)
	if err != nil {
		return nil, err
	}
	dataSummary, err := s.runHostCallDemoData(request.PluginID, uuidValue)
	if err != nil {
		return nil, err
	}
	networkSummary := s.runHostCallDemoNetwork(request, uuidValue)

	return &hostCallDemoPayload{
		VisitCount: visitCount,
		PluginID:   request.PluginID,
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

func (s *serviceImpl) runHostCallDemoNetwork(request *pluginbridge.BridgeRequestEnvelopeV1, demoKey string) *hostCallDemoNetworkPayload {
	result := &hostCallDemoNetworkPayload{
		URL:         hostCallDemoNetworkURL,
		Skipped:     false,
		StatusCode:  0,
		ContentType: "",
		BodyPreview: "",
		Error:       "",
	}
	if hasHostCallDemoFlag(request, "skipNetwork") {
		result.Skipped = true
		return result
	}

	response, err := s.httpSvc.Request(hostCallDemoNetworkURL, &pluginbridge.HostServiceNetworkRequest{
		Method: hostCallDemoNetworkMethodGet,
		Headers: map[string]string{
			"x-request-id": request.RequestID + "-" + demoKey,
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

func hasHostCallDemoFlag(request *pluginbridge.BridgeRequestEnvelopeV1, key string) bool {
	if request == nil || request.Route == nil || len(request.Route.QueryValues) == 0 {
		return false
	}
	values := request.Route.QueryValues[key]
	for _, value := range values {
		switch strings.ToLower(strings.TrimSpace(value)) {
		case "1", "true", "yes", "on":
			return true
		}
	}
	return false
}

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
