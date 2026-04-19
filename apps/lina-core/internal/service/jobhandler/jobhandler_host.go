// This file registers host-provided scheduled-job handlers.

package jobhandler

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/service/jobmeta"
)

// LogCleaner defines the host cleanup capability used by the built-in
// cleanup-job-logs handler.
type LogCleaner interface {
	// CleanupDueLogs removes logs that exceed the effective retention policy.
	CleanupDueLogs(ctx context.Context) (int64, error)
}

// RegisterHostHandlers installs host-provided handlers required by the current iteration.
func RegisterHostHandlers(registry Registry, cleaner LogCleaner) error {
	if registry == nil {
		return gerror.New("任务处理器注册表不能为空")
	}
	if cleaner == nil {
		return gerror.New("任务日志清理器不能为空")
	}

	if err := registerCleanupLogsHandler(registry, cleaner); err != nil {
		return err
	}
	return registerWaitHandler(registry)
}

// registerCleanupLogsHandler installs the built-in cleanup-job-logs handler.
func registerCleanupLogsHandler(registry Registry, cleaner LogCleaner) error {
	return registry.Register(HandlerDef{
		Ref:          "host:cleanup-job-logs",
		DisplayName:  "任务日志清理",
		Description:  "按系统默认或任务覆盖策略清理过期的定时任务执行日志",
		ParamsSchema: `{"type":"object","properties":{}}`,
		Source:       jobmeta.HandlerSourceHost,
		Invoke: func(ctx context.Context, _ json.RawMessage) (result any, err error) {
			deleted, cleanupErr := cleaner.CleanupDueLogs(ctx)
			if cleanupErr != nil {
				return nil, cleanupErr
			}
			return map[string]any{
				"deletedCount": deleted,
			}, nil
		},
	})
}

// waitHandlerParams stores the supported payload for the built-in wait handler.
type waitHandlerParams struct {
	Seconds int `json:"seconds"`
}

// registerWaitHandler installs one host-side diagnostic handler used to verify
// timeout, cancellation, and scheduler execution flow without side effects.
func registerWaitHandler(registry Registry) error {
	return registry.Register(HandlerDef{
		Ref:          "host:wait",
		DisplayName:  "等待指定时长",
		Description:  "按参数等待指定秒数后返回，用于验证调度链路、超时控制和取消行为",
		ParamsSchema: `{"type":"object","properties":{"seconds":{"type":"integer","description":"等待秒数"}},"required":["seconds"]}`,
		Source:       jobmeta.HandlerSourceHost,
		Invoke:       invokeWaitHandler,
	})
}

// invokeWaitHandler blocks for the requested duration unless the execution
// context is cancelled or times out first.
func invokeWaitHandler(ctx context.Context, params json.RawMessage) (any, error) {
	var input waitHandlerParams
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, gerror.Wrap(err, "解析等待处理器参数失败")
	}
	if input.Seconds <= 0 {
		return nil, gerror.New("等待秒数必须大于0")
	}

	timer := time.NewTimer(time.Duration(input.Seconds) * time.Second)
	defer timer.Stop()

	select {
	case <-timer.C:
		return map[string]any{
			"waitSeconds": input.Seconds,
		}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
