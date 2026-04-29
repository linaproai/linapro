// This file defines the public frontend-config API DTOs exposed to unauthenticated clients.

package v1

import "github.com/gogf/gf/v2/frame/g"

// FrontendReq defines the request for fetching public frontend config.
type FrontendReq struct {
	g.Meta `path:"/config/public/frontend" method:"get" tags:"Public Configuration" summary:"Get public frontend configuration" dc:"Return to the login page and the whitelist of brand, login display and interface style configuration that can be safely and publicly read during the startup phase of the management background"`
}

// FrontendRes defines the public frontend config response.
type FrontendRes struct {
	App  FrontendAppRes  `json:"app" dc:"Apply brand display configuration" eg:"{}"`
	Auth FrontendAuthRes `json:"auth" dc:"Login page display configuration" eg:"{}"`
	User FrontendUserRes `json:"user" dc:"User display configuration" eg:"{}"`
	UI   FrontendUIRes   `json:"ui" dc:"UI style configuration" eg:"{}"`
	Cron FrontendCronRes `json:"cron" dc:"Scheduled job frontend capability configuration" eg:"{}"`
}

// FrontendAppRes stores brand-related public settings.
type FrontendAppRes struct {
	Name     string `json:"name" dc:"Application name, used for browser title and workbench logo copywriting" eg:"LinaPro.AI"`
	Logo     string `json:"logo" dc:"Default logo image address" eg:"/logo.png"`
	LogoDark string `json:"logoDark" dc:"Dark theme logo image address" eg:"/logo.png"`
}

// FrontendAuthRes stores login-page public copy settings.
type FrontendAuthRes struct {
	PageTitle     string `json:"pageTitle" dc:"Login page main title copywriting" eg:"An AI-native full-stack framework engineered for sustainable delivery"`
	PageDesc      string `json:"pageDesc" dc:"Login page description copy" eg:"Facing business evolution, it provides out-of-the-box management portals and flexible pluggable expansion mechanisms."`
	LoginSubtitle string `json:"loginSubtitle" dc:"Login form subtitle copywriting" eg:"Please enter your account information to enter the LinaPro hosting workspace"`
	PanelLayout   string `json:"panelLayout" dc:"Login box layout: panel-left=left panel-center=center panel-right=right" eg:"panel-right"`
}

// FrontendUserRes stores user-facing public fallback settings.
type FrontendUserRes struct {
	DefaultAvatar string `json:"defaultAvatar" dc:"The default avatar address used when the user does not set an avatar" eg:"/avatar.webp"`
}

// FrontendUIRes stores public-safe theme and layout preferences.
type FrontendUIRes struct {
	ThemeMode        string `json:"themeMode" dc:"Theme mode: light=light dark=dark auto=follow the system" eg:"light"`
	Layout           string `json:"layout" dc:"Backend default layout: sidebar-nav, sidebar-mixed-nav, header-nav, header-sidebar-nav, header-mixed-nav, mixed-nav, full-content" eg:"sidebar-nav"`
	WatermarkEnabled bool   `json:"watermarkEnabled" dc:"Whether to enable watermark: true=enable false=disable" eg:"false"`
	WatermarkContent string `json:"watermarkContent" dc:"Watermark copy content" eg:"LinaPro"`
}

// FrontendCronRes stores public-safe scheduled-job capability flags.
type FrontendCronRes struct {
	LogRetention FrontendCronLogRetentionRes `json:"logRetention" dc:"System-level scheduled job log retention policy" eg:"{}"`
	Shell        FrontendCronShellRes        `json:"shell" dc:"Shell task frontend capability switch" eg:"{}"`
	Timezone     FrontendCronTimezoneRes     `json:"timezone" dc:"Default time zone configuration for scheduled jobs" eg:"{}"`
}

// FrontendCronLogRetentionRes stores the frontend-visible default log-retention policy.
type FrontendCronLogRetentionRes struct {
	Mode  string `json:"mode" dc:"System-level log retention mode: days=retain by day count=retain by number of entries none=do not automatically clean up" eg:"days"`
	Value int64  `json:"value" dc:"System-level log retention threshold; greater than 0 when mode=days or count, 0 when mode=none" eg:"30"`
}

// FrontendCronShellRes stores the shell-job gate exposed to the frontend.
type FrontendCronShellRes struct {
	Enabled           bool   `json:"enabled" dc:"Whether to allow creation and execution of Shell tasks: true=allowed false=not allowed" eg:"false"`
	Supported         bool   `json:"supported" dc:"Whether the current platform supports Shell tasks: true=supported false=not supported" eg:"true"`
	DisabledReason    string `json:"disabledReason" dc:"English fallback explaining why the Shell task is unavailable" eg:"The current platform does not support shell mode"`
	DisabledReasonKey string `json:"disabledReasonKey" dc:"Runtime i18n key for localizing disabledReason on the frontend" eg:"config.cron.shell.unsupportedReason"`
}

// FrontendCronTimezoneRes stores the default timezone exposed to the frontend.
type FrontendCronTimezoneRes struct {
	Current string `json:"current" dc:"The current host system time zone identifier, used as the default value for the new task time zone" eg:"Asia/Shanghai"`
}
