// base/endpoints.go
// 集中所有外部服务的 URL、主机与端点路径，避免分散在多个文件中。
// 两个独立的服务域：
//   - 业务 API 网关（api.ucloud.cn）：承载全部云资源 RPC 调用，AK/SK 签名或 Bearer。
//   - OAuth 授权服务（oauth2.ucloud.cn）：承载浏览器登录的 /authorize 与 /token 端点。
//
// 回环回调常量供本地 loopback 自动捕获 / 手工粘贴使用。
package base

const (
	// ---- 业务 API 网关 ----
	// DefaultBaseURL location of api server
	DefaultBaseURL = "https://api.ucloud.cn/"

	// ---- OAuth 授权服务 ----
	// defaultOAuthBaseURL 内置默认 OAuth 域名（profile 的 oauth_base_url 可覆盖，见 GetOAuthBaseURL）
	defaultOAuthBaseURL = "https://oauth2.ucloud.cn"
	oauthAuthorizePath  = "/authorize"
	oauthTokenPath      = "/token"
	oauthScope          = "openid email offline_access full_access"

	// OAuth 客户端凭据（public client：secret 嵌入二进制为知情裁定，同 gcloud/gh）
	oauthClientID     = "WP77AwxvUgWt2JqaRCKn"
	oauthClientSecret = "mksUQLod9VaUKMt3wESdgteTFCgVasiUwLSPqq5e"

	// ---- Loopback 回调 ----
	// LoopbackListenHost 本地回调 server 的监听地址（回环 IP，导出供 cmd/callback.go 复用）。
	// loopbackRedirectHost redirect_uri 里的 host 必须是字面量 localhost——后端拒 127.0.0.1 形式的 redirect_uri。
	// 两者指向同一回环地址，但写法必须如此区分。
	LoopbackListenHost   = "127.0.0.1"
	loopbackRedirectHost = "localhost"
	// OAuthRedirectPath redirect_uri 与回调 server mux 共用的路径，必须一致（导出供 cmd/callback.go 复用）。
	OAuthRedirectPath = "/authorization"
)
