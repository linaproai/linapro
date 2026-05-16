// This file contains small value helpers shared by source-plugin payload and
// snapshot wrappers.

package pluginhost

// stringValueFromMap returns one string field from a shallow map.
func stringValueFromMap(values map[string]interface{}, key string) string {
	if len(values) == 0 {
		return ""
	}
	value, _ := values[key].(string)
	return value
}

// cloneValueMap returns a shallow copy of the given payload value map.
func cloneValueMap(values map[string]interface{}) map[string]interface{} {
	if len(values) == 0 {
		return map[string]interface{}{}
	}
	cloned := make(map[string]interface{}, len(values))
	for key, value := range values {
		cloned[key] = value
	}
	return cloned
}
