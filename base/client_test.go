// base/client_test.go
package base

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	uhttp "github.com/ucloud/ucloud-sdk-go/private/protocol/http"
)

func injectorHeaders(t *testing.T, cred *CredentialConfig) map[string]string {
	t.Helper()
	h := newCredHeaderInjector(cred)
	req, err := h(nil, uhttp.NewHttpRequest())
	if err != nil {
		t.Fatal(err)
	}
	return req.GetHeaderMap()
}

// oauth 模式：注入 Authorization Bearer
func TestInjectorOAuthBearer(t *testing.T) {
	headers := injectorHeaders(t, &CredentialConfig{AuthMode: AuthModeOAuth, AccessToken: "tok123"})
	if headers["Authorization"] != "Bearer tok123" {
		t.Errorf("Authorization = %q, want Bearer tok123", headers["Authorization"])
	}
}

// CRITICAL 回归：aksk 模式（含 CloudShell Cookie 注入）头部行为零变化
func TestInjectorAkskAndCloudShellUnchanged(t *testing.T) {
	// aksk：Cookie/Csrf-Token 照旧（空值也照旧 set），绝不出现 Authorization
	h1 := injectorHeaders(t, &CredentialConfig{PublicKey: "pub", PrivateKey: "pri"})
	if _, ok := h1["Authorization"]; ok {
		t.Error("aksk mode must NOT inject Authorization header")
	}
	if v, ok := h1["Cookie"]; !ok || v != "" {
		t.Errorf("Cookie header behavior changed: %q %v", v, ok)
	}
	// CloudShell：Cookie/Csrf-Token 注入照旧
	h2 := injectorHeaders(t, &CredentialConfig{Cookie: "ck", CSRFToken: "cs"})
	if h2["Cookie"] != "ck" || h2["Csrf-Token"] != "cs" {
		t.Errorf("cloudshell headers changed: %v", h2)
	}
	if _, ok := h2["Authorization"]; ok {
		t.Error("cloudshell mode must NOT inject Authorization header")
	}
}

// oauth 模式但 token 为空：不注入（让网关报错而不是发送 "Bearer "）
func TestInjectorOAuthEmptyToken(t *testing.T) {
	headers := injectorHeaders(t, &CredentialConfig{AuthMode: AuthModeOAuth})
	if _, ok := headers["Authorization"]; ok {
		t.Error("empty token must not inject Authorization")
	}
}

// recordedRequest 记录业务请求实际携带的 header 与全部参数（query + form 合并）
type recordedRequest struct {
	header http.Header
	params url.Values
}

// bizRecorderServer 模拟业务网关：记录请求并返回成功响应
func bizRecorderServer(t *testing.T, rec *recordedRequest) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Errorf("parse form: %v", err)
		}
		rec.header = r.Header.Clone()
		rec.params = r.Form // r.Form 含 URL query + body form，两路都覆盖
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"RetCode":0,"Action":"GetRegionResponse"}`)
	}))
}

func callGetRegion(t *testing.T, ac *AggConfig, rec *recordedRequest) {
	t.Helper()
	// GetBizClient 会改写包级全局 ClientConfig/AuthCredential，恢复现场避免测试顺序耦合
	oldClientConfig, oldAuthCredential := ClientConfig, AuthCredential
	t.Cleanup(func() {
		ClientConfig, AuthCredential = oldClientConfig, oldAuthCredential
	})
	bc, err := GetBizClient(ac)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := bc.GetRegion(bc.NewGetRegionRequest()); err != nil {
		t.Fatalf("GetRegion failed: %v", err)
	}
	if rec.params == nil {
		t.Fatal("server did not record any request")
	}
}

// CRITICAL 缺陷回归（RetCode 171）：oauth profile 残留 AK/SK（供 logout 恢复）时，
// 请求必须只携带 Bearer 一种凭据，绝不能同时出现 SDK 签名参数。
func TestOAuthProfileWithRetainedKeysDoesNotSign(t *testing.T) {
	rec := &recordedRequest{}
	s := bizRecorderServer(t, rec)
	defer s.Close()

	ac := &AggConfig{
		Profile: "oauth-leftover", BaseURL: s.URL, Timeout: 15, MaxRetryTimes: intPtr(0),
		Region: "cn-bj2", AuthMode: AuthModeOAuth, AccessToken: "tok",
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
		PublicKey: "leftover-pub", PrivateKey: "leftover-pri",
	}
	callGetRegion(t, ac, rec)

	if got := rec.header.Get("Authorization"); got != "Bearer tok" {
		t.Errorf("Authorization = %q, want %q", got, "Bearer tok")
	}
	for _, k := range []string{"Signature", "PublicKey"} {
		if v, ok := rec.params[k]; ok {
			t.Errorf("oauth profile must not send signature param %s=%v (one request carries exactly one credential)", k, v)
		}
	}
}

// 401 自动重放矩阵（D6 反应式兜底）：401→刷新→重放成功；aksk 模式不重放。
// 注意 SDK 行为：HttpClient.Send 对 status>=400 返回 (nil, StatusError)，
// 401 的 body 在 handler 层不可见，鉴权失败只能从 err 判定。
func TestOAuthRetryHandler(t *testing.T) {
	apiCalls := 0
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCalls++
		if r.Header.Get("Authorization") == "Bearer good" {
			fmt.Fprint(w, `{"RetCode":0,"Action":"GetRegionResponse"}`)
			return
		}
		w.WriteHeader(401)
		fmt.Fprint(w, `{"RetCode":170,"Message":"token expired"}`)
	}))
	defer api.Close()
	oauth := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"access_token":"good","refresh_token":"rt2","expires_in":3600}`)
	}))
	defer oauth.Close()

	ac := &AggConfig{
		Profile: "pr", Active: true, BaseURL: api.URL, Timeout: 15, MaxRetryTimes: intPtr(0),
		AuthMode: AuthModeOAuth, AccessToken: "bad", RefreshToken: "rt1",
		ExpiresAt: 9999999999, OAuthBaseURL: oauth.URL, // 未到期 → 不触发主动刷新，逼出反应式路径
	}
	m := newTestManager(t, ac)
	prevIns, prevList := ConfigIns, AggConfigListIns
	prevCC, prevAC := ClientConfig, AuthCredential
	ConfigIns, AggConfigListIns = ac, m
	t.Cleanup(func() {
		ConfigIns, AggConfigListIns = prevIns, prevList
		ClientConfig, AuthCredential = prevCC, prevAC
	})

	client, err := GetBizClient(ac)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.GetRegion(client.NewGetRegionRequest())
	if err != nil {
		t.Fatalf("replay should succeed: %v", err)
	}
	if resp.GetRetCode() != 0 {
		t.Errorf("RetCode = %d", resp.GetRetCode())
	}
	if apiCalls != 2 {
		t.Errorf("expect 1 fail + 1 replay = 2 api calls, got %d", apiCalls)
	}
	var creds []CredentialConfig
	raw, _ := ioutil.ReadFile(".ucloud/credential.json")
	json.Unmarshal(raw, &creds)
	var persisted *CredentialConfig
	for i := range creds {
		if creds[i].Profile == "pr" {
			persisted = &creds[i]
		}
	}
	if persisted == nil || persisted.AccessToken != "good" || persisted.RefreshToken != "rt2" {
		t.Errorf("refreshed token and rotated refresh_token must be persisted: %s", raw)
	}
}

// RetCode 白名单路径（实测网关行为）：鉴权失败返回 HTTP 200 + RetCode 174 "Token Not Exists"
// （无效与过期 Bearer 同码，2026-06-11 实测）。SDK 管道：200 时 Send 返回 (resp, nil)，
// 默认 errorHTTPHandler 不动 err==nil，body 在本 handler 可读 → 走 isAuthFailure 的
// resp-body 白名单分支：刷新 → 重放一次成功。
func TestOAuthRetryHandlerRetCode174(t *testing.T) {
	apiCalls := 0
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCalls++
		w.Header().Set("Content-Type", "application/json")
		if r.Header.Get("Authorization") == "Bearer good" {
			fmt.Fprint(w, `{"RetCode":0,"Action":"GetRegionResponse"}`)
			return
		}
		fmt.Fprint(w, `{"RetCode":174,"Message":"Token Not Exists"}`)
	}))
	defer api.Close()
	oauth := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"access_token":"good","refresh_token":"rt2","expires_in":3600}`)
	}))
	defer oauth.Close()

	ac := &AggConfig{
		Profile: "p174", Active: true, BaseURL: api.URL, Timeout: 15, MaxRetryTimes: intPtr(0),
		AuthMode: AuthModeOAuth, AccessToken: "bad", RefreshToken: "rt1",
		ExpiresAt: 9999999999, OAuthBaseURL: oauth.URL, // 未到期 → 不触发主动刷新，逼出反应式路径
	}
	m := newTestManager(t, ac)
	prevIns, prevList := ConfigIns, AggConfigListIns
	prevCC, prevAC := ClientConfig, AuthCredential
	ConfigIns, AggConfigListIns = ac, m
	t.Cleanup(func() {
		ConfigIns, AggConfigListIns = prevIns, prevList
		ClientConfig, AuthCredential = prevCC, prevAC
	})

	client, err := GetBizClient(ac)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.GetRegion(client.NewGetRegionRequest())
	if err != nil {
		t.Fatalf("replay should succeed: %v", err)
	}
	if resp.GetRetCode() != 0 {
		t.Errorf("RetCode = %d", resp.GetRetCode())
	}
	if apiCalls != 2 {
		t.Errorf("expect 1 fail + 1 replay = 2 api calls, got %d", apiCalls)
	}
	var creds []CredentialConfig
	raw, _ := ioutil.ReadFile(".ucloud/credential.json")
	json.Unmarshal(raw, &creds)
	var persisted *CredentialConfig
	for i := range creds {
		if creds[i].Profile == "p174" {
			persisted = &creds[i]
		}
	}
	if persisted == nil || persisted.AccessToken != "good" || persisted.RefreshToken != "rt2" {
		t.Errorf("refreshed token and rotated refresh_token must be persisted: %s", raw)
	}
}

// 负路径（174 持续）：重放后仍 174 → 只重放一次（共 2 次 api 调用，不循环），
// RetCode 由 SDK 默认 errorHandler 转为 ServerCodeError 上浮。
func TestOAuthRetryHandlerRetCode174Persists(t *testing.T) {
	apiCalls := 0
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCalls++
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"RetCode":174,"Message":"Token Not Exists"}`)
	}))
	defer api.Close()
	oauth := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"access_token":"good","refresh_token":"rt2","expires_in":3600}`)
	}))
	defer oauth.Close()

	ac := &AggConfig{
		Profile: "p174s", Active: true, BaseURL: api.URL, Timeout: 15, MaxRetryTimes: intPtr(0),
		AuthMode: AuthModeOAuth, AccessToken: "bad", RefreshToken: "rt1",
		ExpiresAt: 9999999999, OAuthBaseURL: oauth.URL,
	}
	m := newTestManager(t, ac)
	prevList, prevCC, prevAC := AggConfigListIns, ClientConfig, AuthCredential
	AggConfigListIns = m
	t.Cleanup(func() { AggConfigListIns, ClientConfig, AuthCredential = prevList, prevCC, prevAC })

	client, err := GetBizClient(ac)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.GetRegion(client.NewGetRegionRequest()); err == nil {
		t.Error("persistent RetCode 174 must surface an error after single replay")
	}
	if apiCalls != 2 {
		t.Errorf("exactly one replay allowed (no loop): got %d api calls", apiCalls)
	}
}

// 负路径 (a)：刷新失败（invalid_grant）→ 原始 401 错误上浮，不重放、不 panic
func TestOAuthRetryHandlerRefreshFails(t *testing.T) {
	apiCalls := 0
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCalls++
		w.WriteHeader(401)
		fmt.Fprint(w, `{"RetCode":170,"Message":"token expired"}`)
	}))
	defer api.Close()
	oauth := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		fmt.Fprint(w, `{"error":"invalid_grant"}`)
	}))
	defer oauth.Close()

	ac := &AggConfig{
		Profile: "prf", Active: true, BaseURL: api.URL, Timeout: 15, MaxRetryTimes: intPtr(0),
		AuthMode: AuthModeOAuth, AccessToken: "bad", RefreshToken: "rt1",
		ExpiresAt: 9999999999, OAuthBaseURL: oauth.URL,
	}
	m := newTestManager(t, ac)
	prevList, prevCC, prevAC := AggConfigListIns, ClientConfig, AuthCredential
	AggConfigListIns = m
	t.Cleanup(func() { AggConfigListIns, ClientConfig, AuthCredential = prevList, prevCC, prevAC })

	client, err := GetBizClient(ac)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.GetRegion(client.NewGetRegionRequest()); err == nil {
		t.Error("refresh failure must surface the original 401 error")
	}
	if apiCalls != 1 {
		t.Errorf("refresh failed, must not replay: got %d api calls", apiCalls)
	}
}

// 负路径 (b)：重放后仍 401 → 只重放一次（共 2 次 api 调用，不循环），错误上浮
func TestOAuthRetryHandlerReplayStill401(t *testing.T) {
	apiCalls := 0
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCalls++
		w.WriteHeader(401)
		fmt.Fprint(w, `{"RetCode":170,"Message":"still no"}`)
	}))
	defer api.Close()
	oauth := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"access_token":"good","refresh_token":"rt2","expires_in":3600}`)
	}))
	defer oauth.Close()

	ac := &AggConfig{
		Profile: "prs", Active: true, BaseURL: api.URL, Timeout: 15, MaxRetryTimes: intPtr(0),
		AuthMode: AuthModeOAuth, AccessToken: "bad", RefreshToken: "rt1",
		ExpiresAt: 9999999999, OAuthBaseURL: oauth.URL,
	}
	m := newTestManager(t, ac)
	prevList, prevCC, prevAC := AggConfigListIns, ClientConfig, AuthCredential
	AggConfigListIns = m
	t.Cleanup(func() { AggConfigListIns, ClientConfig, AuthCredential = prevList, prevCC, prevAC })

	client, err := GetBizClient(ac)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.GetRegion(client.NewGetRegionRequest()); err == nil {
		t.Error("replay still 401 must surface error")
	}
	if apiCalls != 2 {
		t.Errorf("exactly one replay allowed (no loop): got %d api calls", apiCalls)
	}
}

func TestOAuthRetryHandlerSkipsAksk(t *testing.T) {
	apiCalls := 0
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCalls++
		w.WriteHeader(401)
		fmt.Fprint(w, `{"RetCode":170,"Message":"x"}`)
	}))
	defer api.Close()
	ac := &AggConfig{
		Profile: "pa", Active: true, BaseURL: api.URL, Timeout: 15, MaxRetryTimes: intPtr(0),
		PublicKey: "pub", PrivateKey: "pri",
	}
	_ = newTestManager(t, ac)
	prevCC, prevAC := ClientConfig, AuthCredential
	t.Cleanup(func() { ClientConfig, AuthCredential = prevCC, prevAC })
	client, err := GetBizClient(ac)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.GetRegion(client.NewGetRegionRequest()); err == nil {
		t.Error("aksk 401 should surface error, not replay-refresh")
	}
	if apiCalls != 1 {
		t.Errorf("aksk mode must not replay, got %d calls", apiCalls)
	}
}

// 反向护栏：aksk profile 照旧签名（Signature + PublicKey 必须在场，且无 Bearer）
func TestAkskProfileStillSigns(t *testing.T) {
	rec := &recordedRequest{}
	s := bizRecorderServer(t, rec)
	defer s.Close()

	ac := &AggConfig{
		Profile: "aksk", BaseURL: s.URL, Timeout: 15, MaxRetryTimes: intPtr(0),
		Region: "cn-bj2", PublicKey: "pub", PrivateKey: "pri",
	}
	callGetRegion(t, ac, rec)

	for _, k := range []string{"Signature", "PublicKey"} {
		if _, ok := rec.params[k]; !ok {
			t.Errorf("aksk profile must sign requests, missing param %s; got params %v", k, rec.params)
		}
	}
	if _, ok := rec.header["Authorization"]; ok {
		t.Error("aksk profile must not send Authorization header")
	}
}
