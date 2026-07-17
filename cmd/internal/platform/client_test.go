// base/client_test.go
package platform

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
	"github.com/ucloud/ucloud-sdk-go/services/uaccount"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"
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

// injectBody 把 body 以指定 Content-Type 过一遍 injector，返回处理后的请求。
func injectBody(t *testing.T, cred *CredentialConfig, contentType, body string) *uhttp.HttpRequest {
	t.Helper()
	req := uhttp.NewHttpRequest()
	if err := req.SetHeader(uhttp.HeaderNameContentType, contentType); err != nil {
		t.Fatal(err)
	}
	if err := req.SetRequestBody([]byte(body)); err != nil {
		t.Fatal(err)
	}
	out, err := newCredHeaderInjector(cred)(nil, req)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// oauth + JSON body：剥离 Signature/PublicKey，其余字段与 Content-Type 原样保留。
// JSON 编码器此前不可达；products/pgsql(#127) 是第一个走这条路的产品（UPgSQL 网关
// 无法把 form 的字符串 "100" unmarshal 进 Go 的 *int，RetCode 214001，故切 JSONEncoder）。
func TestInjectorOAuthStripsJSONSignature(t *testing.T) {
	body := `{"Action":"ListUPgSQLParamTemplate","Count":100,"PublicKey":"pub","Region":"cn-bj2","Signature":"deadbeef"}`
	out := injectBody(t, &CredentialConfig{AuthMode: AuthModeOAuth, AccessToken: "tok123"}, uhttp.MimeJSON, body)

	var got map[string]interface{}
	if err := json.Unmarshal(out.GetRequestBody(), &got); err != nil {
		t.Fatalf("body is not valid json after strip: %v", err)
	}
	if _, ok := got["Signature"]; ok {
		t.Error("Signature must be stripped from oauth json body")
	}
	if _, ok := got["PublicKey"]; ok {
		t.Error("PublicKey must be stripped from oauth json body")
	}
	if got["Action"] != "ListUPgSQLParamTemplate" || got["Region"] != "cn-bj2" {
		t.Errorf("business fields must survive untouched, got %v", got)
	}
	// int 必须仍是 JSON number —— 产品切 JSONEncoder 的全部意义就在于此，
	// 剥离过程若把它变回字符串就重新踩回 214001。
	if got["Count"] != float64(100) {
		t.Errorf("Count = %#v, want JSON number 100", got["Count"])
	}
	if out.GetHeaderMap()[uhttp.HeaderNameContentType] != uhttp.MimeJSON {
		t.Error("Content-Type must stay application/json")
	}
	if out.GetHeaderMap()["Authorization"] != "Bearer tok123" {
		t.Error("Bearer must still be injected for json body")
	}
}

// CRITICAL 回归：非 oauth（aksk）模式下 JSON body 必须逐字节不变 ——
// AK/SK 路径的签名就活在 body 里，碰一下就验签失败。
func TestInjectorAkskJSONBodyUntouched(t *testing.T) {
	body := `{"Action":"X","PublicKey":"pub","Signature":"deadbeef"}`
	out := injectBody(t, &CredentialConfig{PublicKey: "pub", PrivateKey: "pri"}, uhttp.MimeJSON, body)
	if string(out.GetRequestBody()) != body {
		t.Errorf("aksk json body must be byte-identical\n got: %s\nwant: %s", out.GetRequestBody(), body)
	}
}

// CRITICAL 回归：oauth + form body 行为与历史完全一致（本次只加分支，不动 form）。
func TestInjectorOAuthFormUnchanged(t *testing.T) {
	body := "Action=X&PublicKey=pub&Region=cn-bj2&Signature=deadbeef"
	out := injectBody(t, &CredentialConfig{AuthMode: AuthModeOAuth, AccessToken: "tok"}, uhttp.MimeFormURLEncoded, body)
	vals, err := url.ParseQuery(string(out.GetRequestBody()))
	if err != nil {
		t.Fatal(err)
	}
	if vals.Has("Signature") || vals.Has("PublicKey") {
		t.Errorf("form signature params must be stripped, got %s", out.GetRequestBody())
	}
	if vals.Get("Action") != "X" || vals.Get("Region") != "cn-bj2" {
		t.Errorf("form business fields changed: %s", out.GetRequestBody())
	}
}

// 不认识的 Content-Type：不碰 body（盲目重编码会毁掉它），但 Bearer 照常注入。
func TestInjectorOAuthUnknownContentTypeBodyUntouched(t *testing.T) {
	body := `<xml><Signature>deadbeef</Signature></xml>`
	out := injectBody(t, &CredentialConfig{AuthMode: AuthModeOAuth, AccessToken: "tok"}, "application/xml", body)
	if string(out.GetRequestBody()) != body {
		t.Errorf("unknown content-type body must be untouched, got %s", out.GetRequestBody())
	}
	if out.GetHeaderMap()["Authorization"] != "Bearer tok" {
		t.Error("Bearer must still be injected for unknown content-type")
	}
}

// 端到端：真实 SDK JSONEncoder + 真实 oauth 凭据，确认
// (a) SDK 即使凭据为空也会附加 Signature（这正是必须剥离的原因）；
// (b) injector 之后 body 内再无签名参数，只剩 Bearer 一种凭据机制。
func TestInjectorOAuthJSONEndToEndWithSDKEncoder(t *testing.T) {
	credConfig := &CredentialConfig{AuthMode: AuthModeOAuth, AccessToken: "tok"}
	cfg := ucloud.NewConfig()
	cred := BuildCredentialFrom(credConfig)

	req := &request.CommonBase{}
	if err := req.SetAction("ListUPgSQLParamTemplate"); err != nil {
		t.Fatal(err)
	}
	if err := req.SetRegion("cn-bj2"); err != nil {
		t.Fatal(err)
	}
	httpReq, err := request.NewJSONEncoder(&cfg, cred).Encode(req)
	if err != nil {
		t.Fatal(err)
	}

	var before map[string]interface{}
	if err := json.Unmarshal(httpReq.GetRequestBody(), &before); err != nil {
		t.Fatal(err)
	}
	if _, ok := before["Signature"]; !ok {
		t.Fatal("premise broken: SDK JSONEncoder no longer attaches Signature for empty credential")
	}

	out, err := newCredHeaderInjector(credConfig)(nil, httpReq)
	if err != nil {
		t.Fatal(err)
	}
	var after map[string]interface{}
	if err := json.Unmarshal(out.GetRequestBody(), &after); err != nil {
		t.Fatal(err)
	}
	if _, ok := after["Signature"]; ok {
		t.Error("Signature survived the injector on a real SDK-encoded json body")
	}
	if _, ok := after["PublicKey"]; ok {
		t.Error("PublicKey survived the injector on a real SDK-encoded json body")
	}
	if after["Action"] != "ListUPgSQLParamTemplate" {
		t.Errorf("Action lost: %v", after)
	}
	if out.GetHeaderMap()["Authorization"] != "Bearer tok" {
		t.Error("Bearer missing after injector")
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
	// InitClientRuntime 会改写包级全局 ClientConfig/AuthCredential，恢复现场避免测试顺序耦合
	oldClientConfig, oldAuthCredential := ClientConfig, AuthCredential
	t.Cleanup(func() {
		ClientConfig, AuthCredential = oldClientConfig, oldAuthCredential
	})
	if err := InitClientRuntime(ac); err != nil {
		t.Fatal(err)
	}
	client := uaccount.NewClient(ClientConfig, BuildCredential())
	AttachHandlersWith(client, AuthCredential, ac, AggConfigListIns)
	if _, err := client.GetRegion(client.NewGetRegionRequest()); err != nil {
		t.Fatalf("GetRegion failed: %v", err)
	}
	if rec.params == nil {
		t.Fatal("server did not record any request")
	}
}

func newTestUAccountClient(t *testing.T, ac *AggConfig) *uaccount.UAccountClient {
	t.Helper()
	if err := InitClientRuntime(ac); err != nil {
		t.Fatal(err)
	}
	client := uaccount.NewClient(ClientConfig, BuildCredential())
	AttachHandlersWith(client, AuthCredential, ac, AggConfigListIns)
	return client
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

// channelInjectorHeaders 单独驱动渠道头注入 handler，取其产出的 header map。
// 注意查的是 SetHeader 原样存入的 map（未经 Go 规范化），故 key 必须与注入端字面一致。
func channelInjectorHeaders(t *testing.T, ac *AggConfig) map[string]string {
	t.Helper()
	h := newChannelHeaderInjector(ac)
	req, err := h(nil, uhttp.NewHttpRequest())
	if err != nil {
		t.Fatal(err)
	}
	return req.GetHeaderMap()
}

// 配了 channel_key 的 profile：注入 channel-key 头（AC1）
func TestChannelInjectorSetsHeader(t *testing.T) {
	headers := channelInjectorHeaders(t, &AggConfig{ChannelKey: "ch_combo_test"})
	if headers["channel-key"] != "ch_combo_test" {
		t.Errorf("channel-key = %q, want ch_combo_test", headers["channel-key"])
	}
}

// CRITICAL 零回归（AC2）：未配 channel_key 时该头必须完全不存在——不是空值，是键不存在。
// 主站用户与独立域名专属云渠道恒走此路径，线路字节必须与引入本特性前逐字节一致。
// 与同文件 Cookie/Csrf-Token「空值也照旧 set」的存量契约刻意相反：那是历史行为，
// 本头是新增的，注入空头会给全部存量用户构成回归。
func TestChannelInjectorAbsentWhenEmpty(t *testing.T) {
	for _, tc := range []struct {
		name string
		ac   *AggConfig
	}{
		{"empty channel key", &AggConfig{}},
		{"nil agg config", nil}, // AttachHandlersWith(sc, nil, nil, nil) 降级路径确实存在
	} {
		t.Run(tc.name, func(t *testing.T) {
			if v, ok := channelInjectorHeaders(t, tc.ac)["channel-key"]; ok {
				t.Errorf("channel-key header must be absent, got present with value %q", v)
			}
		})
	}
}

// channel-key 与凭据机制正交（AC3）：auth_mode 不影响其注入。
// spec auth-guidelines 的「一个请求只携带一种凭据机制」约束的是凭据，channel-key 是
// 渠道路由标识，不在其列。2026-07-16 真实网关实测：Bearer 与 channel-key 同时上行被接受。
func TestChannelInjectorOrthogonalToAuthMode(t *testing.T) {
	for _, tc := range []struct {
		name string
		ac   *AggConfig
	}{
		{"aksk", &AggConfig{ChannelKey: "ch_x", PublicKey: "pub", PrivateKey: "pri"}},
		{"oauth", &AggConfig{ChannelKey: "ch_x", AuthMode: AuthModeOAuth, AccessToken: "tok"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := channelInjectorHeaders(t, tc.ac)["channel-key"]; got != "ch_x" {
				t.Errorf("channel-key = %q, want ch_x", got)
			}
		})
	}
}

// 端到端：经完整 handler 链发真实请求，断言服务端侧确实收到该头。
// 这是唯一能验证「Go 的 Header.Set 规范化为 Channel-Key 后网关仍可取到」的用例
// （单测查的 map 未规范化，验证不到线路形式），也顺带证明 handler 确实被挂上了。
func TestChannelKeyHeaderEndToEnd(t *testing.T) {
	rec := &recordedRequest{}
	s := bizRecorderServer(t, rec)
	defer s.Close()

	ac := &AggConfig{
		Profile: "combo", BaseURL: s.URL, Timeout: 15, MaxRetryTimes: intPtr(0),
		Region: "hk", ChannelKey: "ch_combo_e2e",
		PublicKey: "pub", PrivateKey: "pri",
	}
	callGetRegion(t, ac, rec)

	// http.Header.Get 大小写不敏感：线路上是 Channel-Key，此处照样取得到
	if got := rec.header.Get("channel-key"); got != "ch_combo_e2e" {
		t.Errorf("server received channel-key = %q, want ch_combo_e2e", got)
	}
}

// AttachHandlers（读包级 ConfigIns 的那条路径，GetUserInfo 等在用）同样注入 channel-key。
// 该路径实测难以触发（仅 agree_upload_log=true 的 DAS 日志上传会走 GetUserInfo），故以单测钉死。
func TestAttachHandlersInjectsChannelKeyFromConfigIns(t *testing.T) {
	rec := &recordedRequest{}
	s := bizRecorderServer(t, rec)
	defer s.Close()

	oldConfig, oldClientConfig, oldCred := ConfigIns, ClientConfig, AuthCredential
	t.Cleanup(func() { ConfigIns, ClientConfig, AuthCredential = oldConfig, oldClientConfig, oldCred })

	ac := &AggConfig{
		Profile: "combo", BaseURL: s.URL, Timeout: 15, MaxRetryTimes: intPtr(0),
		Region: "hk", ChannelKey: "ch_from_configins",
		PublicKey: "pub", PrivateKey: "pri",
	}
	ConfigIns = ac
	if err := InitClientRuntime(ac); err != nil {
		t.Fatal(err)
	}
	client := uaccount.NewClient(ClientConfig, BuildCredential())
	AttachHandlers(client) // 包级路径：等价于 AttachHandlersWith(sc, AuthCredential, ConfigIns, ...)
	if _, err := client.GetRegion(client.NewGetRegionRequest()); err != nil {
		t.Fatalf("GetRegion failed: %v", err)
	}
	if got := rec.header.Get("channel-key"); got != "ch_from_configins" {
		t.Errorf("AttachHandlers must inject channel-key from ConfigIns, got %q", got)
	}
}

// 端到端零回归：未配 channel_key 时服务端不得看到该头
func TestChannelKeyHeaderAbsentEndToEnd(t *testing.T) {
	rec := &recordedRequest{}
	s := bizRecorderServer(t, rec)
	defer s.Close()

	ac := &AggConfig{
		Profile: "mainsite", BaseURL: s.URL, Timeout: 15, MaxRetryTimes: intPtr(0),
		Region: "cn-bj2", PublicKey: "pub", PrivateKey: "pri",
	}
	callGetRegion(t, ac, rec)

	if _, ok := rec.header["Channel-Key"]; ok {
		t.Errorf("main-site profile must not send channel-key header, got %q", rec.header.Get("channel-key"))
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

	client := newTestUAccountClient(t, ac)
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

	client := newTestUAccountClient(t, ac)
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

	client := newTestUAccountClient(t, ac)
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

	client := newTestUAccountClient(t, ac)
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

	client := newTestUAccountClient(t, ac)
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
	client := newTestUAccountClient(t, ac)
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
