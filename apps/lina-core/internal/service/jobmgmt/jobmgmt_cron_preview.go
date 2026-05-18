// This file implements cron expression preview for scheduled jobs.

package jobmgmt

import (
	"context"
	"time"
)

// PreviewCron returns the next five fire times for one cron expression.
func (s *serviceImpl) PreviewCron(_ context.Context, expr string, timezone string) ([]time.Time, error) {
	_, location, err := normalizeJobTimezone(timezone)
	if err != nil {
		return nil, err
	}
	_, schedule, err := normalizeCronExpression(expr)
	if err != nil {
		return nil, err
	}

	current := time.Now().In(location)
	result := make([]time.Time, 0, 5)
	for len(result) < 5 {
		current = schedule.Next(current)
		result = append(result, current)
	}
	return result, nil
}
