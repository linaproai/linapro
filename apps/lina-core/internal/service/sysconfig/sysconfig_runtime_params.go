// This file coordinates protected system-parameter mutations with the host
// protected-config cache revision.

package sysconfig

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/internal/dao"
	hostconfig "lina-core/internal/service/config"
)

// withConfigMutation runs one sysconfig mutation inside the shared transaction
// boundary used for runtime-param refresh coordination.
func (s *serviceImpl) withConfigMutation(ctx context.Context, handler func(ctx context.Context) error) error {
	return dao.SysConfig.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		return handler(ctx)
	})
}

// refreshRuntimeParamSnapshotIfNeeded marks the protected runtime/public
// frontend config snapshot dirty when a managed value changes.
func (s *serviceImpl) refreshRuntimeParamSnapshotIfNeeded(
	ctx context.Context,
	key string,
	previousValue string,
	currentValue string,
	created bool,
) error {
	if !hostconfig.IsProtectedConfigParam(key) {
		return nil
	}
	if !created && previousValue == currentValue {
		return nil
	}
	return s.configSvc.MarkRuntimeParamsChanged(ctx)
}
