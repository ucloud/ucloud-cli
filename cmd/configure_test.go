package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/cmd/internal/platform"
)

// 回归：oauth profile 已存 AK/SK 时（auth login 保留密钥的常见形态），
// init 确认切回 AK/SK 后必须把 auth_mode/token 清除并落盘，否则下次启动仍走 OAuth
func TestSwitchProfileToAKSKPersistsToDisk(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")
	credPath := filepath.Join(dir, "credential.json")
	cliJSON := `[{"profile":"oa","active":true,"region":"cn-bj2","zone":"cn-bj2-04","base_url":"https://api.ucloud.cn/","timeout_sec":15,"max_retry_times":3}]`
	credJSON := `[{"public_key":"pub","private_key":"pri","profile":"oa","auth_mode":"oauth","access_token":"at","refresh_token":"rt","expires_at":1234567890}]`
	if err := ioutil.WriteFile(cfgPath, []byte(cliJSON), platform.LocalFileMode); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(credPath, []byte(credJSON), platform.LocalFileMode); err != nil {
		t.Fatal(err)
	}

	m, err := platform.NewAggConfigManager(cfgPath, credPath)
	if err != nil {
		t.Fatal(err)
	}
	cfg, ok := m.GetAggConfigByProfile("oa")
	if !ok {
		t.Fatal("profile oa missing")
	}

	oldM, oldC := platform.AggConfigListIns, platform.ConfigIns
	platform.AggConfigListIns, platform.ConfigIns = m, cfg
	defer func() { platform.AggConfigListIns, platform.ConfigIns = oldM, oldC }()

	if err := switchProfileToAKSK(cfg); err != nil {
		t.Fatal(err)
	}

	// 重新读盘验证持久化，而非只看内存
	m2, err := platform.NewAggConfigManager(cfgPath, credPath)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := m2.GetAggConfigByProfile("oa")
	if !ok {
		t.Fatal("profile oa missing after reload")
	}
	if got.AuthMode != "" || got.AccessToken != "" || got.RefreshToken != "" || got.ExpiresAt != 0 {
		t.Errorf("oauth state must be cleared on disk, got auth_mode=%q access_token=%q refresh_token=%q expires_at=%d",
			got.AuthMode, got.AccessToken, got.RefreshToken, got.ExpiresAt)
	}
	if got.PublicKey != "pub" || got.PrivateKey != "pri" {
		t.Errorf("AK/SK must survive the switch, got public_key=%q private_key=%q", got.PublicKey, got.PrivateKey)
	}
}

// 假网关响应体。网关以 HTTP 200 + RetCode 表达错误，假网关须照做。
const (
	respRegionOK         = `{"RetCode":0,"Action":"GetRegionResponse","Regions":[{"Region":"cn-bj2","Zone":"cn-bj2-04","IsDefault":true}]}`
	respProjectOK        = `{"RetCode":0,"Action":"GetProjectListResponse","ProjectSet":[{"ProjectId":"org-123","ProjectName":"Default","IsDefault":true}]}`
	respSignatureFail    = `{"RetCode":171,"Message":"Signature VerifyAC Error"}`
	respProjectNoDefault = `{"RetCode":0,"Action":"GetProjectListResponse","ProjectSet":[{"ProjectId":"org-123","ProjectName":"P","IsDefault":false}]}`
)

// gatewayBehavior 指定假网关对各 action 的响应，零值即全部成功。
// 分 action 控制使「region 校验通过但 project 校验失败」这类组合可精确构造。
type gatewayBehavior struct {
	regionResp  string // 空 = respRegionOK
	projectResp string // 空 = respProjectOK
}

// poisonGateway 任何请求都判失败——用于断言某路径根本不该发起远程校验
// （本地检查应先于远程校验拦截，不打网络）。
func poisonGateway(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("unexpected remote call %s %s — this path must be rejected locally before any validation", r.Method, r.URL.RawQuery)
		w.WriteHeader(500)
	}))
}

// fakeGatewayServer 模拟业务网关：响应远程校验所需的 GetRegion/GetProjectList
func fakeGatewayServer(t *testing.T) *httptest.Server {
	t.Helper()
	return fakeGatewayServerWith(t, gatewayBehavior{})
}

// fakeGatewayServerWith 同 fakeGatewayServer，但可为单个 action 指定失败响应
func fakeGatewayServerWith(t *testing.T, b gatewayBehavior) *httptest.Server {
	t.Helper()
	if b.regionResp == "" {
		b.regionResp = respRegionOK
	}
	if b.projectResp == "" {
		b.projectResp = respProjectOK
	}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		payload := r.URL.RawQuery + string(body)
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(payload, "GetRegion"):
			fmt.Fprint(w, b.regionResp)
		case strings.Contains(payload, "GetProjectList"):
			fmt.Fprint(w, b.projectResp)
		default:
			fmt.Fprint(w, `{"RetCode":230,"Message":"unexpected action"}`)
		}
	}))
}

// 回归：config update --base-url 必须在远程校验（getReasonableRegionZone 等）之前生效，
// 否则旧 base_url 指向坏网关时校验永远打到坏网关，新地址无法保存（鸡生蛋死锁）。
func TestConfigUpdateAppliesBaseURLBeforeValidation(t *testing.T) {
	gateway := fakeGatewayServer(t)
	defer gateway.Close()

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")
	credPath := filepath.Join(dir, "credential.json")
	// 存量 base_url 指向必然连不通的地址，复现坏网关现场
	cliJSON := `[{"profile":"up","active":true,"project_id":"org-123","region":"cn-bj2","zone":"cn-bj2-04","base_url":"http://127.0.0.1:1/","timeout_sec":3,"max_retry_times":0}]`
	credJSON := `[{"public_key":"pub","private_key":"pri","profile":"up"}]`
	if err := ioutil.WriteFile(cfgPath, []byte(cliJSON), platform.LocalFileMode); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(credPath, []byte(credJSON), platform.LocalFileMode); err != nil {
		t.Fatal(err)
	}

	m, err := platform.NewAggConfigManager(cfgPath, credPath)
	if err != nil {
		t.Fatal(err)
	}

	// GetBizClient 会改写包级全局 ClientConfig/AuthCredential，恢复现场避免测试顺序耦合
	oldM, oldCC, oldAC := platform.AggConfigListIns, platform.ClientConfig, platform.AuthCredential
	platform.AggConfigListIns = m
	defer func() {
		platform.AggConfigListIns, platform.ClientConfig, platform.AuthCredential = oldM, oldCC, oldAC
	}()

	cmd := NewCmdConfigUpdate()
	if err := cmd.Flags().Set("profile", "up"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("base-url", gateway.URL); err != nil {
		t.Fatal(err)
	}
	cmd.Run(cmd, nil)

	// 重新读盘验证持久化，而非只看内存
	m2, err := platform.NewAggConfigManager(cfgPath, credPath)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := m2.GetAggConfigByProfile("up")
	if !ok {
		t.Fatal("profile up missing after reload")
	}
	if got.BaseURL != gateway.URL {
		t.Errorf("base_url on disk = %q, want new gateway %q (remote validation must run against the NEW base-url)", got.BaseURL, gateway.URL)
	}
}

// 回归：OAuth-only profile（auth_mode=oauth 且未存 AK/SK，auth login 直接创建的形态）
// 执行 init 确认切回 AK/SK 并走完整配置流程后，末尾持久化不能因 profile 已存在而失败，
// 否则整套新配置（密钥、region、project）全部不落盘
func TestInitSaveOverwritesExistingOAuthOnlyProfile(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")
	credPath := filepath.Join(dir, "credential.json")
	cliJSON := `[{"profile":"oa","active":true,"base_url":"https://api.ucloud.cn/","timeout_sec":15,"max_retry_times":3}]`
	credJSON := `[{"public_key":"","private_key":"","profile":"oa","auth_mode":"oauth","access_token":"at","refresh_token":"rt","expires_at":1234567890}]`
	if err := ioutil.WriteFile(cfgPath, []byte(cliJSON), platform.LocalFileMode); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(credPath, []byte(credJSON), platform.LocalFileMode); err != nil {
		t.Fatal(err)
	}

	m, err := platform.NewAggConfigManager(cfgPath, credPath)
	if err != nil {
		t.Fatal(err)
	}
	cfg, ok := m.GetAggConfigByProfile("oa")
	if !ok {
		t.Fatal("profile oa missing")
	}

	oldM, oldC := platform.AggConfigListIns, platform.ConfigIns
	platform.AggConfigListIns, platform.ConfigIns = m, cfg
	defer func() { platform.AggConfigListIns, platform.ConfigIns = oldM, oldC }()

	// 模拟 NewCmdInit Run 完整配置路径对 ConfigIns（即 manager map 内同一指针）的写入
	clearOAuthState(cfg)
	cfg.PublicKey = "newpub"
	cfg.PrivateKey = "newpri"
	cfg.Region = "cn-bj2"
	cfg.Zone = "cn-bj2-04"
	cfg.ProjectID = "org-new"
	cfg.Timeout = platform.DefaultTimeoutSec
	cfg.BaseURL = platform.DefaultBaseURL
	cfg.Active = true

	if err := saveInitProfile(cfg); err != nil {
		t.Fatalf("save must overwrite existing profile instead of failing, got: %v", err)
	}

	// 重新读盘验证持久化，而非只看内存
	m2, err := platform.NewAggConfigManager(cfgPath, credPath)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := m2.GetAggConfigByProfile("oa")
	if !ok {
		t.Fatal("profile oa missing after reload")
	}
	if got.PublicKey != "newpub" || got.PrivateKey != "newpri" {
		t.Errorf("new AK/SK must land on disk, got public_key=%q private_key=%q", got.PublicKey, got.PrivateKey)
	}
	if got.Region != "cn-bj2" || got.Zone != "cn-bj2-04" || got.ProjectID != "org-new" {
		t.Errorf("region/zone/project must land on disk, got region=%q zone=%q project_id=%q", got.Region, got.Zone, got.ProjectID)
	}
	if got.AuthMode != "" {
		t.Errorf("auth_mode must be cleared on disk, got %q", got.AuthMode)
	}
	// 切回 AK/SK 后 token 必须清除，口径与 switchProfileToAKSK / 'ucloud auth logout' 一致
	if got.AccessToken != "" || got.RefreshToken != "" || got.ExpiresAt != 0 {
		t.Errorf("oauth tokens must be cleared on disk, got access_token=%q refresh_token=%q expires_at=%d",
			got.AccessToken, got.RefreshToken, got.ExpiresAt)
	}
}

// newTestConfigFiles 写入 config/credential 到临时目录，让包级 AggConfigListIns 指向它们，
// 并在 t.Cleanup 中恢复被 GetBizClient 改写的包级全局，避免测试顺序耦合。
func newTestConfigFiles(t *testing.T, cliJSON, credJSON string) (cfgPath, credPath string) {
	t.Helper()
	dir := t.TempDir()
	cfgPath = filepath.Join(dir, "config.json")
	credPath = filepath.Join(dir, "credential.json")
	if err := ioutil.WriteFile(cfgPath, []byte(cliJSON), platform.LocalFileMode); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(credPath, []byte(credJSON), platform.LocalFileMode); err != nil {
		t.Fatal(err)
	}
	m, err := platform.NewAggConfigManager(cfgPath, credPath)
	if err != nil {
		t.Fatal(err)
	}
	oldM, oldCC, oldAC := platform.AggConfigListIns, platform.ClientConfig, platform.AuthCredential
	platform.AggConfigListIns = m
	t.Cleanup(func() {
		platform.AggConfigListIns, platform.ClientConfig, platform.AuthCredential = oldM, oldCC, oldAC
	})
	return cfgPath, credPath
}

// reloadProfile 重新读盘取 profile —— 断言持久化结果，而非内存态
func reloadProfile(t *testing.T, cfgPath, credPath, profile string) (*platform.AggConfig, bool) {
	t.Helper()
	m, err := platform.NewAggConfigManager(cfgPath, credPath)
	if err != nil {
		t.Fatal(err)
	}
	return m.GetAggConfigByProfile(profile)
}

func setFlags(t *testing.T, cmd *cobra.Command, kv ...string) {
	t.Helper()
	for i := 0; i+1 < len(kv); i += 2 {
		if err := cmd.Flags().Set(kv[i], kv[i+1]); err != nil {
			t.Fatalf("set --%s=%s: %v", kv[i], kv[i+1], err)
		}
	}
}

// 回归 AC1：config add 远程校验失败时不得创建 profile。
// 现状 getReasonableRegionZone 出错后只 HandleError 不 return，随即把空 region/zone
// 赋回，照常 Append，落盘一个 region=” zone=” project_id=” 的残缺 profile。
//
// 预置一个 active profile 而非从空配置起步：AggConfigManager.Load 规定「有 profile
// 就必须有 active」（config.go:403），否则这里落盘的 bad(active=false) 会让重新读盘
// 直接失败，断言根本跑不到。
func TestConfigAddRejectsProfileWhenValidationFails(t *testing.T) {
	// 失败路径必调 HandleError → LogErrorTo 会碰单测中恒为 nil 的包级 logger
	t.Setenv("COMP_LINE", "1")
	gateway := fakeGatewayServerWith(t, gatewayBehavior{regionResp: respSignatureFail})
	t.Cleanup(gateway.Close)

	cliJSON := `[{"profile":"good","active":true,"project_id":"org-123","region":"cn-bj2","zone":"cn-bj2-04","base_url":"https://api.ucloud.cn/","timeout_sec":15,"max_retry_times":3}]`
	credJSON := `[{"public_key":"pub","private_key":"pri","profile":"good"}]`
	cfgPath, credPath := newTestConfigFiles(t, cliJSON, credJSON)

	cmd := NewCmdConfigAdd()
	setFlags(t, cmd,
		"profile", "bad",
		"public-key", "FAKE",
		"private-key", "FAKE",
		"base-url", gateway.URL,
		"region", "cn-bj2",
		"zone", "cn-bj2-04",
	)
	cmd.Run(cmd, nil)

	if got, ok := reloadProfile(t, cfgPath, credPath, "bad"); ok {
		t.Errorf("profile must not be created when remote validation fails; got region=%q zone=%q project_id=%q",
			got.Region, got.Zone, got.ProjectID)
	}
}

// 回归 AC2：失败的 config add --active true 不得夺走原有 active profile。
// 现状坏 profile 不仅建成，Append 还会把原 active 踢下去，此后每条命令都用
// 那个 region 为空、密钥为假的 profile，且无任何线索指向病因。
func TestConfigAddFailureDoesNotHijackActiveProfile(t *testing.T) {
	t.Setenv("COMP_LINE", "1")
	gateway := fakeGatewayServerWith(t, gatewayBehavior{regionResp: respSignatureFail})
	t.Cleanup(gateway.Close)

	cliJSON := `[{"profile":"good","active":true,"project_id":"org-123","region":"cn-bj2","zone":"cn-bj2-04","base_url":"https://api.ucloud.cn/","timeout_sec":15,"max_retry_times":3}]`
	credJSON := `[{"public_key":"pub","private_key":"pri","profile":"good"}]`
	cfgPath, credPath := newTestConfigFiles(t, cliJSON, credJSON)

	cmd := NewCmdConfigAdd()
	setFlags(t, cmd,
		"profile", "bad",
		"public-key", "FAKE",
		"private-key", "FAKE",
		"base-url", gateway.URL,
		"active", "true",
	)
	cmd.Run(cmd, nil)

	good, ok := reloadProfile(t, cfgPath, credPath, "good")
	if !ok {
		t.Fatal("profile good missing after reload")
	}
	if !good.Active {
		t.Error("a failed 'config add --active true' must not deactivate the existing active profile")
	}
	if bad, ok := reloadProfile(t, cfgPath, credPath, "bad"); ok {
		t.Errorf("failed 'config add' must not create the profile; got active=%v region=%q", bad.Active, bad.Region)
	}
}

// 回归 AC3：config update 远程校验失败时不得写入任何字段，含 AK/SK。
// 现状「先更新让其生效」的提前 UpdateAggConfig 已把新 AK/SK 落盘，其后 region
// 校验失败虽 return，profile 已停在「密钥已换、region/project 未换」的半更新态。
func TestConfigUpdateWritesNothingWhenValidationFails(t *testing.T) {
	t.Setenv("COMP_LINE", "1")
	gateway := fakeGatewayServerWith(t, gatewayBehavior{regionResp: respSignatureFail})
	t.Cleanup(gateway.Close)

	cliJSON := fmt.Sprintf(`[{"profile":"up","active":true,"project_id":"org-123","region":"cn-bj2","zone":"cn-bj2-04","base_url":%q,"timeout_sec":15,"max_retry_times":3}]`, gateway.URL)
	credJSON := `[{"public_key":"OLD_PUB","private_key":"OLD_PRI","profile":"up"}]`
	cfgPath, credPath := newTestConfigFiles(t, cliJSON, credJSON)

	cmd := NewCmdConfigUpdate()
	setFlags(t, cmd,
		"profile", "up",
		"public-key", "NEW_PUB",
		"private-key", "NEW_PRI",
	)
	cmd.Run(cmd, nil)

	got, ok := reloadProfile(t, cfgPath, credPath, "up")
	if !ok {
		t.Fatal("profile up missing after reload")
	}
	if got.PublicKey != "OLD_PUB" || got.PrivateKey != "OLD_PRI" {
		t.Errorf("a failed 'config update' must write nothing; got public_key=%q private_key=%q, want OLD_PUB/OLD_PRI",
			got.PublicKey, got.PrivateKey)
	}
}

// 回归 AC4（config-cmd-audit）：config update 的 project 校验失败不得清空 ProjectID
// （此前同函数内 region 记得 return、project 忘了 return，把 "" 赋回并落盘）。
// 本任务改为「按需校验」后，仅传 --profile 不再触发校验；凭据变更(--public-key)会触发
// 对存量 region+project 的校验。网关 GetRegion 返成功、GetProjectList 返失败，精确
// 构造「project 校验触发且失败」：fail-closed → 不落盘 → 存量 org-123 不被清空。
func TestConfigUpdateKeepsProjectIDWhenProjectValidationFails(t *testing.T) {
	t.Setenv("COMP_LINE", "1")
	gateway := fakeGatewayServerWith(t, gatewayBehavior{projectResp: respSignatureFail})
	t.Cleanup(gateway.Close)

	cliJSON := fmt.Sprintf(`[{"profile":"up","active":true,"project_id":"org-123","region":"cn-bj2","zone":"cn-bj2-04","base_url":%q,"timeout_sec":15,"max_retry_times":3}]`, gateway.URL)
	credJSON := `[{"public_key":"pub","private_key":"pri","profile":"up"}]`
	cfgPath, credPath := newTestConfigFiles(t, cliJSON, credJSON)

	cmd := NewCmdConfigUpdate()
	setFlags(t, cmd, "profile", "up", "public-key", "NEWKEY")
	cmd.Run(cmd, nil)

	got, ok := reloadProfile(t, cfgPath, credPath, "up")
	if !ok {
		t.Fatal("profile up missing after reload")
	}
	if got.ProjectID != "org-123" {
		t.Errorf("project validation failure must not clear project_id; got %q, want org-123", got.ProjectID)
	}
}

// 回归 AC6（config-cmd-audit）：config 主命令改存量、确实发起的校验失败时不得落盘。
// 本任务改为「按需校验」后，纯元数据编辑不再触发校验，故此处用 --region 触发校验并
// bundle 一个 --timeout-sec：校验失败时，同一条命令里的元数据改动也必须一并不落盘
// （fail-closed 不因本任务回退，prd R4）。
func TestConfigMainCommandWritesNothingWhenValidationFails(t *testing.T) {
	t.Setenv("COMP_LINE", "1")
	gateway := fakeGatewayServerWith(t, gatewayBehavior{regionResp: respSignatureFail})
	t.Cleanup(gateway.Close)

	cliJSON := fmt.Sprintf(`[{"profile":"main","active":true,"project_id":"org-123","region":"cn-bj2","zone":"cn-bj2-04","base_url":%q,"timeout_sec":15,"max_retry_times":3}]`, gateway.URL)
	credJSON := `[{"public_key":"pub","private_key":"pri","profile":"main"}]`
	cfgPath, credPath := newTestConfigFiles(t, cliJSON, credJSON)

	cmd := NewCmdConfig()
	setFlags(t, cmd, "profile", "main", "region", "cn-sh2", "timeout-sec", "30")
	cmd.Run(cmd, nil)

	got, ok := reloadProfile(t, cfgPath, credPath, "main")
	if !ok {
		t.Fatal("profile main missing after reload")
	}
	if got.Timeout != 15 {
		t.Errorf("a failed 'ucloud config' must write nothing; got timeout_sec=%d, want 15 (unchanged)", got.Timeout)
	}
	if got.Region != "cn-bj2" {
		t.Errorf("a failed region validation must not persist the new region; got %q, want cn-bj2 (unchanged)", got.Region)
	}
}

// 回归 D10：config add 的本地 timeout 检查必须先于远程校验。
// timeout <= 0 是纯本地约束；此前它排在 getReasonableRegionZone 之后，坏 timeout
// 会先被喂进校验请求（0 → 无超时 → 照打网关；负值 → 请求瞬败），坏 timeout 反而
// 让人看到 region 错、看不到真正的 timeout 错。用 --timeout-sec 0 + 毒网关：修复后
// 本地检查先拦，网关零调用。
func TestConfigAddChecksTimeoutBeforeValidation(t *testing.T) {
	t.Setenv("COMP_LINE", "1")
	gateway := poisonGateway(t) // 任何远程调用都判失败
	t.Cleanup(gateway.Close)

	cliJSON := `[{"profile":"good","active":true,"project_id":"org-123","region":"cn-bj2","zone":"cn-bj2-04","base_url":"https://api.ucloud.cn/","timeout_sec":15,"max_retry_times":3}]`
	credJSON := `[{"public_key":"pub","private_key":"pri","profile":"good"}]`
	cfgPath, credPath := newTestConfigFiles(t, cliJSON, credJSON)

	cmd := NewCmdConfigAdd()
	setFlags(t, cmd,
		"profile", "bt",
		"public-key", "pub",
		"private-key", "pri",
		"base-url", gateway.URL,
		"region", "cn-bj2",
		"zone", "cn-bj2-04",
		"timeout-sec", "0", // 非法本地值：0 会让 http.Client 变成「无超时」，照打网关
	)
	cmd.Run(cmd, nil)

	// timeout=0 非法，profile 不该建成；关键由毒网关断言：拒绝发生在任何远程调用之前
	if _, ok := reloadProfile(t, cfgPath, credPath, "bt"); ok {
		t.Error("profile 'bt' must not be created with timeout_sec=0")
	}
}

// 回归 D9：config add 必须归一化补全格式的 project-id（org-xxx/Name）。
// getProjectList 补全吐出的正是斜杠形式，而 config add 此前不做 PickResourceID
// （主命令与 update 都做）——合法的补全值会被 getReasonableProject 判为「不存在」，
// 叠加 fail-closed 后直接硬拦一个真实存在的 project。
func TestConfigAddNormalizesSlashProjectID(t *testing.T) {
	t.Setenv("COMP_LINE", "1")
	// respProjectOK 的 ProjectSet 含 org-123（map 键为裸 id）
	gateway := fakeGatewayServer(t)
	t.Cleanup(gateway.Close)

	cliJSON := `[{"profile":"good","active":true,"project_id":"org-123","region":"cn-bj2","zone":"cn-bj2-04","base_url":"https://api.ucloud.cn/","timeout_sec":15,"max_retry_times":3}]`
	credJSON := `[{"public_key":"pub","private_key":"pri","profile":"good"}]`
	cfgPath, credPath := newTestConfigFiles(t, cliJSON, credJSON)

	cmd := NewCmdConfigAdd()
	setFlags(t, cmd,
		"profile", "probe",
		"public-key", "pub",
		"private-key", "pri",
		"base-url", gateway.URL,
		"region", "cn-bj2",
		"zone", "cn-bj2-04",
		"project-id", "org-123/Default", // 补全格式，project 真实存在
	)
	cmd.Run(cmd, nil)

	got, ok := reloadProfile(t, cfgPath, credPath, "probe")
	if !ok {
		t.Fatal("a valid completion-form project-id must not block the save; profile probe was not created")
	}
	if got.ProjectID != "org-123" {
		t.Errorf("project-id must be normalized to the bare id; got %q, want org-123", got.ProjectID)
	}
}

// 回归 AC9：fail-closed 不得误伤「有项目但未设默认」的账号。
// 该场景下 getDefaultProjectWithConfig 返回 errNoDefaultProject（良性），init 认得
// 并放行；config add 必须同口径 —— 建成 profile 且 project_id 留空。
// 本用例守护修复不越界，故在修复前即应通过。
func TestConfigAddAllowsAccountWithoutDefaultProject(t *testing.T) {
	t.Setenv("COMP_LINE", "1")
	gateway := fakeGatewayServerWith(t, gatewayBehavior{projectResp: respProjectNoDefault})
	t.Cleanup(gateway.Close)

	// 同 AC1：预置 active profile，否则新建的 nodef(active=false) 会让重新读盘失败
	cliJSON := `[{"profile":"good","active":true,"project_id":"org-123","region":"cn-bj2","zone":"cn-bj2-04","base_url":"https://api.ucloud.cn/","timeout_sec":15,"max_retry_times":3}]`
	credJSON := `[{"public_key":"pub","private_key":"pri","profile":"good"}]`
	cfgPath, credPath := newTestConfigFiles(t, cliJSON, credJSON)

	cmd := NewCmdConfigAdd()
	setFlags(t, cmd,
		"profile", "nodef",
		"public-key", "pub",
		"private-key", "pri",
		"base-url", gateway.URL,
		"region", "cn-bj2",
		"zone", "cn-bj2-04",
	)
	cmd.Run(cmd, nil)

	got, ok := reloadProfile(t, cfgPath, credPath, "nodef")
	if !ok {
		t.Fatal("an account without a default project must still be configurable; profile nodef was not created")
	}
	if got.ProjectID != "" {
		t.Errorf("project_id should stay empty when the account has no default project, got %q", got.ProjectID)
	}
	if got.Region != "cn-bj2" || got.Zone != "cn-bj2-04" {
		t.Errorf("region/zone must survive; got region=%q zone=%q", got.Region, got.Zone)
	}
}

// AC1（复现→守卫后转绿）：config update 只改元数据（--active）时跳过远程校验，离线可改。
// 存量 base_url 指向毒网关：一旦发起校验就会打到它 → t.Errorf 判红。守卫前无条件校验必红
// （复现成立），守卫后跳过校验、零远程调用、落盘成功转绿。
// 需第二个 active profile "keep" 作锚点：up 由 active→inactive 后仍须有 active profile，
// 否则重新读盘因「no active config found」失败（config.go:403）。
func TestConfigUpdateSkipsValidationForMetadataOnly(t *testing.T) {
	t.Setenv("COMP_LINE", "1")
	gateway := poisonGateway(t)
	t.Cleanup(gateway.Close)

	cliJSON := fmt.Sprintf(`[{"profile":"keep","active":true,"project_id":"org-123","region":"cn-bj2","zone":"cn-bj2-04","base_url":"https://api.ucloud.cn/","timeout_sec":15,"max_retry_times":3},{"profile":"up","active":true,"project_id":"org-123","region":"cn-bj2","zone":"cn-bj2-04","base_url":%q,"timeout_sec":15,"max_retry_times":3}]`, gateway.URL)
	credJSON := `[{"public_key":"pub","private_key":"pri","profile":"keep"},{"public_key":"pub","private_key":"pri","profile":"up"}]`
	cfgPath, credPath := newTestConfigFiles(t, cliJSON, credJSON)

	cmd := NewCmdConfigUpdate()
	setFlags(t, cmd, "profile", "up", "active", "false")
	cmd.Run(cmd, nil)

	got, ok := reloadProfile(t, cfgPath, credPath, "up")
	if !ok {
		t.Fatal("profile up missing after reload")
	}
	if got.Active {
		t.Error("only-metadata 'config update --active false' must persist offline without any remote call; got active=true")
	}
}

// AC2（复现→守卫后转绿）：config update 只改 --timeout-sec（元数据）时跳过远程校验，
// 即便 base_url 指向坏地址也能离线改成功。守卫前无条件校验会打坏网关 → 失败 → 硬拦、
// timeout 改不了（复现红）；守卫后跳过校验、落盘成功转绿。单 active profile 全程不变，
// 重新读盘无「no active config found」之虞。
func TestConfigUpdateOfflineTimeoutChange(t *testing.T) {
	t.Setenv("COMP_LINE", "1")
	// 存量 base_url 指向必然连不通的地址：一旦发起校验必失败
	cliJSON := `[{"profile":"up","active":true,"project_id":"org-123","region":"cn-bj2","zone":"cn-bj2-04","base_url":"http://127.0.0.1:1/","timeout_sec":15,"max_retry_times":3}]`
	credJSON := `[{"public_key":"pub","private_key":"pri","profile":"up"}]`
	cfgPath, credPath := newTestConfigFiles(t, cliJSON, credJSON)

	cmd := NewCmdConfigUpdate()
	setFlags(t, cmd, "profile", "up", "timeout-sec", "30")
	cmd.Run(cmd, nil)

	got, ok := reloadProfile(t, cfgPath, credPath, "up")
	if !ok {
		t.Fatal("profile up missing after reload")
	}
	if got.Timeout != 30 {
		t.Errorf("only-metadata 'config update --timeout-sec 30' must persist offline; got timeout_sec=%d, want 30", got.Timeout)
	}
}

// AC5（复现→守卫后转绿）：主命令 config 改存量 profile 且只改元数据（--active）时跳过校验。
// p1 已存在 → ok==true → validateRegion/Project 皆 false → 毒网关零调用。keep 作 active
// 锚点，使 p1 由 inactive→active 为可见变更，且切换后仍有 active profile 供重新读盘。
func TestConfigMainSkipsValidationForExistingMetadataOnly(t *testing.T) {
	t.Setenv("COMP_LINE", "1")
	gateway := poisonGateway(t)
	t.Cleanup(gateway.Close)

	cliJSON := fmt.Sprintf(`[{"profile":"keep","active":true,"project_id":"org-123","region":"cn-bj2","zone":"cn-bj2-04","base_url":"https://api.ucloud.cn/","timeout_sec":15,"max_retry_times":3},{"profile":"p1","active":false,"project_id":"org-123","region":"cn-bj2","zone":"cn-bj2-04","base_url":%q,"timeout_sec":15,"max_retry_times":3}]`, gateway.URL)
	credJSON := `[{"public_key":"pub","private_key":"pri","profile":"keep"},{"public_key":"pub","private_key":"pri","profile":"p1"}]`
	cfgPath, credPath := newTestConfigFiles(t, cliJSON, credJSON)

	cmd := NewCmdConfig()
	setFlags(t, cmd, "profile", "p1", "active", "true")
	cmd.Run(cmd, nil)

	got, ok := reloadProfile(t, cfgPath, credPath, "p1")
	if !ok {
		t.Fatal("profile p1 missing after reload")
	}
	if !got.Active {
		t.Error("main 'config --active true' on an existing profile must persist offline without validation; got active=false")
	}
}

// AC3（反向，守卫前后都应绿）：config update 传了 --region 时仍必须校验——防止把该校验的
// 也跳了。网关 GetRegion 返 171，传 --region cn-sh2：校验触发且失败 → fail-closed 不落盘，
// 存量 region 不变。
func TestConfigUpdateStillValidatesOnRegionChange(t *testing.T) {
	t.Setenv("COMP_LINE", "1")
	gateway := fakeGatewayServerWith(t, gatewayBehavior{regionResp: respSignatureFail})
	t.Cleanup(gateway.Close)

	cliJSON := fmt.Sprintf(`[{"profile":"up","active":true,"project_id":"org-123","region":"cn-bj2","zone":"cn-bj2-04","base_url":%q,"timeout_sec":15,"max_retry_times":3}]`, gateway.URL)
	credJSON := `[{"public_key":"pub","private_key":"pri","profile":"up"}]`
	cfgPath, credPath := newTestConfigFiles(t, cliJSON, credJSON)

	cmd := NewCmdConfigUpdate()
	setFlags(t, cmd, "profile", "up", "region", "cn-sh2")
	cmd.Run(cmd, nil)

	got, ok := reloadProfile(t, cfgPath, credPath, "up")
	if !ok {
		t.Fatal("profile up missing after reload")
	}
	if got.Region != "cn-bj2" {
		t.Errorf("a --region change must still be validated; a failed validation must not persist. got region=%q, want cn-bj2 (unchanged)", got.Region)
	}
}

// AC4（反向，最关键，守卫前后都应绿）：config update 改凭据（--public-key）必须同时触发
// region 与 project 两个校验。网关 GetRegion 返成功、GetProjectList 返 171：守卫正确时
// 凭据变更会触发 project 校验并失败 → fail-closed 不落盘，新 public-key 不落盘。
// 若守卫漏了「凭据变更 → 校验 project」，project 不被校验 → 命令直接落盘新 key → 断言判红。
func TestConfigUpdateValidatesBothOnCredChange(t *testing.T) {
	t.Setenv("COMP_LINE", "1")
	gateway := fakeGatewayServerWith(t, gatewayBehavior{projectResp: respSignatureFail})
	t.Cleanup(gateway.Close)

	cliJSON := fmt.Sprintf(`[{"profile":"up","active":true,"project_id":"org-123","region":"cn-bj2","zone":"cn-bj2-04","base_url":%q,"timeout_sec":15,"max_retry_times":3}]`, gateway.URL)
	credJSON := `[{"public_key":"OLD_PUB","private_key":"pri","profile":"up"}]`
	cfgPath, credPath := newTestConfigFiles(t, cliJSON, credJSON)

	cmd := NewCmdConfigUpdate()
	setFlags(t, cmd, "profile", "up", "public-key", "NEW_PUB")
	cmd.Run(cmd, nil)

	got, ok := reloadProfile(t, cfgPath, credPath, "up")
	if !ok {
		t.Fatal("profile up missing after reload")
	}
	if got.PublicKey != "OLD_PUB" {
		t.Errorf("a credential change must trigger project validation; a failed project validation must not persist. got public_key=%q, want OLD_PUB (unchanged)", got.PublicKey)
	}
}

// AC5 补充（反向，守卫前后都应绿）：主命令 config 新建 profile（!ok）时必须无条件校验。
// 新 profile 的 region/project 须从零建立，不得因「无 Changed 标志」而跳过。
// 网关 GetRegion 返 171 → 新建被拒、不落盘。
func TestConfigMainStillValidatesNewProfile(t *testing.T) {
	t.Setenv("COMP_LINE", "1")
	gateway := fakeGatewayServerWith(t, gatewayBehavior{regionResp: respSignatureFail})
	t.Cleanup(gateway.Close)

	cliJSON := `[{"profile":"keep","active":true,"project_id":"org-123","region":"cn-bj2","zone":"cn-bj2-04","base_url":"https://api.ucloud.cn/","timeout_sec":15,"max_retry_times":3}]`
	credJSON := `[{"public_key":"pub","private_key":"pri","profile":"keep"}]`
	cfgPath, credPath := newTestConfigFiles(t, cliJSON, credJSON)

	cmd := NewCmdConfig()
	setFlags(t, cmd,
		"profile", "fresh",
		"public-key", "pub",
		"private-key", "pri",
		"base-url", gateway.URL,
		"region", "cn-bj2",
		"zone", "cn-bj2-04",
	)
	cmd.Run(cmd, nil)

	if got, ok := reloadProfile(t, cfgPath, credPath, "fresh"); ok {
		t.Errorf("a new profile must be validated unconditionally; a failed validation must not create it. got region=%q", got.Region)
	}
}
