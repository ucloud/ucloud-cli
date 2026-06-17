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

	"github.com/ucloud/ucloud-cli/base"
)

// 回归：oauth profile 已存 AK/SK 时（auth login 保留密钥的常见形态），
// init 确认切回 AK/SK 后必须把 auth_mode/token 清除并落盘，否则下次启动仍走 OAuth
func TestSwitchProfileToAKSKPersistsToDisk(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")
	credPath := filepath.Join(dir, "credential.json")
	cliJSON := `[{"profile":"oa","active":true,"region":"cn-bj2","zone":"cn-bj2-04","base_url":"https://api.ucloud.cn/","timeout_sec":15,"max_retry_times":3}]`
	credJSON := `[{"public_key":"pub","private_key":"pri","profile":"oa","auth_mode":"oauth","access_token":"at","refresh_token":"rt","expires_at":1234567890}]`
	if err := ioutil.WriteFile(cfgPath, []byte(cliJSON), base.LocalFileMode); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(credPath, []byte(credJSON), base.LocalFileMode); err != nil {
		t.Fatal(err)
	}

	m, err := base.NewAggConfigManager(cfgPath, credPath)
	if err != nil {
		t.Fatal(err)
	}
	cfg, ok := m.GetAggConfigByProfile("oa")
	if !ok {
		t.Fatal("profile oa missing")
	}

	oldM, oldC := base.AggConfigListIns, base.ConfigIns
	base.AggConfigListIns, base.ConfigIns = m, cfg
	defer func() { base.AggConfigListIns, base.ConfigIns = oldM, oldC }()

	if err := switchProfileToAKSK(cfg); err != nil {
		t.Fatal(err)
	}

	// 重新读盘验证持久化，而非只看内存
	m2, err := base.NewAggConfigManager(cfgPath, credPath)
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

// fakeGatewayServer 模拟业务网关：响应远程校验所需的 GetRegion/GetProjectList
func fakeGatewayServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		payload := r.URL.RawQuery + string(body)
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(payload, "GetRegion"):
			fmt.Fprint(w, `{"RetCode":0,"Action":"GetRegionResponse","Regions":[{"Region":"cn-bj2","Zone":"cn-bj2-04","IsDefault":true}]}`)
		case strings.Contains(payload, "GetProjectList"):
			fmt.Fprint(w, `{"RetCode":0,"Action":"GetProjectListResponse","ProjectSet":[{"ProjectId":"org-123","ProjectName":"Default","IsDefault":true}]}`)
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
	if err := ioutil.WriteFile(cfgPath, []byte(cliJSON), base.LocalFileMode); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(credPath, []byte(credJSON), base.LocalFileMode); err != nil {
		t.Fatal(err)
	}

	m, err := base.NewAggConfigManager(cfgPath, credPath)
	if err != nil {
		t.Fatal(err)
	}

	// GetBizClient 会改写包级全局 ClientConfig/AuthCredential，恢复现场避免测试顺序耦合
	oldM, oldCC, oldAC := base.AggConfigListIns, base.ClientConfig, base.AuthCredential
	base.AggConfigListIns = m
	defer func() {
		base.AggConfigListIns, base.ClientConfig, base.AuthCredential = oldM, oldCC, oldAC
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
	m2, err := base.NewAggConfigManager(cfgPath, credPath)
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
	if err := ioutil.WriteFile(cfgPath, []byte(cliJSON), base.LocalFileMode); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(credPath, []byte(credJSON), base.LocalFileMode); err != nil {
		t.Fatal(err)
	}

	m, err := base.NewAggConfigManager(cfgPath, credPath)
	if err != nil {
		t.Fatal(err)
	}
	cfg, ok := m.GetAggConfigByProfile("oa")
	if !ok {
		t.Fatal("profile oa missing")
	}

	oldM, oldC := base.AggConfigListIns, base.ConfigIns
	base.AggConfigListIns, base.ConfigIns = m, cfg
	defer func() { base.AggConfigListIns, base.ConfigIns = oldM, oldC }()

	// 模拟 NewCmdInit Run 完整配置路径对 ConfigIns（即 manager map 内同一指针）的写入
	clearOAuthState(cfg)
	cfg.PublicKey = "newpub"
	cfg.PrivateKey = "newpri"
	cfg.Region = "cn-bj2"
	cfg.Zone = "cn-bj2-04"
	cfg.ProjectID = "org-new"
	cfg.Timeout = base.DefaultTimeoutSec
	cfg.BaseURL = base.DefaultBaseURL
	cfg.Active = true

	if err := saveInitProfile(cfg); err != nil {
		t.Fatalf("save must overwrite existing profile instead of failing, got: %v", err)
	}

	// 重新读盘验证持久化，而非只看内存
	m2, err := base.NewAggConfigManager(cfgPath, credPath)
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
