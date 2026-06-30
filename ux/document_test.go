package ux

import (
	"bytes"
	"testing"
)

func TestNewDocumentBindsWriterAndAutoDisablesNonTTY(t *testing.T) {
	var buf bytes.Buffer

	d := NewDocument(&buf)

	if d == nil {
		t.Fatal("NewDocument returned nil")
	}
	if d.out != &buf {
		t.Fatal("NewDocument did not bind the provided writer")
	}
	// A *bytes.Buffer is not a TTY → rendering must be auto-disabled so the
	// 20fps spinner goroutine never starts and machine output stays clean.
	if !d.disable {
		t.Fatal("non-TTY writer must auto-disable rendering")
	}
}
