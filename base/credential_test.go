// base/credential_test.go
package base

import (
	"testing"

	"github.com/ucloud/ucloud-sdk-go/services/udb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

// oauth 模式：残留的 AK/SK 绝不能进入签名凭据（否则签名参数与 Bearer 同时上行 → 网关 RetCode 171）。
func TestBuildCredentialOAuthEmpty(t *testing.T) {
	cc := &CredentialConfig{AuthMode: AuthModeOAuth, PublicKey: "pk", PrivateKey: "sk"}
	cred := buildCredential(cc)
	if cred.PublicKey != "" {
		t.Errorf("oauth credential PublicKey = %q, want empty", cred.PublicKey)
	}
	if cred.PrivateKey != "" {
		t.Errorf("oauth credential PrivateKey = %q, want empty", cred.PrivateKey)
	}
}

// AK/SK 模式（非 oauth）：真实公私钥必须进入签名凭据，否则 SDK 签名器拿不到密钥 → 网关 RetCode 171。
func TestBuildCredentialAkskKeys(t *testing.T) {
	// AuthMode 留空即 AK/SK（没有 AuthModeAKSK 常量；config.go:896 显式置空表示走签名）。
	cc := &CredentialConfig{AuthMode: "", PublicKey: "pk", PrivateKey: "sk"}
	cred := buildCredential(cc)
	if cred.PublicKey != "pk" {
		t.Errorf("aksk credential PublicKey = %q, want pk", cred.PublicKey)
	}
	if cred.PrivateKey != "sk" {
		t.Errorf("aksk credential PrivateKey = %q, want sk", cred.PrivateKey)
	}
}

// 包级 wrapper BuildCredential 必须与 buildCredential 走同一逻辑/分支（cli.NewServiceClient 依赖它）。
func TestBuildCredentialPackageWrapper(t *testing.T) {
	old := AuthCredential
	t.Cleanup(func() { AuthCredential = old })

	// AK/SK：返回真实密钥
	AuthCredential = &CredentialConfig{AuthMode: "", PublicKey: "pk", PrivateKey: "sk"}
	cred := BuildCredential()
	if cred.PublicKey != "pk" || cred.PrivateKey != "sk" {
		t.Errorf("BuildCredential aksk = {%q,%q}, want {pk,sk}", cred.PublicKey, cred.PrivateKey)
	}

	// oauth：返回空凭据
	AuthCredential = &CredentialConfig{AuthMode: AuthModeOAuth, PublicKey: "pk", PrivateKey: "sk"}
	cred = BuildCredential()
	if cred.PublicKey != "" || cred.PrivateKey != "" {
		t.Errorf("BuildCredential oauth = {%q,%q}, want empty", cred.PublicKey, cred.PrivateKey)
	}
}

// AttachHandlers 在真实 sub-client 上挂载三个 handler，不 panic、不报错；挂载后 client 仍可用。
// （handler 内部行为已由 client_test.go 的 12 个鉴权用例覆盖，此处只验证 wiring 不破。）
func TestAttachHandlersDoesNotPanic(t *testing.T) {
	oldCred, oldIns := AuthCredential, ConfigIns
	t.Cleanup(func() { AuthCredential, ConfigIns = oldCred, oldIns })
	AuthCredential = &CredentialConfig{AuthMode: AuthModeOAuth, AccessToken: "tok"}
	ConfigIns = &AggConfig{Profile: "p"}

	c := udb.NewClient(&sdk.Config{}, &auth.Credential{})
	// 不 panic 即通过；sdk 的 AddXxxHandler 恒返回 nil，但以防回归仍断言 client 非空。
	AttachHandlers(c)
	if c == nil || c.Client == nil {
		t.Fatal("AttachHandlers must leave the client usable")
	}
}
