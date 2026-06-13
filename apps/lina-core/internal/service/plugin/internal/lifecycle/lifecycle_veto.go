// This file contains shared lifecycle veto formatting used by source-plugin
// and dynamic-plugin lifecycle decisions.

package lifecycle

import (
	"strings"
)

// lifecycleVetoDecision is the shared projection used to summarize source and
// dynamic lifecycle decision outputs without duplicating formatting logic.
type lifecycleVetoDecision struct {
	// PluginID identifies the plugin that produced the decision.
	PluginID string
	// OK reports whether the lifecycle participant allowed the operation.
	OK bool
	// Reason is the participant-provided machine reason or message key.
	Reason string
	// Err carries an execution failure when no reason was provided.
	Err error
}

// summarizeLifecycleVetoDecisionReasons applies an optional translator to reason
// keys while preserving the existing plugin-prefixed reason format.
func summarizeLifecycleVetoDecisionReasons(
	decisions []lifecycleVetoDecision,
	translate func(key string) string,
) string {
	includePluginPrefix := translate == nil || countLifecycleVetoDecisions(decisions) > 1
	items := make([]string, 0, len(decisions))
	for _, decision := range decisions {
		if decision.OK {
			continue
		}
		reason := strings.TrimSpace(decision.Reason)
		if reason == "" && decision.Err != nil {
			reason = decision.Err.Error()
		}
		pluginID := strings.TrimSpace(decision.PluginID)
		if reason == "" {
			reason = "plugin." + pluginID + ".lifecycle.vetoed"
		}
		if translate != nil {
			if translated := strings.TrimSpace(translate(reason)); translated != "" {
				reason = translated
			}
		}
		if includePluginPrefix && pluginID != "" {
			items = append(items, pluginID+":"+reason)
			continue
		}
		items = append(items, reason)
	}
	if len(items) == 0 {
		return "unknown"
	}
	return strings.Join(items, ";")
}

// countLifecycleVetoDecisions returns how many lifecycle decisions blocked the action.
func countLifecycleVetoDecisions(decisions []lifecycleVetoDecision) int {
	count := 0
	for _, decision := range decisions {
		if !decision.OK {
			count++
		}
	}
	return count
}
