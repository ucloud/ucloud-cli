package base

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/ux"
)

// Sspoll (concurrent poller) must suppress spinner animation on a non-TTY
// writer, exactly like Spoll, emitting a single terminal-state line instead.
func TestSspollNonTTY_Done(t *testing.T) {
	describeFunc := func(resourceID string, _ *request.CommonBase) (interface{}, error) {
		return &instDone{State: "DONE"}, nil
	}

	buf := &bytes.Buffer{}
	p := NewSpoller(describeFunc, buf)
	p.Timeout = 30 * time.Second

	ret := p.Sspoll("res-001", "creating", []string{"DONE"}, ux.NewBlock(), &request.CommonBase{})

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
