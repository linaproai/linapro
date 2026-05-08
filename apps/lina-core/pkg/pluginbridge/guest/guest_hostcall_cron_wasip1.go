//go:build wasip1

// This file provides guest-side helpers for the governed cron registration
// host service used by dynamic-plugin scheduled-job discovery.

package guest

import "github.com/gogf/gf/v2/errors/gerror"

// CronHostService exposes guest-side helpers for the cron registration host
// service.
type CronHostService interface {
	// Register submits one dynamic-plugin cron declaration to the current
	// host-side discovery collector.
	Register(contract *CronContract) error
}

// cronHostService is the default guest-side cron host-service client.
type cronHostService struct{}

// defaultCronHostService stores the singleton cron host-service client used by
// package-level helpers.
var defaultCronHostService CronHostService = &cronHostService{}

// Cron returns the cron host service guest client.
func Cron() CronHostService {
	return defaultCronHostService
}

// Register submits one dynamic-plugin cron declaration to the current
// host-side discovery collector.
func (s *cronHostService) Register(contract *CronContract) error {
	if contract == nil {
		return gerror.New("cron contract cannot be nil")
	}
	contractSnapshot := *contract
	_, err := invokeHostService(
		HostServiceCron,
		HostServiceMethodCronRegister,
		"",
		"",
		MarshalHostServiceCronRegisterRequest(&HostServiceCronRegisterRequest{Contract: &contractSnapshot}),
	)
	return err
}
