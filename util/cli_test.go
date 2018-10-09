package util

import "testing"

func TestGetHomePath(t *testing.T) {
	home := GetHomePath()
	if home == "" {
		t.Errorf("util.GetHomePath(), home shoud not be empty. Got :%q", home)
	}
}
