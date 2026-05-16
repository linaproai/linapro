// This file contains small value helpers shared by source-plugin payload wrappers.

package pluginhost

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
