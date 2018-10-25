package base

import "testing"

func TestGetHomePath(t *testing.T) {
	home := GetHomePath()
	if home == "" {
		t.Errorf("base.GetHomePath(), home shoud not be empty. Got :%q", home)
	}
}
