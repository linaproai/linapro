// This file defines built-in cron handler business error codes and their i18n
// metadata.

package cron

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeCronSessionCleanupDependencyMissing reports that session cleanup dependencies are missing.
	CodeCronSessionCleanupDependencyMissing = bizerr.MustDefine(
		"CRON_SESSION_CLEANUP_DEPENDENCY_MISSING",
		"Online-session cleanup dependency is not initialized",
		gcode.CodeInternalError,
	)
	// CodeCronAccessTopologyDependencyMissing reports that access topology sync dependencies are missing.
	CodeCronAccessTopologyDependencyMissing = bizerr.MustDefine(
		"CRON_ACCESS_TOPOLOGY_DEPENDENCY_MISSING",
		"Access topology sync dependency is not initialized",
		gcode.CodeInternalError,
	)
	// CodeCronRuntimeParamDependencyMissing reports that runtime parameter sync dependencies are missing.
	CodeCronRuntimeParamDependencyMissing = bizerr.MustDefine(
		"CRON_RUNTIME_PARAM_DEPENDENCY_MISSING",
		"Runtime parameter sync dependency is not initialized",
		gcode.CodeInternalError,
	)
	// CodeCronKVCacheDependencyMissing reports that KV cache cleanup dependencies are missing.
	CodeCronKVCacheDependencyMissing = bizerr.MustDefine(
		"CRON_KVCACHE_DEPENDENCY_MISSING",
		"KV cache cleanup dependency is not initialized",
		gcode.CodeInternalError,
	)
)
