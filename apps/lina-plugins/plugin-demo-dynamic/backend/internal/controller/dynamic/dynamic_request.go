// This file provides low-level bridge request readers shared by the dynamic
// plugin controllers.

package dynamic

import (
	"encoding/json"
	"strconv"
	"strings"

	"lina-core/pkg/pluginbridge"
	dynamicservice "lina-plugin-demo-dynamic/backend/internal/service/dynamic"
)

func decodeDemoRecordMutationBody(request *pluginbridge.BridgeRequestEnvelopeV1) (*dynamicservice.DemoRecordMutationInput, error) {
	if request == nil || request.Request == nil || len(request.Request.Body) == 0 {
		return nil, dynamicservice.NewDemoRecordInvalidInputError("请求体不能为空")
	}

	input := &dynamicservice.DemoRecordMutationInput{}
	if err := json.Unmarshal(request.Request.Body, input); err != nil {
		return nil, dynamicservice.NewDemoRecordInvalidInputError("请求体 JSON 无法解析")
	}
	return input, nil
}

func readDynamicPathParam(request *pluginbridge.BridgeRequestEnvelopeV1, key string) string {
	if request == nil || request.Route == nil || len(request.Route.PathParams) == 0 {
		return ""
	}
	return strings.TrimSpace(request.Route.PathParams[key])
}

func readDynamicQueryValue(request *pluginbridge.BridgeRequestEnvelopeV1, key string) string {
	if request == nil || request.Route == nil || len(request.Route.QueryValues) == 0 {
		return ""
	}

	values := request.Route.QueryValues[key]
	if len(values) == 0 {
		return ""
	}
	return strings.TrimSpace(values[0])
}

func readDynamicQueryInt(request *pluginbridge.BridgeRequestEnvelopeV1, key string) int {
	value := readDynamicQueryValue(request, key)
	if value == "" {
		return 0
	}

	parsedValue, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return parsedValue
}

func hasDynamicQueryFlag(request *pluginbridge.BridgeRequestEnvelopeV1, key string) bool {
	if request == nil || request.Route == nil || len(request.Route.QueryValues) == 0 {
		return false
	}

	for _, value := range request.Route.QueryValues[key] {
		switch strings.ToLower(strings.TrimSpace(value)) {
		case "1", "true", "yes", "on":
			return true
		}
	}
	return false
}
