package cli

import (
	"bytes"
	"strings"
	"testing"
	"time"

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

// pollNoop 是仅用于构造 poller 的空 describe 函数。
func pollNoop(string, *request.CommonBase) (interface{}, error) { return nil, nil }

func TestNewPollerDefaultTimeout(t *testing.T) {
	p, ok := NewPoller(pollNoop, &bytes.Buffer{}).(*poller)
	if !ok {
		t.Fatalf("NewPoller did not return *poller")
	}
	if p.timeout != 10*time.Minute {
		t.Errorf("default poll timeout = %v, want 10m", p.timeout)
	}
}

func TestSetUserPollTimeoutOverride(t *testing.T) {
	defer func() { userPollTimeout = 0 }() // 直接复位；SetUserPollTimeout(0) 会被守卫忽略
	SetUserPollTimeout(30 * time.Minute)
	p := NewPoller(pollNoop, &bytes.Buffer{}).(*poller)
	if p.timeout != 30*time.Minute {
		t.Errorf("after SetUserPollTimeout(30m), timeout = %v, want 30m", p.timeout)
	}
}

func TestSetUserPollTimeoutIgnoresNonPositive(t *testing.T) {
	defer func() { userPollTimeout = 0 }()
	SetUserPollTimeout(30 * time.Minute)
	SetUserPollTimeout(0)                // 忽略
	SetUserPollTimeout(-5 * time.Minute) // 忽略
	p := NewPoller(pollNoop, &bytes.Buffer{}).(*poller)
	if p.timeout != 30*time.Minute {
		t.Errorf("non-positive SetUserPollTimeout must be ignored, timeout = %v, want 30m", p.timeout)
	}
}

func TestEffectivePollTimeoutPriority(t *testing.T) {
	defer func() { userPollTimeout = 0 }()

	// 都未设 → builtin
	userPollTimeout = 0
	if got := effectivePollTimeout(0); got != 10*time.Minute {
		t.Errorf("no user, no command: got %v, want 10m", got)
	}
	// 仅命令自设 → command
	userPollTimeout = 0
	if got := effectivePollTimeout(20 * time.Minute); got != 20*time.Minute {
		t.Errorf("command only: got %v, want 20m", got)
	}
	// 用户已设 → user 覆盖命令
	userPollTimeout = 15 * time.Minute
	if got := effectivePollTimeout(20 * time.Minute); got != 15*time.Minute {
		t.Errorf("user overrides command: got %v, want 15m", got)
	}
}

func TestWithTimeoutSetsCommandTimeout(t *testing.T) {
	defer func() { userPollTimeout = 0 }()
	userPollTimeout = 0
	p := NewPoller(pollNoop, &bytes.Buffer{}, WithTimeout(30*time.Minute)).(*poller)
	if p.commandTimeout != 30*time.Minute {
		t.Errorf("commandTimeout = %v, want 30m", p.commandTimeout)
	}
	if p.timeout != 30*time.Minute {
		t.Errorf("effective timeout with command option = %v, want 30m", p.timeout)
	}
}

func TestWithTimeoutIgnoresNonPositive(t *testing.T) {
	defer func() { userPollTimeout = 0 }()
	userPollTimeout = 0
	p := NewPoller(pollNoop, &bytes.Buffer{}, WithTimeout(0), WithTimeout(-5*time.Minute)).(*poller)
	if p.timeout != 10*time.Minute {
		t.Errorf("non-positive WithTimeout must be ignored, timeout = %v, want builtin 10m", p.timeout)
	}
}

func TestUserFlagOverridesCommandOption(t *testing.T) {
	defer func() { userPollTimeout = 0 }()
	SetUserPollTimeout(20 * time.Minute)
	p := NewPoller(pollNoop, &bytes.Buffer{}, WithTimeout(30*time.Minute)).(*poller)
	if p.timeout != 20*time.Minute {
		t.Errorf("user flag must override command option: timeout = %v, want 20m", p.timeout)
	}
}
