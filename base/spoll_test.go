package base

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/ucloud/ucloud-sdk-go/ucloud/request"
)

// instDone is a minimal struct whose State field signals completion.
type instDone struct {
	State string
}

func TestSpollNonTTY_Done(t *testing.T) {
	describeFunc := func(resourceID string, _ *request.CommonBase) (interface{}, error) {
		return &instDone{State: "DONE"}, nil
	}

	buf := &bytes.Buffer{}
	p := NewSpoller(describeFunc, buf)
	p.Timeout = 30 * time.Second

	p.Spoll("res-001", "creating", []string{"DONE"})

	out := buf.String()

	if !strings.Contains(out, "creating...done\n") {
		t.Errorf("expected 'creating...done\\n' in output, got: %q", out)
	}
	// Spinner frames should NOT be present (non-TTY suppression).
	if strings.ContainsRune(out, '⣾') {
		t.Errorf("spinner frame rune '⣾' leaked into non-TTY output: %q", out)
	}
}

// TestSpollNonTTY_NoSpinnerFrames verifies that no ANSI spinner frames are
// emitted to a non-TTY writer regardless of outcome.  We keep a separate
// done-path test here with an immediate target-state match so the test is
// fast and deterministic.
func TestSpollNonTTY_NoSpinnerFrames(t *testing.T) {
	calls := 0
	describeFunc := func(resourceID string, _ *request.CommonBase) (interface{}, error) {
		calls++
		return &instDone{State: "ACTIVE"}, nil
	}

	buf := &bytes.Buffer{}
	p := NewSpoller(describeFunc, buf)
	p.Timeout = 30 * time.Second

	p.Spoll("res-003", "activating", []string{"ACTIVE"})

	out := buf.String()
	if !strings.Contains(out, "activating...done\n") {
		t.Errorf("expected 'activating...done\\n' in output, got: %q", out)
	}
	for _, r := range []rune{'⣾', '⣽', '⣻', '⢿', '⡿', '⣟', '⣯', '⣷'} {
		if strings.ContainsRune(out, r) {
			t.Errorf("spinner frame rune %q leaked into non-TTY output: %q", r, out)
		}
	}
}
