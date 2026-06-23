package cli_test

import (
	"testing"

	"github.com/ucloud/ucloud-cli/base"
	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-sdk-go/services/udb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
)

// saveBaseVars 保存并恢复 NewServiceClient 读取的包级全局，避免测试顺序耦合。
func saveBaseVars(t *testing.T) {
	t.Helper()
	oldCfg, oldCred, oldIns := base.ClientConfig, base.AuthCredential, base.ConfigIns
	t.Cleanup(func() {
		base.ClientConfig, base.AuthCredential, base.ConfigIns = oldCfg, oldCred, oldIns
	})
}

// AK/SK profile：NewServiceClient 返回非 nil client，且 BuildCredential 暴露真实密钥
// （凭据与 base.NewClient 等价是 §9 的签名正确性保证）。
func TestNewServiceClientAksk(t *testing.T) {
	saveBaseVars(t)
	base.ClientConfig = &sdk.Config{}
	base.AuthCredential = &base.CredentialConfig{AuthMode: "", PublicKey: "pk", PrivateKey: "sk"}
	base.ConfigIns = &base.AggConfig{}

	c := cli.NewServiceClient(cli.NewContext(cli.Deps{}), udb.NewClient)
	if c == nil {
		t.Fatal("NewServiceClient returned nil")
	}
	cred := base.BuildCredential()
	if cred.PublicKey != "pk" || cred.PrivateKey != "sk" {
		t.Errorf("aksk signing credential = {%q,%q}, want {pk,sk}", cred.PublicKey, cred.PrivateKey)
	}
}

// oauth profile：NewServiceClient 返回非 nil client，且 BuildCredential 返回空凭据
// （残留 AK/SK 不得进入签名器，Bearer 唯一凭据，§9 防 RetCode 171）。
func TestNewServiceClientOAuth(t *testing.T) {
	saveBaseVars(t)
	base.ClientConfig = &sdk.Config{}
	base.AuthCredential = &base.CredentialConfig{AuthMode: base.AuthModeOAuth, AccessToken: "tok", PublicKey: "pk", PrivateKey: "sk"}
	base.ConfigIns = &base.AggConfig{}

	c := cli.NewServiceClient(cli.NewContext(cli.Deps{}), udb.NewClient)
	if c == nil {
		t.Fatal("NewServiceClient returned nil")
	}
	cred := base.BuildCredential()
	if cred.PublicKey != "" || cred.PrivateKey != "" {
		t.Errorf("oauth signing credential = {%q,%q}, want empty", cred.PublicKey, cred.PrivateKey)
	}
}
