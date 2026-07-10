package cmd

import (
	"os"
	"strings"
	"testing"
)

func TestExtCommandMigratedOutOfPlatformCmd(t *testing.T) {
	if _, err := os.Stat("ext.go"); err == nil {
		t.Fatal("cmd/ext.go must be removed after ext migrates to products/eip/internal/ext")
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat ext.go: %v", err)
	}

	src, err := os.ReadFile("root.go")
	if err != nil {
		t.Fatalf("read root.go: %v", err)
	}
	if contains := string(src); contains == "" {
		t.Fatal("root.go is unexpectedly empty")
	} else if strings.Contains(contains, "NewCmdExt(") {
		t.Fatal("cmd/root.go must not register NewCmdExt after ext migrates to products/eip")
	}
}

func TestExtCommandDoesNotDependOnCompatShims(t *testing.T) {
	for _, path := range []string{"eip_compat.go", "uhost_compat.go"} {
		if _, err := os.Stat(path); err == nil {
			t.Fatalf("%s must be removed after ext owns its SDK helpers", path)
		} else if !os.IsNotExist(err) {
			t.Fatalf("stat %s: %v", path, err)
		}
	}
}
