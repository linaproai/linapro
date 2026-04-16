package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// System Info API

type GetInfoReq struct {
	g.Meta `path:"/system/info" method:"get" tags:"系统信息" summary:"获取系统运行信息" dc:"获取系统运行时信息，包括Go版本、GoFrame版本、操作系统、数据库版本、启动时间、运行时长以及前后端组件列表" permission:"about:system:list"`
}

// ComponentInfo Component information
type ComponentInfo struct {
	Name        string `json:"name" dc:"组件名称" eg:"GoFrame"`
	Version     string `json:"version" dc:"组件版本" eg:"v2.10.0"`
	Url         string `json:"url" dc:"组件主页URL" eg:"https://goframe.org"`
	Description string `json:"description" dc:"组件描述" eg:"Go语言应用开发框架"`
}

// GetInfoRes System runtime info response
type GetInfoRes struct {
	GoVersion          string          `json:"goVersion" dc:"Go版本" eg:"go1.22.0"`
	GfVersion          string          `json:"gfVersion" dc:"GoFrame版本" eg:"v2.10.0"`
	Os                 string          `json:"os" dc:"操作系统" eg:"linux"`
	Arch               string          `json:"arch" dc:"系统架构" eg:"amd64"`
	DbVersion          string          `json:"dbVersion" dc:"数据库版本" eg:"MySQL 8.0.36"`
	StartTime          string          `json:"startTime" dc:"系统启动时间" eg:"2025-01-01 08:00:00"`
	RunDuration        string          `json:"runDuration" dc:"系统运行时长" eg:"3天5小时20分钟"`
	BackendComponents  []ComponentInfo `json:"backendComponents" dc:"后端组件列表" eg:"[]"`
	FrontendComponents []ComponentInfo `json:"frontendComponents" dc:"前端组件列表" eg:"[]"`
}
