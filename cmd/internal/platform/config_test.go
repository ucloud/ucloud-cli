package platform

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

const cliConfigJSON = `[
	{"project_id":"org-bdks4e","region":"cn-bj2","zone":"cn-bj2-04","base_url":"https://api.ucloud.cn/","timeout_sec":15,"profile":"uweb","active":true},
	{"project_id":"org-oxjwoi","region":"hk","zone":"hk-02","base_url":"https://api.ucloud.cn/","timeout_sec":15,"profile":"test","active":false}
]`

const credentialJSON = `[
	{"public_key":"4E9UU*****3ZAPWQ==","private_key":"6945*****a0d45","profile":"uweb"},
	{"public_key":"YSQG*****zgnCRQ=","private_key":"jtma*****Avms","profile":"test"}
]`

func TestAggConfigManager(t *testing.T) {
	os.MkdirAll(".ucloud", 0700)
	err := ioutil.WriteFile(".ucloud/config.json", []byte(cliConfigJSON), LocalFileMode)
	if err != nil {
		t.Error(err)
	}
	err = ioutil.WriteFile(".ucloud/credential.json", []byte(credentialJSON), LocalFileMode)
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err := os.RemoveAll(".ucloud")
		if err != nil {
			t.Error(err)
		}
	}()

	acManager, err := NewAggConfigManager(".ucloud/config.json", ".ucloud/credential.json")
	if err != nil {
		t.Error(err)
	}

	if len(acManager.configs) != 2 {
		t.Errorf("expect length of configs is 2, accpet %d", len(acManager.configs))
	}

}

func TestEmptyAggConfigManager(t *testing.T) {
	os.MkdirAll(".ucloud", 0700)
	defer func() {
		err := os.RemoveAll(".ucloud")
		if err != nil {
			t.Error(err)
		}
	}()

	acManager, err := NewAggConfigManager(".ucloud/config.json", ".ucloud/credential.json")
	if err != nil {
		t.Error(err)
	}

	err = acManager.Load()
	if err != nil {
		t.Fatal(err)
	}

	if len(acManager.configs) != 0 {
		t.Errorf("expect length of configs is 2, accpet %d", len(acManager.configs))
	}
}

// CRITICAL 回归：旧 credential.json（无 oauth 字段）必须照常加载且 Save 后不丢数据
func TestOldCredentialCompat(t *testing.T) {
	os.MkdirAll(".ucloud", 0700)
	defer os.RemoveAll(".ucloud")
	ioutil.WriteFile(".ucloud/config.json", []byte(cliConfigJSON), LocalFileMode)
	ioutil.WriteFile(".ucloud/credential.json", []byte(credentialJSON), LocalFileMode)

	m, err := NewAggConfigManager(".ucloud/config.json", ".ucloud/credential.json")
	if err != nil {
		t.Fatal(err)
	}
	ac, ok := m.GetAggConfigByProfile("uweb")
	if !ok {
		t.Fatal("profile uweb missing")
	}
	if ac.AuthMode != "" || ac.AccessToken != "" {
		t.Errorf("old file should yield empty oauth fields, got %+v", ac)
	}
	if ac.PublicKey == "" {
		t.Error("aksk fields must survive")
	}
	if err := m.Save(); err != nil {
		t.Fatal(err)
	}
}

// oauth 字段写入后能读回（含轮换写回场景的字段完整性）
func TestOAuthFieldsRoundTrip(t *testing.T) {
	os.MkdirAll(".ucloud", 0700)
	defer os.RemoveAll(".ucloud")
	m, err := NewAggConfigManager(".ucloud/config.json", ".ucloud/credential.json")
	if err != nil {
		t.Fatal(err)
	}
	ac := &AggConfig{
		Profile: "oauthp", Active: true, BaseURL: DefaultBaseURL, Timeout: 15,
		MaxRetryTimes: intPtr(3),
		AuthMode:      AuthModeOAuth, AccessToken: "at", RefreshToken: "rt", ExpiresAt: 1234567890,
		OAuthBaseURL: "https://oauth.example.com",
	}
	if err := m.Append(ac); err != nil {
		t.Fatal(err)
	}

	// 重新读盘验证
	m2, err := NewAggConfigManager(".ucloud/config.json", ".ucloud/credential.json")
	if err != nil {
		t.Fatal(err)
	}
	got, ok := m2.GetAggConfigByProfile("oauthp")
	if !ok {
		t.Fatal("profile oauthp missing after reload")
	}
	if got.AuthMode != AuthModeOAuth || got.AccessToken != "at" || got.RefreshToken != "rt" ||
		got.ExpiresAt != 1234567890 || got.OAuthBaseURL != "https://oauth.example.com" {
		t.Errorf("oauth fields lost on round trip: %+v", got)
	}
}

// channel_key 写入后能读回（AC4）。覆盖跨层链路 copyToCLIConfig → config.json → Load，
// 任一层漏改都会表现为「能存不能读」或「存了不落盘」。
func TestChannelKeyRoundTrip(t *testing.T) {
	os.MkdirAll(".ucloud", 0700)
	defer os.RemoveAll(".ucloud")
	m, err := NewAggConfigManager(".ucloud/config.json", ".ucloud/credential.json")
	if err != nil {
		t.Fatal(err)
	}
	ac := &AggConfig{
		Profile: "combo", Active: true, BaseURL: "https://api.ucloud-global.com/", Timeout: 15,
		MaxRetryTimes: intPtr(3), ChannelKey: "ch_combo_roundtrip",
	}
	if err := m.Append(ac); err != nil {
		t.Fatal(err)
	}

	m2, err := NewAggConfigManager(".ucloud/config.json", ".ucloud/credential.json")
	if err != nil {
		t.Fatal(err)
	}
	got, ok := m2.GetAggConfigByProfile("combo")
	if !ok {
		t.Fatal("profile combo missing after reload")
	}
	if got.ChannelKey != "ch_combo_roundtrip" {
		t.Errorf("ChannelKey lost on round trip: got %q, want ch_combo_roundtrip", got.ChannelKey)
	}

	// channel-key 是接入点配置而非凭据：绝不能落进 credential.json
	raw, err := os.ReadFile(".ucloud/credential.json")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "ch_combo_roundtrip") {
		t.Error("channel_key must not be persisted to credential.json (it is an endpoint config, not a credential)")
	}
}

// 向后兼容（AC6）：不含 channel_key 字段的旧 config.json 照常加载，且该值为空。
func TestConfigWithoutChannelKeyLoads(t *testing.T) {
	os.MkdirAll(".ucloud", 0700)
	defer os.RemoveAll(".ucloud")
	if err := os.WriteFile(".ucloud/config.json",
		[]byte(`[{"profile":"legacy","active":true,"base_url":"https://api.ucloud.cn/","timeout_sec":15,"max_retry_times":3}]`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(".ucloud/credential.json",
		[]byte(`[{"profile":"legacy","public_key":"pub","private_key":"pri"}]`), 0600); err != nil {
		t.Fatal(err)
	}
	m, err := NewAggConfigManager(".ucloud/config.json", ".ucloud/credential.json")
	if err != nil {
		t.Fatal(err)
	}
	got, ok := m.GetAggConfigByProfile("legacy")
	if !ok {
		t.Fatal("legacy profile missing")
	}
	if got.ChannelKey != "" {
		t.Errorf("ChannelKey = %q, want empty for a legacy config", got.ChannelKey)
	}
}

// UpdateAggConfig 必须以传入的 config 为准：当传入指针与 map 内条目不是同一个对象时
// （如 `ucloud --profile <不存在>` 回退到包级默认 ConfigIns 而盘上已有同名 profile），
// 不能静默把 map 里的旧数据存盘、丢掉调用方的数据。
func TestUpdateAggConfigPointerMismatch(t *testing.T) {
	os.MkdirAll(".ucloud", 0700)
	defer os.RemoveAll(".ucloud")
	m, err := NewAggConfigManager(".ucloud/config.json", ".ucloud/credential.json")
	if err != nil {
		t.Fatal(err)
	}
	old := &AggConfig{
		Profile: "x", Active: true, Region: "cn-bj2", Zone: "cn-bj2-04",
		PublicKey: "oldpub", PrivateKey: "oldpri",
		BaseURL: DefaultBaseURL, Timeout: 15, MaxRetryTimes: intPtr(3),
	}
	if err := m.Append(old); err != nil {
		t.Fatal(err)
	}

	// 独立构造的另一个指针，同 Profile、不同字段值
	fresh := &AggConfig{
		Profile: "x", Active: true, Region: "hk", Zone: "hk-02",
		PublicKey: "newpub", PrivateKey: "newpri",
		BaseURL: DefaultBaseURL, Timeout: 30, MaxRetryTimes: intPtr(5),
	}
	if err := m.UpdateAggConfig(fresh); err != nil {
		t.Fatal(err)
	}

	m2, err := NewAggConfigManager(".ucloud/config.json", ".ucloud/credential.json")
	if err != nil {
		t.Fatal(err)
	}
	got, ok := m2.GetAggConfigByProfile("x")
	if !ok {
		t.Fatal("profile x missing after reload")
	}
	if got.Region != "hk" || got.Zone != "hk-02" || got.PublicKey != "newpub" ||
		got.PrivateKey != "newpri" || got.Timeout != 30 {
		t.Errorf("passed config was silently dropped, stale data persisted: %+v", got)
	}
}

func intPtr(i int) *int { return &i }
