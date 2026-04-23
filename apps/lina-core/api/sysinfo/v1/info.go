// This file defines DTOs for the system information API payloads.

package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// System Info API

// GetInfoReq requests the current runtime system information payload.
type GetInfoReq struct {
	g.Meta `path:"/system/info" method:"get" tags:"系统信息" summary:"获取系统运行信息" dc:"获取系统运行时信息，包括Go版本、GoFrame版本、操作系统、数据库版本、启动时间、运行时长以及前后端组件列表" permission:"about:system:list"`
}

// ComponentInfo Component information
type ComponentInfo struct {
	Name        string `json:"name" dc:"组件名称" eg:"GoFrame"`
	Version     string `json:"version" dc:"组件版本" eg:"v2.10.0"`
	Url         string `json:"url" dc:"组件主页URL" eg:"https://goframe.org"`
	Description string `json:"description" dc:"组件描述" eg:"Go语言开发框架"`
}

// FrameworkInfo Framework information
type FrameworkInfo struct {
	Name          string `json:"name" dc:"框架名称" eg:"LinaPro"`
	Version       string `json:"version" dc:"框架版本号" eg:"v0.5.0"`
	Description   string `json:"description" dc:"框架介绍" eg:"AI驱动的全栈开发框架"`
	Homepage      string `json:"homepage" dc:"项目官网" eg:"https://linapro.ai"`
	RepositoryURL string `json:"repositoryUrl" dc:"仓库地址" eg:"https://github.com/gqcn/linapro"`
	License       string `json:"license" dc:"开源协议" eg:"MIT"`
}

// GetInfoRes System runtime info response
type GetInfoRes struct {
	Framework          FrameworkInfo   `json:"framework" dc:"框架信息" eg:"{}"`
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
