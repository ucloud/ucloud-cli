// base/oauth_http_test.go
package base

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func tokenServer(t *testing.T, status int, body string, gotForm *map[string]string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/token" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		r.ParseForm()
		if gotForm != nil {
			m := map[string]string{}
			for k := range r.PostForm {
				m[k] = r.PostForm.Get(k)
			}
			*gotForm = m
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		fmt.Fprint(w, body)
	}))
}

// 换 token 3 分支：成功 / invalid_grant / 5xx
func TestExchangeToken(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var form map[string]string
		s := tokenServer(t, 200, `{"access_token":"at1","refresh_token":"rt1","id_token":"idt","expires_in":3600,"token_type":"Bearer"}`, &form)
		defer s.Close()
		tr, err := ExchangeToken(s.URL, "http://localhost:8723/authorization", "code1")
		if err != nil {
			t.Fatal(err)
		}
		if tr.AccessToken != "at1" || tr.RefreshToken != "rt1" || tr.ExpiresIn != 3600 {
			t.Errorf("unexpected token response: %+v", tr)
		}
		if form["grant_type"] != "authorization_code" || form["code"] != "code1" || form["redirect_uri"] != "http://localhost:8723/authorization" {
			t.Errorf("bad form: %v", form)
		}
	})
	t.Run("invalid_grant translated", func(t *testing.T) {
		s := tokenServer(t, 400, `{"error":"invalid_grant","error_description":"code expired"}`, nil)
		defer s.Close()
		_, err := ExchangeToken(s.URL, "http://localhost:8723/authorization", "old")
		if err == nil || !strings.Contains(err.Error(), "ucloud auth login") || !strings.Contains(err.Error(), "expired or already used") {
			t.Errorf("invalid_grant should translate to actionable message, got %v", err)
		}
	})
	t.Run("server 5xx", func(t *testing.T) {
		s := tokenServer(t, 500, `oops`, nil)
		defer s.Close()
		_, err := ExchangeToken(s.URL, "http://localhost:8723/authorization", "c")
		if err == nil || !strings.Contains(err.Error(), "server error") {
			t.Errorf("5xx should say server error + retry, got %v", err)
		}
	})
}

// 刷新 3 分支：成功(轮换) / refresh 失效 / 网络不可达
func TestRefreshToken(t *testing.T) {
	t.Run("success with rotation", func(t *testing.T) {
		var form map[string]string
		s := tokenServer(t, 200, `{"access_token":"at2","refresh_token":"rt2-rotated","expires_in":3600}`, &form)
		defer s.Close()
		tr, err := RefreshToken(s.URL, "rt1")
		if err != nil {
			t.Fatal(err)
		}
		if tr.RefreshToken != "rt2-rotated" {
			t.Errorf("rotated refresh token not surfaced: %+v", tr)
		}
		if form["grant_type"] != "refresh_token" || form["refresh_token"] != "rt1" {
			t.Errorf("bad form: %v", form)
		}
	})
	t.Run("invalid refresh token", func(t *testing.T) {
		s := tokenServer(t, 400, `{"error":"invalid_grant","error_description":"refresh token revoked"}`, nil)
		defer s.Close()
		if _, err := RefreshToken(s.URL, "dead"); err == nil {
			t.Error("expect error for revoked refresh token")
		}
	})
	t.Run("unreachable", func(t *testing.T) {
		_, err := RefreshToken("http://127.0.0.1:1", "rt")
		if err == nil || !strings.Contains(err.Error(), "cannot reach oauth server") {
			t.Errorf("network error should be distinguished, got %v", err)
		}
	})
}

// 钉死：oauthHTTPClient 必须遵守 HTTPS_PROXY 等代理环境变量（默认 Transport 或显式 ProxyFromEnvironment）
func TestOAuthClientHonorsProxyEnv(t *testing.T) {
	if oauthHTTPClient.Transport == nil {
		return // nil Transport == http.DefaultTransport，自带 ProxyFromEnvironment
	}
	tr, ok := oauthHTTPClient.Transport.(*http.Transport)
	if !ok || tr.Proxy == nil {
		t.Error("oauthHTTPClient custom transport must set Proxy: http.ProxyFromEnvironment")
	}
}
