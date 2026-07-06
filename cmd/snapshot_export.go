package cmd

import (
	"github.com/spf13/cobra"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/base"
	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// AddChildrenForSnapshot builds the full command tree for the structure golden,
// without InitConfig/network side effects. Test-only helper.
//
// Some NewCmdXxx constructors call base.BizClient.NewXxxRequest() and
// base.ClientConfig fields at construction time, so both must be non-nil.
// We initialise them with zero-credential stubs when InitConfig was skipped.
func AddChildrenForSnapshot(root *cobra.Command) {
	if base.ClientConfig == nil {
		base.ClientConfig = &sdk.Config{BaseUrl: base.DefaultBaseURL}
	}
	if base.AuthCredential == nil {
		base.AuthCredential = &base.CredentialConfig{}
	}
	if base.BizClient == nil {
		base.BizClient = base.NewClient(base.ClientConfig, base.AuthCredential, nil)
	}
	addChildren(root)
}

// ProductsForSnapshot exposes the registered product list to the snapshot
// golden tests (hack/snapshot): each product's subtree is rendered against
// the golden the product team owns (products/<name>/testdata/). Test-only.
func ProductsForSnapshot() []cli.Product { return registeredProducts() }
