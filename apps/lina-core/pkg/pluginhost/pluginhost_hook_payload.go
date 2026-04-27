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
	// HookPayloadKeyReason stores the stable reason code for auth hook events.
	HookPayloadKeyReason HookPayloadKey = "reason"
)

// Stable reason codes published with host authentication lifecycle events.
const (
	// AuthHookReasonLoginSuccessful identifies successful login events.
	AuthHookReasonLoginSuccessful = "loginSuccessful"
	// AuthHookReasonLoginFailed identifies generic failed login events.
	AuthHookReasonLoginFailed = "loginFailed"
	// AuthHookReasonLogoutSuccessful identifies successful logout events.
	AuthHookReasonLogoutSuccessful = "logoutSuccessful"
	// AuthHookReasonInvalidCredentials identifies invalid credential events.
	AuthHookReasonInvalidCredentials = "invalidCredentials"
	// AuthHookReasonUserDisabled identifies disabled account events.
	AuthHookReasonUserDisabled = "userDisabled"
	// AuthHookReasonIPBlacklisted identifies blocked login IP events.
	AuthHookReasonIPBlacklisted = "ipBlacklisted"
)

// AuthHookPayloadInput defines the published auth hook payload fields.
type AuthHookPayloadInput struct {
	// UserName is the authenticated username.
	UserName string
	// Status is the login status code associated with the auth event.
	Status int
	// IP is the client IP address.
	IP string
	// ClientType identifies the login client type.
	ClientType string
	// Browser is the detected browser description.
	Browser string
	// OS is the detected operating-system description.
	OS string
	// Message is the English fallback audit message delivered to plugins.
	Message string
	// Reason is the stable auth lifecycle reason code delivered to plugins.
	Reason string
}

// PluginLifecycleHookPayloadInput defines the published plugin lifecycle hook fields.
type PluginLifecycleHookPayloadInput struct {
	// PluginID is the immutable plugin identifier.
	PluginID string
	// Name is the plugin display name.
	Name string
	// Version is the plugin semantic version string.
	Version string
	// Status is the optional plugin enabled status code.
	Status *int
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
		HookPayloadKeyReason.String():     input.Reason,
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
