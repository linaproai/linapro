// This file defines the public cron registrar contract exposed to source
// plugins and the guarded host-side implementation used at runtime.

package pluginhost

import (
	"context"
	"lina-core/pkg/logger"

	"github.com/gogf/gf/v2/os/gcron"
)

// PrimaryNodeChecker defines one host callback that reports whether the current node is the primary node.
type PrimaryNodeChecker func() bool

// CronJobHandler defines one plugin-owned cron job callback.
type CronJobHandler func(ctx context.Context) error

// CronRegistrar exposes host cron registration and node-role inspection for one plugin.
type CronRegistrar interface {
	// Add registers one guarded cron job.
	Add(ctx context.Context, pattern string, name string, handler CronJobHandler) error
	// IsPrimaryNode reports whether the current host node is the primary node.
	IsPrimaryNode() bool
}

type cronRegistrar struct {
	pluginID           string
	enabledChecker     PluginEnabledChecker
	primaryNodeChecker PrimaryNodeChecker
}

// NewCronRegistrar creates and returns a new CronRegistrar instance.
func NewCronRegistrar(
	pluginID string,
	enabledChecker PluginEnabledChecker,
	primaryNodeChecker PrimaryNodeChecker,
) CronRegistrar {
	return &cronRegistrar{
		pluginID:           pluginID,
		enabledChecker:     enabledChecker,
		primaryNodeChecker: primaryNodeChecker,
	}
}

// Add registers one guarded cron job.
func (r *cronRegistrar) Add(
	ctx context.Context,
	pattern string,
	name string,
	handler CronJobHandler,
) error {
	if handler == nil {
		panic("pluginhost: cron handler is nil")
	}

	_, err := gcron.Add(ctx, pattern, func(jobCtx context.Context) {
		if r.enabledChecker != nil && !r.enabledChecker(r.pluginID) {
			return
		}
		// Guard every cron callback at runtime so disabling a plugin immediately stops
		// future executions without requiring host restart or plugin re-registration.
		if runErr := handler(jobCtx); runErr != nil {
			logger.Warningf(jobCtx, "plugin cron failed plugin=%s name=%s err=%v", r.pluginID, name, runErr)
		}
	}, name)
	return err
}

// IsPrimaryNode reports whether the current host node is the primary node.
func (r *cronRegistrar) IsPrimaryNode() bool {
	if r == nil || r.primaryNodeChecker == nil {
		return true
	}
	return r.primaryNodeChecker()
}
