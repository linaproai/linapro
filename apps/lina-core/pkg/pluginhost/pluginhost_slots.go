// This file declares published backend extension points, execution modes, and
// hook-action enums that together form the stable pluginhost contract.

package pluginhost

// ExtensionPoint defines one published backend plugin extension point.
type ExtensionPoint string

// ExtensionKind defines the category of one backend extension point.
type ExtensionKind string

// CallbackExecutionMode defines how the host executes one registered callback.
type CallbackExecutionMode string

// ExtensionPointDefinition describes one published backend extension point.
type ExtensionPointDefinition struct {
	// DefaultMode is the default callback execution mode for the current point.
	DefaultMode CallbackExecutionMode
	// Kind is the category of the current extension point.
	Kind ExtensionKind
}

const (
	// ExtensionKindHook defines one event-driven hook extension point.
	ExtensionKindHook ExtensionKind = "hook"
	// ExtensionKindRegistrar defines one callback-registration extension point.
	ExtensionKindRegistrar ExtensionKind = "registrar"
)

const (
	// CallbackExecutionModeBlocking executes the callback in the current host flow.
	CallbackExecutionModeBlocking CallbackExecutionMode = "blocking"
	// CallbackExecutionModeAsync executes the callback in background without blocking the current host flow.
	CallbackExecutionModeAsync CallbackExecutionMode = "async"
)

const (
	// ExtensionPointAuthLoginSucceeded is fired after user login succeeds.
	ExtensionPointAuthLoginSucceeded ExtensionPoint = "auth.login.succeeded"
	// ExtensionPointAuthLoginFailed is fired after user login fails.
	ExtensionPointAuthLoginFailed ExtensionPoint = "auth.login.failed"
	// ExtensionPointAuthLogoutSucceeded is fired after user logout succeeds.
	ExtensionPointAuthLogoutSucceeded ExtensionPoint = "auth.logout.succeeded"
	// ExtensionPointPluginInstalled is fired after a dynamic plugin is installed.
	ExtensionPointPluginInstalled ExtensionPoint = "plugin.installed"
	// ExtensionPointPluginEnabled is fired after a plugin is enabled.
	ExtensionPointPluginEnabled ExtensionPoint = "plugin.enabled"
	// ExtensionPointPluginDisabled is fired after a plugin is disabled.
	ExtensionPointPluginDisabled ExtensionPoint = "plugin.disabled"
	// ExtensionPointPluginUninstalled is fired after a dynamic plugin is uninstalled.
	ExtensionPointPluginUninstalled ExtensionPoint = "plugin.uninstalled"
	// ExtensionPointSystemStarted is fired after host HTTP server startup.
	ExtensionPointSystemStarted ExtensionPoint = "system.started"
	// ExtensionPointHTTPRouteRegister registers plugin-owned HTTP routes at host startup.
	ExtensionPointHTTPRouteRegister ExtensionPoint = "http.route.register"
	// ExtensionPointHTTPRequestAfterAuth dispatches one callback after host authentication succeeds.
	ExtensionPointHTTPRequestAfterAuth ExtensionPoint = "http.request.after-auth"
	// ExtensionPointCronRegister registers plugin-owned cron jobs.
	ExtensionPointCronRegister ExtensionPoint = "cron.register"
	// ExtensionPointMenuFilter filters host menus.
	ExtensionPointMenuFilter ExtensionPoint = "menu.filter"
	// ExtensionPointPermissionFilter filters host permissions.
	ExtensionPointPermissionFilter ExtensionPoint = "permission.filter"
)

// HookAction defines one supported plugin hook action.
type HookAction string

const (
	// HookActionInsert inserts one row into plugin-owned table.
	HookActionInsert HookAction = "insert"
	// HookActionSleep blocks for the declared duration until the host timeout fires or the delay completes.
	HookActionSleep HookAction = "sleep"
	// HookActionError returns one configured error immediately for runtime isolation verification.
	HookActionError HookAction = "error"
)

var publishedExtensionPoints = map[ExtensionPoint]ExtensionPointDefinition{
	ExtensionPointAuthLoginSucceeded:  {Kind: ExtensionKindHook, DefaultMode: CallbackExecutionModeBlocking},
	ExtensionPointAuthLoginFailed:     {Kind: ExtensionKindHook, DefaultMode: CallbackExecutionModeBlocking},
	ExtensionPointAuthLogoutSucceeded: {Kind: ExtensionKindHook, DefaultMode: CallbackExecutionModeBlocking},
	ExtensionPointPluginInstalled:     {Kind: ExtensionKindHook, DefaultMode: CallbackExecutionModeBlocking},
	ExtensionPointPluginEnabled:       {Kind: ExtensionKindHook, DefaultMode: CallbackExecutionModeBlocking},
	ExtensionPointPluginDisabled:      {Kind: ExtensionKindHook, DefaultMode: CallbackExecutionModeBlocking},
	ExtensionPointPluginUninstalled:   {Kind: ExtensionKindHook, DefaultMode: CallbackExecutionModeBlocking},
	ExtensionPointSystemStarted:       {Kind: ExtensionKindHook, DefaultMode: CallbackExecutionModeBlocking},
	ExtensionPointHTTPRouteRegister:   {Kind: ExtensionKindRegistrar, DefaultMode: CallbackExecutionModeBlocking},
	ExtensionPointHTTPRequestAfterAuth: {
		Kind:        ExtensionKindRegistrar,
		DefaultMode: CallbackExecutionModeBlocking,
	},
	ExtensionPointCronRegister:     {Kind: ExtensionKindRegistrar, DefaultMode: CallbackExecutionModeBlocking},
	ExtensionPointMenuFilter:       {Kind: ExtensionKindRegistrar, DefaultMode: CallbackExecutionModeBlocking},
	ExtensionPointPermissionFilter: {Kind: ExtensionKindRegistrar, DefaultMode: CallbackExecutionModeBlocking},
}

var supportedExtensionPointModes = map[ExtensionPoint]map[CallbackExecutionMode]struct{}{
	ExtensionPointAuthLoginSucceeded: {
		CallbackExecutionModeBlocking: {},
		CallbackExecutionModeAsync:    {},
	},
	ExtensionPointAuthLoginFailed: {
		CallbackExecutionModeBlocking: {},
		CallbackExecutionModeAsync:    {},
	},
	ExtensionPointAuthLogoutSucceeded: {
		CallbackExecutionModeBlocking: {},
		CallbackExecutionModeAsync:    {},
	},
	ExtensionPointPluginInstalled: {
		CallbackExecutionModeBlocking: {},
		CallbackExecutionModeAsync:    {},
	},
	ExtensionPointPluginEnabled: {
		CallbackExecutionModeBlocking: {},
		CallbackExecutionModeAsync:    {},
	},
	ExtensionPointPluginDisabled: {
		CallbackExecutionModeBlocking: {},
		CallbackExecutionModeAsync:    {},
	},
	ExtensionPointPluginUninstalled: {
		CallbackExecutionModeBlocking: {},
		CallbackExecutionModeAsync:    {},
	},
	ExtensionPointSystemStarted: {
		CallbackExecutionModeBlocking: {},
		CallbackExecutionModeAsync:    {},
	},
	ExtensionPointHTTPRouteRegister: {
		CallbackExecutionModeBlocking: {},
	},
	ExtensionPointHTTPRequestAfterAuth: {
		CallbackExecutionModeBlocking: {},
	},
	ExtensionPointCronRegister: {
		CallbackExecutionModeBlocking: {},
	},
	ExtensionPointMenuFilter: {
		CallbackExecutionModeBlocking: {},
	},
	ExtensionPointPermissionFilter: {
		CallbackExecutionModeBlocking: {},
	},
}

var publishedCallbackExecutionModes = map[CallbackExecutionMode]struct{}{
	CallbackExecutionModeBlocking: {},
	CallbackExecutionModeAsync:    {},
}

var publishedHookActions = map[HookAction]struct{}{
	HookActionInsert: {},
	HookActionSleep:  {},
	HookActionError:  {},
}

// String returns the canonical backend extension point key.
func (point ExtensionPoint) String() string {
	return string(point)
}

// String returns the canonical backend extension kind.
func (kind ExtensionKind) String() string {
	return string(kind)
}

// String returns the canonical callback execution mode.
func (mode CallbackExecutionMode) String() string {
	return string(mode)
}

// DefaultCallbackExecutionMode returns the default execution mode of one published backend extension point.
func DefaultCallbackExecutionMode(point ExtensionPoint) CallbackExecutionMode {
	definition, ok := publishedExtensionPoints[point]
	if !ok {
		return ""
	}
	return definition.DefaultMode
}

// IsPublishedExtensionPoint reports whether the backend extension point is part of the published host contract.
func IsPublishedExtensionPoint(point ExtensionPoint) bool {
	_, ok := publishedExtensionPoints[point]
	return ok
}

// IsHookExtensionPoint reports whether the backend extension point is an event-driven hook.
func IsHookExtensionPoint(point ExtensionPoint) bool {
	definition, ok := publishedExtensionPoints[point]
	return ok && definition.Kind == ExtensionKindHook
}

// IsRegistrationExtensionPoint reports whether the backend extension point is a callback-registration point.
func IsRegistrationExtensionPoint(point ExtensionPoint) bool {
	definition, ok := publishedExtensionPoints[point]
	return ok && definition.Kind == ExtensionKindRegistrar
}

// IsPublishedCallbackExecutionMode reports whether the execution mode is supported by current host runtime.
func IsPublishedCallbackExecutionMode(mode CallbackExecutionMode) bool {
	_, ok := publishedCallbackExecutionModes[mode]
	return ok
}

// IsExtensionPointExecutionModeSupported reports whether the backend extension point supports the given execution mode.
func IsExtensionPointExecutionModeSupported(
	point ExtensionPoint,
	mode CallbackExecutionMode,
) bool {
	modes, ok := supportedExtensionPointModes[point]
	if !ok {
		return false
	}
	_, ok = modes[mode]
	return ok
}

// String returns the canonical hook action key.
func (action HookAction) String() string {
	return string(action)
}

// IsSupportedHookAction reports whether the hook action is supported by current host runtime.
func IsSupportedHookAction(action HookAction) bool {
	_, ok := publishedHookActions[action]
	return ok
}
