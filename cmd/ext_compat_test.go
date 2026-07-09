package cmd

import (
	"os"
	"strings"
	"testing"
)

func TestExtCommandDoesNotDependOnCompatShims(t *testing.T) {
	src, err := os.ReadFile("ext.go")
	if err != nil {
		t.Fatalf("read ext.go: %v", err)
	}
	for _, helper := range []string{
		"describeUHostByID(",
		"getUhostList(",
		"sbindEIP(",
		"unbindEIP(",
	} {
		if strings.Contains(string(src), helper) {
			t.Fatalf("ext.go must not call compat helper %s", helper)
		}
	}
	for _, path := range []string{"eip_compat.go", "uhost_compat.go"} {
		if _, err := os.Stat(path); err == nil {
			t.Fatalf("%s must be removed after ext owns its SDK helpers", path)
		} else if !os.IsNotExist(err) {
			t.Fatalf("stat %s: %v", path, err)
		}
	}
}
