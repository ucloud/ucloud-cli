package tidb

import "testing"

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
