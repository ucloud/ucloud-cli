package utils

import (
	"fmt"
	"strings"
)

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

// IsStringIn will return if the value is contains by an array
func IsStringIn(val string, avaliables []string) bool {
	for _, choice := range avaliables {
		if val == choice {
			return true
		}
	}

	return false
}

// CheckStringIn will check if the value is contains by an array
func CheckStringIn(val string, avaliables []string) error {
	if IsStringIn(val, avaliables) {
		return nil
	}
	return fmt.Errorf("got %s, should be one of %s", val, strings.Join(avaliables, ","))
}
