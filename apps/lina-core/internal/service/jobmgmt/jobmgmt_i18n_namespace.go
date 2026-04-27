// This file registers scheduled-job source-text namespaces with the host i18n
// missing-message diagnostics.

package jobmgmt

import (
	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/internal/service/jobmeta"
)

// init registers job-management source-text-backed translation namespaces.
func init() {
	i18nsvc.RegisterSourceTextNamespace(
		jobmeta.HandlerI18nKeyPrefix+".",
		"scheduled-job handler metadata is supplied by host or plugin source definitions",
	)
	i18nsvc.RegisterSourceTextNamespace(
		"job.group.default.",
		"the default scheduled-job group metadata is supplied by job management source definitions",
	)
}
