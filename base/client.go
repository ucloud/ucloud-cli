package base

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/ucloud/ucloud-sdk-go/private/protocol/http"
	ppathx "github.com/ucloud/ucloud-sdk-go/private/services/pathx"
	pudb "github.com/ucloud/ucloud-sdk-go/private/services/udb"
	puhost "github.com/ucloud/ucloud-sdk-go/private/services/uhost"
	pumem "github.com/ucloud/ucloud-sdk-go/private/services/umem"
	"github.com/ucloud/ucloud-sdk-go/services/pathx"
	"github.com/ucloud/ucloud-sdk-go/services/uaccount"
	"github.com/ucloud/ucloud-sdk-go/services/ucompshare"
	"github.com/ucloud/ucloud-sdk-go/services/udb"
	"github.com/ucloud/ucloud-sdk-go/services/udisk"
	"github.com/ucloud/ucloud-sdk-go/services/udpn"
	"github.com/ucloud/ucloud-sdk-go/services/uhost"
	"github.com/ucloud/ucloud-sdk-go/services/ulb"
	"github.com/ucloud/ucloud-sdk-go/services/umem"
	"github.com/ucloud/ucloud-sdk-go/services/unet"
	"github.com/ucloud/ucloud-sdk-go/services/uphost"
	"github.com/ucloud/ucloud-sdk-go/services/vpc"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
	uerr "github.com/ucloud/ucloud-sdk-go/ucloud/error"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"
)

// PrivateUHostClient 私有模块的uhost client 即未在官网开放的接口
type PrivateUHostClient = puhost.UHostClient

// PrivateUDBClient 私有模块的udb client 即未在官网开放的接口
type PrivateUDBClient = pudb.UDBClient

// PrivateUMemClient 私有模块的umem client 即未在官网开放的接口
type PrivateUMemClient = pumem.UMemClient

// PrivatePathxClient 私有模块的pathx client 即未在官网开放的接口
type PrivatePathxClient = ppathx.PathXClient

// Client aggregate client for business
type Client struct {
	uaccount.UAccountClient
	uhost.UHostClient
	unet.UNetClient
	vpc.VPCClient
	udpn.UDPNClient
	pathx.PathXClient
	udisk.UDiskClient
	ulb.ULBClient
	udb.UDBClient
	umem.UMemClient
	uphost.UPHostClient
	PrivateUHostClient
	PrivateUDBClient
	PrivateUMemClient PrivateUMemClient
	PrivatePathxClient
	ucompshare.UCompShareClient
}

// newCredHeaderInjector 返回凭据头注入 handler。
// aksk/CloudShell 行为与历史完全一致（Cookie/Csrf-Token 始终 set，含空值）；
// auth_mode==oauth 时剥离 SDK 编码器无条件附加的签名参数（Credential.Apply 即使
// 空密钥也会算出 Signature），并在 token 非空时追加 Authorization: Bearer，
// 保证 oauth 请求只携带 Bearer 一种凭据机制（凭据模型见 spec §2）。
func newCredHeaderInjector(credConfig *CredentialConfig) sdk.HttpRequestHandler {
	return func(c *sdk.Client, req *http.HttpRequest) (*http.HttpRequest, error) {
		if err := req.SetHeader("Cookie", credConfig.Cookie); err != nil {
			return req, err
		}
		if err := req.SetHeader("Csrf-Token", credConfig.CSRFToken); err != nil {
			return req, err
		}
		if credConfig.AuthMode == AuthModeOAuth {
			// 仅对 form-urlencoded body 剥离（JSON 编码器虽当前不可达，但
			// url.ParseQuery 对 JSON 往往"成功"，重编码会悄悄毁掉 body）
			if req.GetHeaderMap()[http.HeaderNameContentType] == http.MimeFormURLEncoded {
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

// authRetCodeWhitelist 鉴权类 RetCode 白名单（D6）。实测网关（2026-06-11 实探）：
// 鉴权失败以 HTTP 200 + RetCode 返回，401 仅作防御性分支保留。
// 174 "Token Not Exists"：伪造与已过期的 Bearer 同为 174（已实测确认）；属网关
// 前置鉴权拒绝，业务必未执行，重放一次安全。网关团队书面确认仍待补档（spec §7）。
// 170（缺签名，oauth 请求恒带 Bearer 不会触发）、171/172（AK/SK 路径）不入列。
var authRetCodeWhitelist = map[int]bool{
	174: true, // Token Not Exists：无效或过期 Bearer
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
func newOAuthRetryHandler(credConfig *CredentialConfig, ac *AggConfig) sdk.HttpResponseHandler {
	return func(c *sdk.Client, req *http.HttpRequest, resp *http.HttpResponse, err error) (*http.HttpResponse, error) {
		if ac == nil || credConfig.AuthMode != AuthModeOAuth || credConfig.AccessToken == "" {
			return resp, err
		}
		if !isAuthFailure(resp, err) {
			return resp, err
		}
		// 刷新（flock 串行化 + 拿锁后重读，见 refreshAndSave）
		if rerr := refreshAndSave(ac, AggConfigListIns); rerr != nil {
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
	credential := &auth.Credential{}
	if credConfig.AuthMode != AuthModeOAuth {
		credential.PublicKey = credConfig.PublicKey
		credential.PrivateKey = credConfig.PrivateKey
	}
	return credential
}

// BuildCredential 从包级 AuthCredential（由 InitConfig/GetBizClient 填充）构造签名凭据。
// 供 cli.NewServiceClient 使用——与 NewClient 走完全相同的 buildCredential 逻辑/分支，
// oauth 与 AK/SK profile 共用一条代码路径（不分叉，§9 无鉴权回归）。
func BuildCredential() *auth.Credential {
	return buildCredential(AuthCredential)
}

// attachHandlers 把三个平台 handler 挂到 service client 上：
// project-id 归一化、凭据头注入、oauth 反应式重试。
// credConfig 与 ac 显式传入：NewClient 借此传它自己的构造来源 profile（ac），
// 重试目标必须是构造本 client 的 profile，而非包级 ConfigIns
// （详见 newOAuthRetryHandler 的注释：os.Args 扫描识别不了 -p X/--profile=X，
// ConfigIns 可能指向另一个 profile，错刷会把别人的 Bearer 重放到当前请求）。
func attachHandlers(sc sdk.ServiceClient, credConfig *CredentialConfig, ac *AggConfig) {
	sc.AddRequestHandler(func(c *sdk.Client, req request.Common) (request.Common, error) {
		err := req.SetProjectId(PickResourceID(req.GetProjectId()))
		return req, err
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
	sc.AddHttpResponseHandler(newOAuthRetryHandler(credConfig, ac))
}

// AttachHandlers 用包级 AuthCredential/ConfigIns（由 InitConfig 填充）把平台 handler
// 挂到 sc 上。供 cli.NewServiceClient 使用——此时活动 profile 就是 ConfigIns，
// 它正是正确的反应式刷新目标。
func AttachHandlers(sc sdk.ServiceClient) {
	attachHandlers(sc, AuthCredential, ConfigIns)
}

// NewClient will return a aggregate client.
// ac 是构造来源 profile（oauth 401 反应式刷新的对象），允许为 nil（此时不重放）。
func NewClient(config *sdk.Config, credConfig *CredentialConfig, ac *AggConfig) *Client {
	credential := buildCredential(credConfig)
	var (
		uaccountClient = *uaccount.NewClient(config, credential)
		uhostClient    = *uhost.NewClient(config, credential)
		unetClient     = *unet.NewClient(config, credential)
		vpcClient      = *vpc.NewClient(config, credential)
		udpnClient     = *udpn.NewClient(config, credential)
		pathxClient    = *pathx.NewClient(config, credential)
		udiskClient    = *udisk.NewClient(config, credential)
		ulbClient      = *ulb.NewClient(config, credential)
		udbClient      = *udb.NewClient(config, credential)
		umemClient     = *umem.NewClient(config, credential)
		uphostClient   = *uphost.NewClient(config, credential)
		puhostClient   = *puhost.NewClient(config, credential)
		pudbClient     = *pudb.NewClient(config, credential)
		pumemClient    = *pumem.NewClient(config, credential)
		ppathxClient   = *ppathx.NewClient(config, credential)
		ulhostClient   = *ucompshare.NewClient(config, credential)
	)

	attachHandlers(&uaccountClient, credConfig, ac)
	attachHandlers(&uhostClient, credConfig, ac)
	attachHandlers(&unetClient, credConfig, ac)
	attachHandlers(&vpcClient, credConfig, ac)
	attachHandlers(&udpnClient, credConfig, ac)
	attachHandlers(&pathxClient, credConfig, ac)
	attachHandlers(&udiskClient, credConfig, ac)
	attachHandlers(&ulbClient, credConfig, ac)
	attachHandlers(&udbClient, credConfig, ac)
	attachHandlers(&umemClient, credConfig, ac)
	attachHandlers(&uphostClient, credConfig, ac)
	attachHandlers(&puhostClient, credConfig, ac)
	attachHandlers(&pudbClient, credConfig, ac)
	attachHandlers(&pumemClient, credConfig, ac)
	attachHandlers(&ppathxClient, credConfig, ac)
	attachHandlers(&ulhostClient, credConfig, ac)

	return &Client{
		uaccountClient,
		uhostClient,
		unetClient,
		vpcClient,
		udpnClient,
		pathxClient,
		udiskClient,
		ulbClient,
		udbClient,
		umemClient,
		uphostClient,
		puhostClient,
		pudbClient,
		pumemClient,
		ppathxClient,
		ulhostClient,
	}
}
