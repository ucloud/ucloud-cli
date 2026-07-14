package cmd

import (
	"github.com/spf13/cobra"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/cmd/internal/platform"
	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// AddChildrenForSnapshot builds the full command tree for the structure golden,
// without InitConfig/network side effects. Test-only helper.
//
// Some NewCmdXxx constructors create service-specific SDK requests at
// construction time, so runtime SDK config and credential must be non-nil. We
// initialise them with zero-credential stubs when InitConfig was skipped.
func AddChildrenForSnapshot(root *cobra.Command) {
	runtimeAutoStub = true
	if platform.ClientConfig == nil {
		platform.ClientConfig = &sdk.Config{BaseUrl: platform.DefaultBaseURL}
	}
	if platform.AuthCredential == nil {
		platform.AuthCredential = &platform.CredentialConfig{}
	}
	setActiveRuntimeFromBaseGlobals()
	addChildren(root)
}

// DisableRuntimeForSnapshotCompletion poisons runtime-backed dynamic
// completions after command construction, so snapshot rendering does not issue
// real network calls. It mirrors the old test behavior of nil-ing platform.BizClient
// after AddChildrenForSnapshot.
func DisableRuntimeForSnapshotCompletion() {
	platform.ClientConfig = nil
	platform.AuthCredential = nil
	runtimeAutoStub = false
	activeRuntime = buildRuntimeFromBaseGlobals()
}

// ProductsForSnapshot exposes the registered product list to the snapshot
// golden tests (hack/snapshot): each product's subtree is rendered against
// the golden the product team owns (products/<name>/testdata/). Test-only.
func ProductsForSnapshot() []cli.Product { return registeredProducts() }
