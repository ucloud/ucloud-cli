package common

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestDateTimeLayout(t *testing.T) {
	if DateTimeLayout != "2006-01-02/15:04:05" {
		t.Fatalf("DateTimeLayout = %q", DateTimeLayout)
	}
}

func TestFormatDateTime(t *testing.T) {
	want := time.Unix(int64(1700000000), 0).Format("2006-01-02/15:04:05")
	if got := FormatDateTime(1700000000); got != want {
		t.Fatalf("FormatDateTime(1700000000) = %q, want %q", got, want)
	}
}

func TestFormatDate(t *testing.T) {
	want := time.Unix(int64(1609459200), 0).Format("2006-01-02")
	got := FormatDate(1609459200)
	if got != want {
		t.Fatalf("FormatDate(1609459200) = %q, want %q", got, want)
	}
	if len(got) != len("2006-01-02") {
		t.Fatalf("FormatDate(1609459200) length = %d, want %d (%q)", len(got), len("2006-01-02"), got)
	}
}

func TestIsBase64Encoded(t *testing.T) {
	if !IsBase64Encoded([]byte("aGVsbG8=")) {
		t.Fatalf("IsBase64Encoded(%q) = false, want true", "aGVsbG8=")
	}
	if IsBase64Encoded([]byte("not base64!!!")) {
		t.Fatalf("IsBase64Encoded(%q) = true, want false", "not base64!!!")
	}
}

func TestGetHomePath(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix-only home semantics")
	}
	t.Setenv("HOME", "/tmp/uctest-home")
	if got := GetHomePath(); got != "/tmp/uctest-home" {
		t.Fatalf("GetHomePath() = %q, want /tmp/uctest-home", got)
	}
}

func TestGetFileList(t *testing.T) {
	dir := t.TempDir()
	for _, n := range []string{"a.cnf", "b.txt"} {
		if err := os.WriteFile(filepath.Join(dir, n), nil, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	// Last token of COMP_LINE is the directory prefix to complete from.
	t.Setenv("COMP_LINE", "ucloud mysql conf upload "+dir)

	got := GetFileList(".cnf")
	want := filepath.Join(dir, "a.cnf")
	if len(got) != 1 || got[0] != want {
		t.Fatalf("GetFileList(.cnf) = %v, want [%s]", got, want)
	}
}

func TestGetFileListNoMatchOrMissingDir(t *testing.T) {
	t.Setenv("COMP_LINE", "ucloud mysql conf upload /no/such/dir/xyz")
	if got := GetFileList(".cnf"); got != nil {
		t.Fatalf("GetFileList on missing dir = %v, want nil", got)
	}
}
