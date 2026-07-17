package platform

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/ucloud/ucloud-sdk-go/private/protocol/http"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
	uerr "github.com/ucloud/ucloud-sdk-go/ucloud/error"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"
)

// newCredHeaderInjector 返回凭据头注入 handler。
// aksk/CloudShell 行为与历史完全一致（Cookie/Csrf-Token 始终 set，含空值）；
// auth_mode==oauth 时剥离 SDK 编码器无条件附加的签名参数（Credential.Apply 即使
// 空密钥也会算出 Signature），并在 token 非空时追加 Authorization: Bearer，
// 保证 oauth 请求只携带 Bearer 一种凭据机制（凭据模型见 spec §2）。
func newCredHeaderInjector(credConfig *CredentialConfig) sdk.HttpRequestHandler {
	if credConfig == nil {
		credConfig = &CredentialConfig{}
	}
	return func(c *sdk.Client, req *http.HttpRequest) (*http.HttpRequest, error) {
		if err := req.SetHeader("Cookie", credConfig.Cookie); err != nil {
			return req, err
		}
		if err := req.SetHeader("Csrf-Token", credConfig.CSRFToken); err != nil {
			return req, err
		}
		if credConfig.AuthMode == AuthModeOAuth {
			// 按 Content-Type 分派剥离：编码器决定 body 形态，剥离方式必须与之配对。
			// 不能只按一种形态盲剥 —— url.ParseQuery 对 JSON 往往"成功"，重编码会
			// 悄悄毁掉 body；不认识的 Content-Type 一律不碰 body。
			switch req.GetHeaderMap()[http.HeaderNameContentType] {
			case http.MimeFormURLEncoded:
				vals, err := url.ParseQuery(string(req.GetRequestBody()))
				if err != nil {
					// 剥不掉就明确失败：客户端报错优于网关 171
					return req, fmt.Errorf("strip signature params from oauth request failed: %w", err)
				}
				vals.Del("Signature")
				vals.Del("PublicKey")
				if err := req.SetRequestBody([]byte(vals.Encode())); err != nil {
					return req, err
				}
			case http.MimeJSON:
				// JSONEncoder 把 cred.Apply 附加的 Signature/PublicKey 放在 body 顶层
				// （SDK ucloud/request/encoder_json.go），与 form 分支同构剥离。
				var payload map[string]interface{}
				if err := json.Unmarshal(req.GetRequestBody(), &payload); err != nil {
					return req, fmt.Errorf("strip signature params from oauth json request failed: %w", err)
				}
				delete(payload, "Signature")
				delete(payload, "PublicKey")
				bs, err := json.Marshal(payload)
				if err != nil {
					return req, fmt.Errorf("re-encode oauth json request failed: %w", err)
				}
				if err := req.SetRequestBody(bs); err != nil {
					return req, err
				}
			}
			if credConfig.AccessToken != "" {
				if err := req.SetHeader("Authorization", "Bearer "+credConfig.AccessToken); err != nil {
					return req, err
				}
			}
		}
		return req, nil
	}
}

// newChannelHeaderInjector 返回专属云渠道头注入 handler。
//
// 网关据 channel-key 识别「复用主站域名」的专属云渠道（如 api.ucloud-global.com）；
// 独立域名渠道与主站用户没有这个 key，故空值必须完全不注入该头 —— 与本文件中
// Cookie/Csrf-Token「空值也照旧 set」的历史行为刻意相反：那是存量契约，而这是新增头，
// 给全部存量用户平白加一个空头即构成回归。
//
// 取值源是 ac 而非 credConfig：channel-key 是接入点配置（与 base_url 同类、成对配置），
// 不是凭据机制，与 auth_mode 正交（spec auth-guidelines「一个请求只携带一种凭据机制」
// 说的是凭据，不含本头）。且 newServiceClientForConfig 传入的 ac 正是用户此刻正在配置
// 的 profile，config 子命令自身的 region/project 远程校验请求因此天然带上正确的 key，
// 不会出现「配置时校验失败 → 配不上」的死锁。
//
// 实测（2026-07-16，真实 combo 账号 + 真实网关）：同一 token 同一域名，唯一变量为本头，
// RetCode 174 → 0；且 Go 的 Header.Set 规范化为 Channel-Key 后网关照常接受。
func newChannelHeaderInjector(ac *AggConfig) sdk.HttpRequestHandler {
	return func(c *sdk.Client, req *http.HttpRequest) (*http.HttpRequest, error) {
		if ac == nil || ac.ChannelKey == "" {
			return req, nil
		}
		return req, req.SetHeader("channel-key", ac.ChannelKey)
	}
}

// authRetCodeWhitelist 鉴权类 RetCode 白名单（D6）。实测网关（2026-06-11 实探）：
// 鉴权失败以 HTTP 200 + RetCode 返回，401 仅作防御性分支保留。
// 174 "Token Not Exists"：伪造与已过期的 Bearer 同为 174（已实测确认）；属网关
// 前置鉴权拒绝，业务必未执行，重放一次安全。网关团队书面确认仍待补档（spec §7）。
// 170（缺签名，oauth 请求恒带 Bearer 不会触发）、171/172（AK/SK 路径）不入列。
//
// 174 的第三种成因（2026-07-16 实测发现，已知且刻意保留现状）：channel-key 与账号
// 所属渠道不匹配时网关同样返回 174（同一有效 token 换成别的渠道 key 即报 174）。
// 于是本白名单会把「配错 channel-key」误判为 token 过期 → 刷新（refresh_token 轮转、
// 旧的立即作废）→ 重放 → 仍 174。后果是每次请求白耗一轮刷新，且错误文案把用户导向
// 重新登录 —— 而重新登录永远治不好。本次不改控制流（改了会波及正常的过期刷新路径）。
// 排障：OAuth profile 报 174 且重新登录无效时，优先检查 channel_key 是否与账号所属
// 渠道匹配。
var authRetCodeWhitelist = map[int]bool{
	174: true, // Token Not Exists：无效或过期 Bearer（亦可能是 channel-key 不匹配，见上）
}

// isAuthFailure 判定是否鉴权类失败：HTTP 401 或 body RetCode 在白名单（网关前置鉴权，业务必未执行）。
// 注意 SDK 行为：HttpClient.Send 对 status>=400 返回 (nil, StatusError)（vendor
// private/protocol/http/client.go），且默认 errorHTTPHandler 先于本 handler 把它
// 转成 uerr.ServerError —— 401 只会出现在 err 里、resp 必为 nil；resp 路径仅作
// RetCode 白名单（HTTP 200 + 鉴权 RetCode）的判定入口。
func isAuthFailure(resp *http.HttpResponse, err error) bool {
	switch e := err.(type) {
	case http.StatusError:
		if e.StatusCode == 401 {
			return true
		}
	case uerr.ServerError:
		if e.StatusCode() == 401 {
			return true
		}
	}
	if resp == nil {
		return false
	}
	var body struct {
		RetCode int `json:"RetCode"`
	}
	if jerr := json.Unmarshal(resp.GetBody(), &body); jerr == nil {
		return authRetCodeWhitelist[body.RetCode]
	}
	return false
}

// newOAuthRetryHandler 反应式兜底（D6，Google 式）：鉴权失败 → 刷新 → 自动重放一次。
// 重放直接走 httpClient.Send，不再经过本 handler，天然不会循环。
// 刷新对象是构造本 client 的 ac（而非 ConfigIns）：cmd/root.go 的 os.Args 扫描
// 识别不了 -p X/--profile=X 等形式，ConfigIns 可能指向另一个 profile，错刷会把
// 别人的 Bearer 重放到当前请求上。ac 在所有 oauth 路径上都是 manager 持有的指针
// （GetAggConfigByProfile/Append 直接存取同一指针），refreshAndSave 的写回因此可靠。
// req 的 Authorization 由 SetHeader 以 map 赋值覆盖（不会叠加重复头），且 body 中
// 的签名参数已被 newCredHeaderInjector 剥离，重放仍满足「oauth 请求只带 Bearer」不变式。
func newOAuthRetryHandler(credConfig *CredentialConfig, ac *AggConfig, manager *AggConfigManager) sdk.HttpResponseHandler {
	return func(c *sdk.Client, req *http.HttpRequest, resp *http.HttpResponse, err error) (*http.HttpResponse, error) {
		if ac == nil || credConfig.AuthMode != AuthModeOAuth || credConfig.AccessToken == "" {
			return resp, err
		}
		if !isAuthFailure(resp, err) {
			return resp, err
		}
		if manager == nil {
			LogWarn("oauth reactive refresh skipped: config manager is not initialized")
			return resp, err
		}
		// 刷新（flock 串行化 + 拿锁后重读，见 refreshAndSave）
		if rerr := refreshAndSave(ac, manager); rerr != nil {
			LogWarn(fmt.Sprintf("oauth reactive refresh failed: %v", Redact(rerr.Error())))
			return resp, err
		}
		credConfig.AccessToken = ac.AccessToken
		_ = req.SetHeader("Authorization", "Bearer "+credConfig.AccessToken) // SetHeader 恒返回 nil
		LogInfo("auth failure detected, token refreshed, replaying request once")
		hc := http.NewHttpClient()
		nresp, nerr := hc.Send(req)
		if serr, ok := nerr.(http.StatusError); ok {
			// 本 handler 位于链尾，重放结果不会再经过默认 errorHTTPHandler，
			// 在此对齐其行为：StatusError → uerr.ServerError
			nerr = uerr.NewServerStatusError(serr.StatusCode, serr.Message)
		}
		return nresp, nerr
	}
}

// buildCredential 构造 SDK 签名凭据。
// 不变式：一个请求只携带一种凭据机制（auth_mode 唯一决定走哪种）。
// oauth profile 会保留旧 AK/SK 在磁盘上（供 auth logout 恢复），但它们必须
// 对 SDK 签名器不可见——否则签名参数与 Bearer 同时上行，网关先验签名
// 直接报 RetCode 171 Signature VerifyAC Error。oauth 模式下凭据留空；
// 注意 SDK 编码器对空密钥仍会附加 Signature 参数，由 newCredHeaderInjector
// 剥离，最终 Bearer 是唯一凭据。AK/SK 模式填真实公私钥，SDK 签名器据此签名。
func buildCredential(credConfig *CredentialConfig) *auth.Credential {
	if credConfig == nil {
		credConfig = &CredentialConfig{}
	}
	credential := &auth.Credential{}
	if credConfig.AuthMode != AuthModeOAuth {
		credential.PublicKey = credConfig.PublicKey
		credential.PrivateKey = credConfig.PrivateKey
	}
	return credential
}

// BuildCredential 从包级 AuthCredential（由 InitConfig/InitClientRuntime 填充）构造签名凭据。
// 供 cli.NewServiceClient 使用——与 NewClient 走完全相同的 buildCredential 逻辑/分支，
// oauth 与 AK/SK profile 共用一条代码路径（不分叉，§9 无鉴权回归）。
func BuildCredential() *auth.Credential {
	return BuildCredentialFrom(AuthCredential)
}

// BuildCredentialFrom constructs an SDK signing credential from an explicit
// credential config, without reading package-level runtime state.
func BuildCredentialFrom(credConfig *CredentialConfig) *auth.Credential {
	return buildCredential(credConfig)
}

// normalizeProjectID 把 project-id 从补全带出的 "id/name" 形态还原成纯 id
// （补全候选由 getProjectList 生成，形如 "org-xxx/ProjectName"）。
//
// typed 请求 SetProjectId 写进 CommonBase 即可。GenericRequest 要多一步：SDK 的
// BaseGenericRequest 把 GetProjectId() override 成 payload 优先，却没有 override
// SetProjectId —— SetProjectId 写的是 CommonBase，而 GetPayload() 末尾用 payload
// 覆盖 CommonBase（SDK ucloud/request/generic.go），于是归一化被 payload 里的原值
// 吃掉，"org-x/Name" 原样上行，网关报 RetCode 292 Project [org-x/Name] not exists
// （2026-07-14 对 api.ucloud.cn 实测）。凡是把 ProjectId 放进 payload map 的产品
// 都会中招，故在此一并同步 payload。
//
// 未发生归一化时（已是纯 id 或为空，即绝大多数调用）直接返回，payload 一字不动 ——
// 行为与历史逐字节一致。
func normalizeProjectID(req request.Common) (request.Common, error) {
	raw := req.GetProjectId()
	normalized := PickResourceID(raw)
	if err := req.SetProjectId(normalized); err != nil {
		return req, err
	}
	if raw == normalized {
		return req, nil
	}
	gr, ok := req.(request.GenericRequest)
	if !ok {
		return req, nil
	}
	payload := gr.GetPayload()
	if _, exists := payload["ProjectId"]; !exists {
		return req, nil
	}
	payload["ProjectId"] = normalized
	if err := gr.SetPayload(payload); err != nil {
		return req, err
	}
	return req, nil
}

// attachHandlers 把平台 handler 挂到 service client 上：
// project-id 归一化、请求日志、凭据头注入、专属云渠道头注入、oauth 反应式重试。
// credConfig 与 ac 显式传入：NewClient 借此传它自己的构造来源 profile（ac），
// 重试目标必须是构造本 client 的 profile，而非包级 ConfigIns
// （详见 newOAuthRetryHandler 的注释：os.Args 扫描识别不了 -p X/--profile=X，
// ConfigIns 可能指向另一个 profile，错刷会把别人的 Bearer 重放到当前请求）。
func attachHandlersWithManager(sc sdk.ServiceClient, credConfig *CredentialConfig, ac *AggConfig, manager *AggConfigManager) {
	if credConfig == nil {
		credConfig = &CredentialConfig{}
	}
	sc.AddRequestHandler(func(c *sdk.Client, req request.Common) (request.Common, error) {
		return normalizeProjectID(req)
	})
	// Platform request logging: every API request is logged uniformly at the SDK
	// layer (replaces per-command hand-rolled logging; products no longer build
	// "api:..." lines with ToQueryMap). logToFile writes to local cli.log only
	// (NO DAS upload) and skips completion (COMP_LINE) — see batch-1 plan Part 0
	// Task 0.2 (decision A: keep request logs local, don't inflate telemetry).
	sc.AddRequestHandler(func(c *sdk.Client, req request.Common) (request.Common, error) {
		logToFile(requestLogLine(req))
		return req, nil
	})
	sc.AddHttpRequestHandler(newCredHeaderInjector(credConfig))
	sc.AddHttpRequestHandler(newChannelHeaderInjector(ac))
	sc.AddHttpResponseHandler(newOAuthRetryHandler(credConfig, ac, manager))
}

// AttachHandlers 用包级 AuthCredential/ConfigIns（由 InitConfig 填充）把平台 handler
// 挂到 sc 上。供 cli.NewServiceClient 使用——此时活动 profile 就是 ConfigIns，
// 它正是正确的反应式刷新目标。
func AttachHandlers(sc sdk.ServiceClient) {
	AttachHandlersWith(sc, AuthCredential, ConfigIns, AggConfigListIns)
}

// AttachHandlersWith attaches platform handlers using explicit runtime state,
// so callers do not need the old aggregate base client singleton.
func AttachHandlersWith(sc sdk.ServiceClient, credConfig *CredentialConfig, ac *AggConfig, manager *AggConfigManager) {
	attachHandlersWithManager(sc, credConfig, ac, manager)
}
