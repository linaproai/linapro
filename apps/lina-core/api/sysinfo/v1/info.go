// This file defines DTOs for the system information API payloads.

package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// System Info API

// GetInfoReq requests the current runtime system information payload.
type GetInfoReq struct {
	g.Meta `path:"/system/info" method:"get" tags:"System Information" summary:"Get system runtime information" dc:"Obtain system runtime information, including Go version, GoFrame version, operating system, database version, startup time, running time, frontend and backend component lists, and cache coordination diagnostics" permission:"about:system:list"`
}

// ComponentInfo Component information
type ComponentInfo struct {
	Name        string `json:"name" dc:"Component name" eg:"GoFrame"`
	Version     string `json:"version" dc:"Component version" eg:"v2.10.0"`
	Url         string `json:"url" dc:"Component home page URL" eg:"https://goframe.org"`
	Description string `json:"description" dc:"Component description" eg:"Go language development framework"`
}

// FrameworkInfo Framework information
type FrameworkInfo struct {
	Name          string `json:"name" dc:"Framework name" eg:"LinaPro"`
	Version       string `json:"version" dc:"Framework version number" eg:"v0.5.0"`
	Description   string `json:"description" dc:"Framework introduction" eg:"An AI-native full-stack framework engineered for sustainable delivery"`
	Homepage      string `json:"homepage" dc:"Project official website" eg:"https://linapro.ai"`
	RepositoryURL string `json:"repositoryUrl" dc:"Repository URL" eg:"https://github.com/linaproai/linapro"`
	License       string `json:"license" dc:"Open source license" eg:"MIT"`
}

// CacheCoordinationInfo describes one cache coordination domain and scope.
type CacheCoordinationInfo struct {
	Domain           string `json:"domain" dc:"Cache coordination domain identifier" eg:"runtime-config"`
	Scope            string `json:"scope" dc:"Explicit cache invalidation scope inside the domain" eg:"global"`
	AuthoritySource  string `json:"authoritySource" dc:"Canonical data source used to rebuild the cache domain" eg:"sys_config protected runtime parameters"`
	ConsistencyModel string `json:"consistencyModel" dc:"Consistency model used by the cache domain" eg:"shared-revision"`
	MaxStaleSeconds  int64  `json:"maxStaleSeconds" dc:"Maximum acceptable stale window in seconds" eg:"10"`
	FailureStrategy  string `json:"failureStrategy" dc:"Caller-visible degradation strategy when freshness cannot be confirmed" eg:"return-visible-error"`
	LocalRevision    int64  `json:"localRevision" dc:"Latest revision consumed by this host process" eg:"5"`
	SharedRevision   int64  `json:"sharedRevision" dc:"Latest shared revision observed from the coordination store" eg:"5"`
	LastSyncedAt     string `json:"lastSyncedAt" dc:"Latest successful local synchronization time, empty when not yet synchronized" eg:"2025-01-01 08:00:00"`
	RecentError      string `json:"recentError" dc:"Most recent coordination failure message, empty when healthy" eg:""`
	StaleSeconds     int64  `json:"staleSeconds" dc:"Seconds elapsed since the latest successful local synchronization" eg:"1"`
}

// GetInfoRes System runtime info response
type GetInfoRes struct {
	Framework          FrameworkInfo           `json:"framework" dc:"frame information" eg:"{}"`
	GoVersion          string                  `json:"goVersion" dc:"Go version" eg:"go1.22.0"`
	GfVersion          string                  `json:"gfVersion" dc:"GoFrame version" eg:"v2.10.0"`
	Os                 string                  `json:"os" dc:"operating system" eg:"linux"`
	Arch               string                  `json:"arch" dc:"System architecture" eg:"amd64"`
	DbVersion          string                  `json:"dbVersion" dc:"Database version" eg:"PostgreSQL 14.0"`
	StartTime          string                  `json:"startTime" dc:"System startup time" eg:"2025-01-01 08:00:00"`
	RunDuration        string                  `json:"runDuration" dc:"System running time" eg:"3 days, 5 hours and 20 minutes"`
	RunDurationSeconds int64                   `json:"runDurationSeconds" dc:"System running time represented as total seconds for client-side structured formatting" eg:"12345"`
	BackendComponents  []ComponentInfo         `json:"backendComponents" dc:"Backend component list" eg:"[]"`
	FrontendComponents []ComponentInfo         `json:"frontendComponents" dc:"Front-end component list" eg:"[]"`
	CacheCoordination  []CacheCoordinationInfo `json:"cacheCoordination" dc:"Cache coordination diagnostics for critical runtime cache domains" eg:"[]"`
}
