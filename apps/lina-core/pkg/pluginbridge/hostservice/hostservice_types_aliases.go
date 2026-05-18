// This file aliases bridge contract types used by host-service codecs.

package hostservice

import "lina-core/pkg/pluginbridge/contract"

type (
	CronContract    = contract.CronContract
	CronScope       = contract.CronScope
	CronConcurrency = contract.CronConcurrency
)

const (
	DefaultCronContractTimezone       = contract.DefaultCronContractTimezone
	DefaultCronContractTimeoutSeconds = contract.DefaultCronContractTimeoutSeconds
	CronScopeMasterOnly               = contract.CronScopeMasterOnly
	CronScopeAllNode                  = contract.CronScopeAllNode
	CronConcurrencySingleton          = contract.CronConcurrencySingleton
	CronConcurrencyParallel           = contract.CronConcurrencyParallel
)
