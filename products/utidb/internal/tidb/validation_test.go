package tidb

import (
	"strings"
	"testing"
)

func TestValidateNodeCount(t *testing.T) {
	tests := []struct {
		count   int
		wantErr bool
	}{
		{3, true},
		{2, true},
		{4, false},
		{5, false},
	}
	for _, tt := range tests {
		err := validateNodeCount(tt.count, "tikv")
		if tt.wantErr && err == nil {
			t.Fatalf("count=%d: want error", tt.count)
		}
		if !tt.wantErr && err != nil {
			t.Fatalf("count=%d: unexpected error: %v", tt.count, err)
		}
		if err != nil && !strings.Contains(err.Error(), "大于 3") {
			t.Fatalf("error should include Chinese hint: %v", err)
		}
	}
}

func TestValidateServerType(t *testing.T) {
	for _, st := range []string{"tidb", "TiKV", "PD", "tiflash"} {
		if err := validateServerType(st); err != nil {
			t.Fatalf("ServerType %q: %v", st, err)
		}
	}
	if err := validateServerType("mysql"); err == nil {
		t.Fatal("want error for unknown ServerType")
	}
}
