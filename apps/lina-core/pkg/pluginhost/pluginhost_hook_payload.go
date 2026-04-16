package pluginhost

// HookPayloadKey defines one published field name inside a host hook payload.
type HookPayloadKey string

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
