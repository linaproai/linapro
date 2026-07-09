// This file contains test-only helpers for runtime i18n scanner assertions.

package runtimei18n

// scanRuntimeI18N scans source files and returns all non-allowlisted findings.
func scanRuntimeI18N(repoRoot string, options scanOptions) ([]scanFinding, error) {
	report, err := scanRuntimeI18NReport(repoRoot, options)
	if err != nil {
		return nil, err
	}
	return report.Findings, nil
}
