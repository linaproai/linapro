// This file declares published hook payload keys and helper constructors shared
// by host hook dispatchers and source plugins.

package pluginhost

// HookPayloadKey defines one published field name inside a host hook payload.
type HookPayloadKey string

// Published hook payload field names.
const (
	// HookPayloadKeyPluginID identifies the current plugin targeted by lifecycle events.
	HookPayloadKeyPluginID HookPayloadKey = "pluginId"
	// HookPayloadKeyPluginName stores the plugin display name for lifecycle events.
	HookPayloadKeyPluginName HookPayloadKey = "name"
	// HookPayloadKeyPluginVersion stores the plugin version for lifecycle events.
	HookPayloadKeyPluginVersion HookPayloadKey = "version"
	// HookPayloadKeyStatus stores the status code associated with the current event.
	HookPayloadKeyStatus HookPayloadKey = "status"
	// HookPayloadKeyUserName stores the authenticated username for auth hook events.
	HookPayloadKeyUserName HookPayloadKey = "userName"
	// HookPayloadKeyIP stores the client IP for auth hook events.
	HookPayloadKeyIP HookPayloadKey = "ip"
	// HookPayloadKeyClientType stores the client type for auth hook events.
	HookPayloadKeyClientType HookPayloadKey = "clientType"
	// HookPayloadKeyBrowser stores the browser description for auth hook events.
	HookPayloadKeyBrowser HookPayloadKey = "browser"
	// HookPayloadKeyOS stores the operating-system description for auth hook events.
	HookPayloadKeyOS HookPayloadKey = "os"
	// HookPayloadKeyMessage stores the audit message for auth hook events.
	HookPayloadKeyMessage HookPayloadKey = "message"
	// HookPayloadKeyTitle stores the audit title for request audit hook events.
	HookPayloadKeyTitle HookPayloadKey = "title"
	// HookPayloadKeyOperSummary stores the audit operation summary.
	HookPayloadKeyOperSummary HookPayloadKey = "operSummary"
	// HookPayloadKeyOperType stores the audit operation type code.
	HookPayloadKeyOperType HookPayloadKey = "operType"
	// HookPayloadKeyMethod stores the routed handler path or method marker.
	HookPayloadKeyMethod HookPayloadKey = "method"
	// HookPayloadKeyRequestMethod stores the HTTP request method.
	HookPayloadKeyRequestMethod HookPayloadKey = "requestMethod"
	// HookPayloadKeyOperName stores the operator username recorded by the audit event.
	HookPayloadKeyOperName HookPayloadKey = "operName"
	// HookPayloadKeyOperURL stores the full request URL captured by the audit event.
	HookPayloadKeyOperURL HookPayloadKey = "operUrl"
	// HookPayloadKeyOperParam stores the sanitized request payload captured by the audit event.
	HookPayloadKeyOperParam HookPayloadKey = "operParam"
	// HookPayloadKeyJSONResult stores the serialized response body captured by the audit event.
	HookPayloadKeyJSONResult HookPayloadKey = "jsonResult"
	// HookPayloadKeyErrorMsg stores the error summary captured by the audit event.
	HookPayloadKeyErrorMsg HookPayloadKey = "errorMsg"
	// HookPayloadKeyCostTime stores the elapsed request time in milliseconds.
	HookPayloadKeyCostTime HookPayloadKey = "costTime"
)

// AuthHookPayloadInput defines the published auth hook payload fields.
type AuthHookPayloadInput struct {
	UserName   string
	Status     int
	IP         string
	ClientType string
	Browser    string
	OS         string
	Message    string
}

// PluginLifecycleHookPayloadInput defines the published plugin lifecycle hook fields.
type PluginLifecycleHookPayloadInput struct {
	PluginID string
	Name     string
	Version  string
	Status   *int
}

// AuditHookPayloadInput defines the published request-audit hook payload fields.
type AuditHookPayloadInput struct {
	Title         string
	OperSummary   string
	OperType      int
	Method        string
	RequestMethod string
	OperName      string
	OperURL       string
	OperIP        string
	OperParam     string
	JSONResult    string
	Status        int
	ErrorMsg      string
	CostTime      int
}

// String returns the canonical published hook payload field name.
func (key HookPayloadKey) String() string {
	return string(key)
}

// BuildAuthHookPayloadValues creates the published auth-event payload map.
func BuildAuthHookPayloadValues(input AuthHookPayloadInput) map[string]interface{} {
	return map[string]interface{}{
		HookPayloadKeyUserName.String():   input.UserName,
		HookPayloadKeyStatus.String():     input.Status,
		HookPayloadKeyIP.String():         input.IP,
		HookPayloadKeyClientType.String(): input.ClientType,
		HookPayloadKeyBrowser.String():    input.Browser,
		HookPayloadKeyOS.String():         input.OS,
		HookPayloadKeyMessage.String():    input.Message,
	}
}

// BuildPluginLifecycleHookPayloadValues creates the published plugin lifecycle payload map.
func BuildPluginLifecycleHookPayloadValues(input PluginLifecycleHookPayloadInput) map[string]interface{} {
	values := map[string]interface{}{
		HookPayloadKeyPluginID.String():      input.PluginID,
		HookPayloadKeyPluginName.String():    input.Name,
		HookPayloadKeyPluginVersion.String(): input.Version,
	}
	if input.Status != nil {
		values[HookPayloadKeyStatus.String()] = *input.Status
	}
	return values
}

// BuildAuditHookPayloadValues creates the published request-audit payload map.
func BuildAuditHookPayloadValues(input AuditHookPayloadInput) map[string]interface{} {
	return map[string]interface{}{
		HookPayloadKeyTitle.String():         input.Title,
		HookPayloadKeyOperSummary.String():   input.OperSummary,
		HookPayloadKeyOperType.String():      input.OperType,
		HookPayloadKeyMethod.String():        input.Method,
		HookPayloadKeyRequestMethod.String(): input.RequestMethod,
		HookPayloadKeyOperName.String():      input.OperName,
		HookPayloadKeyOperURL.String():       input.OperURL,
		HookPayloadKeyIP.String():            input.OperIP,
		HookPayloadKeyOperParam.String():     input.OperParam,
		HookPayloadKeyJSONResult.String():    input.JSONResult,
		HookPayloadKeyStatus.String():        input.Status,
		HookPayloadKeyErrorMsg.String():      input.ErrorMsg,
		HookPayloadKeyCostTime.String():      input.CostTime,
	}
}

// CloneHookPayloadValues returns a shallow copy of published hook payload values.
func CloneHookPayloadValues(values map[string]interface{}) map[string]interface{} {
	return cloneValueMap(values)
}

// HookPayloadStringValue extracts one string payload field from the published map.
func HookPayloadStringValue(values map[string]interface{}, key HookPayloadKey) string {
	if len(values) == 0 {
		return ""
	}
	value, _ := values[key.String()].(string)
	return value
}

// HookPayloadIntValue extracts one int payload field from the published map.
func HookPayloadIntValue(values map[string]interface{}, key HookPayloadKey) (int, bool) {
	if len(values) == 0 {
		return 0, false
	}
	value, ok := values[key.String()].(int)
	return value, ok
}
