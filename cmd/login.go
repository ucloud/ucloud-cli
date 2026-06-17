// cmd/login.go
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/uaccount"

	"github.com/ucloud/ucloud-cli/base"
)

const loginLongHelp = `Log in to UCloud via your browser (OAuth authorization code flow).

How it works (default):
  1. ucloud-cli opens your browser at the UCloud authorization page.
  2. You log in and approve. The browser is redirected to a local callback
     that ucloud-cli is listening on — captured automatically, no copy-paste.
  3. ucloud-cli exchanges the code for tokens, saves them to
     ~/.ucloud/credential.json (0600), and auto-configures the default
     region/zone/project for this profile.

Headless / SSH: pass --no-browser. ucloud-cli prints the authorization URL;
open it on any device, log in, then copy the FULL callback URL from the
address bar and paste it back into the terminal.

Tokens are valid for about 1 hour and renew silently via the refresh token.
OAuth login targets interactive human use. For scripts and CI/CD, use an
AK/SK profile instead: ucloud config --profile <name> --public-key ... --private-key ...`

// oauthHelpTmpl 在全局 helpTmpl（不渲染 Long）前面补上 Long 段，仅作用于 login/logout
const oauthHelpTmpl = `{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}

{{end}}` + helpTmpl

// NewCmdAuth ucloud auth 命令组：浏览器登录相关子命令
func NewCmdAuth() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authenticate ucloud-cli via browser (OAuth)",
		Long:  "Browser-based OAuth authentication for ucloud-cli. Subcommands: login, logout",
	}
	cmd.SetHelpTemplate(oauthHelpTmpl)
	cmd.AddCommand(NewCmdLogin())
	cmd.AddCommand(NewCmdLogout())
	return cmd
}

// NewCmdLogin ucloud auth login
func NewCmdLogin() *cobra.Command {
	var noBrowser bool
	var oauthBaseURL string
	cmd := &cobra.Command{
		Use:     "login",
		Short:   "Log in to UCloud via browser (OAuth)",
		Long:    loginLongHelp,
		Args:    cobra.NoArgs,
		Example: "ucloud auth login\nucloud auth login --no-browser",
		Run: func(cmd *cobra.Command, args []string) {
			runLogin(noBrowser, oauthBaseURL)
		},
	}
	cmd.Flags().BoolVar(&noBrowser, "no-browser", false, "Print the authorization URL instead of opening a browser (for headless/SSH environments)")
	cmd.Flags().StringVar(&oauthBaseURL, "oauth-base-url", "", "Override the OAuth authorization server URL (for non-default environments; persisted to the profile)")
	cmd.SetHelpTemplate(oauthHelpTmpl)
	return cmd
}

// resolveLoginOAuthBase 决定登录使用的 OAuth 域：--oauth-base-url flag 最优先，
// 给定时写回 cfg.OAuthBaseURL 以便登录成功后随 profile 持久化（后续刷新沿用）；
// 未给定则回退到 profile 配置或内置默认（GetOAuthBaseURL）。
func resolveLoginOAuthBase(cfg *base.AggConfig, flagVal string) (string, error) {
	if flagVal != "" {
		cfg.OAuthBaseURL = strings.TrimSuffix(flagVal, "/")
	}
	oauthBase, err := base.GetOAuthBaseURL(cfg)
	if err == nil && cfg.OAuthBaseURL == "" {
		cfg.OAuthBaseURL = oauthBase
	}
	return oauthBase, err
}

func runLogin(noBrowser bool, oauthBaseURL string) {
	// AP-1：非 TTY fail-fast
	if !base.IsStdinTTY() {
		fmt.Fprintln(os.Stderr, "'ucloud auth login' requires an interactive terminal. For automation/CI, use an AK/SK profile: ucloud config --profile <name> --public-key <pub> --private-key <pri>")
		os.Exit(1)
	}

	cfg := base.ConfigIns
	oauthBase, err := resolveLoginOAuthBase(cfg, oauthBaseURL)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	state, err := base.GenerateState()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var code, redirectURI string
	if noBrowser {
		code, redirectURI = runLoginManual(oauthBase, state)
	} else {
		code, redirectURI = runLoginAuto(oauthBase, state)
	}

	tr, err := base.ExchangeToken(oauthBase, redirectURI, code)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// D5：已有 AK/SK 时打印一行告知 + 回退指引
	if cfg.PublicKey != "" || cfg.PrivateKey != "" {
		fmt.Printf("Note: profile '%s' had AK/SK configured; it now switches to OAuth (auth_mode=oauth). AK/SK keys are kept; to switch back, run 'ucloud auth logout' then 'ucloud init'\n", cfg.Profile)
	}

	base.ApplyTokenResponse(cfg, tr)
	cfg.Active = true
	if _, ok := base.AggConfigListIns.GetAggConfigByProfile(cfg.Profile); ok {
		err = base.AggConfigListIns.UpdateAggConfig(cfg)
	} else {
		err = base.AggConfigListIns.Append(cfg)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "save credential failed: %v\n", err)
		os.Exit(1)
	}

	// AP-2：首登补链——自动配置 region/zone/project（Bearer 调用，复用 init 逻辑）
	if cfg.Region == "" || cfg.Zone == "" {
		region, rerr := fetchRegionWithConfig(cfg)
		if rerr != nil {
			fmt.Printf("Warning: fetch default region failed (%v). Set it later: ucloud config update --profile %s --region <region> --zone <zone>\n", rerr, cfg.Profile)
		} else {
			cfg.Region = region.DefaultRegion
			cfg.Zone = region.DefaultZone
			fmt.Printf("Configured default region:%s zone:%s\n", cfg.Region, cfg.Zone)
		}
	}
	// 既有 project_id 也要用新账号的项目列表校验：跨账号/跨站点遗留的 project_id
	// 若原样保留，后续业务命令全部 RetCode 292 "Project not exists"
	if projects, perr := fetchProjectListWithConfig(cfg); perr != nil {
		fmt.Printf("Warning: fetch project list failed (%v). Set it later: ucloud config update --profile %s --project-id <id>\n", perr, cfg.Profile)
	} else if id, notice, rerr := resolveLoginProject(cfg.ProjectID, projects); rerr != nil {
		fmt.Printf("Warning: resolve default project failed (%v). Set it later: ucloud config update --profile %s --project-id <id>\n", rerr, cfg.Profile)
	} else {
		cfg.ProjectID = id
		if notice != "" {
			fmt.Println(notice)
		}
	}
	if err := base.AggConfigListIns.UpdateAggConfig(cfg); err != nil {
		fmt.Printf("Warning: saving default region/project failed (%v). Set them later: ucloud config update --profile %s --region <region> --zone <zone> --project-id <id>\n", err, cfg.Profile)
	}

	// ⑥ 输出 email + 过期时间（id_token 仅解析不落盘）
	until := time.Unix(cfg.ExpiresAt, 0).Format("15:04")
	if email, eerr := base.ParseIDTokenEmail(tr.IDToken); eerr == nil && email != "" {
		fmt.Printf("Logged in as %s, token valid until %s\n", email, until)
	} else {
		fmt.Printf("Logged in, token valid until %s\n", until)
	}
}

// resolveLoginProject 决定登录后 profile 应使用的 project（AP-2 的校验补丁）：
// existing 为空 → 选账号默认项目（首登补链）；existing 在列表内 → 保持不变，无提示；
// existing 不在列表内（跨账号/跨站点遗留）→ 切到默认项目并返回提示。
// 返回 (projectID, notice)；notice 非空时调用方原样打印。列表无默认项目时返回 errNoDefaultProject。
func resolveLoginProject(existing string, projects []uaccount.ProjectListInfo) (string, string, error) {
	var defaultID, defaultName string
	for _, p := range projects {
		if existing != "" && p.ProjectId == existing {
			return existing, "", nil
		}
		if p.IsDefault {
			defaultID, defaultName = p.ProjectId, p.ProjectName
		}
	}
	if defaultID == "" {
		return "", "", errNoDefaultProject
	}
	if existing == "" {
		return defaultID, fmt.Sprintf("Configured default project:%s %s", defaultID, defaultName), nil
	}
	notice := fmt.Sprintf("Existing project '%s' does not belong to this account; switching to default project '%s' %s", existing, defaultID, defaultName)
	return defaultID, notice, nil
}

// loginCallbackTimeout 自动捕获的等待上限；超时回退到手工粘贴
const loginCallbackTimeout = 3 * time.Minute

// runLoginManual --no-browser 手工模式：分配一个 >=1024 端口（仅取号，立即释放 listener），
// 打印 URL，从 stdin 读回调 URL。返回 (code, redirectURI)；出错时直接退出。
func runLoginManual(oauthBase, state string) (string, string) {
	ln, port, err := allocateLoopbackListener()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	ln.Close()
	redirectURI := base.BuildLoopbackRedirectURI(port)
	authorizeURL := base.BuildAuthorizeURL(oauthBase, redirectURI, state)

	fmt.Println("Logging in via browser (manual paste). 3 steps:")
	fmt.Println("  1. Open the URL below and finish login & authorization.")
	fmt.Println("  2. The browser will be redirected to a localhost page that CANNOT")
	fmt.Printf("     open (%s?...). THIS IS EXPECTED.\n", redirectURI)
	fmt.Println("  3. Copy the FULL URL from the address bar and paste it here.")
	fmt.Println()
	fmt.Printf("Open this URL in your browser:\n\n  %s\n\n", authorizeURL)

	code, err := readCallbackCode(state)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return code, redirectURI
}

// runLoginAuto 默认模式：起本地回调 server 自动捕获 code，超时回退手工粘贴。
// 返回 (code, redirectURI)；遇到主动错误（拒绝授权/state 不匹配）直接退出。
func runLoginAuto(oauthBase, state string) (string, string) {
	ln, port, err := allocateLoopbackListener()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	redirectURI := base.BuildLoopbackRedirectURI(port)
	authorizeURL := base.BuildAuthorizeURL(oauthBase, redirectURI, state)

	srv, ch := startCallbackServer(ln, state)

	fmt.Println("A browser window will open; finish the login there and return here — no copy-paste needed.")
	fmt.Printf("If it does not open, visit:\n\n  %s\n\n", authorizeURL)
	openbrowser(authorizeURL)

	select {
	case res := <-ch:
		srv.Close()
		if res.err != nil {
			fmt.Fprintln(os.Stderr, res.err)
			os.Exit(1)
		}
		return res.code, redirectURI
	case <-time.After(loginCallbackTimeout):
		srv.Close()
		fmt.Fprintln(os.Stderr, "Automatic capture timed out. Paste the callback URL here as a fallback:")
		code, err := readCallbackCode(state)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return code, redirectURI
	}
}

// readCallbackCode 读回调 URL：容忍折行（粘贴的多行一次到达时合并），允许重试 3 次
func readCallbackCode(state string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	for attempt := 1; attempt <= 3; attempt++ {
		fmt.Print("Paste the full callback URL here: ")
		raw, err := readWrappedLine(reader)
		if err != nil {
			return "", fmt.Errorf("read input failed: %v", err)
		}
		code, perr := base.ParseCallbackURL(raw, state)
		if perr == nil {
			return code, nil
		}
		fmt.Fprintln(os.Stderr, perr)
	}
	return "", fmt.Errorf("too many invalid inputs. Run 'ucloud auth login' again")
}

// readWrappedLine 读一行；若粘贴内容因终端折行带来多行（缓冲区中仍有数据），继续读完合并
func readWrappedLine(r *bufio.Reader) (string, error) {
	line, err := r.ReadString('\n')
	if err != nil && line == "" {
		return "", err
	}
	for r.Buffered() > 0 {
		next, nerr := r.ReadString('\n')
		line += next
		if nerr != nil {
			break
		}
	}
	return line, nil
}

// NewCmdLogout ucloud auth logout
func NewCmdLogout() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "logout",
		Short:   "Log out: remove local OAuth tokens of the current profile",
		Long:    "Log out: remove local OAuth tokens (access_token/refresh_token) of the current profile from ~/.ucloud/credential.json",
		Args:    cobra.NoArgs,
		Example: "ucloud auth logout",
		Run: func(cmd *cobra.Command, args []string) {
			cfg := base.ConfigIns
			if cfg.AuthMode != base.AuthModeOAuth && cfg.AccessToken == "" {
				fmt.Printf("Profile '%s' is not logged in via OAuth, nothing to do\n", cfg.Profile)
				return
			}
			clearOAuthState(cfg)
			if err := base.AggConfigListIns.UpdateAggConfig(cfg); err != nil {
				base.HandleError(err)
				return
			}
			// AP-4：不加服务端有效期提示（用户裁定，spec 风险 #5）
			fmt.Printf("Logged out: local tokens of profile '%s' removed\n", cfg.Profile)
		},
	}
	cmd.SetHelpTemplate(oauthHelpTmpl)
	return cmd
}
