// This file defines the public frontend-config API DTOs exposed to unauthenticated clients.

package v1

import "github.com/gogf/gf/v2/frame/g"

// FrontendReq defines the request for fetching public frontend config.
type FrontendReq struct {
	g.Meta `path:"/config/public/frontend" method:"get" tags:"公共配置" summary:"获取公共前端配置" dc:"返回登录页与管理后台启动阶段可安全公开读取的品牌、登录展示和界面风格配置白名单"`
}

// FrontendRes defines the public frontend config response.
type FrontendRes struct {
	App  FrontendAppRes  `json:"app" dc:"应用品牌展示配置" eg:"{}"`
	Auth FrontendAuthRes `json:"auth" dc:"登录页展示配置" eg:"{}"`
	UI   FrontendUIRes   `json:"ui" dc:"界面风格配置" eg:"{}"`
	Cron FrontendCronRes `json:"cron" dc:"定时任务前端能力配置" eg:"{}"`
}

// FrontendAppRes stores brand-related public settings.
type FrontendAppRes struct {
	Name     string `json:"name" dc:"应用名称，用于浏览器标题和工作台 Logo 文案" eg:"LinaPro"`
	Logo     string `json:"logo" dc:"默认 Logo 图片地址" eg:"https://unpkg.com/@vbenjs/static-source@0.1.7/source/logo-v1.webp"`
	LogoDark string `json:"logoDark" dc:"深色主题 Logo 图片地址" eg:"https://unpkg.com/@vbenjs/static-source@0.1.7/source/logo-v1.webp"`
}

// FrontendAuthRes stores login-page public copy settings.
type FrontendAuthRes struct {
	PageTitle     string `json:"pageTitle" dc:"登录页主标题文案" eg:"AI驱动的全栈开发框架"`
	PageDesc      string `json:"pageDesc" dc:"登录页说明文案" eg:"面向业务演进，提供开箱即用的管理入口与灵活可插拔的扩展机制"`
	LoginSubtitle string `json:"loginSubtitle" dc:"登录表单副标题文案" eg:"请输入您的帐户信息以进入 LinaPro 宿主工作区"`
	PanelLayout   string `json:"panelLayout" dc:"登录框布局：panel-left=居左 panel-center=居中 panel-right=居右" eg:"panel-right"`
}

// FrontendUIRes stores public-safe theme and layout preferences.
type FrontendUIRes struct {
	ThemeMode        string `json:"themeMode" dc:"主题模式：light=浅色 dark=深色 auto=跟随系统" eg:"light"`
	Layout           string `json:"layout" dc:"后台默认布局：sidebar-nav、sidebar-mixed-nav、header-nav、header-sidebar-nav、header-mixed-nav、mixed-nav、full-content" eg:"sidebar-nav"`
	WatermarkEnabled bool   `json:"watermarkEnabled" dc:"是否启用水印：true=启用 false=关闭" eg:"false"`
	WatermarkContent string `json:"watermarkContent" dc:"水印文案内容" eg:"LinaPro"`
}

// FrontendCronRes stores public-safe scheduled-job capability flags.
type FrontendCronRes struct {
	LogRetention FrontendCronLogRetentionRes `json:"logRetention" dc:"系统级定时任务日志保留策略" eg:"{}"`
	Shell        FrontendCronShellRes        `json:"shell" dc:"Shell 任务前端能力开关" eg:"{}"`
	Timezone     FrontendCronTimezoneRes     `json:"timezone" dc:"定时任务默认时区配置" eg:"{}"`
}

// FrontendCronLogRetentionRes stores the frontend-visible default log-retention policy.
type FrontendCronLogRetentionRes struct {
	Mode  string `json:"mode" dc:"系统级日志保留模式：days=按天保留 count=按条数保留 none=不自动清理" eg:"days"`
	Value int64  `json:"value" dc:"系统级日志保留阈值；mode=days 或 count 时大于0，mode=none 时为0" eg:"30"`
}

// FrontendCronShellRes stores the shell-job gate exposed to the frontend.
type FrontendCronShellRes struct {
	Enabled        bool   `json:"enabled" dc:"是否允许创建和执行 Shell 任务：true=允许 false=不允许" eg:"false"`
	Supported      bool   `json:"supported" dc:"当前平台是否支持 Shell 任务：true=支持 false=不支持" eg:"true"`
	DisabledReason string `json:"disabledReason" dc:"Shell 任务不可用时的原因说明" eg:"当前平台不支持 shell 模式"`
}

// FrontendCronTimezoneRes stores the default timezone exposed to the frontend.
type FrontendCronTimezoneRes struct {
	Current string `json:"current" dc:"当前宿主系统时区标识，作为新增任务时区的默认值" eg:"Asia/Shanghai"`
}
