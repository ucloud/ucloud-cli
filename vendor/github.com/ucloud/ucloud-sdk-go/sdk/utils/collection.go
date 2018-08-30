package utils

// MergeMap will merge two map and return a new map
func MergeMap(args ...map[string]string) map[string]string {
	m := map[string]string{}
	for _, kv := range args {
		for k, v := range kv {
			m[k] = v
		}
	}
	return m
}

// SetMapIfNotExists will set a
func SetMapIfNotExists(m map[string]string, k string, v string) {
	if _, ok := m[k]; !ok && v != "" {
		m[k] = v
	}
}
