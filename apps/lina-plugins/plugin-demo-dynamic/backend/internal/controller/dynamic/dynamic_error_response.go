// This file translates dynamic demo business errors into bridge responses
// via the shared pluginbridge ErrorClassifier composition.

package dynamic

import (
	"lina-core/pkg/pluginbridge"
	dynamicservice "lina-plugin-demo-dynamic/backend/internal/service/dynamic"
)

// dynamicErrorClassifier maps dynamic sample business errors to normalized
// bridge responses. BindJSON sentinels are handled by pluginbridge itself.
var dynamicErrorClassifier = pluginbridge.NewErrorClassifier(
	pluginbridge.NewErrorCase(dynamicservice.IsDemoRecordInvalidInput, pluginbridge.NewBadRequestResponse),
	pluginbridge.NewErrorCase(dynamicservice.IsDemoRecordNotFound, pluginbridge.NewNotFoundResponse),
)

// wrapDynamicError converts one dynamic sample business error into a
// prebuilt bridge response error so typed guest controllers can return it
// through the standard error channel.
func wrapDynamicError(err error) error {
	if err == nil {
		return nil
	}
	return pluginbridge.NewResponseError(dynamicErrorClassifier.Classify(err))
}
