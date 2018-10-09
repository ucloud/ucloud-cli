package vpc

import (
	"github.com/ucloud/ucloud-sdk-go/sdk"
	"github.com/ucloud/ucloud-sdk-go/sdk/auth"
)

type VPCClient struct {
	client *sdk.Client
}

func NewClient(config *sdk.Config, credential *auth.Credential) *VPCClient {
	client := sdk.NewClient(config, credential)
	return &VPCClient{
		client: client,
	}
}
