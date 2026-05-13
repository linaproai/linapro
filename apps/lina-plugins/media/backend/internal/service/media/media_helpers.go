// This file implements shared media service helpers.

package media

import (
	"context"

	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/pkg/bizerr"
	"lina-plugin-media/backend/internal/dao"
)

// normalizePagination applies list paging defaults and max-page-size limits.
func normalizePagination(pageNum int, pageSize int) (int, int) {
	if pageNum <= 0 {
		pageNum = defaultPageNum
	}
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	return pageNum, pageSize
}

// normalizeSwitchValue validates and normalizes a numeric switch value.
func normalizeSwitchValue(value int, defaultValue SwitchValue) (int, error) {
	if value == 0 {
		return int(defaultValue), nil
	}
	switch SwitchValue(value) {
	case SwitchOn, SwitchOff:
		return value, nil
	default:
		return 0, bizerr.NewCode(CodeMediaSwitchValueInvalid)
	}
}

// normalizeBinaryValue validates and normalizes a numeric yes/no value.
func normalizeBinaryValue(value int, defaultValue BinaryValue) (int, error) {
	if value == 0 {
		return int(defaultValue), nil
	}
	switch BinaryValue(value) {
	case BinaryYes:
		return value, nil
	default:
		return 0, bizerr.NewCode(CodeMediaBinaryValueInvalid)
	}
}

// currentActorID returns the user ID to persist in creator/updater fields.
func (s *serviceImpl) currentActorID(ctx context.Context) int64 {
	current := s.bizCtxSvc.Current(ctx)
	if current.ActingUserID > 0 {
		return int64(current.ActingUserID)
	}
	return int64(current.UserID)
}

// formatTime formats one optional GoFrame time value for API output.
func formatTime(value *gtime.Time) string {
	if value == nil {
		return ""
	}
	return value.String()
}

// validateMediaTablesReady verifies plugin-owned tables exist before business operations continue.
func validateMediaTablesReady(ctx context.Context) error {
	tableNames := []string{
		dao.MediaStrategy.Table(),
		dao.MediaStrategyDevice.Table(),
		dao.MediaStrategyDeviceTenant.Table(),
		dao.MediaStrategyTenant.Table(),
		dao.MediaStreamAlias.Table(),
	}
	for _, tableName := range tableNames {
		fields, err := dao.MediaStrategy.DB().TableFields(ctx, tableName)
		if err != nil {
			return bizerr.WrapCode(err, CodeMediaTableCheckFailed)
		}
		if len(fields) == 0 {
			return bizerr.NewCode(CodeMediaTableNotInstalled)
		}
	}
	return nil
}
