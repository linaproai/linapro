// This file centralizes the optional install-time mock-data load helper used
// by both the source-plugin install path and the dynamic-plugin reconciler so
// the bizerr wrapping for a rolled-back load lives in one place.

package plugin

import (
	"errors"

	"lina-core/internal/service/plugin/internal/migration"
	"lina-core/pkg/bizerr"
)

// wrapMockDataLoadError converts a migration.MockDataLoadError into the stable
// user-facing bizerr that carries all parameters into i18n templates. Returns
// the original err unchanged when the chain does not contain a mock-data load
// error so callers can pass through arbitrary install errors safely.
func wrapMockDataLoadError(err error) error {
	if err == nil {
		return nil
	}
	var mockErr *migration.MockDataLoadError
	if !errors.As(err, &mockErr) {
		return err
	}
	causeText := ""
	if mockErr.Cause != nil {
		causeText = mockErr.Cause.Error()
	}
	return bizerr.NewCode(
		CodePluginInstallMockDataFailed,
		bizerr.P("pluginId", mockErr.PluginID),
		bizerr.P("failedFile", mockErr.FailedFile),
		bizerr.P("rolledBackFiles", mockErr.RolledBackFiles),
		bizerr.P("cause", causeText),
	)
}

// isMockDataLoadError reports whether err represents an install that succeeded
// except for the optional mock-data load phase.
func isMockDataLoadError(err error) bool {
	var mockErr *migration.MockDataLoadError
	return errors.As(err, &mockErr)
}
