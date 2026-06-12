// cmd/callback_test.go
package cmd

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

// setupCallback 在已分配 listener 上起 callback server，返回端口与结果 channel。
func setupCallback(t *testing.T, expectState string) (int, <-chan callbackResult) {
	t.Helper()
	ln, port, err := allocateLoopbackListener()
	if err != nil {
		t.Fatalf("allocate listener: %v", err)
	}
	srv, ch := startCallbackServer(ln, expectState)
	t.Cleanup(func() { srv.Close() })
	return port, ch
}

// get 向 callback server 发一个回调请求，返回 HTTP 状态码。
func get(t *testing.T, port int, query string) int {
	t.Helper()
	url := fmt.Sprintf("http://127.0.0.1:%d/authorization?%s", port, query)
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("GET callback: %v", err)
	}
	resp.Body.Close()
	return resp.StatusCode
}

// drive 起 callback server，发一个回调请求，返回投递的结果（仅用于必然投递的场景）。
func drive(t *testing.T, expectState, query string) callbackResult {
	t.Helper()
	port, ch := setupCallback(t, expectState)
	get(t, port, query)

	select {
	case res := <-ch:
		return res
	case <-time.After(3 * time.Second):
		t.Fatal("callback result not delivered")
		return callbackResult{}
	}
}

// assertNoDelivery 断言 channel 在短窗口内保持为空（噪音请求不得投递）。
func assertNoDelivery(t *testing.T, ch <-chan callbackResult) {
	t.Helper()
	select {
	case res := <-ch:
		t.Fatalf("noise request must not deliver a result, got %+v", res)
	case <-time.After(100 * time.Millisecond):
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

// 噪音请求（state 不匹配，如陈旧标签页的旧回调）：回 400 但不投递，继续等待真正的回调。
func TestCallbackStateMismatch(t *testing.T) {
	port, ch := setupCallback(t, "st")
	if status := get(t, port, "code=x&state=WRONG"); status != http.StatusBadRequest {
		t.Errorf("state-mismatch noise status = %d, want 400", status)
	}
	assertNoDelivery(t, ch)

	// 同一 server 上真正的回调仍然成功
	if status := get(t, port, "code=real&state=st"); status != http.StatusOK {
		t.Errorf("genuine callback status = %d, want 200", status)
	}
	select {
	case res := <-ch:
		if res.err != nil {
			t.Fatalf("genuine callback after noise: unexpected err %v", res.err)
		}
		if res.code != "real" {
			t.Errorf("code = %q, want real", res.code)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("genuine callback not delivered after noise request")
	}
}

// 噪音请求（有 state 无 code，如本地探针）：回 400 但不投递，继续等待真正的回调。
func TestCallbackStateWithoutCode(t *testing.T) {
	port, ch := setupCallback(t, "st")
	if status := get(t, port, "state=st"); status != http.StatusBadRequest {
		t.Errorf("missing-code noise status = %d, want 400", status)
	}
	assertNoDelivery(t, ch)

	// 同一 server 上真正的回调仍然成功
	if status := get(t, port, "code=abc&state=st"); status != http.StatusOK {
		t.Errorf("genuine callback status = %d, want 200", status)
	}
	select {
	case res := <-ch:
		if res.err != nil {
			t.Fatalf("genuine callback after noise: unexpected err %v", res.err)
		}
		if res.code != "abc" {
			t.Errorf("code = %q, want abc", res.code)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("genuine callback not delivered after noise request")
	}
}

// error 参数（用户在浏览器里拒绝授权）是明确的失败信号：必须投递并中止登录。
func TestCallbackAccessDenied(t *testing.T) {
	res := drive(t, "st", "error=access_denied&state=st")
	if res.err == nil {
		t.Fatal("expected access_denied error")
	}
}
