// demo_record_helper.go implements shared helpers for plugin-demo-source record controllers.

package demo

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
