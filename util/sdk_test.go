package util

import "testing"

func TestGetHomePath(t *testing.T) {
	home := GetHomePath()
	if home == "" {
		t.Errorf("home shoud not be empty")
	}
}
