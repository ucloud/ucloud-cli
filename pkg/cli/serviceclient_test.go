package cli_test

import (
	"testing"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-sdk-go/services/udb"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

func TestNewServiceClientUsesInjectedProviders(t *testing.T) {
	attached := false
	ctx := cli.NewContext(cli.Deps{
		ClientConfig: func() *ucloud.Config { return &ucloud.Config{} },
		BuildCredential: func() *auth.Credential {
			return &auth.Credential{PublicKey: "pk", PrivateKey: "sk"}
		},
		AttachHandlers: func(sc ucloud.ServiceClient) {
			attached = true
		},
	})

	c := cli.NewServiceClient(ctx, udb.NewClient)
	if c == nil {
		t.Fatal("NewServiceClient returned nil")
	}
	if !attached {
		t.Fatal("NewServiceClient did not call AttachHandlers provider")
	}
}

func TestNewServiceClientRequiresInjectedProviders(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("NewServiceClient without providers did not panic")
		}
	}()

	_ = cli.NewServiceClient(cli.NewContext(cli.Deps{}), udb.NewClient)
}
