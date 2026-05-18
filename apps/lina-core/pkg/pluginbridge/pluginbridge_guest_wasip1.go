//go:build wasip1

// This file forwards WASI-only guest host-service client helpers to the guest
// subcomponent while preserving the historical root package entrypoints.

package pluginbridge

import "lina-core/pkg/pluginbridge/guest"

type (
	RuntimeHostService    = guest.RuntimeHostService
	StorageHostService    = guest.StorageHostService
	HTTPHostService       = guest.HTTPHostService
	DataHostService       = guest.DataHostService
	CacheHostService      = guest.CacheHostService
	LockHostService       = guest.LockHostService
	ConfigHostService     = guest.ConfigHostService
	NotifyHostService     = guest.NotifyHostService
	CronHostService       = guest.CronHostService
	HostDBQueryResult     = guest.HostDBQueryResult
	DataListResult        = guest.DataListResult
	DataGetResult         = guest.DataGetResult
	DataMutationResult    = guest.DataMutationResult
	DataTransactionInput  = guest.DataTransactionInput
	DataTransactionResult = guest.DataTransactionResult
)

var (
	Runtime         = guest.Runtime
	Storage         = guest.Storage
	HTTP            = guest.HTTP
	Data            = guest.Data
	Cache           = guest.Cache
	Lock            = guest.Lock
	Config          = guest.Config
	Notify          = guest.Notify
	Cron            = guest.Cron
	HostLog         = guest.HostLog
	HostStateGet    = guest.HostStateGet
	HostStateSet    = guest.HostStateSet
	HostStateDelete = guest.HostStateDelete
	HostStateGetInt = guest.HostStateGetInt
	HostStateSetInt = guest.HostStateSetInt
	HostDBQuery     = guest.HostDBQuery
	HostDBExecute   = guest.HostDBExecute
)
