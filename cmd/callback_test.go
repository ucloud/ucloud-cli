// cmd/callback_test.go
package cmd

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

// drive 在已分配 listener 上起 callback server，向其发一个回调请求，返回投递的结果。
func drive(t *testing.T, expectState, query string) callbackResult {
	t.Helper()
	ln, port, err := allocateLoopbackListener()
	if err != nil {
		t.Fatalf("allocate listener: %v", err)
	}
	srv, ch := startCallbackServer(ln, expectState)
	defer srv.Close()

	url := fmt.Sprintf("http://127.0.0.1:%d/authorization?%s", port, query)
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("GET callback: %v", err)
	}
	resp.Body.Close()

	select {
	case res := <-ch:
		return res
	case <-time.After(3 * time.Second):
		t.Fatal("callback result not delivered")
		return callbackResult{}
	}
}

func TestCallbackSuccess(t *testing.T) {
	res := drive(t, "st", "code=abc&state=st")
	if res.err != nil {
		t.Fatalf("expected success, got err %v", res.err)
	}
	if res.code != "abc" {
		t.Errorf("code = %q, want abc", res.code)
	}
}

func TestCallbackStateMismatch(t *testing.T) {
	res := drive(t, "st", "code=x&state=WRONG")
	if res.err == nil {
		t.Fatal("expected state mismatch error")
	}
	if res.code != "" {
		t.Errorf("code should be empty on error, got %q", res.code)
	}
}

func TestCallbackStateWithoutCode(t *testing.T) {
	res := drive(t, "st", "state=st")
	if res.err == nil {
		t.Fatal("expected missing-code error")
	}
	if got, want := res.err.Error(), "callback carried no authorization code. Run 'ucloud auth login' to try again"; got != want {
		t.Errorf("err = %q, want %q", got, want)
	}
	if res.code != "" {
		t.Errorf("code should be empty on error, got %q", res.code)
	}
}

func TestCallbackAccessDenied(t *testing.T) {
	res := drive(t, "st", "error=access_denied&state=st")
	if res.err == nil {
		t.Fatal("expected access_denied error")
	}
}
