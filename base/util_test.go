package base

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetHomePath(t *testing.T) {
	home := GetHomePath()
	if home == "" {
		t.Errorf("base.GetHomePath(), home shoud not be empty. Got :%q", home)
	}
}

func TestWriteJSONFileAtomic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cred.json")
	if err := WriteJSONFileAtomic([]map[string]string{{"k": "v1"}}, path); err != nil {
		t.Fatal(err)
	}
	raw, err := ioutil.ReadFile(path)
	if err != nil || !strings.Contains(string(raw), "v1") {
		t.Fatalf("content wrong: %s %v", raw, err)
	}
	fi, _ := os.Stat(path)
	if fi.Mode().Perm() != 0600 {
		t.Errorf("perm = %v, want 0600", fi.Mode().Perm())
	}
	// 覆盖写
	if err := WriteJSONFileAtomic([]map[string]string{{"k": "v2"}}, path); err != nil {
		t.Fatal(err)
	}
	raw, _ = ioutil.ReadFile(path)
	if !strings.Contains(string(raw), "v2") {
		t.Errorf("overwrite failed: %s", raw)
	}
	// 同目录无残留临时文件
	entries, _ := ioutil.ReadDir(dir)
	if len(entries) != 1 {
		t.Errorf("temp files left behind: %v", entries)
	}
}
