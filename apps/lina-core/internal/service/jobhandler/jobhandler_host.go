// This file registers host-provided scheduled-job handlers.

package jobhandler

import (
	"context"
	"encoding/json"
	"time"

	"lina-core/internal/service/jobmeta"
	"lina-core/pkg/bizerr"
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
		return bizerr.NewCode(CodeJobHandlerRegistryRequired)
	}
	if cleaner == nil {
		return bizerr.NewCode(CodeJobHandlerLogCleanerRequired)
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
		DisplayName:  "Job Log Cleanup",
		Description:  "Cleans up scheduled-job execution logs according to global and job-level retention policies.",
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
		DisplayName:  "Wait for Duration",
		Description:  "Waits for the requested number of seconds to verify scheduling, timeout, and cancellation flows.",
		ParamsSchema: `{"type":"object","properties":{"seconds":{"type":"integer","description":"Wait seconds"}},"required":["seconds"]}`,
		Source:       jobmeta.HandlerSourceHost,
		Invoke:       invokeWaitHandler,
	})
}

// invokeWaitHandler blocks for the requested duration unless the execution
// context is cancelled or times out first.
func invokeWaitHandler(ctx context.Context, params json.RawMessage) (any, error) {
	var input waitHandlerParams
	if err := json.Unmarshal(params, &input); err != nil {
		return nil, bizerr.WrapCode(err, CodeJobHandlerWaitParamsParseFailed)
	}
	if input.Seconds <= 0 {
		return nil, bizerr.NewCode(CodeJobHandlerWaitSecondsInvalid)
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
