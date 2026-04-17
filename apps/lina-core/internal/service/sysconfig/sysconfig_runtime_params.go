// This file coordinates protected runtime-parameter mutations with the host
// runtime-parameter cache revision.

package sysconfig

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/internal/dao"
	hostconfig "lina-core/internal/service/config"
)

func (s *serviceImpl) withConfigMutation(ctx context.Context, handler func(ctx context.Context) error) error {
	return dao.SysConfig.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		return handler(ctx)
	})
}

func (s *serviceImpl) refreshRuntimeParamSnapshotIfNeeded(
	ctx context.Context,
	key string,
	previousValue string,
	currentValue string,
	created bool,
) error {
	if !hostconfig.IsProtectedRuntimeParam(key) {
		return nil
	}
	if !created && previousValue == currentValue {
		return nil
	}
	return s.configSvc.MarkRuntimeParamsChanged(ctx)
}
