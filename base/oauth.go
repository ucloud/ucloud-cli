// base/oauth.go
package base

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gofrs/flock"
	"github.com/mattn/go-isatty"
)

// OAuth 常量（spec D2：client_secret 嵌入二进制为知情裁定，同 gcloud/gh）。
const (
	defaultOAuthBaseURL = "https://oauth2.ucloud.cn"
	oauthClientID       = "WP77AwxvUgWt2JqaRCKn"
	oauthClientSecret   = "mksUQLod9VaUKMt3wESdgteTFCgVasiUwLSPqq5e"
	oauthRedirectPath   = "/authorization"
	oauthScope          = "openid email offline_access full_access"
)

// BuildLoopbackRedirectURI 按后端规则拼 loopback redirect_uri：host 必须是字面量 localhost
// （127.0.0.1 会被后端拒），端口为内核分配的临时端口（>=1024）。
func BuildLoopbackRedirectURI(port int) string {
	return fmt.Sprintf("http://localhost:%d%s", port, oauthRedirectPath)
}

// AuthModeOAuth auth_mode 取值：OAuth 浏览器登录。空串/其他值一律视为 AK/SK 签名模式。
const AuthModeOAuth = "oauth"

// TokenExpirySkew 主动刷新的时钟偏斜余量（D6）
const TokenExpirySkew = 5 * time.Minute

// GenerateState 生成 CSRF state：32 字节随机 base64url
func GenerateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate state failed: %v", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// GetOAuthBaseURL 生效的 OAuth 域名：profile 配置优先，否则内置默认（D9.2）
func GetOAuthBaseURL(cfg *AggConfig) (string, error) {
	if cfg.OAuthBaseURL != "" {
		return strings.TrimSuffix(cfg.OAuthBaseURL, "/"), nil
	}
	return defaultOAuthBaseURL, nil
}

// BuildAuthorizeURL 拼授权 URL（流程步骤①）
func BuildAuthorizeURL(oauthBase, redirectURI, state string) string {
	v := url.Values{}
	v.Set("response_type", "code")
	v.Set("client_id", oauthClientID)
	v.Set("redirect_uri", redirectURI)
	v.Set("scope", oauthScope)
	v.Set("state", state)
	return fmt.Sprintf("%s/authorize?%s", oauthBase, v.Encode())
}

// SanitizeCallbackInput 容忍前后空白/引号/终端折行引入的内部空白与换行（D7 输入容错）
func SanitizeCallbackInput(input string) string {
	s := strings.TrimSpace(input)
	s = strings.Trim(s, `"'`)
	return strings.Map(func(r rune) rune {
		switch r {
		case '\n', '\r', ' ', '\t':
			return -1
		}
		return r
	}, s)
}

const callbackFormatHint = "expected format: http://localhost/authorization?code=xxx&state=yyy"

// ParseCallbackURL 校验 state 并提取 code（流程步骤③）
func ParseCallbackURL(input, expectState string) (string, error) {
	s := SanitizeCallbackInput(input)
	u, err := url.Parse(s)
	if err != nil {
		return "", fmt.Errorf("cannot parse the pasted URL, no authorization code found; %s", callbackFormatHint)
	}
	q := u.Query()
	if e := q.Get("error"); e != "" {
		if e == "access_denied" {
			return "", fmt.Errorf("authorization was denied in the browser. Run 'ucloud auth login' to try again")
		}
		return "", fmt.Errorf("oauth server returned error %q. Run 'ucloud auth login' to try again", e)
	}
	code := q.Get("code")
	if code == "" {
		return "", fmt.Errorf("no authorization code in the pasted URL; %s", callbackFormatHint)
	}
	if q.Get("state") != expectState {
		return "", fmt.Errorf("state mismatch: the pasted URL likely comes from a previous login attempt. Run 'ucloud auth login' again and paste the URL from THIS attempt")
	}
	return code, nil
}

// TokenExpiredAt 判断 token 是否需要刷新（留 TokenExpirySkew 余量）
func TokenExpiredAt(expiresAt int64, now time.Time) bool {
	if expiresAt == 0 {
		return true
	}
	return now.Add(TokenExpirySkew).Unix() >= expiresAt
}

// TokenExpired TokenExpiredAt 的当前时间封装
func TokenExpired(expiresAt int64) bool {
	return TokenExpiredAt(expiresAt, time.Now())
}

// ParseIDTokenEmail 解 id_token payload 取 email。不验签，仅用于 UI 展示（D2 知情裁定）；id_token 不落盘。
func ParseIDTokenEmail(idToken string) (string, error) {
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("malformed id_token")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		// 部分 OIDC 实现会输出带 '=' 填充的 base64url，去填充后重试一次
		payload, err = base64.RawURLEncoding.DecodeString(strings.TrimRight(parts[1], "="))
		if err != nil {
			return "", fmt.Errorf("decode id_token payload failed: %v", err)
		}
	}
	var claims struct {
		Email string `json:"email"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return "", fmt.Errorf("parse id_token payload failed: %v", err)
	}
	return claims.Email, nil
}

// redactPatterns 覆盖 query 参数、JSON 字段、HTTP 头三种形态的敏感值
var redactPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)((?:^|[?&\s])code=)[^&\s"']+`),
	regexp.MustCompile(`(?i)((?:^|[?&\s])state=)[^&\s"']+`),
	regexp.MustCompile(`(?i)(access_token"?\s*[:=]\s*"?)[^,}&\s"']+`),
	regexp.MustCompile(`(?i)(refresh_token"?\s*[:=]\s*"?)[^,}&\s"']+`),
	regexp.MustCompile(`(?i)(id_token"?\s*[:=]\s*"?)[^,}&\s"']+`),
	regexp.MustCompile(`(?i)(authorization:?\s*bearer\s+)\S+`),
}

// Redact 脱敏 code/token/authorization（D7 最小脱敏，UC1 提前到 Phase 1）
func Redact(s string) string {
	for _, p := range redactPatterns {
		s = p.ReplaceAllString(s, "${1}********")
	}
	return s
}

// IsStdinTTY 判断 stdin 是否为交互终端（AP-1）。
// 不能用 os.ModeCharDevice：/dev/null 也是字符设备，cron/CI 重定向会被误判为交互。
// go-isatty 走真实终端检查（unix ioctl / windows console API），Cygwin/mintty 下 stdin 是管道，单独判。
func IsStdinTTY() bool {
	fd := os.Stdin.Fd()
	return isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd)
}

// OAuthLoginRequiredHint oauth 模式但 token 缺失时的提示（AP-1/AP-3，走 stderr）
func OAuthLoginRequiredHint(profile string, isTTY bool) string {
	if isTTY {
		return fmt.Sprintf("Profile '%s' uses OAuth login but has no token. Run 'ucloud auth login' first", profile)
	}
	return fmt.Sprintf("Profile '%s' uses OAuth login, which cannot work in a non-interactive environment. For automation/CI, use an AK/SK profile: ucloud config --profile <name> --public-key <pub> --private-key <pri>", profile)
}

// OAuthRefreshFailedHint refresh_token 失效/刷新失败时的提示（AP-3 模板）
func OAuthRefreshFailedHint(profile string, isTTY bool, err error) string {
	if isTTY {
		return fmt.Sprintf("Login expired for profile '%s' (%s). Run 'ucloud auth login' again", profile, Redact(err.Error()))
	}
	return fmt.Sprintf("OAuth login for profile '%s' cannot be renewed in a non-interactive environment (%s). For unattended scenarios, use an AK/SK profile instead", profile, Redact(err.Error()))
}

// CheckOAuthRunnable oauth 模式启动检查；ok=false 时调用方应将 msg 输出到 stderr 并以非零码退出。
// 此处刻意忽略 ExpiresAt——过期由 EnsureFreshToken（Task 6）处理，本函数只检查 token 是否存在。
func CheckOAuthRunnable(cfg *AggConfig, isTTY bool) (string, bool) {
	if cfg.AccessToken == "" || cfg.RefreshToken == "" {
		return OAuthLoginRequiredHint(cfg.Profile, isTTY), false
	}
	return "", true
}

// TokenResponse /token 端点响应（流程步骤④）
type TokenResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	IDToken          string `json:"id_token"`
	ExpiresIn        int64  `json:"expires_in"`
	TokenType        string `json:"token_type"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// oauthHTTPClient 使用默认 Transport：自动遵守 HTTPS_PROXY/HTTP_PROXY/NO_PROXY（ProxyFromEnvironment）
var oauthHTTPClient = &http.Client{Timeout: 30 * time.Second}

func requestToken(oauthBase string, form url.Values) (*TokenResponse, error) {
	endpoint := strings.TrimSuffix(oauthBase, "/") + "/token"
	resp, err := oauthHTTPClient.PostForm(endpoint, form)
	if err != nil {
		return nil, fmt.Errorf("cannot reach oauth server %s (check network or proxy settings): %v", endpoint, err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read oauth server response failed: %v", err)
	}
	var tr TokenResponse
	if jerr := json.Unmarshal(body, &tr); jerr != nil {
		if resp.StatusCode >= 500 {
			return nil, fmt.Errorf("oauth server error (HTTP %d), retry later", resp.StatusCode)
		}
		return nil, fmt.Errorf("unexpected oauth server response (HTTP %d): %s", resp.StatusCode, Redact(string(body)))
	}
	if tr.Error != "" {
		return nil, translateOAuthError(tr.Error, tr.ErrorDescription)
	}
	if resp.StatusCode >= 500 {
		return nil, fmt.Errorf("oauth server error (HTTP %d), retry later", resp.StatusCode)
	}
	if tr.AccessToken == "" {
		return nil, fmt.Errorf("oauth server returned no access_token (HTTP %d)", resp.StatusCode)
	}
	return &tr, nil
}

// translateOAuthError 按 AP-3 模板翻译 OAuth 错误码：原因 + 下一步命令
func translateOAuthError(code, desc string) error {
	switch code {
	case "invalid_grant":
		return fmt.Errorf("authorization code or refresh token expired or already used (each code works only once). Run 'ucloud auth login' again and paste the URL promptly")
	case "access_denied":
		return fmt.Errorf("authorization was denied. Run 'ucloud auth login' to try again")
	default:
		return fmt.Errorf("oauth server rejected the request: %s (%s). Run 'ucloud auth login' to start over", code, Redact(desc))
	}
}

// ExchangeToken 授权码换 token（流程步骤④）
func ExchangeToken(oauthBase, redirectURI, code string) (*TokenResponse, error) {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("client_id", oauthClientID)
	form.Set("client_secret", oauthClientSecret)
	form.Set("redirect_uri", redirectURI)
	return requestToken(oauthBase, form)
}

// RefreshToken 刷新 access_token；响应中的新 refresh_token 表示轮换（D3：旧的立即作废，必须写回）
func RefreshToken(oauthBase, refreshToken string) (*TokenResponse, error) {
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)
	form.Set("client_id", oauthClientID)
	form.Set("client_secret", oauthClientSecret)
	return requestToken(oauthBase, form)
}

// ApplyTokenResponse 把 /token 响应写入 cfg；轮换语义：响应带新 refresh_token 则覆盖（D3）
func ApplyTokenResponse(cfg *AggConfig, tr *TokenResponse) {
	cfg.AuthMode = AuthModeOAuth
	cfg.AccessToken = tr.AccessToken
	if tr.RefreshToken != "" {
		cfg.RefreshToken = tr.RefreshToken
	}
	cfg.ExpiresAt = time.Now().Unix() + tr.ExpiresIn
}

// EnsureFreshToken 主动刷新（D6）：过期（含 5min 偏斜余量）则 refresh 并写回。
// 「刷新+写回」由 refreshAndSave 内的 flock 串行化，拿锁后重读磁盘（Task 11）。
func EnsureFreshToken(cfg *AggConfig, manager *AggConfigManager) error {
	if !TokenExpired(cfg.ExpiresAt) {
		return nil
	}
	return refreshAndSave(cfg, manager)
}

// credentialLockPath flock 锁文件路径；包级变量便于测试注入
var credentialLockPath = ""

func getCredentialLockPath() string {
	if credentialLockPath != "" {
		return credentialLockPath
	}
	return GetConfigDir() + "/credential.lock"
}

// credentialLockTimeout 拿锁超时（D3：超时明确报错）
const credentialLockTimeout = 10 * time.Second

// refreshAndSave 串行化「刷新+写回」临界区（D3/D9.4）：
// flock 跨进程互斥 → 拿锁后重读磁盘（他进程可能已刷新并轮换）→ 仍过期才真正刷新。
func refreshAndSave(cfg *AggConfig, manager *AggConfigManager) error {
	staleToken := cfg.AccessToken

	fl := flock.New(getCredentialLockPath())
	ctx, cancel := context.WithTimeout(context.Background(), credentialLockTimeout)
	defer cancel()
	ok, err := fl.TryLockContext(ctx, 200*time.Millisecond)
	if err != nil && !errors.Is(err, context.DeadlineExceeded) {
		// 硬错误（如锁文件无权限），与拿锁超时是两回事，必须带上原始错误
		return fmt.Errorf("acquire credential lock %s failed: %v", getCredentialLockPath(), err)
	}
	if !ok {
		return fmt.Errorf("timed out acquiring credential lock %s after %v: another ucloud process may be refreshing, retry later", getCredentialLockPath(), credentialLockTimeout)
	}
	defer fl.Unlock()

	// 拿锁后重读：他进程已刷新则直接采用，避免用已作废的 refresh_token 二次刷新
	if disk, derr := readCredentialFromDisk(manager.credPath, cfg.Profile); derr == nil && disk != nil {
		if disk.AccessToken != "" && disk.AccessToken != staleToken && !TokenExpired(disk.ExpiresAt) {
			cfg.AccessToken = disk.AccessToken
			cfg.RefreshToken = disk.RefreshToken
			cfg.ExpiresAt = disk.ExpiresAt
			cfg.AuthMode = AuthModeOAuth
			return nil
		}
		if disk.RefreshToken != "" {
			cfg.RefreshToken = disk.RefreshToken // 轮换后的最新 refresh_token 以磁盘为准
		}
	}

	oauthBase, err := GetOAuthBaseURL(cfg)
	if err != nil {
		return err
	}
	tr, err := RefreshToken(oauthBase, cfg.RefreshToken)
	if err != nil {
		return err
	}
	ApplyTokenResponse(cfg, tr)
	// Save() 会把内存里全部 profile 整写落盘，而本进程内存是 t0 快照：他进程可能已在
	// t0 之后轮换了其它 profile 的 refresh_token（D3 旧的立即作废）。落盘前重读磁盘，
	// 把「非当前 profile」的 oauth 字段以磁盘为准合并，否则会把轮换结果覆盖回陈旧值，
	// 导致对方 profile 下次刷新 invalid_grant（被迫重新登录）。
	if creds, rerr := readAllCredentialsFromDisk(manager.credPath); rerr == nil {
		mergeOtherProfilesOAuthFromDisk(manager.configs, creds, cfg.Profile)
	}
	if err := manager.Save(); err != nil {
		return fmt.Errorf("token refreshed but saving credential failed: %v", err)
	}
	return nil
}

// mergeOtherProfilesOAuthFromDisk 把磁盘版凭据中「非当前 profile」的 oauth 四字段
// （auth_mode/access_token/refresh_token/expires_at）合并进内存。只合并这四个字段：
// flock 临界区内唯一的合法并发写就是 oauth 刷新轮换，AK/SK、cookie 等字段不受锁保护、
// 不在此处静默采纳。当前 profile 保持本次刷新后的内存值。
func mergeOtherProfilesOAuthFromDisk(configs map[string]*AggConfig, diskCreds []CredentialConfig, currentProfile string) {
	for i := range diskCreds {
		dc := &diskCreds[i]
		if dc.Profile == currentProfile {
			continue
		}
		ac, ok := configs[dc.Profile]
		if !ok {
			continue
		}
		ac.AuthMode = dc.AuthMode
		ac.AccessToken = dc.AccessToken
		ac.RefreshToken = dc.RefreshToken
		ac.ExpiresAt = dc.ExpiresAt
	}
}

// readAllCredentialsFromDisk 重新读盘取全部 profile 的最新凭据（不经 manager 缓存）
func readAllCredentialsFromDisk(credPath string) ([]CredentialConfig, error) {
	raw, err := ioutil.ReadFile(credPath)
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 {
		return nil, nil
	}
	var creds []CredentialConfig
	if err := json.Unmarshal(raw, &creds); err != nil {
		return nil, err
	}
	return creds, nil
}

// readCredentialFromDisk 重新读盘取指定 profile 的最新凭据（不经 manager 缓存）
func readCredentialFromDisk(credPath, profile string) (*CredentialConfig, error) {
	creds, err := readAllCredentialsFromDisk(credPath)
	if err != nil {
		return nil, err
	}
	for i := range creds {
		if creds[i].Profile == profile {
			return &creds[i], nil
		}
	}
	return nil, nil
}
