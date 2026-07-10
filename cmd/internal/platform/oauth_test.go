// base/oauth_test.go
package platform

import (
	"os"
	"strings"
	"testing"
	"time"
)

// /dev/null 是字符设备但不是终端；cron/CI 常用 `ucloud xxx </dev/null`，必须走非交互分支。
func TestIsStdinTTYDevNull(t *testing.T) {
	devNull, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatalf("open %s: %v", os.DevNull, err)
	}
	orig := os.Stdin
	os.Stdin = devNull
	t.Cleanup(func() {
		os.Stdin = orig
		devNull.Close()
	})

	if IsStdinTTY() {
		t.Errorf("IsStdinTTY() = true with stdin = %s, want false", os.DevNull)
	}
}

func TestGenerateState(t *testing.T) {
	s1, err := GenerateState()
	if err != nil {
		t.Fatal(err)
	}
	s2, _ := GenerateState()
	if s1 == s2 {
		t.Error("two states should differ")
	}
	if len(s1) < 32 {
		t.Errorf("state too short: %d", len(s1))
	}
}

func TestBuildAuthorizeURL(t *testing.T) {
	u := BuildAuthorizeURL("https://oauth.example.com", "http://localhost:8723/authorization", "st123")
	for _, want := range []string{
		"https://oauth.example.com/authorize?",
		"response_type=code",
		"state=st123",
		"redirect_uri=http%3A%2F%2Flocalhost%3A8723%2Fauthorization",
		"scope=openid+email+offline_access+full_access",
	} {
		if !strings.Contains(u, want) {
			t.Errorf("authorize url missing %q, got %s", want, u)
		}
	}
}

func TestBuildLoopbackRedirectURI(t *testing.T) {
	got := BuildLoopbackRedirectURI(54321)
	if got != "http://localhost:54321/authorization" {
		t.Errorf("redirect uri = %q", got)
	}
	if strings.Contains(got, "127.0.0.1") {
		t.Errorf("redirect uri must use literal localhost, not 127.0.0.1: %s", got)
	}
}

// 回调解析 6 分支：正常 / 输入容错 / 缺 code / state 不匹配 / 畸形 / access_denied
func TestParseCallbackURL(t *testing.T) {
	const state = "st123"
	cases := []struct {
		name    string
		input   string
		wantErr string // 空串表示期望成功
		want    string
	}{
		{"normal", "http://localhost/authorization?code=abc&state=st123", "", "abc"},
		{"tolerant", "  \"http://localhost/authorization?code=a\nbc&state=st123'  \n", "", "abc"},
		{"missing code", "http://localhost/authorization?state=st123", "no authorization code", ""},
		{"state mismatch", "http://localhost/authorization?code=abc&state=OLD", "state mismatch", ""},
		{"malformed", "not a url at all", "no authorization code", ""},
		{"access denied", "http://localhost/authorization?error=access_denied&state=st123", "denied", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			code, err := ParseCallbackURL(c.input, state)
			if c.wantErr == "" {
				if err != nil {
					t.Fatalf("expect ok, got %v", err)
				}
				if code != c.want {
					t.Errorf("code = %q, want %q", code, c.want)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), c.wantErr) {
				t.Errorf("err = %v, want contains %q", err, c.wantErr)
			}
		})
	}
}

// 过期判断 4 分支
func TestTokenExpiredAt(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	cases := []struct {
		name      string
		expiresAt int64
		want      bool
	}{
		{"zero means expired", 0, true},
		{"already past", now.Unix() - 10, true},
		{"inside 5min skew", now.Add(4 * time.Minute).Unix(), true},
		{"fresh", now.Add(30 * time.Minute).Unix(), false},
	}
	for _, c := range cases {
		if got := TokenExpiredAt(c.expiresAt, now); got != c.want {
			t.Errorf("%s: got %v want %v", c.name, got, c.want)
		}
	}
}

func TestParseIDTokenEmail(t *testing.T) {
	// header.payload.sig，payload = {"email":"u@ucloud.cn"} 的 base64url（RawURLEncoding 无填充）
	idToken := "eyJhbGciOiJSUzI1NiJ9.eyJlbWFpbCI6InVAdWNsb3VkLmNuIn0.c2ln"
	email, err := ParseIDTokenEmail(idToken)
	if err != nil {
		t.Fatal(err)
	}
	if email != "u@ucloud.cn" {
		t.Errorf("email = %q", email)
	}
	if _, err := ParseIDTokenEmail("only-one-part"); err == nil {
		t.Error("malformed id_token should error")
	}
}

func TestRedact(t *testing.T) {
	cases := []struct{ in, mustHide string }{
		{"http://localhost/authorization?code=SECRET1&state=SECRET2", "SECRET1"},
		{"http://localhost/authorization?code=SECRET1&state=SECRET2", "SECRET2"},
		{`{"access_token":"SECRET3","refresh_token":"SECRET4"}`, "SECRET3"},
		{`{"access_token":"SECRET3","refresh_token":"SECRET4"}`, "SECRET4"},
		{"Authorization: Bearer SECRET5", "SECRET5"},
		{`id_token=SECRET6`, "SECRET6"},
	}
	for _, c := range cases {
		out := Redact(c.in)
		if strings.Contains(out, c.mustHide) {
			t.Errorf("Redact(%q) leaked %q: %s", c.in, c.mustHide, out)
		}
	}
	if Redact("plain text without secrets") != "plain text without secrets" {
		t.Error("non-sensitive text should pass through")
	}
	if out := Redact("zip_code=12345"); !strings.Contains(out, "12345") {
		t.Errorf("zip_code should not be redacted: %s", out)
	}
	if out := Redact("us_state=CA"); !strings.Contains(out, "CA") {
		t.Errorf("us_state should not be redacted: %s", out)
	}
}

func TestOAuthHints(t *testing.T) {
	if msg := OAuthLoginRequiredHint("p1", true); !strings.Contains(msg, "ucloud auth login") || !strings.Contains(msg, "p1") {
		t.Errorf("tty hint should mention next command and profile: %s", msg)
	}
	if msg := OAuthLoginRequiredHint("p1", false); !strings.Contains(msg, "AK/SK") {
		t.Errorf("non-tty hint should point to AK/SK: %s", msg)
	}
}
