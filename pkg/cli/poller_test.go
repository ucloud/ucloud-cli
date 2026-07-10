package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/pkg/ui"
)

type pollerDoneInstance struct {
	State string
}

func TestPollerSpollNonTTYDone(t *testing.T) {
	describeFunc := func(resourceID string, _ *request.CommonBase) (interface{}, error) {
		return &pollerDoneInstance{State: "DONE"}, nil
	}

	buf := &bytes.Buffer{}
	NewPoller(describeFunc, buf).Spoll("res-001", "creating", []string{"DONE"})

	out := buf.String()
	if !strings.Contains(out, "creating...done\n") {
		t.Errorf("expected 'creating...done\\n' in output, got: %q", out)
	}
	if strings.ContainsRune(out, '⣾') {
		t.Errorf("spinner frame rune '⣾' leaked into non-TTY output: %q", out)
	}
}

func TestPollerSpollNonTTYNoSpinnerFrames(t *testing.T) {
	describeFunc := func(resourceID string, _ *request.CommonBase) (interface{}, error) {
		return &pollerDoneInstance{State: "ACTIVE"}, nil
	}

	buf := &bytes.Buffer{}
	NewPoller(describeFunc, buf).Spoll("res-003", "activating", []string{"ACTIVE"})

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

func TestPollerSspollNonTTYDone(t *testing.T) {
	describeFunc := func(resourceID string, _ *request.CommonBase) (interface{}, error) {
		return &pollerDoneInstance{State: "DONE"}, nil
	}

	buf := &bytes.Buffer{}
	ret := NewPoller(describeFunc, buf).Sspoll("res-001", "creating", []string{"DONE"}, ui.NewBlock(), &request.CommonBase{})

	if ret == nil || !ret.Done {
		t.Fatalf("Sspoll non-TTY: want Done=true, got %+v", ret)
	}
	out := buf.String()
	if !strings.Contains(out, "creating...done\n") {
		t.Errorf("expected 'creating...done\\n' in output, got: %q", out)
	}
	for _, r := range []rune{'⣾', '⣽', '⣻', '⢿', '⡿', '⣟', '⣯', '⣷'} {
		if strings.ContainsRune(out, r) {
			t.Errorf("spinner frame %q leaked into non-TTY Sspoll output: %q", r, out)
		}
	}
}
