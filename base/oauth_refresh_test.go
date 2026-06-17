// base/oauth_refresh_test.go
package base

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func newTestManager(t *testing.T, ac *AggConfig) *AggConfigManager {
	t.Helper()
	os.MkdirAll(".ucloud", 0700)
	t.Cleanup(func() { os.RemoveAll(".ucloud") })
	credentialLockPath = ".ucloud/credential.lock"
	t.Cleanup(func() { credentialLockPath = "" })
	m, err := NewAggConfigManager(".ucloud/config.json", ".ucloud/credential.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := m.Append(ac); err != nil {
		t.Fatal(err)
	}
	return m
}

func TestApplyTokenResponse(t *testing.T) {
	cfg := &AggConfig{Profile: "p", RefreshToken: "old-rt"}
	ApplyTokenResponse(cfg, &TokenResponse{AccessToken: "at", ExpiresIn: 3600})
	if cfg.AuthMode != AuthModeOAuth || cfg.AccessToken != "at" {
		t.Errorf("token not applied: %+v", cfg)
	}
	if cfg.RefreshToken != "old-rt" {
		t.Error("empty refresh_token in response must keep the old one")
	}
	if cfg.ExpiresAt < time.Now().Unix()+3500 || cfg.ExpiresAt > time.Now().Unix()+3700 {
		t.Errorf("expires_at wrong: %d", cfg.ExpiresAt)
	}
	// 轮换：新 refresh_token 覆盖旧（D3）
	ApplyTokenResponse(cfg, &TokenResponse{AccessToken: "at2", RefreshToken: "new-rt", ExpiresIn: 3600})
	if cfg.RefreshToken != "new-rt" {
		t.Error("rotated refresh_token must overwrite")
	}
}

func TestEnsureFreshToken(t *testing.T) {
	refreshCalls := 0
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		refreshCalls++
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"access_token":"new-at","refresh_token":"new-rt","expires_in":3600}`)
	}))
	defer s.Close()

	t.Run("expired triggers refresh and persists", func(t *testing.T) {
		ac := &AggConfig{
			Profile: "p1", Active: true, BaseURL: DefaultBaseURL, Timeout: 15, MaxRetryTimes: intPtr(3),
			AuthMode: AuthModeOAuth, AccessToken: "old-at", RefreshToken: "old-rt",
			ExpiresAt: time.Now().Unix() - 100, OAuthBaseURL: s.URL,
		}
		m := newTestManager(t, ac)
		if err := EnsureFreshToken(ac, m); err != nil {
			t.Fatal(err)
		}
		if ac.AccessToken != "new-at" || ac.RefreshToken != "new-rt" {
			t.Errorf("token not refreshed in memory: %+v", ac)
		}
		raw, _ := ioutil.ReadFile(".ucloud/credential.json")
		if !strings.Contains(string(raw), "new-rt") {
			t.Errorf("rotated refresh token not persisted: %s", raw)
		}
	})

	t.Run("fresh token skips refresh", func(t *testing.T) {
		before := refreshCalls
		ac := &AggConfig{
			Profile: "p2", Active: true, BaseURL: DefaultBaseURL, Timeout: 15, MaxRetryTimes: intPtr(3),
			AuthMode: AuthModeOAuth, AccessToken: "at", RefreshToken: "rt",
			ExpiresAt: time.Now().Add(time.Hour).Unix(), OAuthBaseURL: s.URL,
		}
		m := newTestManager(t, ac)
		if err := EnsureFreshToken(ac, m); err != nil {
			t.Fatal(err)
		}
		if refreshCalls != before {
			t.Error("fresh token must not hit /token")
		}
	})
}

// 跨 profile 凭据保护：进程 A（t0 加载 X/Y）刷新 Y 落盘时，不得用内存里的陈旧 X
// 覆盖他进程 B（t1）已轮换写盘的 X 凭据——否则 X 下次刷新必 invalid_grant（D3 旧 refresh_token 立即作废）。
func TestRefreshAndSaveKeepsOtherProfilesRotatedTokens(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"access_token":"y-new-at","refresh_token":"y-new-rt","expires_in":3600}`)
	}))
	defer s.Close()

	// t0：进程 A 加载两个 oauth profile（X 未过期，Y 已过期）
	acX := &AggConfig{
		Profile: "px", Active: true, BaseURL: DefaultBaseURL, Timeout: 15, MaxRetryTimes: intPtr(3),
		AuthMode: AuthModeOAuth, AccessToken: "x-old-at", RefreshToken: "x-old-rt",
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
	}
	m := newTestManager(t, acX)
	acY := &AggConfig{
		Profile: "py", BaseURL: DefaultBaseURL, Timeout: 15, MaxRetryTimes: intPtr(3),
		AuthMode: AuthModeOAuth, AccessToken: "y-old-at", RefreshToken: "y-old-rt",
		ExpiresAt: time.Now().Unix() - 100, OAuthBaseURL: s.URL,
	}
	if err := m.Append(acY); err != nil {
		t.Fatal(err)
	}

	// t1：模拟进程 B 刷新 X 并轮换 refresh_token，直接写盘（不经 A 的 manager）
	raw, err := ioutil.ReadFile(".ucloud/credential.json")
	if err != nil {
		t.Fatal(err)
	}
	var creds []CredentialConfig
	if err := json.Unmarshal(raw, &creds); err != nil {
		t.Fatal(err)
	}
	for i := range creds {
		if creds[i].Profile == "px" {
			creds[i].AccessToken = "x-rotated-at"
			creds[i].RefreshToken = "x-rotated-rt"
			creds[i].ExpiresAt = time.Now().Add(2 * time.Hour).Unix()
		}
	}
	out, err := json.Marshal(creds)
	if err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(".ucloud/credential.json", out, 0600); err != nil {
		t.Fatal(err)
	}

	// t2：进程 A 刷新 Y 并 Save
	if err := EnsureFreshToken(acY, m); err != nil {
		t.Fatal(err)
	}

	diskX, err := readCredentialFromDisk(".ucloud/credential.json", "px")
	if err != nil || diskX == nil {
		t.Fatalf("reload px from disk failed: %v (%v)", diskX, err)
	}
	if diskX.AccessToken != "x-rotated-at" || diskX.RefreshToken != "x-rotated-rt" {
		t.Errorf("process B's rotated X tokens were overwritten by A's stale copy: access=%s refresh=%s",
			diskX.AccessToken, diskX.RefreshToken)
	}
	diskY, err := readCredentialFromDisk(".ucloud/credential.json", "py")
	if err != nil || diskY == nil {
		t.Fatalf("reload py from disk failed: %v (%v)", diskY, err)
	}
	if diskY.AccessToken != "y-new-at" || diskY.RefreshToken != "y-new-rt" {
		t.Errorf("Y's refreshed tokens not persisted: access=%s refresh=%s", diskY.AccessToken, diskY.RefreshToken)
	}
}

// 并发刷新仅一次轮换（D3）：两个并发 EnsureFreshToken 只允许打一次 /token
func TestConcurrentRefreshSingleRotation(t *testing.T) {
	var mu sync.Mutex
	refreshCalls := 0
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		refreshCalls++
		n := refreshCalls
		mu.Unlock()
		time.Sleep(100 * time.Millisecond) // 放大竞争窗口
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"access_token":"at-%d","refresh_token":"rt-%d","expires_in":3600}`, n, n)
	}))
	defer s.Close()

	ac := &AggConfig{
		Profile: "pc", Active: true, BaseURL: DefaultBaseURL, Timeout: 15, MaxRetryTimes: intPtr(3),
		AuthMode: AuthModeOAuth, AccessToken: "old-at", RefreshToken: "old-rt",
		ExpiresAt: time.Now().Unix() - 100, OAuthBaseURL: s.URL,
	}
	m := newTestManager(t, ac)

	// 模拟两个进程：各自持有独立的 AggConfig 副本与 manager 视图。
	// 注意必须用独立 manager（m2）：若复用 m，ac2 的刷新结果不在 m.configs 中，
	// Save() 不会落盘，磁盘重读看到的仍是旧凭据，测不出真实的跨进程行为。
	m2, err := NewAggConfigManager(".ucloud/config.json", ".ucloud/credential.json")
	if err != nil {
		t.Fatal(err)
	}
	ac2, ok := m2.GetAggConfigByProfile("pc")
	if !ok {
		t.Fatal("profile pc not loaded by second manager")
	}

	var wg sync.WaitGroup
	errs := make([]error, 2)
	wg.Add(2)
	go func() { defer wg.Done(); errs[0] = EnsureFreshToken(ac, m) }()
	go func() { defer wg.Done(); errs[1] = EnsureFreshToken(ac2, m2) }()
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("refresher %d failed: %v", i, err)
		}
	}
	if refreshCalls != 1 {
		t.Errorf("expect exactly 1 rotation, got %d", refreshCalls)
	}
}
